"""Embedded, conservative multi-client MCP config editor. Python 3.11+, no dependencies.

Input/output is JSON. Never return existing configuration values, secrets, or parser errors.
The journal stores only Chuli's public addition, never the user's full config.
"""
import copy
import hashlib
import json
import os
from pathlib import Path, PureWindowsPath
import re
import stat
import sys
import tempfile
import tomllib
from urllib.parse import urlsplit

SAFE_ID = re.compile(r"^[A-Za-z0-9][A-Za-z0-9_-]{0,48}$")
VALID_CLIENTS = {"codex", "claude", "agy"}


class SafeError(ValueError):
    """A diagnostic authored here, safe to show without configuration values."""


def safe_path(path):
    for part in (path, *path.parents):
        try:
            info = part.lstat()
        except FileNotFoundError:
            continue
        # st_file_attributes is available on Python 3.11 Windows too, unlike
        # Path.is_junction (3.12+). Refuse all reparse points conservatively.
        reparse = getattr(info, "st_file_attributes", 0) & getattr(stat, "FILE_ATTRIBUTE_REPARSE_POINT", 0x400)
        if stat.S_ISLNK(info.st_mode) or reparse:
            raise SafeError("Ruta simbólica no admitida; usá un destino local regular.")
        if part.exists() and part != path and not part.is_dir():
            raise SafeError("El destino no tiene una estructura de directorios válida.")


def read(path, limit=2 * 1024 * 1024):
    safe_path(path)
    if not path.exists():
        return b""
    if not stat.S_ISREG(path.stat().st_mode) or path.stat().st_size > limit:
        raise SafeError("Archivo no regular o demasiado grande.")
    return path.read_bytes()


def parse_toml(raw):
    try:
        data = tomllib.loads(raw.decode("utf-8-sig"))
        if not isinstance(data.get("mcp_servers", {}), dict) or not isinstance(data.get("plugins", {}), dict):
            raise ValueError()
        return data
    except (ValueError, UnicodeError):
        raise SafeError("config.toml inválido; corregilo antes de continuar.") from None


def parse_json_cfg(raw, filename):
    try:
        text = raw.decode("utf-8-sig").strip()
        if not text:
            return {}
        data = json.loads(text)
        if not isinstance(data, dict) or not isinstance(data.get("mcpServers", {}), dict):
            raise ValueError()
        return data
    except (ValueError, UnicodeError):
        raise SafeError(f"{filename} inválido; corregilo antes de continuar.") from None


def endpoint(value):
    if not isinstance(value, str):
        return None
    try:
        url = urlsplit(value)
        return (url.scheme.lower(), url.netloc.lower(), url.path.rstrip("/"))
    except ValueError:
        return None


def related_servers(servers, plugins, name, entry):
    if name in servers:
        return True
    target_ep = endpoint(entry.get("url")) if entry and entry.get("url") else None
    for value in servers.values():
        if not isinstance(value, dict):
            continue
        raw_url = value.get("url") or value.get("serverUrl")
        if target_ep and endpoint(raw_url) == target_ep:
            return True
        command = value.get("command", "")
        if name == "engram" and isinstance(command, str) and PureWindowsPath(command).stem.lower() == "engram":
            return True
    return any(name in key.lower() and isinstance(value, dict) and value.get("enabled", True)
               for key, value in plugins.items())


def atomic(path, content):
    safe_path(path)
    path.parent.mkdir(parents=True, exist_ok=True)
    fd, temporary = tempfile.mkstemp(prefix=".chuli-", dir=path.parent)
    try:
        with os.fdopen(fd, "wb") as stream:
            stream.write(content)
            stream.flush()
            os.fsync(stream.fileno())
        os.replace(temporary, path)
    finally:
        if os.path.exists(temporary):
            os.unlink(temporary)


def addition(name, entry):
    block = "\n# chuli integration: " + name + "\n[mcp_servers." + name + "]\n"
    return block + "".join(key + " = " + json.dumps(value, ensure_ascii=False) + "\n"
                           for key, value in sorted(entry.items()))


def json_entry_for_client(entry, client):
    if entry.get("url"):
        out = {"url": entry["url"]}
        if entry.get("bearer_token_env_var"):
            env_name = entry["bearer_token_env_var"]
            out["headers"] = {"Authorization": f"Bearer ${{{env_name}}}"}
        return out
    out = {"command": entry["command"]}
    if entry.get("args"):
        out["args"] = list(entry["args"])
    return out


