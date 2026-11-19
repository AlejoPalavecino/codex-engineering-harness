"""Prepare checksums or reverify a complete candidate before manual publication."""

import argparse
import hashlib
import json
from pathlib import Path
import re
import subprocess
import sys
import tarfile
from tempfile import NamedTemporaryFile
import zipfile

from package_release import profile_files

ROOT = Path(__file__).resolve().parents[1]
CHECKSUMS = "SHA256SUMS.txt"


def verify_metadata(version):
    actual = (ROOT / "VERSION").read_text(encoding="utf-8").strip()
    if not re.fullmatch(r"\d+\.\d+\.\d+", version, flags=re.ASCII) or version != actual:
        raise ValueError("Requested version must match the repository VERSION")
    changelog = (ROOT / "CHANGELOG.md").read_text(encoding="utf-8").splitlines()
    prefix = f"## [{version}]"
    if not any(line == prefix or line.startswith(prefix + " ") for line in changelog):
        raise ValueError(f"CHANGELOG.md needs an entry for {version}")
    notes = ROOT / "docs" / "releases" / f"v{version}.md"
    if not notes.is_file() or not notes.read_text(encoding="utf-8").strip():
        raise ValueError(f"Missing delivery notes: {notes.name}")


def verify_zip(path, version=None):
    with zipfile.ZipFile(path) as archive:
        damaged = archive.testzip()
        if damaged:
            raise ValueError(f"Invalid ZIP member: {damaged}")
        if version is None:
            if archive.namelist() != ["chuli.exe"] or not archive.getinfo("chuli.exe").file_size:
                raise ValueError("Expected a single non-empty chuli.exe")
        else:
            if sorted(archive.namelist()) != profile_files():
                raise ValueError("Incorrect profile contents for this checkout")
            if archive.read("VERSION").decode().strip() != version:
                raise ValueError("Incorrect profile VERSION")


def verify_tar(path):
    with tarfile.open(path, "r:gz") as archive:
        members = archive.getmembers()
        if len(members) != 1 or members[0].name != "chuli" or not members[0].isfile() or not members[0].size:
            raise ValueError("Expected a single non-empty chuli executable")
        if not members[0].mode & 0o100:
            raise ValueError("chuli is missing owner execute permission")
        while archive.fileobj.read(1024 * 1024):
            pass  # Read through the gzip trailer to verify its checksum.


def digest(path):
    with path.open("rb") as asset:
        return hashlib.file_digest(asset, "sha256").hexdigest()


def verify_published(directory, metadata):
    expected = {path.name: "sha256:" + digest(path) for path in directory.iterdir()}
    assets = metadata.get("assets") if isinstance(metadata, dict) else None
    if (not isinstance(assets, list) or len(assets) != len(expected)
            or metadata.get("draft") is not False or metadata.get("prerelease") is not False
            or not all(isinstance(asset, dict) and isinstance(asset.get("name"), str) for asset in assets)
            or {asset["name"]: asset.get("digest") for asset in assets} != expected):
        raise ValueError("Existing release needs human review: status or asset digests do not match")


def candidate_manifest(version, directory, check):
    verify_metadata(version)
    files = (
        f"chuli-profile-v{version}.zip", f"chuli_v{version}_windows_amd64.zip",
        f"chuli_v{version}_linux_amd64.tar.gz", f"chuli_v{version}_darwin_arm64.tar.gz",
        "install-chuli-windows.ps1",
    )
    expected = {*files, CHECKSUMS} if check else set(files)
    actual = {path.name for path in directory.iterdir()}
    missing, extra = expected - actual, actual - expected
    if not check:
        extra.discard(CHECKSUMS)  # A repeated preparation may replace an earlier manifest.
    if missing or extra:
        raise ValueError(f"Invalid asset set (missing: {sorted(missing)}, unexpected: {sorted(extra)})")
    for name in expected:
        path = directory / name
        if path.is_symlink() or not path.is_file() or not path.stat().st_size:
            raise ValueError(f"Invalid release asset: {name}")

    profile, windows, linux, macos, installer = (directory / name for name in files)
    for path, validate in ((profile, lambda path: verify_zip(path, version)),
                           (windows, verify_zip), (linux, verify_tar), (macos, verify_tar)):
        try:
            validate(path)
        except (OSError, ValueError, EOFError, zipfile.BadZipFile, tarfile.TarError) as exc:
            raise ValueError(f"{path.name}: {exc}") from exc
    if installer.read_bytes() != (ROOT / "scripts/install-manager.ps1").read_bytes():
        raise ValueError("Packaged installer differs from the reviewed source")
    return "".join(f"{digest(directory / name)}  {name}\n" for name in files).encode("ascii")


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("version")
    parser.add_argument("directory", type=Path)
    parser.add_argument("--check", action="store_true", help="Compare with the prepared checksum manifest")
    parser.add_argument("--published", action="store_true", help="Also verify existing release metadata from stdin")
    args = parser.parse_args()
    if args.published and not args.check:
        parser.error("--published requires --check")
    try:
        manifest = candidate_manifest(args.version, args.directory, args.check)
        destination = args.directory / CHECKSUMS
        if args.check:
            if destination.read_bytes() != manifest:
                raise ValueError("Prepared SHA256SUMS.txt does not match the assets")
            if args.published:
                verify_published(args.directory, json.load(sys.stdin))
        else:
            with NamedTemporaryFile(dir=args.directory, delete=False) as temporary:
                staging = Path(temporary.name)
            try:
                staging.write_bytes(manifest)
                staging.replace(destination)
            finally:
                staging.unlink(missing_ok=True)
    except (OSError, ValueError, subprocess.CalledProcessError) as exc:
        parser.exit(1, f"Release candidate rejected: {exc}\n")
    print(f"Release candidate verified for v{args.version}")


if __name__ == "__main__":
    main()
