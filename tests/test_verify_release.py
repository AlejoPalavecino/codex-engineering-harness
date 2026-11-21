"""Synthetic release candidates exercise rejection and retry without GitHub access."""

from hashlib import sha256
from io import BytesIO
import json
from pathlib import Path
import subprocess
import sys
import tarfile
import tempfile
import unittest
import zipfile

ROOT = Path(__file__).resolve().parents[1]
sys.path.insert(0, str(ROOT / "scripts"))
from package_release import profile_files

PROFILE_FILES = profile_files()
VERSION = (ROOT / "VERSION").read_text().strip()
OTHER_VERSION = "0.0.0"
NAMES = (
    f"chuli-profile-v{VERSION}.zip", f"chuli_v{VERSION}_windows_amd64.zip",
    f"chuli_v{VERSION}_linux_amd64.tar.gz", f"chuli_v{VERSION}_darwin_arm64.tar.gz",
    "install-chuli-windows.ps1",
)


class VerifyRelease(unittest.TestCase):
    def setUp(self):
        folder = tempfile.TemporaryDirectory(prefix="chuli-release-")
        self.addCleanup(folder.cleanup)
        self.assets = Path(folder.name)
        self.checksums = self.assets / "SHA256SUMS.txt"
        self.profile_zip(VERSION)
        self.write_zip(NAMES[1], "chuli.exe", b"Windows Manager")
        for name in NAMES[2:4]:
            self.write_tar(name, 0o755)
        (self.assets / NAMES[4]).write_bytes((ROOT / "scripts/install-manager.ps1").read_bytes())

    def write_zip(self, filename, member, content):
        with zipfile.ZipFile(self.assets / filename, "w") as archive:
            archive.writestr(member, content)

    def write_tar(self, filename, mode):
        with tarfile.open(self.assets / filename, "w:gz") as archive:
            info = tarfile.TarInfo("chuli")
            data = b"Unix Manager"
            info.mode = mode
            info.size = len(data)
            archive.addfile(info, BytesIO(data))

    def profile_zip(self, version):
        with zipfile.ZipFile(self.assets / NAMES[0], "w") as archive:
            for name in PROFILE_FILES:
                archive.writestr(name, f"{version}\n" if name == "VERSION" else b"synthetic resource")

    def prepare(self, version=VERSION, *options, metadata=None):
        return subprocess.run(
            [sys.executable, str(ROOT / "scripts/verify_release.py"), version, str(self.assets), *options],
            capture_output=True, text=True, input=None if metadata is None else json.dumps(metadata),
        )

    def assert_ready(self):
        result = self.prepare()
        self.assertEqual(result.returncode, 0, result.stderr)
        expected = [f"{sha256((self.assets / name).read_bytes()).hexdigest()}  {name}" for name in NAMES]
        self.assertEqual(self.checksums.read_text().splitlines(), expected)

    def assert_rejected(self, reason, version=VERSION):
        result = self.prepare(version)
        self.assertNotEqual(result.returncode, 0)
        self.assertIn(reason, result.stderr)
        self.assertFalse(self.checksums.exists())

    def test_prepares_and_can_repeat(self):
        self.assert_ready()
        self.assert_ready()
        result = self.prepare(VERSION, "--check")
        self.assertEqual(result.returncode, 0, result.stderr)

    def test_check_rejects_replaced_archive_without_rewriting_manifest(self):
        self.assert_ready()
        manifest = self.checksums.read_bytes()
        self.write_zip(NAMES[1], "chuli.exe", b"Different bytes in valid archive")
        result = self.prepare(VERSION, "--check")
        self.assertNotEqual(result.returncode, 0)
        self.assertIn("does not match the assets", result.stderr)
        self.assertEqual(self.checksums.read_bytes(), manifest)

    def test_retry_checks_published_bytes_including_checksum_manifest(self):
        self.assert_ready()
        metadata = {"draft": False, "prerelease": False, "assets": [
            {"name": path.name, "digest": "sha256:" + sha256(path.read_bytes()).hexdigest()}
            for path in self.assets.iterdir()
        ]}
        options = ("--check", "--published")
        result = self.prepare(VERSION, *options, metadata=metadata)
        self.assertEqual(result.returncode, 0, result.stderr)
        published_manifest = next(asset for asset in metadata["assets"] if asset["name"] == self.checksums.name)
        for digest in (None, "sha256:" + "0" * 64):
            with self.subTest(digest=digest):
                published_manifest["digest"] = digest
                result = self.prepare(VERSION, *options, metadata=metadata)
                self.assertNotEqual(result.returncode, 0)
                self.assertIn("needs human review", result.stderr)

    def test_rejects_missing_artifact_and_recovers_after_replacement(self):
        path = self.assets / NAMES[1]
        data = path.read_bytes()
        path.unlink()
        self.assert_rejected("missing")
        path.write_bytes(data)
        self.assert_ready()

    def test_rejects_wrong_version_before_writing(self):
        self.profile_zip(OTHER_VERSION)
        self.assert_rejected("profile VERSION")

    def test_rejects_incomplete_profile(self):
        self.write_zip(NAMES[0], "VERSION", f"{VERSION}\n")
        self.assert_rejected("profile contents")

    def test_rejects_unapproved_version(self):
        self.assert_rejected("repository VERSION", OTHER_VERSION)

    def test_rejects_corrupt_archive(self):
        (self.assets / NAMES[2]).write_bytes(b"incomplete archive")
        self.assert_rejected(NAMES[2])

    def test_rejects_manager_without_unix_execute_permission(self):
        self.write_tar(NAMES[2], 0o644)
        self.assert_rejected("execute permission")

    def test_packager_sets_unix_execute_permission_from_non_executable_source(self):
        binary = self.assets / "built-manager"
        binary.write_bytes(b"synthetic executable")
        binary.chmod(0o644)
        result = subprocess.run([
            sys.executable, str(ROOT / "scripts/package_release.py"), "manager", str(self.assets),
            "--binary", str(binary), "--goos", "linux", "--goarch", "amd64",
        ], capture_output=True, text=True)
        self.assertEqual(result.returncode, 0, result.stderr)
        with tarfile.open(self.assets / NAMES[2], "r:gz") as package:
            self.assertEqual(package.getmember("chuli").mode, 0o755)

    def test_rejects_unexpected_artifact(self):
        (self.assets / "unexpected.txt").touch()
        self.assert_rejected("unexpected")

    def test_rejects_altered_installer(self):
        (self.assets / NAMES[4]).write_bytes(b"different installer")
        self.assert_rejected("installer differs")

    def test_profile_files_excludes_go_sources(self):
        from package_release import select_profile_files

        selected = select_profile_files([*PROFILE_FILES, "config/models_embed.go"])
        self.assertIn("config/models.json", selected)
        self.assertNotIn("config/models_embed.go", selected)
        self.assertFalse(any(name.endswith(".go") for name in selected))

    def test_profile_budget_constants_match_go_archive_limits(self):
        import re
        import verify_release as vr

        archive_go = (ROOT / "internal/manager/archive.go").read_text(encoding="utf-8")
        expected = {
            "maxArchive": vr.MAX_PROFILE_ARCHIVE_BYTES,
            "maxExpanded": vr.MAX_PROFILE_EXPANDED_BYTES,
            "maxEntry": vr.MAX_PROFILE_ENTRY_BYTES,
            "maxEntries": vr.MAX_PROFILE_ENTRIES,
        }
        for name, value in expected.items():
            match = re.search(rf"\b{name}\s*=\s*(\d+)(?:\s*<<\s*(\d+))?", archive_go)
            self.assertIsNotNone(match, f"Missing {name} in internal/manager/archive.go")
            go_val = int(match.group(1)) << int(match.group(2) or 0)
            self.assertEqual(value, go_val, f"{name} mismatch between verify_release.py and archive.go")

    def test_profile_budget_rejects_limit_violations(self):
        import verify_release as vr

        profile_path = self.assets / NAMES[0]
        with zipfile.ZipFile(profile_path) as archive:
            infos = archive.infolist()

        vr.verify_profile_budget(profile_path, infos)
        with self.assertRaisesRegex(ValueError, "compressed size"):
            vr.verify_profile_budget(profile_path, infos, max_archive=1)
        with self.assertRaisesRegex(ValueError, "expanded size"):
            vr.verify_profile_budget(profile_path, infos, max_expanded=1)
        with self.assertRaisesRegex(ValueError, "entry .* exceeds"):
            vr.verify_profile_budget(profile_path, infos, max_entry=1)
        with self.assertRaisesRegex(ValueError, "entry count"):
            vr.verify_profile_budget(profile_path, infos, max_entries=1)

    def test_real_packaged_profile_fits_budget(self):
        import verify_release as vr

        result = subprocess.run(
            [sys.executable, str(ROOT / "scripts/package_release.py"), "profile", str(self.assets)],
            capture_output=True, text=True,
        )
        self.assertEqual(result.returncode, 0, result.stderr)
        with zipfile.ZipFile(self.assets / NAMES[0]) as archive:
            stats = vr.verify_profile_budget(self.assets / NAMES[0], archive.infolist())
        self.assertGreater(stats["entries"], 0)
        self.assertLessEqual(stats["archive_bytes"], vr.MAX_PROFILE_ARCHIVE_BYTES)
        self.assertLessEqual(stats["expanded_bytes"], vr.MAX_PROFILE_EXPANDED_BYTES)