def resolve_target(home, client, name):
    if client == "codex":
        return home / "config.toml", home / "chuli-integrations" / f"{name}.json", "toml"
    if client == "agy":
        return home / "mcp_config.json", home / "chuli-integrations" / f"agy-{name}.json", "json"
    # Claude Code stores user-scope mcpServers in ~/.claude.json alongside ~/.claude,
    # while isolated test sandboxes keep .claude.json inside the sandbox directory.
    cfg = (home.parent / ".claude.json") if home.name == ".claude" else (home / ".claude.json")
    return cfg, home / "chuli-integrations" / f"claude-{name}.json", "json"


def summarize_servers(home, client, data, fmt):
    servers = data.get("mcp_servers", {}) if fmt == "toml" else data.get("mcpServers", {})
    out = []
    for srv_name in sorted(servers):
        val = servers[srv_name]
        if not isinstance(val, dict):
            continue
        _, journal, _ = resolve_target(home, client, srv_name)
        managed = False
        token_env = ""
        if journal.is_file():
            try:
                rec = json.loads(read(journal, 65536))
                managed = isinstance(rec, dict) and rec.get("entry") == val
                if managed:
                    if isinstance(val.get("bearer_token_env_var"), str):
                        token_env = val["bearer_token_env_var"]
                    elif isinstance(val.get("headers"), dict):
                        auth = str(val["headers"].get("Authorization", ""))
                        m_env = re.fullmatch(r"Bearer \$\{([A-Z][A-Z0-9_]{1,60})\}", auth)
                        if m_env:
                            token_env = m_env.group(1)
            except SafeError:
                managed = False
        raw_url = val.get("url") or val.get("serverUrl")
        if isinstance(raw_url, str) and raw_url:
            out.append({"id": srv_name, "transport": "http", "target": raw_url, "managed": managed, "token_env": token_env})
        else:
            cmd = str(val.get("command", ""))
            args = val.get("args", [])
            arg_str = " ".join(str(a) for a in args) if isinstance(args, list) else ""
            out.append({"id": srv_name, "transport": "stdio", "target": (cmd + " " + arg_str).strip(), "managed": managed, "token_env": ""})
    return out


