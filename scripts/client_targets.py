#!/usr/bin/env python3
"""Multi-client model catalog, agent renderer, and JSON config editor for Chuli v2.0.

Supports OpenAI Codex (codex), Claude Code (claude), and Google Antigravity CLI (agy).
Requires Python 3.11+ (stdlib only).
"""
from __future__ import annotations

import argparse
import copy
import hashlib
import json
from pathlib import Path
import re
import sys
import tomllib

ROOT = Path(__file__).resolve().parents[1]
MODELS_PATH = ROOT / "config/models.json"
BOM = b"\xef\xbb\xbf"
SAFE_ID = re.compile(r"^[A-Za-z0-9_.:/-]{1,80}$")

READONLY_ROLES = {"chuli_explorer", "solution_architect", "quality_reviewer"}
EXEC_READONLY_ROLES = {"solution_architect", "quality_reviewer"}

AGY_DISPLAY_NAMES = {
    "orchestrator-argon": "Orchestrator (Argon)",
    "orchestrator-carbon": "Orchestrator (Carbon)",
    "gemini-4-carbon": "Gemini 4 Carbon",
    "gemini-4-argon": "Gemini 4 Argon",
    "gemini-4-hydrogen": "Gemini 4 Hydrogen",
}
AGY_EFFORT_LABELS = {
    "low": "Low",
    "medium": "Medium",
    "high": "High",
    "xhigh": "Extra",
}
AGY_VALID_AGENT_TIERS = {"inherit", "flash", "flash_lite", "pro"}
AGY_AGENT_TIERS = {
    "orchestrator-carbon": "pro",
    "orchestrator-argon": "pro",
    "gemini-4-carbon": "pro",
    "gemini-4-argon": "pro",
    "gemini-4-hydrogen": "flash",
    "inherit": "inherit",
}


def load_catalog() -> dict:
    data = json.loads(MODELS_PATH.read_text(encoding="utf-8"))
    if not isinstance(data, dict) or "clients" not in data or "roles" not in data:
        raise ValueError("Invalid config/models.json structure")
    return data


def role_ids(catalog: dict | None = None) -> list[str]:
    cat = catalog or load_catalog()
    return [entry["id"] for entry in cat["roles"]]


def agy_agent_tier(model: str, effort: str) -> str:
    norm = (model or "").strip().lower()
    if norm in AGY_VALID_AGENT_TIERS:
        return norm
    if norm == "gemini-4-hydrogen" and (effort or "").strip().lower() == "low":
        return "flash_lite"
    if norm == "gemini-4-argon" and (effort or "").strip().lower() == "low":
        return "flash"
    # Historical manifests may contain IDs previously installed as inherit.
    return AGY_AGENT_TIERS.get(norm, "inherit")


def valid_agy_label(label: str) -> bool:
    return bool(label) and len(label.encode("utf-8")) <= 160 and all(
        c.isalnum() or c in " -_.()/+:" for c in label
    )


def agy_orchestrator_label(spec: dict[str, str]) -> str:
    label = spec.get("label", "")
    if label:
        if not isinstance(label, str) or not valid_agy_label(label):
            raise ValueError("Invalid AGY model label")
        return label
    model = spec["model"]
    if model not in AGY_DISPLAY_NAMES and model != "inherit":
        raise ValueError(f"AGY model {model!r} requires its exact CLI label")
    return format_agy_model(model, spec["effort"])


def format_agy_model(model: str, effort: str) -> str:
    if model in ("orchestrator-argon", "orchestrator-carbon"):
        return AGY_DISPLAY_NAMES[model]
    if model == "inherit":
        return "inherit"
    base = AGY_DISPLAY_NAMES.get(model)
    if base:
        label = AGY_EFFORT_LABELS.get(effort.lower(), "Medium")
        return f"{base} ({label})"
    return model


