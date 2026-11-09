"""The standalone Windows installer is shipped from reviewed source bytes."""
from pathlib import Path
import subprocess
import sys
import tempfile
import unittest

ROOT = Path(__file__).resolve().parents[1]


class ManagerInstallerPackage(unittest.TestCase):
    def test_installer_asset_is_exact_source_and_only_expected_file(self):
        with tempfile.TemporaryDirectory(prefix="chuli-installer-package-") as folder:
            result = subprocess.run(
                [sys.executable, "-X", "utf8", "-B", str(ROOT / "scripts/package_release.py"), "installer", folder],
                capture_output=True, text=True,
            )
            self.assertEqual(result.returncode, 0, result.stdout + result.stderr)
            target = Path(folder)
            self.assertEqual([p.name for p in target.iterdir()], ["install-chuli-windows.ps1"])
            self.assertEqual((target / "install-chuli-windows.ps1").read_bytes(), (ROOT / "scripts/install-manager.ps1").read_bytes())
