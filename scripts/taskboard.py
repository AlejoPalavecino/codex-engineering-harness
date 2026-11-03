#!/usr/bin/env python3
"""Optional, deterministic task-state checker for TTP v1.0. Python 3.11+.

Usage:
    python taskboard.py check odd/tasks/feature.md
    python taskboard.py show odd/tasks/feature.md
    python taskboard.py set odd/tasks/feature.md TASK-001 IN_PROGRESS --expect READY
"""
from __future__ import annotations

import argparse
from collections import Counter
from dataclasses import dataclass
import os
from pathlib import Path
import re
import sys
import tempfile

HEADER = re.compile(r"^### ([A-Z][A-Z0-9]*-\d{3,}) [—-] (.+)$", re.MULTILINE)
FIELD = re.compile(r"^- \*\*(State|Depends on|Owner|Acceptance|Evidence|Blocker):\*\* (.*)$", re.MULTILINE)
META = re.compile(r"^- \*\*(Spec|Design|Approval):\*\* (.*)$", re.MULTILINE)
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
    approval = metadata["Approval"].strip().lower()
    if approval not in {"pending", "approved"}:
        errors.append("Approval debe ser pending o approved")
    if not metadata["Spec"] or not metadata["Design"]:
        errors.append("Spec y Design deben indicar referencias o 'n/a' justificable")
    by_id: dict[str, Task] = {}
    for task in tasks:
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
    for name in ("check", "show", "set"):
        command = sub.add_parser(name)
        command.add_argument("file", type=Path)
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
            atomic_write(args.file, result)
            source = result
        issues = errors_for(source)
        if issues:
            for issue in issues:
                print("ERROR:", issue, file=sys.stderr)
            return 2
        metadata, tasks = parse(source)
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
