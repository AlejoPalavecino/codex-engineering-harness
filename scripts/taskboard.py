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
from dataclasses import dataclass
import os
from pathlib import Path
import re
import sys
import tempfile

HEADER = re.compile(r"^### ([A-Z][A-Z0-9]*-\d{3,}) [—-] (.+)$", re.MULTILINE)
FIELD = re.compile(r"^- \*\*(State|Depends on|Owner|Acceptance|Evidence|Blocker|Scope):\*\* (.*)$", re.MULTILINE)
META = re.compile(r"^- \*\*(Spec|Design|Approval|Evidence mode):\*\* (.*)$", re.MULTILINE)
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

    tasks: list[Task] = []
    for idx, match in enumerate(matches):
        end = matches[idx + 1].start() if idx + 1 < len(matches) else len(source)
        body = source[match.end():end]
        fields: dict[str, str] = {}
        for entry in FIELD.finditer(body):
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
    if mode not in (None, "candidate-v1"):
        errors.append("Evidence mode debe ser candidate-v1 (o ausente en tableros históricos)")
    approval = metadata["Approval"].strip().lower()
    if approval not in {"pending", "approved"}:
        errors.append("Approval debe ser pending o approved")
    if not metadata["Spec"] or not metadata["Design"]:
        errors.append("Spec y Design deben indicar referencias o 'n/a' justificable")
    by_id: dict[str, Task] = {}
    for task in tasks:
        if mode == "candidate-v1":
            try:
                validate_scope(json.loads(task.fields.get("Scope", "null")))
                if task.state == "DONE":
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


def candidate_errors(source: str, root: Path, board: Path, ident: str | None = None) -> list[str]:
    metadata, tasks = parse(source)
    if metadata.get("Evidence mode") != "candidate-v1":
        return []
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
            actual = fingerprint(root, board, json.loads(task.fields["Scope"]))
            if evidence_record(task.fields["Evidence"])["candidate"] != actual:
                errors.append(f"{task.ident}: candidato cambió; repetir verificación antes de DONE")
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


def main(argv: list[str] | None = None) -> int:
    parser = argparse.ArgumentParser(description="TTP: validación y transición de tareas (sin dependencias)")
    sub = parser.add_subparsers(dest="action", required=True)
    for name in ("check", "show", "set", "fingerprint"):
        command = sub.add_parser(name)
        command.add_argument("file", type=Path)
        command.add_argument("--root", type=Path, default=Path.cwd())
        if name == "check":
            command.add_argument("--current", action="store_true")
        if name == "fingerprint":
            command.add_argument("id")
        if name == "set":
            command.add_argument("id")
            command.add_argument("state", choices=STATES)
            command.add_argument("--expect", choices=STATES)
            command.add_argument("--evidence")
            command.add_argument("--blocker")
            command.add_argument("--reason")
    args = parser.parse_args(argv)
    try:
        source = args.file.read_text(encoding="utf-8")
        if args.action == "set":
            result = transition(source, args.id, args.state, args.expect,
                                args.evidence, args.blocker, args.reason)
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
        if args.action == "fingerprint":
            task = next((task for task in tasks if task.ident == args.id), None)
            if metadata.get("Evidence mode") != "candidate-v1" or task is None:
                raise ValueError("fingerprint requiere candidate-v1 y un ID existente")
            print(fingerprint(args.root, args.file, json.loads(task.fields["Scope"])))
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