def resolve_role_models(
    client: str,
    preset: str = "",
    overrides: dict[str, dict[str, str]] | None = None,
    catalog: dict | None = None,
) -> tuple[str, dict[str, dict[str, str]]]:
    cat = catalog or load_catalog()
    clients = cat["clients"]
    if client not in clients:
        raise ValueError(f"Unsupported client: {client!r}")
    client_spec = clients[client]
    presets = client_spec["presets"]
    chosen_preset = preset.strip() if preset else client_spec["default_preset"]
    if chosen_preset not in presets and chosen_preset != "custom":
        raise ValueError(f"Unknown preset {chosen_preset!r} for client {client!r}")
    base_preset = chosen_preset if chosen_preset in presets else client_spec["default_preset"]
    resolved = copy.deepcopy(presets[base_preset]["roles"])
    valid_roles = set(role_ids(cat))
    valid_efforts = set(client_spec["efforts"])
    if overrides:
        if not isinstance(overrides, dict):
            raise ValueError("Role model overrides must be an object")
        for role, spec in overrides.items():
            if role not in valid_roles or not isinstance(spec, dict):
                raise ValueError(f"Invalid role override: {role!r}")
            model = str(spec.get("model", "")).strip()
            effort = str(spec.get("effort", resolved[role]["effort"])).strip().lower()
            if not model or not SAFE_ID.fullmatch(model):
                raise ValueError(f"Invalid model identifier for {role}: {model!r}")
            if effort not in valid_efforts:
                raise ValueError(f"Invalid effort {effort!r} for {role} on {client}")
            resolved[role] = {"model": model, "effort": effort}
            if client == "agy":
                if role == "orchestrator" and "label" in spec:
                    label = spec["label"]
                    if not isinstance(label, str) or not valid_agy_label(label):
                        raise ValueError("Invalid AGY model label")
                    resolved[role]["label"] = label
        if resolved != presets[base_preset]["roles"] and not preset:
            chosen_preset = "custom"
    return chosen_preset, resolved


def parse_role_spec(items: list[str], client: str) -> dict[str, dict[str, str]]:
    """Parse ['orchestrator=gpt-6.1-sol:medium', ...] into overrides dict."""
    cat = load_catalog()
    default_preset = cat["clients"][client]["default_preset"]
    defaults = cat["clients"][client]["presets"][default_preset]["roles"]
    overrides: dict[str, dict[str, str]] = {}
    for item in items:
        if "=" not in item:
            raise ValueError(f"Expected ROLE=MODEL[:EFFORT], got {item!r}")
        role, rest = item.split("=", 1)
        role = role.strip()
        if role not in defaults:
            raise ValueError(f"Unknown role {role!r}")
        if ":" in rest:
            model, effort = rest.split(":", 1)
        else:
            model, effort = rest, defaults[role]["effort"]
        if client == "agy" and role != "orchestrator" and model.strip().lower() not in AGY_VALID_AGENT_TIERS | AGY_AGENT_TIERS.keys():
            raise ValueError(f"AGY specialist model {model!r} must use pro, flash, flash_lite or inherit")
        overrides[role] = {"model": model.strip(), "effort": effort.strip().lower()}
    return overrides


def load_source_agents() -> dict[str, dict]:
    agents: dict[str, dict] = {}
    for path in sorted((ROOT / "agents").glob("*.toml")):
        data = tomllib.loads(path.read_bytes().decode("utf-8-sig"))
        agents[path.stem] = data
    return agents


def render_codex_agent(agent: dict, model: str, effort: str) -> bytes:
    lines = [
        f'name = {json.dumps(agent["name"], ensure_ascii=False)}',
        f'description = {json.dumps(agent["description"], ensure_ascii=False)}',
        f'model = {json.dumps(model, ensure_ascii=False)}',
        f'model_reasoning_effort = {json.dumps(effort, ensure_ascii=False)}',
    ]
    if "sandbox_mode" in agent:
        lines.append(f'sandbox_mode = {json.dumps(agent["sandbox_mode"], ensure_ascii=False)}')
    body = agent["developer_instructions"].replace("\r\n", "\n").strip("\n")
    lines.append(f"developer_instructions = '''\n{body}\n'''")
    return ("\n".join(lines) + "\n").encode("utf-8")


def render_claude_agent(agent: dict, model: str, effort: str) -> bytes:
    role = agent["name"]
    lines = [
        "---",
        f"name: {role}",
        f"description: {json.dumps(agent['description'], ensure_ascii=False)}",
        f"model: {model}",
        f"effort: {effort}",
    ]
    if role == "chuli_explorer":
        lines.append("tools: Read, Grep, Glob, WebSearch")
        lines.append("disallowedTools: Write, Edit")
    elif role in EXEC_READONLY_ROLES:
        lines.append("tools: Read, Grep, Glob, Bash, WebSearch")
        lines.append("disallowedTools: Write, Edit")
    lines.append("---")
    body = agent["developer_instructions"].replace("\r\n", "\n").strip("\n")
    lines.append("")
    lines.append(body)
    return ("\n".join(lines) + "\n").encode("utf-8")


