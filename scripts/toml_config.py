#!/usr/bin/env python3
"""Conservative, loss-minimizing TOML config update for Codex Harness.

Requires Python 3.11+ (stdlib tomllib). It refuses unsupported layouts instead
of guessing. It never prints values from the user's config.
"""
from __future__ import annotations

import argparse
import copy
import json
import os
from pathlib import Path
import re
import sys
import tomllib

HEADER = re.compile(r"^\s*\[([^\]\r\n]+)\]\s*(?:#.*)?$")
KEY = re.compile(r"^(\s*)([A-Za-z0-9_-]+)(\s*=\s*)(.*?)(\r?\n)?$")
UNSET = object()

def agent_definitions():
    for path in sorted((Path(__file__).resolve().parents[1] / "agents").glob("*.toml")):
        agent = tomllib.loads(path.read_bytes().decode("utf-8-sig"))
        if not re.fullmatch(r"[a-zA-Z0-9_-]+", path.stem) or agent.get("name") != path.stem:
            raise ValueError(f"Invalid packaged agent identity: {path.name}")
        description = agent.get("description")
        if not isinstance(description, str) or not description:
            raise ValueError(f"Missing agent description: {path.name}")
        yield path.stem, description


def value_at(data: dict, path: tuple[str, ...]):
    current = data
    for component in path:
        if not isinstance(current, dict) or component not in current:
            return UNSET
        current = current[component]
    return current


def put_value(data: dict, path: tuple[str, ...], value):
    target = data
    for component in path[:-1]:
        target = target.setdefault(component, {})
        if not isinstance(target, dict):
            raise ValueError("Table/value conflict at " + ".".join(path))
    target[path[-1]] = value


def section_path(header: str) -> tuple[str, ...]:
    try:
        probe = tomllib.loads("[" + header + "]\n_harness_probe = true\n")
    except tomllib.TOMLDecodeError as exc:
        raise ValueError("Unsupported TOML table header") from exc
    found: list[tuple[str, ...]] = []

    def visit(node, path=()):
        if not isinstance(node, dict):
            return
        if "_harness_probe" in node:
            found.append(path)
        for key, val in node.items():
            visit(val, path + (key,))

    visit(probe)
    if len(found) != 1:
        raise ValueError("Unsupported TOML table header")
    return found[0]


def segments(lines: list[str]):
    headers = [(i, section_path(match.group(1))) for i, line in enumerate(lines)
               if (match := HEADER.fullmatch(line.rstrip("\r\n"))) is not None]
    return headers


def update_text(source: str, path: tuple[str, ...], value) -> str:
    lines = source.splitlines(keepends=True)
    headers = segments(lines)
    path_header = path[:-1]
    key = path[-1]
    if not path_header:
        start = 0
        end = headers[0][0] if headers else len(lines)
    else:
        sections = [(i, p) for i, p in headers if p == path_header]
        if len(sections) > 1:
            raise ValueError("Duplicate TOML section: " + ".".join(path_header))
        if not sections:
            # Refuse inline tables and dotted-key layouts rather than
            # appending a duplicate table declaration.
            existing = value_at(tomllib.loads(source), path_header)
            implicit_parent = any(
                len(section) > len(path_header) and section[:len(path_header)] == path_header
                for _, section in headers
            )
            if existing is not UNSET and not implicit_parent:
                raise ValueError("Unsafe TOML layout: table exists without a supported header")
            return source.rstrip("\r\n") + "\n\n[" + ".".join(path_header) + "]\n" + key + " = " + json.dumps(value, ensure_ascii=True) + "\n"
        start = sections[0][0] + 1
        end = next((i for i, _ in headers if i >= start), len(lines))

    matches = []
    for i in range(start, end):
        match = KEY.fullmatch(lines[i])
        if match is not None and match.group(2) == key:
            matches.append((i, match))
    if len(matches) > 1:
        raise ValueError("Duplicate setting: " + ".".join(path))
    rendered = json.dumps(value, ensure_ascii=True)
    if matches:
        index, match = matches[0]
        # Inline comments on strings or primitive values are intentionally
        # not retained; unknown keys and all other lines remain byte-identical.
        newline = match.group(5) or "\n"
        lines[index] = match.group(1) + key + match.group(3) + rendered + newline
    else:
        lines.insert(end, key + " = " + rendered + "\n")
    return "".join(lines)


