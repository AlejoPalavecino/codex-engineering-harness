#!/usr/bin/env python3
"""Optional, deterministic task-state checker for TTP v1.0. Python 3.11+.

Usage:
    python taskboard.py check odd/tasks/feature.md
    python taskboard.py show odd/tasks/feature.md
    python taskboard.py set odd/tasks/feature.md TASK-001 IN_PROGRESS --expect READY
"""
from __future__ import annotations

import argparse
import hashlib
import json
from collections import Counter
from collections.abc import Callable
from dataclasses import dataclass
from datetime import datetime
from functools import cache
import os
from pathlib import Path
import re
import sys
import tempfile

HEADER = re.compile(r"^### ([A-Z][A-Z0-9]*-\d{3,}) [—-] (.+)$", re.MULTILINE)
FIELD = re.compile(r"^- \*\*(State|Depends on|Owner|Acceptance|Evidence|Blocker|Scope):\*\* (.*)$", re.MULTILINE)
META = re.compile(r"^- \*\*(Spec|Design|Approval|Evidence mode|Trace mode):\*\* (.*)$", re.MULTILINE)
REQUIREMENT = re.compile(r"^- \*\*(REQ-[^:*\r\n]*):\*\*[ \t]*(.*)$", re.MULTILINE)
TRACE_FIELD = re.compile(r"^- \*\*(Requirements|Verification):\*\*[ \t]*(.*)$", re.MULTILINE)
RISK_FIELD = re.compile(r"^- \*\*(Risk):\*\*[ \t]*(.*)$", re.MULTILINE)
QUALITY_MODE = "quality-v1"
CANDIDATE_MODES = ("candidate-v1", QUALITY_MODE)
CASE_KINDS = ("test", "negative", "static", "review")
RISK_MINIMUM = {"low": (), "standard": ("test",), "critical": ("test", "negative", "review")}
RUN_STATES = ("passed", "failed", "incomplete")
STATES = ("PLANNED", "READY", "IN_PROGRESS", "VERIFYING", "BLOCKED", "DONE", "CANCELLED")
TRANSITIONS = {
    "PLANNED": {"READY", "CANCELLED"},
    "READY": {"IN_PROGRESS", "BLOCKED", "CANCELLED"},
    "IN_PROGRESS": {"VERIFYING", "BLOCKED", "READY", "CANCELLED"},
    "VERIFYING": {"DONE", "IN_PROGRESS", "BLOCKED", "CANCELLED"},
    "BLOCKED": {"READY", "CANCELLED"},
    "DONE": {"READY"},
    "CANCELLED": set(),
}
UNSET = {"", "none", "ninguna", "-", "—", "pending"}


@dataclass
class Task:
    ident: str
    title: str
    start: int
    end: int
    fields: dict[str, str]

    @property
    def state(self) -> str:
        return self.fields["State"]

    @property
    def requirements(self) -> list[str]:
        return [item.strip() for item in self.fields.get("Requirements", "").split(",")]

    @property
    def deps(self) -> list[str]:
        value = self.fields["Depends on"].strip()
        return [] if value.lower() in UNSET else [x.strip() for x in value.split(",")]