def render_agy_agent(agent: dict, model: str, effort: str) -> bytes:
    role = agent["name"]
    agy_tier = agy_agent_tier(model, effort)
    if role == "chuli_explorer":
        tools = "[read_file, search_web]"
    elif role in EXEC_READONLY_ROLES:
        tools = "[read_file, run_command, manage_task, search_web]"
    else:
        tools = "[read_file, edit_file, write_file, run_command, manage_task, search_web]"
    lines = [
        "---",
        f"name: {role}",
        f"description: {json.dumps(agent['description'], ensure_ascii=False)}",
        "subagent: true",
        "mainAgent: false",
        f"model: {agy_tier}",
        f"effort: {effort}",
        f"tools: {tools}",
        "---",
    ]
    body = agent["developer_instructions"].replace("\r\n", "\n").strip("\n")
    lines.append("")
    lines.append(body)
    return ("\n".join(lines) + "\n").encode("utf-8")


def render_agy_plugin_manifest(version: str) -> bytes:
    manifest = {
        "name": "chuli",
        "displayName": "Chuli Engineering Harness",
        "version": version,
        "description": "Orquestador ODD/EWP, 6 especialistas y 25 skills de ingeniería para AGY CLI.",
        "repository": "https://github.com/AlejoPalavecino/codex-engineering-harness",
    }
    return (json.dumps(manifest, indent=2, ensure_ascii=False) + "\n").encode("utf-8")


def render_orchestrator_body(
    client: str,
    role_models: dict[str, dict[str, str]] | None = None,
) -> bytes:
    raw = (ROOT / "config/AGENTS.md").read_bytes().decode("utf-8-sig").replace("\r\n", "\n").rstrip("\n")
    _, resolved = resolve_role_models(client, overrides=role_models)
    if client == "claude":
        raw = raw.replace("# Chuli · orquestador de ingeniería para Codex", "# Chuli · orquestador de ingeniería para Claude Code")
        raw = raw.replace("CODEX_HOME", "CLAUDE_CONFIG_DIR")
        raw = raw.replace("~/.codex", "~/.claude")
    elif client == "agy":
        raw = raw.replace("# Chuli · orquestador de ingeniería para Codex", "# Chuli · orquestador de ingeniería para AGY CLI")
        raw = raw.replace("CODEX_HOME", "AGY_HOME")
        raw = raw.replace("~/.codex", "~/.gemini/config")

    def fmt(role: str) -> str:
        spec = resolved[role]
        if client == "agy":
            return spec.get("label") or format_agy_model(spec["model"], spec["effort"])
        return f"{spec['model']} ({spec['effort']})"

    old_table = (
        "| Rol estable | Encargo |\n"
        "| --- | --- |\n"
        "| `chuli_explorer` | Sol medium. Lectura de rutas, consumidores y contratos. |\n"
        "| `chuli_worker` | Sol medium. Implementación y verificación focalizada. |\n"
        "| `solution_architect` | Alternativas de arquitectura, datos y contratos. |\n"
        "| `product_designer` | Flujos, estados, accesibilidad y dirección visual. |\n"
        "| `test_engineer` | Pruebas derivadas de aceptación y regresión. |\n"
        "| `quality_reviewer` | Revisión independiente sobre un candidato identificado. |"
    )
    new_table = (
        f"Modelo configurado para `orchestrator`: **{fmt('orchestrator')}**.\n\n"
        "| Rol estable | Modelo configurado | Encargo |\n"
        "| --- | --- | --- |\n"
        f"| `chuli_explorer` | `{fmt('chuli_explorer')}` | Lectura de rutas, consumidores y contratos. |\n"
        f"| `chuli_worker` | `{fmt('chuli_worker')}` | Implementación y verificación focalizada. |\n"
        f"| `solution_architect` | `{fmt('solution_architect')}` | Alternativas de arquitectura, datos y contratos. |\n"
        f"| `product_designer` | `{fmt('product_designer')}` | Flujos, estados, accesibilidad y dirección visual. |\n"
        f"| `test_engineer` | `{fmt('test_engineer')}` | Pruebas derivadas de aceptación y regresión. |\n"
        f"| `quality_reviewer` | `{fmt('quality_reviewer')}` | Revisión independiente sobre un candidato identificado. |"
    )
    raw = raw.replace(old_table, new_table)
    return raw.encode("utf-8")


