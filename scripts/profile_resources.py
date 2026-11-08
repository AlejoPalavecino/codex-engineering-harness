"""Inventory and preflight of files managed inside one Codex profile (Python 3.11+)."""
from __future__ import annotations

import argparse
import hashlib
import json
import os
from pathlib import Path
import re
import stat
import sys

ROOT = Path(__file__).resolve().parents[1]
BEGIN = b"<!-- chuli:orchestrator -->"
END = b"<!-- /chuli:orchestrator -->"
GENTLE = re.compile(rb"<!--\s*/?gentle-ai:(orchestrator|sdd-orchestrator|agent-routing)\s*-->", re.I)
BOM = b"\xef\xbb\xbf"
LEGACY_AGENTS = {f"agents/{name}.toml" for name in
                 ("architect", "explorer", "reviewer", "slice_worker", "test_runner")}
BACKUP_ID = re.compile(r"\d{8}-\d{6}-[0-9a-f]{8}\Z")
RESERVED = re.compile(r"(?:CON|PRN|AUX|NUL|COM[1-9¹²³]|LPT[1-9¹²³])(?:\.|$)", re.I)


def digest(data: bytes) -> str:
    return hashlib.sha256(data).hexdigest()


def safe_path(path: Path, *, directory: bool = False) -> None:
    """Check broken links, Windows junctions, and hard-linked files before any write."""
    for component in (*reversed(path.parents), path):
        try:
            info = component.lstat()
        except FileNotFoundError:
            return
        # On macOS these system-owned aliases are used by the default temporary directory.
        if sys.platform == "darwin" and component in (Path("/tmp"), Path("/var")) and component.resolve() == Path("/private", component.name):
            continue
        if stat.S_ISLNK(info.st_mode) or getattr(info, "st_file_attributes", 0) & stat.FILE_ATTRIBUTE_REPARSE_POINT:
            raise ValueError(f"Symbolic link or reparse point rejected: {component}")
        if component != path or directory:
            if not stat.S_ISDIR(info.st_mode):
                raise ValueError(f"Expected a directory: {component}")
        elif not stat.S_ISREG(info.st_mode) or info.st_nlink != 1:
            raise ValueError(f"Expected a regular file with one link: {component}")


def valid_relative(name: str) -> bool:
    return isinstance(name, str) and all(
        part and part not in {".", ".."} and not part.endswith((" ", ".")) and
        not RESERVED.match(part) and not re.search(r'[\\:*?"<>|\x00-\x1f]', part)
        for part in name.split("/")
    )


def inventory(root: Path) -> dict[str, str]:
    result = {"AGENTS.md": "config/AGENTS.md", "harness/bin/taskboard.py": "scripts/taskboard.py",
              "harness/templates/feature-tasks.md": "templates/feature-tasks.md",
              "harness/THIRD_PARTY_NOTICES.md": "THIRD_PARTY_NOTICES.md"}
    for folder, target in (("skills", "skills"), ("agents", "agents"),
                           ("docs/protocols", "harness/docs/protocols")):
        base = root / folder
        safe_path(base, directory=True)
        if not base.is_dir():
            raise ValueError(f"Missing source directory: {folder}")
        # os.walk ignores unreadable directories by default. Incomplete packages must fail.
        def fail(error: OSError) -> None:
            raise error
        for parent, dirs, files in os.walk(base, onerror=fail):
            for name in dirs:
                safe_path(Path(parent, name), directory=True)
            if folder == "agents":
                dirs.clear()
            for name in sorted(files):
                path = Path(parent, name)
                if folder == "agents" and path.suffix != ".toml":
                    continue
                relative = path.relative_to(base).as_posix()
                if not valid_relative(relative):
                    raise ValueError(f"Unsafe source name: {relative}")
                result[f"{target}/{relative}"] = path.relative_to(root).as_posix()
    if len({name.casefold() for name in result}) != len(result):
        raise ValueError("Case-insensitive collision in package paths")
    for source in result.values():
        safe_path(root / source)
        if not (root / source).is_file():
            raise ValueError(f"Missing source: {source}")
    return dict(sorted(result.items()))