def parse(source: str) -> tuple[dict[str, str], list[Task]]:
    matches = list(HEADER.finditer(source))
    if not matches:
        raise ValueError("No hay tareas: se esperan encabezados '### TASK-001 — ...'")
    metadata: dict[str, str] = {}
    for entry in META.finditer(source[:matches[0].start()]):
        key, value = entry.group(1), entry.group(2).strip()
        if key in metadata:
            raise ValueError(f"Metadato repetido: {key}")
        metadata[key] = value
    for key in ("Spec", "Design", "Approval"):
        if key not in metadata:
            raise ValueError(f"Falta metadato {key}")

    tracing = "Trace mode" in metadata
    quality = metadata.get("Evidence mode") == QUALITY_MODE
    if tracing:
        if metadata["Trace mode"] != "requirements-v1":
            raise ValueError("Trace mode debe ser requirements-v1")
        for entry in REQUIREMENT.finditer(source[:matches[0].start()]):
            ident, description = entry.groups()
            if not re.fullmatch(r"REQ-[0-9]{3,}", ident) or not description.strip():
                raise ValueError(f"Requisito inválido: {ident}; se espera REQ-001 y descripción o referencia")
            if ident in metadata:
                raise ValueError(f"{ident}: requisito repetido")
            metadata[ident] = description.strip()

    tasks: list[Task] = []
    for idx, match in enumerate(matches):
        end = matches[idx + 1].start() if idx + 1 < len(matches) else len(source)
        body = source[match.end():end]
        fields: dict[str, str] = {}
        entries = list(FIELD.finditer(body))
        if tracing or quality:
            entries.extend(entry for entry in TRACE_FIELD.finditer(body)
                           if tracing or entry.group(1) == "Verification")
        if quality:
            entries.extend(RISK_FIELD.finditer(body))
        for entry in entries:
            key, value = entry.group(1), entry.group(2).strip()
            if key in fields:
                raise ValueError(f"{match.group(1)}: campo repetido {key}")
            fields[key] = value
        for key in ("State", "Depends on", "Owner", "Acceptance", "Evidence", "Blocker"):
            if key not in fields:
                raise ValueError(f"{match.group(1)}: falta campo {key}")
        tasks.append(Task(match.group(1), match.group(2).strip(), match.start(), end, fields))
    return metadata, tasks


def errors_for(source: str) -> list[str]:
    try:
        metadata, tasks = parse(source)
    except ValueError as exc:
        return [str(exc)]
    errors: list[str] = []
    mode = metadata.get("Evidence mode")
    if mode is not None and mode not in CANDIDATE_MODES:
        errors.append("Evidence mode debe ser candidate-v1, quality-v1 o ausente")
    approval = metadata["Approval"].strip().lower()
    if approval not in {"pending", "approved"}:
        errors.append("Approval debe ser pending o approved")
    if not metadata["Spec"] or not metadata["Design"]:
        errors.append("Spec y Design deben indicar referencias o 'n/a' justificable")
    if "Trace mode" in metadata:
        linked = set()
        for task in tasks:
            for ident, count in Counter(task.requirements).items():
                if ident not in metadata or not ident.startswith("REQ-"):
                    errors.append(f"{task.ident}: Requirements requiere un requisito definido: {ident!r}")
                if count > 1:
                    errors.append(f"{task.ident}: requisito repetido {ident}")
                linked.add(ident)
            if task.fields.get("Verification", "").lower() in UNSET:
                errors.append(f"{task.ident}: falta la prueba o inspección prevista en Verification")
        for ident in metadata:
            if ident.startswith("REQ-") and ident not in linked:
                errors.append(f"{ident}: sin tarea asociada")

    by_id: dict[str, Task] = {}
    for task in tasks:
        if mode in CANDIDATE_MODES:
            try:
                validate_scope(json.loads(task.fields.get("Scope", "null")))
                if mode == QUALITY_MODE:
                    cases = quality_cases(task)
                    runs = quality_evidence(task.fields["Evidence"], cases)[1]
                    if task.state == "DONE":
                        blocked = [f"{case}: {status}" for case, status in quality_statuses(cases, runs).items()
                                   if status != "passed"]
                        if blocked:
                            raise ValueError("Quality gate requiere passed: " + ", ".join(blocked))
                elif task.state == "DONE":
                    evidence_record(task.fields["Evidence"])
            except (ValueError, TypeError) as exc:
                errors.append(f"{task.ident}: {exc}")
        if task.ident in by_id:
            errors.append(f"{task.ident}: ID repetido")
        by_id[task.ident] = task
    for task in tasks:
        if task.state not in STATES:
            errors.append(f"{task.ident}: estado desconocido {task.state!r}")
        if not task.title or not task.fields["Owner"] or not task.fields["Acceptance"]:
            errors.append(f"{task.ident}: título, Owner y Acceptance obligatorios")
        if task.state in {"READY", "IN_PROGRESS", "VERIFYING", "DONE"} and approval != "approved":
            errors.append(f"{task.ident}: no se puede ejecutar/cerrar sin Approval: approved")
        if task.state == "DONE" and task.fields["Evidence"].lower() in UNSET:
            errors.append(f"{task.ident}: DONE requiere evidencia")
        if task.state == "BLOCKED" and task.fields["Blocker"].lower() in UNSET:
            errors.append(f"{task.ident}: BLOCKED requiere un bloqueo concreto")
        if task.ident in task.deps:
            errors.append(f"{task.ident}: no puede depender de sí misma")
        for dep in task.deps:
            if dep not in by_id:
                errors.append(f"{task.ident}: dependencia inexistente {dep}")
            elif task.state in {"READY", "IN_PROGRESS", "VERIFYING", "DONE"} and by_id[dep].state != "DONE":
                errors.append(f"{task.ident}: dependencia {dep} todavía no está DONE")

    visiting: set[str] = set()
    visited: set[str] = set()
    def dfs(ident: str) -> None:
        if ident in visiting:
            errors.append(f"Ciclo de dependencias desde {ident}")
            return
        if ident in visited:
            return
        visiting.add(ident)
        for dep in by_id[ident].deps:
            if dep in by_id and dep not in visiting:
                dfs(dep)
            elif dep in visiting:
                errors.append(f"Ciclo de dependencias: {ident} -> {dep}")
        visiting.remove(ident)
        visited.add(ident)
    for ident in by_id:
        dfs(ident)
    return list(dict.fromkeys(errors))