def config_targets(skip_task_progress: bool):
    targets: dict[tuple[str, ...], object] = {}
    for role, description in agent_definitions():
        targets[("agents", role, "description")] = description
        targets[("agents", role, "config_file")] = "./agents/" + role + ".toml"
    if not skip_task_progress:
        targets[("tools", "update_plan", "enabled")] = True
    targets[("model",)] = "gpt-6.1-sol"
    targets[("model_reasoning_effort",)] = "medium"
    targets[("agents", "default_subagent_model")] = "gpt-6.1-sol"
    return targets


def prepare(source: str, *, configure_models: bool = False,
            skip_task_progress: bool = False, replace_managed: bool = False):
    before = tomllib.loads(source)
    expected = copy.deepcopy(before)
    updated = source
    changed: list[str] = []
    preserved: list[str] = []
    tracked: dict[str, object] = {}
    for path, wanted in config_targets(skip_task_progress).items():
        key = ".".join(path)
        current = value_at(expected, path)
        if current is not UNSET and current != wanted and not (replace_managed or (configure_models and path in {("model",), ("model_reasoning_effort",), ("agents", "default_subagent_model")})):
            preserved.append(key)
            tracked[key] = current
            continue
        tracked[key] = wanted
        if current == wanted:
            continue
        updated = update_text(updated, path, wanted)
        put_value(expected, path, wanted)
        changed.append(key)
    after = tomllib.loads(updated)
    if after != expected:
        raise ValueError("Unsafe TOML layout or unexpected changes: original preserved; manually review table styles")
    return updated, {"changed": changed, "preserved": preserved, "tracked": tracked}


def validate(source: str, state: dict | None = None):
    data = tomllib.loads(source)
    drift: list[str] = []
    if state is not None:
        for key, wanted in state.get("managed_config", {}).items():
            found = value_at(data, tuple(key.split(".")))
            if found is UNSET or found != wanted:
                drift.append(key)
    return {"syntax": "valid", "managed_drift": drift}


def main() -> int:
    parser = argparse.ArgumentParser()
    parser.add_argument("action", choices=["prepare", "validate"])
    parser.add_argument("--input", type=Path, required=True)
    parser.add_argument("--output", type=Path)
    parser.add_argument("--report", type=Path)
    parser.add_argument("--manifest", type=Path)
    parser.add_argument("--configure-models", action="store_true")
    parser.add_argument("--skip-task-progress", action="store_true")
    parser.add_argument("--replace-managed", action="store_true")
    args = parser.parse_args()
    try:
        raw = args.input.read_bytes() if args.input.is_file() else b""
        source = raw.decode("utf-8-sig")
        if args.action == "prepare":
            if args.output is None or args.report is None:
                raise ValueError("prepare requires --output and --report")
            updated, report = prepare(
                source, configure_models=args.configure_models,
                skip_task_progress=args.skip_task_progress,
                replace_managed=args.replace_managed)
            encoding = "utf-8-sig" if raw.startswith(b"\xef\xbb\xbf") else "utf-8"
            args.output.write_text(updated, encoding=encoding, newline="")
            if os.name != "nt":
                args.output.chmod(0o600)
            args.report.write_text(json.dumps(report, indent=2), encoding="utf-8")
            print("TOML candidate valid; changed keys:", len(report["changed"]),
                  "; custom managed keys preserved:", len(report["preserved"]))
            for key in report["preserved"]:
                print("PRESERVED MANAGED KEY:", key)
        else:
            manifest = json.loads(args.manifest.read_text(encoding="utf-8")) if args.manifest else None
            report = validate(source, manifest)
            if report["managed_drift"]:
                print("INVALID: managed key drift:", ", ".join(report["managed_drift"]))
                return 2
            print("TOML syntax and recorded managed keys: OK")
        return 0
    except (ValueError, OSError, tomllib.TOMLDecodeError, json.JSONDecodeError) as exc:
        print("TOML validation failed (no configuration written):", str(exc), file=sys.stderr)
        return 2


if __name__ == "__main__":
    raise SystemExit(main())