def preflight(home: Path, action: str) -> dict[str, str]:
    if not home.is_absolute() or ".." in home.parts or home == home.parent:
        raise ValueError("Expected an absolute profile directory below the filesystem root")
    if home == ROOT or home.is_relative_to(ROOT) or ROOT.is_relative_to(home):
        raise ValueError("Profile and source checkout must be separate directories")
    resources = inventory(ROOT)
    allowed = set(resources) | LEGACY_AGENTS | {"config.toml"}
    safe_path(home, directory=True)
    for name in (*allowed, "harness-install/state.json", "AGENTS.override.md"):
        safe_path(home / name)
    safe_path(home / "backups/harness", directory=True)
    if action in {"install", "verify"}:
        override = home / "AGENTS.override.md"
        if override.exists() and override.read_bytes().strip():
            raise ValueError(f"Instruction source {override} overrides AGENTS.md. Choose a separate Chuli profile or move this override.")
        instructions = home / "AGENTS.md"
        if instructions.is_file():
            content = instructions.read_bytes()
            check_gentle(content, instructions)
            if action == "verify" and not block_span(content.removeprefix(BOM)):
                raise ValueError(f"No Chuli orchestrator block in {instructions}. Run the v1 installer to migrate.")
    state_path = home / "harness-install/state.json"
    if state_path.exists():
        state = json.loads(state_path.read_bytes())
        validate_state_paths(home, state, allowed, action == "rollback")
        if action == "verify":
            data = (home / "AGENTS.md").read_bytes().removeprefix(BOM)
            span = block_span(data)
            if not span or digest(data[span[0]:span[1]]) != state.get("orchestrator_sha256"):
                raise ValueError("Chuli orchestrator block differs from the installed profile. Inspect AGENTS.md and run the installer after review.")
        if action == "rollback" and state.get("previous_state_existed"):
            previous_path = Path(state["backup_dir"]) / "_previous_state.json"
            previous_bytes = previous_path.read_bytes()
            if state.get("previous_state_sha256") and digest(previous_bytes) != state["previous_state_sha256"].lower():
                raise ValueError("Previous manifest checksum mismatch")
            previous = json.loads(previous_bytes)
            validate_state_paths(home, previous, allowed, False)
    return resources


def validate_state_paths(home: Path, state: dict, allowed: set[str], rollback: bool) -> None:
    if not isinstance(state, dict) or not isinstance(state.get("files"), list):
        raise ValueError("Invalid installation manifest")
    if not re.fullmatch(r"[0-9]+\.[0-9]+\.[0-9]+", str(state.get("version", ""))):
        raise ValueError("Invalid installation version")
    if any(state.get(key, 1) not in (1, 2) for key in ("schema_version", "rollback_schema_version")):
        raise ValueError("Unsupported installation or rollback manifest schema")
    seen = set()
    for field in ("managed_files", "preserved_files"):
        managed = state.get(field, {})
        if not isinstance(managed, dict) or any(key not in allowed or not isinstance(value, str)
                or not re.fullmatch(r"[0-9a-fA-F]{64}", value) for key, value in managed.items()):
            raise ValueError(f"Invalid resource or digest in {field}")
    for item in state["files"]:
        key = item.get("path") if isinstance(item, dict) else None
        if key not in allowed or key in seen:
            raise ValueError(f"Unknown or duplicate rollback target: {key!r}")
        if not isinstance(item.get("existed"), bool):
            raise ValueError(f"Invalid previous-existence flag: {key}")
        for field in ("installed_sha256", "backup_sha256"):
            value = item.get(field)
            if value is not None and not (isinstance(value, str) and re.fullmatch(r"[a-fA-F0-9]{64}", value)):
                raise ValueError(f"Invalid {field}: {key}")
        seen.add(key)
    raw = state.get("backup_dir")
    if raw is not None or rollback:
        if not isinstance(raw, str) or not raw:
            raise ValueError("Missing backup directory")
        backup = Path(raw)
        if backup.parent != home / "backups/harness" or not BACKUP_ID.fullmatch(backup.name):
            raise ValueError("Expected a backup belonging to the selected profile")
        safe_path(backup, directory=True)
        if rollback:
            safe_path(backup / "_previous_state.json")
            for name in seen:
                safe_path(backup / name)


def block_span(data: bytes) -> tuple[int, int] | None:
    if BEGIN not in data and END not in data:
        return None
    if data.count(BEGIN) != 1 or data.count(END) != 1:
        raise ValueError("Incomplete or duplicate Chuli orchestrator markers in AGENTS.md")
    match = re.search(rb"(?ms)^" + re.escape(BEGIN) + rb"\r?\n.*?^" + re.escape(END) + rb"(?:\r?\n|\Z)", data)
    if not match:
        raise ValueError("Invalid Chuli orchestrator block in AGENTS.md")
    return match.span()