def validate_scope(scope: object) -> list[str]:
    if not isinstance(scope, list) or not scope:
        raise ValueError("Scope requiere un array JSON no vacío de archivos relativos")
    for name in scope:
        if (not isinstance(name, str) or not name or "\\" in name or ":" in name
                or any(part in {"", ".", ".."} for part in name.split("/"))):
            raise ValueError("Scope requiere rutas relativas canónicas con /, sin escapes")
    if len(set(scope)) != len(scope):
        raise ValueError("Scope contiene archivos duplicados")
    return scope


def evidence_record(value: str) -> dict:
    try:
        record = json.loads(value)
    except ValueError as exc:
        raise ValueError("Evidence requiere JSON con command, result y candidate") from exc
    if not isinstance(record, dict) or any(
        not isinstance(record.get(key), str) or not record[key].strip()
        for key in ("command", "result", "candidate")
    ):
        raise ValueError("Evidence requiere command, result y candidate no vacíos")
    if not re.fullmatch(r"sha256:[0-9a-f]{64}", record["candidate"]):
        raise ValueError("Evidence candidate debe ser sha256:<64 hex>")
    return record


def quality_object(value: str, field: str) -> dict:
    def unique_keys(pairs: list) -> dict:
        result = {}
        for key, item in pairs:
            if key in result:
                raise ValueError(f"{field}: clave repetida {key!r}")
            result[key] = item
        return result

    try:
        result = json.loads(value, object_pairs_hook=unique_keys)
    except json.JSONDecodeError as exc:
        raise ValueError(f"{field} requiere un objeto JSON") from exc
    if not isinstance(result, dict):
        raise ValueError(f"{field} requiere un objeto JSON")
    return result


def quality_cases(task: Task) -> dict:
    risk = task.fields.get("Risk")
    if risk not in RISK_MINIMUM:
        raise ValueError("Risk debe ser low, standard o critical")
    cases = quality_object(task.fields.get("Verification", ""), "Verification")
    if not cases:
        raise ValueError("Verification requiere al menos un caso para Acceptance")
    for ident, case in cases.items():
        if (not re.fullmatch(r"[^\W\d][\w-]*", ident) or not isinstance(case, dict)
                or set(case) != {"kind", "target"} or case["kind"] not in CASE_KINDS
                or not isinstance(case["target"], str) or not case["target"].strip()
                or not case["target"].isprintable()):
            raise ValueError(f"Verification: caso inválido {ident!r}; se espera kind y target")
    missing = set(RISK_MINIMUM[risk]) - {case["kind"] for case in cases.values()}
    if missing:
        raise ValueError(f"Risk {risk} requiere casos de tipo " + ", ".join(sorted(missing)))
    return cases


