"""Package the versioned Codex profile or a native Chuli binary using the stdlib."""

import argparse
import subprocess
import tarfile
import zipfile
from pathlib import Path

ROOT = Path(__file__).resolve().parents[1]
PROFILE_DIRS = ("agents/", "config/", "docs/protocols/", "skills/", "licenses/", "LICENSES/", "third_party/")
PROFILE_FILES = {
    "VERSION", "LICENSE", "THIRD_PARTY_NOTICES.md", "templates/feature-tasks.md",
    "docs/protocols/TTP.md", "scripts/install.ps1", "scripts/verify.ps1",
    "scripts/rollback.ps1", "scripts/profile.ps1", "scripts/profile_resources.py", "scripts/toml_config.py",
    "scripts/taskboard.py", "scripts/legacy-managed-hashes.json",
}


def package_profile(output, version):
    tracked = subprocess.check_output(["git", "ls-files", "-z"], cwd=ROOT).decode().split("\0")
    files = sorted(p for p in tracked if p in PROFILE_FILES or p.startswith(PROFILE_DIRS))
    missing = PROFILE_FILES - set(files)
    if missing:
        raise ValueError(f"Missing release resources: {sorted(missing)}")
    with zipfile.ZipFile(output / f"chuli-profile-v{version}.zip", "w", zipfile.ZIP_DEFLATED) as package:
        for name in files:
            source = ROOT / name
            if source.is_symlink() or not source.is_file():
                raise ValueError(f"Invalid release resource: {name}")
            info = zipfile.ZipInfo(name, date_time=(2020, 1, 1, 0, 0, 0))
            info.create_system = 3
            info.external_attr = 0o100644 << 16
            package.writestr(info, source.read_bytes(), compress_type=zipfile.ZIP_DEFLATED)


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("kind", choices=("profile", "manager"))
    parser.add_argument("output", type=Path)
    parser.add_argument("--binary", type=Path)
    parser.add_argument("--goos", choices=("windows", "linux", "darwin"))
    parser.add_argument("--goarch", choices=("amd64", "arm64"))
    args = parser.parse_args()
    version = (ROOT / "VERSION").read_text().strip()
    args.output.mkdir(parents=True, exist_ok=True)
    if args.kind == "profile":
        package_profile(args.output, version)
        return
    if not (args.binary and args.goos and args.goarch):
        parser.error("manager requires --binary, --goos and --goarch")
    base = args.output / f"chuli_v{version}_{args.goos}_{args.goarch}"
    if args.goos == "windows":
        with zipfile.ZipFile(str(base) + ".zip", "w", zipfile.ZIP_DEFLATED) as z:
            z.write(args.binary, "chuli.exe")
    else:
        with tarfile.open(str(base) + ".tar.gz", "w:gz") as t:
            t.add(args.binary, arcname="chuli", recursive=False)


if __name__ == "__main__":
    main()
