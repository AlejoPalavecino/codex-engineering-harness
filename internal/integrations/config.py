"""Embedded, conservative MCP config editor. Python 3.11+, no dependencies.

Input/output is JSON. Never return existing configuration values or parser errors.
The journal stores only Chuli's public addition, never the user's full config.
"""
import hashlib
import json
import os
from pathlib import Path, PureWindowsPath
import stat
import sys
import tempfile
import tomllib
from urllib.parse import urlsplit


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


def parse(raw):
    try:
        data = tomllib.loads(raw.decode("utf-8-sig"))
        if not isinstance(data.get("mcp_servers", {}), dict) or not isinstance(data.get("plugins", {}), dict):
            raise ValueError()
        return data
    except (ValueError, UnicodeError):
        raise SafeError("config.toml inválido; corregilo antes de continuar.") from None


def endpoint(value):
    if not isinstance(value, str):
        return None
    try:
        url = urlsplit(value)
        return (url.scheme.lower(), url.netloc.lower(), url.path.rstrip("/"))
    except ValueError:
        return None


def related(data, name, entry):
    servers = data.get("mcp_servers", {})
    if name in servers:
        return True
    for value in servers.values():
        if not isinstance(value, dict):
            continue
        if entry.get("url") and endpoint(value.get("url")) == endpoint(entry["url"]):
            return True
        command = value.get("command", "")
        if name == "engram" and isinstance(command, str) and PureWindowsPath(command).stem.lower() == "engram":
            return True
    return any(name in key.lower() and isinstance(value, dict) and value.get("enabled", True)
               for key, value in data.get("plugins", {}).items())


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


def execute(request):
    home = Path(request["home"])
    if not home.is_absolute():
        raise SafeError("CODEX_HOME debe ser absoluto.")
    name, entry, action = request["id"], request["entry"], request["action"]
    if name not in {"context7", "exa", "github", "engram"} or action not in {"inspect", "configure", "undo"}:
        raise SafeError("Operación no admitida.")
    config = home / "config.toml"
    journal = home / "chuli-integrations" / (name + ".json")
    raw, record_raw = read(config), read(journal, 65536)
    data = parse(raw)
    record = json.loads(record_raw) if record_raw else None
    if record is not None and (not isinstance(record, dict) or record.get("version") != 1
                               or record.get("entry") != entry or record.get("block") != addition(name, entry)
                               or not isinstance(record.get("original_exists"), bool)):
        raise SafeError("Registro de reversión incompatible; no se modificó el perfil.")
    revision = hashlib.sha256(bytes([config.exists()]) + raw + b"\x00" + record_raw).hexdigest()
    exact = data.get("mcp_servers", {}).get(name) == entry
    exists = related(data, name, entry)
    owned = bool(record and exact and raw.count(record["block"].encode()) == 1)
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
        if action == "configure":
            if exists:
                return dict(message="Ya existe una conexión o plugin relacionado. Se conservó sin cambios.")
            if record:
                raise SafeError("Existe un registro previo sin conexión intacta; requiere revisión manual.")
            block = addition(name, entry)
            proposed = raw + block.encode("utf-8")
            parsed = parse(proposed)
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
        parsed = parse(proposed)
        if name in parsed.get("mcp_servers", {}):
            raise SafeError("La conexión tiene modificaciones; revisala manualmente.")
        if proposed or record["original_exists"]:
            atomic(config, proposed)
        else:
            config.unlink()
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
        # No tracebacks, paths, TOML lines, or user values cross the process boundary.
        print(json.dumps({"error": "No se pudo operar de forma segura. Revisá TOML, permisos, bloqueo y vista previa."}))
        sys.exit(1)