def quality_evidence(value: str, cases: dict) -> tuple[str | None, dict]:
    if value.lower() in UNSET:
        return None, {}
    record = quality_object(value, "Evidence")
    candidate, runs = record.get("candidate"), record.get("runs")
    if (set(record) != {"candidate", "runs"} or not isinstance(candidate, str)
            or not re.fullmatch(r"sha256:[0-9a-f]{64}", candidate) or not isinstance(runs, dict)):
        raise ValueError("Evidence requiere candidate sha256:<64 hex> y un objeto runs")
    details = ("command", "result", "version", "observed_at")
    for ident, run in runs.items():
        if ident not in cases:
            raise ValueError(f"Evidence: caso no declarado {ident!r}")
        if (not isinstance(run, dict) or set(run) != {"status", *details}
                or run["status"] not in RUN_STATES
                or any(not isinstance(run[key], str) or not run[key].strip() for key in details)):
            raise ValueError(f"Evidence {ident}: se espera status, command, result, version y observed_at")
        try:
            timestamp = datetime.fromisoformat(run["observed_at"])
            if "T" not in run["observed_at"] or timestamp.tzinfo is None:
                raise ValueError
        except ValueError as exc:
            raise ValueError(f"Evidence {ident}: observed_at requiere fecha y hora ISO 8601 con zona") from exc
    return candidate, runs


def quality_statuses(cases: dict, runs: dict) -> dict[str, str]:
    return {ident: runs.get(ident, {}).get("status", "not-run") for ident in cases}


def fingerprint(root: Path, board: Path, scope: list[str]) -> str:
    """Hash explicit paths and contents; never execute evidence or follow escapes."""
    root = root.resolve(strict=True)
    board = board.resolve(strict=True)
    if not root.is_dir() or not board.is_relative_to(root):
        raise ValueError("El tablero debe estar dentro de --root")
    entries = []
    for name in sorted(validate_scope(scope)):
        try:
            path = (root / name).resolve(strict=True)
        except (OSError, RuntimeError) as exc:
            raise ValueError(f"Scope: archivo ausente {name}") from exc
        if (not path.is_relative_to(root) or not path.is_file()
                or path.samefile(board)):
            raise ValueError(f"Scope: archivo fuera del alcance o tablero excluido: {name}")
        content = path.read_bytes()
        try:
            content.decode("utf-8")
        except UnicodeDecodeError:
            pass
        else:
            # Git/Windows newline conversions must not invalidate UTF-8 text.
            if b"\0" not in content:
                content = content.replace(b"\r\n", b"\n")
        entries.append([name, hashlib.sha256(content).hexdigest()])
    payload = json.dumps(entries, ensure_ascii=True, separators=(",", ":")).encode("ascii")
    return "sha256:" + hashlib.sha256(payload).hexdigest()


def task_fingerprints(root: Path, board: Path, metadata: dict[str, str], tasks: list[Task]) -> Callable[[str], str]:
    by_id = {task.ident: task for task in tasks}

    @cache
    def compute(ident: str) -> str:
        task = by_id[ident]
        snapshot = fingerprint(root, board, json.loads(task.fields["Scope"]))
        if metadata["Evidence mode"] != QUALITY_MODE:
            return snapshot
        payload = {
            "mode": QUALITY_MODE,
            "snapshot": snapshot,
            "task": ident,
            "dependencies": {dep: compute(dep) for dep in task.deps},
            "spec": metadata["Spec"],
            "design": metadata["Design"],
            "acceptance": task.fields["Acceptance"],
            "requirements": {req: metadata[req] for req in task.requirements} if "Trace mode" in metadata else {},
            "risk": task.fields["Risk"],
            "verification": quality_cases(task),
        }
        encoded = json.dumps(payload, ensure_ascii=True, sort_keys=True, separators=(",", ":")).encode("ascii")
        return "sha256:" + hashlib.sha256(encoded).hexdigest()

    return compute