def prepare_json_settings(
    raw: bytes,
    client: str,
    role_models: dict[str, dict[str, str]],
    *,
    configure_models: bool = False,
    replace_managed: bool = False,
) -> tuple[bytes, dict]:
    """Conservatively update Claude Code or AGY CLI settings.json preserving BOM and unknown keys."""
    has_bom = raw.startswith(BOM)
    text = raw.decode("utf-8-sig").strip()
    if not text:
        before: dict = {}
    else:
        before = json.loads(text)
        if not isinstance(before, dict):
            raise ValueError("settings.json root must be a JSON object")
    updated = copy.deepcopy(before)
    changed: list[str] = []
    preserved: list[str] = []
    tracked: dict[str, object] = {}

    orch = role_models["orchestrator"]
    if client == "claude":
        # Claude Code gives this global override precedence over each agent's model.
        env = updated.get("env")
        override = "CLAUDE_CODE_SUBAGENT_MODEL"
        if isinstance(env, dict) and override in env:
            if configure_models:
                del env[override]
                changed.append(f"env.{override}")
            else:
                preserved.append(f"env.{override}")
        targets: list[tuple[tuple[str, ...], object]] = [
            (("model",), orch["model"]),
            (("env", "CLAUDE_CODE_EFFORT_LEVEL"), orch["effort"]),
        ]
    elif client == "agy":
        targets = [
            (("model",), agy_orchestrator_label(orch)),
        ]
    else:
        raise ValueError(f"JSON settings editor does not support {client!r}")

    for path, wanted in targets:
        key = ".".join(path)
        if len(path) == 1:
            current = updated.get(path[0], None)
            exists = path[0] in updated
        else:
            parent = updated.get(path[0])
            if parent is not None and not isinstance(parent, dict):
                raise ValueError(f"settings.json key {path[0]!r} must be an object")
            exists = isinstance(parent, dict) and path[1] in parent
            current = parent.get(path[1]) if exists else None

        if exists and current != wanted and not (replace_managed or configure_models):
            preserved.append(key)
            tracked[key] = current
            continue
        tracked[key] = wanted
        if exists and current == wanted:
            continue
        if len(path) == 1:
            updated[path[0]] = wanted
        else:
            updated.setdefault(path[0], {})[path[1]] = wanted
        changed.append(key)

    if not changed and text:
        out_bytes = raw
    else:
        rendered = (json.dumps(updated, indent=2, ensure_ascii=False) + "\n").encode("utf-8")
        out_bytes = (BOM + rendered) if has_bom else rendered
    return out_bytes, {"changed": changed, "preserved": preserved, "tracked": tracked}


def output_style_warnings(data: dict, client: str | None) -> list[str]:
    """Advise when a Claude Code output style may override Chuli's CES tone. Never edits settings."""
    if client != "claude":
        return []
    style = data.get("outputStyle")
    if not isinstance(style, str) or style.strip().lower() in ("", "default"):
        return []
    return [
        f'outputStyle "{style}" activo en settings.json: un output style se suma al system prompt '
        "y puede contradecir el tono CES de Chuli (voseo, registro, profundidad). "
        "Si querés el tono de Chuli, quitá la clave outputStyle o elegí el estilo Default."
    ]


def validate_json_settings(raw: bytes, state: dict | None = None, client: str | None = None) -> dict:
    text = raw.decode("utf-8-sig").strip()
    data = json.loads(text) if text else {}
    if not isinstance(data, dict):
        raise ValueError("settings.json root must be an object")
    drift: list[str] = []
    if state is not None:
        for key, wanted in state.get("managed_config", {}).items():
            parts = key.split(".")
            cur: object = data
            found = True
            for part in parts:
                if not isinstance(cur, dict) or part not in cur:
                    found = False
                    break
                cur = cur[part]
            if not found or cur != wanted:
                drift.append(key)
    return {"syntax": "valid", "managed_drift": drift, "warnings": output_style_warnings(data, client)}