def check_gentle(data: bytes, path: Path) -> None:
    match = GENTLE.search(data)
    if match:
        origin = match[1].decode()
        raise ValueError(f"Active Gentle AI instructions ({origin}) in {path}. Choose a separate Chuli profile or remove the Gentle routing block first.")


def prepare_instructions(home: Path, output: Path, replace: bool) -> dict:
    path = home / "AGENTS.md"
    raw = path.read_bytes() if path.is_file() else b""
    raw.decode("utf-8-sig")  # Preserve all surrounding bytes, including the BOM and CRLF.
    prefix = BOM if raw.startswith(BOM) else b""
    data = raw[len(prefix):]
    source = (ROOT / "config/AGENTS.md").read_bytes().decode("utf-8-sig").replace("\r\n", "\n").encode().rstrip(b"\n")
    if BEGIN in source or END in source:
        raise ValueError("config/AGENTS.md must contain only the managed body")
    block = BEGIN + b"\n" + source + b"\n" + END + b"\n"
    span = block_span(data)
    state_path = home / "harness-install/state.json"
    state = json.loads(state_path.read_bytes()) if state_path.exists() else {}
    prior_hash = state.get("orchestrator_baseline_sha256", state.get("orchestrator_sha256", ""))
    canonical_hash = digest(block)
    preserved = False
    if span:
        start, end = span
        existing = data[start:end]
        if existing != block and digest(existing) != prior_hash and not replace:
            block = existing
            preserved = True
        merged = data[:start] + block + data[end:]
    else:
        # The v0.2 releases owned the entire file. Recognize official LF and CRLF copies.
        baselines = json.loads((ROOT / "scripts/legacy-managed-hashes.json").read_bytes())
        known = {entry.get("AGENTS.md", "").lower() for entry in baselines.values()}
        if digest(data.replace(b"\r\n", b"\n")) in known:
            merged = block
        elif data.startswith("# Codex Engineering Harness — instrucciones globales".encode()):
            raise ValueError(f"Customized legacy orchestrator in {path}. Move your personal text outside the old instructions or choose a separate Chuli profile.")
        else:
            separator = b"" if not data or data.endswith(b"\n") else b"\n"
            merged = data + separator + block
    output.write_bytes(prefix + merged)
    return {"orchestrator_sha256": digest(block), "baseline_sha256": canonical_hash,
            "preserved": preserved}


def undo_instructions(home: Path, output: Path) -> None:
    """Restore only the managed block when the user edited neighboring instructions."""
    state = json.loads((home / "harness-install/state.json").read_bytes())
    current = (home / "AGENTS.md").read_bytes()
    bom = BOM if current.startswith(BOM) else b""
    data = current[len(bom):]
    span = block_span(data)
    if not span or digest(data[span[0]:span[1]]) != state.get("orchestrator_sha256"):
        raise ValueError("Managed orchestrator edited after installation; rollback stopped")
    original = Path(state["backup_dir"]) / "AGENTS.md"
    existed = next(record["existed"] for record in state["files"] if record["path"] == "AGENTS.md")
    previous = original.read_bytes().removeprefix(BOM) if existed else b""
    old_span = block_span(previous)
    if old_span:
        previous_block = previous[old_span[0]:old_span[1]]
    elif previous.startswith("# Codex Engineering Harness — instrucciones globales".encode()):
        previous_block = previous
    else:
        previous_block = b""
    before, after = data[:span[0]], data[span[1]:]
    if not old_span and previous and not previous.endswith(b"\n") and before == previous + b"\n" and not after:
        before = previous
    output.write_bytes(bom + before + previous_block + after)


def main() -> None:
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("action", choices=("install", "verify", "rollback", "instructions", "undo-instructions"))
    parser.add_argument("--home", type=Path, required=True)
    parser.add_argument("--output", type=Path)
    parser.add_argument("--replace", action="store_true")
    args = parser.parse_args()
    if args.action == "instructions":
        result = prepare_instructions(args.home, args.output, args.replace)
    elif args.action == "undo-instructions":
        undo_instructions(args.home, args.output)
        return
    else:
        result = preflight(args.home, args.action)
    print(json.dumps(result, ensure_ascii=True))


if __name__ == "__main__":
    try:
        main()
    except (ValueError, OSError, TypeError, KeyError) as exc:
        print(f"Profile preflight failed: {exc}", file=sys.stderr)
        sys.exit(2)