def candidate_errors(source: str, root: Path, board: Path, ident: str | None = None) -> list[str]:
    metadata, tasks = parse(source)
    mode = metadata.get("Evidence mode")
    if mode not in CANDIDATE_MODES:
        return []
    compute = task_fingerprints(root, board, metadata, tasks)
    selected = None
    if ident is not None:
        by_id = {task.ident: task for task in tasks}
        selected = set()
        pending = [ident]
        while pending:
            current = pending.pop()
            if current not in selected:
                selected.add(current)
                pending.extend(by_id[current].deps)
    errors = []
    for task in tasks:
        if task.state == "DONE" and (selected is None or task.ident in selected):
            actual = compute(task.ident)
            evidence = task.fields["Evidence"]
            recorded = (quality_evidence(evidence, quality_cases(task))[0]
                        if mode == QUALITY_MODE else evidence_record(evidence)["candidate"])
            if recorded != actual:
                context = "candidato o criterio cambió" if mode == QUALITY_MODE else "candidato cambió"
                errors.append(f"{task.ident}: {context}; repetir verificación antes de DONE")
    return errors


def replace_field(block: str, key: str, value: str) -> str:
    if "\n" in value or "\r" in value:
        raise ValueError("Los valores de campo deben ocupar una sola línea")
    pattern = re.compile(r"(^- \*\*" + re.escape(key) + r":\*\* )([^\r\n]*)", re.MULTILINE)
    changed, count = pattern.subn(lambda m: m.group(1) + value, block, count=1)
    if count != 1:
        raise ValueError(f"Falta campo {key}")
    return changed


def transition(source: str, ident: str, target: str, expected: str | None,
               evidence: str | None, blocker: str | None, reason: str | None) -> str:
    errs = errors_for(source)
    if errs:
        raise ValueError("El tablero tiene errores previos: " + "; ".join(errs))
    _, tasks = parse(source)
    task = next((t for t in tasks if t.ident == ident), None)
    if task is None:
        raise ValueError(f"Tarea desconocida: {ident}")
    if target not in STATES:
        raise ValueError(f"Estado desconocido: {target}")
    if expected is not None and task.state != expected:
        raise ValueError(f"{ident}: estado esperado {expected}; actual {task.state}")
    if target not in TRANSITIONS[task.state]:
        raise ValueError(f"Transición inválida: {task.state} -> {target}")
    if target == "DONE" and (evidence is None or evidence.strip().lower() in UNSET):
        raise ValueError("DONE requiere --evidence no vacía")
    if target == "BLOCKED" and (blocker is None or blocker.strip().lower() in UNSET):
        raise ValueError("BLOCKED requiere --blocker")
    if task.state == "DONE" and target == "READY" and not reason:
        raise ValueError("Reabrir DONE requiere --reason")

    block = source[task.start:task.end]
    block = replace_field(block, "State", target)
    if target == "DONE":
        block = replace_field(block, "Evidence", evidence.strip())
    elif task.state == "DONE" and target == "READY":
        block = replace_field(block, "Evidence", "none")
        block = block.rstrip() + f"\n- **Reopen reason:** {reason.strip()}\n\n"
    if target == "BLOCKED":
        block = replace_field(block, "Blocker", blocker.strip())
    elif task.state == "BLOCKED":
        block = replace_field(block, "Blocker", "none")
    result = source[:task.start] + block + source[task.end:]
    errs = errors_for(result)
    if errs:
        raise ValueError("Transición rechazada: " + "; ".join(errs))
    return result


def atomic_write(path: Path, source: str) -> None:
    if not path.is_file():
        raise ValueError("El tablero debe existir")
    temp: str | None = None
    try:
        with tempfile.NamedTemporaryFile("w", encoding="utf-8", newline="\n",
                                         dir=path.parent, prefix=".taskboard-", delete=False) as fp:
            fp.write(source)
            temp = fp.name
        os.replace(temp, path)
    finally:
        if temp is not None and os.path.exists(temp):
            os.unlink(temp)