def main(argv: list[str] | None = None) -> int:
    parser = argparse.ArgumentParser(description=__doc__)
    sub = parser.add_subparsers(dest="action", required=True)

    p_resolve = sub.add_parser("resolve")
    p_resolve.add_argument("--client", choices=("codex", "claude", "agy"), default="codex")
    p_resolve.add_argument("--preset", default="")
    p_resolve.add_argument("--role-models-json", default="")
    p_resolve.add_argument("--role-models-file", type=Path)

    p_stage = sub.add_parser("stage-agents")
    p_stage.add_argument("--client", choices=("codex", "claude", "agy"), default="codex")
    p_stage.add_argument("--preset", default="")
    p_stage.add_argument("--role-models-json", default="")
    p_stage.add_argument("--role-models-file", type=Path)
    p_stage.add_argument("--output-dir", type=Path, required=True)

    p_json = sub.add_parser("json-config")
    p_json.add_argument("mode", choices=("prepare", "validate"))
    p_json.add_argument("--client", choices=("claude", "agy"), required=True)
    p_json.add_argument("--input", type=Path, required=True)
    p_json.add_argument("--output", type=Path)
    p_json.add_argument("--report", type=Path)
    p_json.add_argument("--manifest", type=Path)
    p_json.add_argument("--preset", default="")
    p_json.add_argument("--role-models-json", default="")
    p_json.add_argument("--role-models-file", type=Path)
    p_json.add_argument("--configure-models", action="store_true")
    p_json.add_argument("--preserve-models", action="store_true")
    p_json.add_argument("--replace-managed", action="store_true")

    args = parser.parse_args(argv)
    try:
        raw_role_models = getattr(args, "role_models_json", "")
        role_models_file = getattr(args, "role_models_file", None)
        if role_models_file is not None:
            raw_role_models = role_models_file.read_text(encoding="utf-8-sig")
        overrides = json.loads(raw_role_models) if raw_role_models else None

        if args.action == "resolve":
            preset, roles = resolve_role_models(args.client, args.preset, overrides)
            print(json.dumps({"client": args.client, "preset": preset, "role_models": roles}, ensure_ascii=False))
            return 0

        if args.action == "stage-agents":
            preset, roles = resolve_role_models(args.client, args.preset, overrides)
            agents = load_source_agents()
            args.output_dir.mkdir(parents=True, exist_ok=True)
            version = (ROOT / "VERSION").read_text(encoding="utf-8").strip()
            default_codex = load_catalog()["clients"]["codex"]["presets"]["codex-balanced"]["roles"]
            staged: dict[str, str] = {}
            for name, agent in agents.items():
                spec = roles[name]
                if args.client == "codex":
                    rel = f"agents/{name}.toml"
                    out = args.output_dir / rel
                    out.parent.mkdir(parents=True, exist_ok=True)
                    if roles == default_codex:
                        out.write_bytes((ROOT / "agents" / f"{name}.toml").read_bytes())
                    else:
                        out.write_bytes(render_codex_agent(agent, spec["model"], spec["effort"]))
                    staged[rel] = str(out)
                elif args.client == "claude":
                    rel = f"agents/{name}.md"
                    out = args.output_dir / rel
                    out.parent.mkdir(parents=True, exist_ok=True)
                    out.write_bytes(render_claude_agent(agent, spec["model"], spec["effort"]))
                    staged[rel] = str(out)
                elif args.client == "agy":
                    rel = f"plugins/chuli/agents/{name}.md"
                    out = args.output_dir / rel
                    out.parent.mkdir(parents=True, exist_ok=True)
                    out.write_bytes(render_agy_agent(agent, spec["model"], spec["effort"]))
                    staged[rel] = str(out)
            if args.client == "agy":
                manifest_rel = "plugins/chuli/plugin.json"
                manifest_out = args.output_dir / manifest_rel
                manifest_out.parent.mkdir(parents=True, exist_ok=True)
                manifest_out.write_bytes(render_agy_plugin_manifest(version))
                staged[manifest_rel] = str(manifest_out)
            print(json.dumps({"preset": preset, "role_models": roles, "staged": staged}, ensure_ascii=False))
            return 0

        if args.action == "json-config":
            raw = args.input.read_bytes() if args.input.is_file() else b""
            if args.mode == "prepare":
                _, roles = resolve_role_models(args.client, args.preset, overrides)
                updated, report = prepare_json_settings(
                    raw,
                    args.client,
                    roles,
                    configure_models=(args.configure_models or bool(args.preset or raw_role_models)) and not args.preserve_models,
                    replace_managed=args.replace_managed,
                )
                args.output.write_bytes(updated)
                args.report.write_text(json.dumps(report, indent=2), encoding="utf-8")
                print(
                    "JSON candidate valid; changed keys:",
                    len(report["changed"]),
                    "; custom managed keys preserved:",
                    len(report["preserved"]),
                )
            else:
                manifest = json.loads(args.manifest.read_text(encoding="utf-8")) if args.manifest else None
                report = validate_json_settings(raw, manifest, client=args.client)
                for warning in report["warnings"]:
                    print("WARNING:", warning)
                if report["managed_drift"]:
                    print("INVALID: managed key drift:", ", ".join(report["managed_drift"]))
                    return 2
                print("JSON syntax and recorded managed keys: OK")
            return 0
    except (ValueError, OSError, json.JSONDecodeError, UnicodeError) as exc:
        print("Client target error:", str(exc), file=sys.stderr)
        return 2
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