def execute(request):
    home = Path(request["home"])
    if not home.is_absolute():
        raise SafeError("CODEX_HOME debe ser absoluto.")
    client = request.get("client", "codex") or "codex"
    if client not in VALID_CLIENTS:
        raise SafeError("Cliente no admitido.")
    action = request["action"]
    if action not in {"list", "inspect", "configure", "undo"}:
        raise SafeError("Operación no admitida.")

    if action == "list":
        config, _, fmt = resolve_target(home, client, "list")
        raw = read(config)
        data = parse_toml(raw) if fmt == "toml" else parse_json_cfg(raw, config.name)
        return {"servers": summarize_servers(home, client, data, fmt)}

    name = request.get("id", "")
    entry = request.get("entry")
    if not SAFE_ID.fullmatch(name) or not isinstance(entry, dict) or (not entry.get("url") and not entry.get("command")):
        raise SafeError("Operación no admitida.")

    config, journal, fmt = resolve_target(home, client, name)
    raw, record_raw = read(config), read(journal, 65536)
    data = parse_toml(raw) if fmt == "toml" else parse_json_cfg(raw, config.name)
    record = json.loads(record_raw) if record_raw else None

    if fmt == "toml":
        if record is not None and (not isinstance(record, dict) or record.get("version") != 1
                                   or record.get("entry") != entry or record.get("block") != addition(name, entry)
                                   or not isinstance(record.get("original_exists"), bool)):
            raise SafeError("Registro de reversión incompatible; no se modificó el perfil.")
        revision = hashlib.sha256(bytes([config.exists()]) + raw + b"\x00" + record_raw).hexdigest()
        exact = data.get("mcp_servers", {}).get(name) == entry
        exists = related_servers(data.get("mcp_servers", {}), data.get("plugins", {}), name, entry)
        owned = bool(record and exact and raw.count(record["block"].encode()) == 1)
    else:
        target_entry = json_entry_for_client(entry, client)
        if record is not None and (not isinstance(record, dict) or record.get("version") != 1
                                   or record.get("client") != client or record.get("entry") != target_entry
                                   or not isinstance(record.get("original_exists"), bool)):
            raise SafeError("Registro de reversión incompatible; no se modificó el perfil.")
        revision = hashlib.sha256(bytes([config.exists()]) + raw + b"\x00" + record_raw).hexdigest()
        servers = data.get("mcpServers", {})
        exact = servers.get(name) == target_entry
        exists = related_servers(servers, {}, name, entry)
        owned = bool(record and exact)

    if action == "inspect":
        return dict(revision=revision, exists=exists, matched=exact, managed=owned)

    home.mkdir(parents=True, exist_ok=True)
    lock = home / ".chuli-integrations.lock"
    safe_path(lock)
    try:
        lock.mkdir()
    except FileExistsError:
        raise SafeError("Otro asistente está trabajando. Revisá el bloqueo antes de reintentar.") from None
    try:
        # Recheck after acquiring our lock. External editors must still be closed.
        if request.get("expected") != revision or read(config) != raw or read(journal, 65536) != record_raw:
            raise SafeError("La configuración cambió; revisá una nueva vista previa.")

        if fmt == "toml":
            if action == "configure":
                if exists:
                    return dict(message="Ya existe una conexión o plugin relacionado. Se conservó sin cambios.")
                if record:
                    raise SafeError("Existe un registro previo sin conexión intacta; requiere revisión manual.")
                block = addition(name, entry)
                proposed = raw + block.encode("utf-8")
                parsed = parse_toml(proposed)
                if parsed.get("mcp_servers", {}).get(name) != entry:
                    raise SafeError("No se puede agregar una tabla a este formato TOML sin reescribirlo.")
                record = dict(version=1, entry=entry, block=block, original_exists=config.exists())
                atomic(journal, json.dumps(record, ensure_ascii=False).encode("utf-8"))
                try:
                    atomic(config, proposed)
                except OSError:
                    journal.unlink()
                    raise
                return dict(message="Conexión configurada. Todavía falta comprobarla; abrí una sesión nueva de Codex.")
            if not owned:
                raise SafeError("No hay un bloque propio intacto para deshacer. Se preservó la configuración.")
            proposed = raw.replace(record["block"].encode("utf-8"), b"", 1)
            parsed = parse_toml(proposed)
            if name in parsed.get("mcp_servers", {}):
                raise SafeError("La conexión tiene modificaciones; revisala manualmente.")
            if proposed or record["original_exists"]:
                atomic(config, proposed)
            else:
                config.unlink()
            journal.unlink()
            return dict(message="Se deshizo sólo la conexión agregada por Chuli. Credenciales y programas se conservan.")

        # JSON branch (agy / claude)
        target_entry = json_entry_for_client(entry, client)
        if action == "configure":
            if exists:
                return dict(message="Ya existe una conexión relacionada. Se conservó sin cambios.")
            if record:
                raise SafeError("Existe un registro previo sin conexión intacta; requiere revisión manual.")
            updated = copy.deepcopy(data)
            updated.setdefault("mcpServers", {})[name] = target_entry
            proposed = (json.dumps(updated, indent=2, ensure_ascii=False) + "\n").encode("utf-8")
            parse_json_cfg(proposed, config.name)
            record = dict(version=1, client=client, entry=target_entry, original_exists=config.exists())
            atomic(journal, json.dumps(record, ensure_ascii=False).encode("utf-8"))
            try:
                atomic(config, proposed)
            except OSError:
                journal.unlink()
                raise
            return dict(message=f"Conexión configurada en {config.name}. Abrí una sesión nueva para usarla.")

        if not owned:
            raise SafeError("No hay una conexión propia intacta para deshacer. Se preservó la configuración.")
        updated = copy.deepcopy(data)
        updated.get("mcpServers", {}).pop(name, None)
        if not updated.get("mcpServers") and len(updated) == 1 and not record["original_exists"]:
            if config.exists():
                config.unlink()
        else:
            proposed = (json.dumps(updated, indent=2, ensure_ascii=False) + "\n").encode("utf-8")
            parse_json_cfg(proposed, config.name)
            atomic(config, proposed)
        journal.unlink()
        return dict(message="Se deshizo sólo la conexión agregada por Chuli. Credenciales y programas se conservan.")
    finally:
        lock.rmdir()


if __name__ == "__main__":
    try:
        print(json.dumps(execute(json.load(sys.stdin)), ensure_ascii=False))
    except SafeError as error:
        print(json.dumps({"error": str(error)}, ensure_ascii=False))
        sys.exit(1)
    except (OSError, ValueError, KeyError, TypeError):
        # No tracebacks, paths, TOML/JSON lines, or user values cross the process boundary.
        print(json.dumps({"error": "No se pudo operar de forma segura. Revisá configuración, permisos, bloqueo y vista previa."}))
        sys.exit(1)