def show_impact(filename: str, metadata: dict[str, str], tasks: list[Task]) -> None:
    filename = filename.replace("\\", "/").removeprefix("./")
    validate_scope([filename])
    matches = []
    unscoped = 0
    for task in tasks:
        raw_scope = task.fields.get("Scope")
        if raw_scope is None:
            unscoped += 1
            continue
        try:
            scope = validate_scope(json.loads(raw_scope))
        except ValueError as exc:
            raise ValueError(f"{task.ident}: Scope inválido: {exc}") from exc
        if filename in scope:
            matches.append(task)

    print(f"Archivo: {filename}")
    quality = metadata.get("Evidence mode") == QUALITY_MODE
    affected = {task.ident for task in matches}
    if quality:
        by_dependency: dict[str, list[str]] = {}
        for task in tasks:
            for dependency in task.deps:
                by_dependency.setdefault(dependency, []).append(task.ident)
        pending = list(affected)
        while pending:
            for dependent in by_dependency.get(pending.pop(), []):
                if dependent not in affected:
                    pending.append(dependent)
                    affected.add(dependent)
    direct = {task.ident for task in matches}
    for task in tasks:
        if task.ident not in affected:
            continue
        print(f"{task.ident}: {task.state} — {task.title}")
        if "Trace mode" in metadata:
            for ident in task.requirements:
                print(f"  {ident}: {metadata[ident]}")
        else:
            suffix = "" if quality else " y verificación prevista"
            print(f"  Requisitos{suffix}: sin vínculo declarado")
        if quality:
            origin = "Scope" if task.ident in direct else "dependencia de tarea"
            print(f"  Regresión ({origin}, Risk: {task.fields['Risk']})")
            for ident, case in quality_cases(task).items():
                print(f"    {ident} [{case['kind']}]: {case['target']}")
        elif "Trace mode" in metadata:
            print(f"  Verification: {task.fields['Verification']}")
        print(f"  Evidence: {task.fields['Evidence']}")
    if not matches:
        print("Sin coincidencias en los alcances declarados.")
    if unscoped:
        print(f"Tareas sin Scope: {unscoped}. Su impacto queda indeterminado.")


def report_quality(metadata: dict[str, str], task: Task, evidence: str, compute: Callable[[str], str]) -> int:
    cases = quality_cases(task)
    candidate, runs = quality_evidence(evidence, cases)
    stale = candidate is not None and candidate != compute(task.ident)
    statuses = quality_statuses(cases, runs)
    print(f"{task.ident}: {task.fields['Acceptance']}")
    for ident, status in statuses.items():
        shown = f"stale (registro: {status})" if stale and ident in runs else status
        print(f"  {ident}: {shown} [{cases[ident]['kind']}] {cases[ident]['target']}")
    passed = not stale and all(status == "passed" for status in statuses.values())
    print(f"Quality gate: {'passed' if passed else 'blocked'}")
    print(f"Approval: {metadata['Approval']}")
    print("Autorización de entrega: decisión humana por separado")
    return 0 if passed else 1


def main(argv: list[str] | None = None) -> int:
    parser = argparse.ArgumentParser(description="TTP: validación y transición de tareas (sin dependencias)")
    sub = parser.add_subparsers(dest="action", required=True)
    for name in ("check", "show", "set", "fingerprint", "trace", "impact", "quality"):
        command = sub.add_parser(name)
        command.add_argument("file", type=Path)
        command.add_argument("--root", type=Path, default=Path.cwd())
        if name == "impact":
            command.add_argument("path", help="Ruta del archivo relativa a la raíz del proyecto")
        if name == "check":
            command.add_argument("--current", action="store_true")
        if name in ("fingerprint", "trace", "quality"):
            command.add_argument("id")
        if name == "quality":
            command.add_argument("--evidence-file", type=Path, help="Registro de ejecución para cotejar sin escribir el tablero")
        if name == "set":
            command.add_argument("id")
            command.add_argument("state", choices=STATES)
            command.add_argument("--expect", choices=STATES)
            evidence_input = command.add_mutually_exclusive_group()
            evidence_input.add_argument("--evidence")
            evidence_input.add_argument("--evidence-file", type=Path, help="Objeto JSON de evidencia para DONE")
            command.add_argument("--blocker")
            command.add_argument("--reason")
    args = parser.parse_args(argv)
    try:
        source = args.file.read_text(encoding="utf-8")
        if args.action == "set":
            evidence = args.evidence
            if args.evidence_file:
                if args.state != "DONE":
                    raise ValueError("--evidence-file requiere una transición a DONE")
                raw = args.evidence_file.read_text(encoding="utf-8-sig")
                evidence = json.dumps(quality_object(raw, "Evidence"), ensure_ascii=True, separators=(",", ":"))
            result = transition(source, args.id, args.state, args.expect,
                                evidence, args.blocker, args.reason)
            structural = errors_for(result)
            if structural:
                raise ValueError("; ".join(structural))
            candidate = candidate_errors(result, args.root, args.file, args.id) if args.state == "DONE" else []
            if candidate:
                raise ValueError("; ".join(candidate))
            atomic_write(args.file, result)
            source = result
        issues = errors_for(source)
        if issues:
            for issue in issues:
                print("ERROR:", issue, file=sys.stderr)
            return 2
        metadata, tasks = parse(source)
        if args.action == "impact":
            show_impact(args.path, metadata, tasks)
            return 0
        if args.action == "trace":
            if "Trace mode" not in metadata:
                raise ValueError("El tablero no declara trazabilidad de requisitos")
            if not args.id.startswith("REQ-") or args.id not in metadata:
                raise ValueError(f"Requisito desconocido: {args.id}")
            print(f"{args.id}: {metadata[args.id]}")
            for task in tasks:
                if args.id in task.requirements:
                    print(f"{task.ident}: {task.state} — {task.title}")
                    for key in ("Acceptance", "Scope", "Verification", "Evidence"):
                        if key in task.fields:
                            print(f"  {key}: {task.fields[key]}")
            return 0
        if args.action in ("quality", "fingerprint"):
            task = next((task for task in tasks if task.ident == args.id), None)
            if task is None:
                raise ValueError(f"Tarea desconocida: {args.id}")
            compute = task_fingerprints(args.root, args.file, metadata, tasks)
            if args.action == "quality":
                if metadata.get("Evidence mode") != QUALITY_MODE:
                    raise ValueError("quality requiere Evidence mode: quality-v1")
                evidence = args.evidence_file.read_text(encoding="utf-8-sig").strip() if args.evidence_file else task.fields["Evidence"]
                return report_quality(metadata, task, evidence, compute)
            if metadata.get("Evidence mode") not in CANDIDATE_MODES:
                raise ValueError("fingerprint requiere candidate-v1 o quality-v1")
            print(compute(task.ident))
            return 0
        candidate = candidate_errors(source, args.root, args.file) if args.action == "check" and args.current else []
        if candidate:
            raise ValueError("; ".join(candidate))
        if args.action == "show":
            counts = Counter(t.state for t in tasks)
            print(f"Spec: {metadata['Spec']} | Approval: {metadata['Approval']}")
            print(f"DONE: {counts['DONE']}/{len(tasks)}")
            for task in tasks:
                print(f"{task.ident}: {task.state} — {task.title} [Owner: {task.fields['Owner']}]")
                if task.state == "BLOCKED":
                    print(f"  Bloqueo: {task.fields['Blocker']}")
            ready = [t.ident for t in tasks if t.state == "READY"]
            print("Próximas READY:", ", ".join(ready) if ready else "ninguna")
        elif args.action == "set":
            print(f"{args.id}: {args.state} (tablero actualizado)")
        else:
            print(f"OK: {len(tasks)} tareas válidas")
        return 0
    except (OSError, ValueError) as exc:
        print("ERROR:", exc, file=sys.stderr)
        return 2


if __name__ == "__main__":
    raise SystemExit(main())
