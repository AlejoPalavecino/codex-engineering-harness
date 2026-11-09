"""Pinned upstream resources and portable, safe diagram exports."""
from pathlib import Path
import hashlib
import importlib.util
import json
import subprocess
import sys
import tempfile
import unittest
import xml.etree.ElementTree as ET

ROOT = Path(__file__).resolve().parents[1]
SKILL = ROOT / "skills/diagram-design"
spec = importlib.util.spec_from_file_location("diagram_profile", ROOT / "scripts/profile_resources.py")
profile = importlib.util.module_from_spec(spec)
spec.loader.exec_module(profile)


class DiagramDesignResources(unittest.TestCase):
    def test_pinned_bytes_and_inventory(self):
        manifest = json.loads((SKILL / "UPSTREAM.json").read_text(encoding="utf-8"))
        self.assertEqual(manifest["commit"], "6a7f19fcc132b39b126bea90167679b46b14bcc4")
        self.assertEqual(len(manifest["files"]), 293)
        resources = profile.inventory(ROOT)
        for relative, digest in manifest["files"].items():
            with self.subTest(resource=relative):
                self.assertEqual(hashlib.sha256((SKILL / relative).read_bytes()).hexdigest(), digest)
                self.assertIn("skills/diagram-design/" + relative, resources)
        self.assertEqual(hashlib.sha256((SKILL / "LICENSE").read_bytes()).hexdigest(), manifest["license_sha256"])
        self.assertIn("skills/diagram-design/LICENSE", resources)

    def test_portable_export_and_safety_gate(self):
        with tempfile.TemporaryDirectory(prefix="chuli-diagram-test-") as directory:
            source = Path(directory) / "example.html"
            source.write_text(
                '<!doctype html><html lang="es"><head><meta charset="utf-8"><title>Prueba</title></head>'
                '<body><svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 320 160" role="img" '
                'aria-labelledby="example-title example-desc"><title id="example-title">Prueba</title>'
                '<desc id="example-desc">Un resultado legible.</desc>'
                '<text x="24" y="48" font-size="16">Resultado</text></svg></body></html>',
                encoding="utf-8",
            )
            checker = [sys.executable, "-B", str(SKILL / "scripts/self_check.py"), str(source), "--offline"]
            valid = subprocess.run(checker, capture_output=True, text=True)
            self.assertEqual(valid.returncode, 0, valid.stdout + valid.stderr)
            target = source.with_suffix(".svg")
            result = subprocess.run(
                [sys.executable, "-B", str(SKILL / "scripts/export_svg.py"), str(source), str(target), "--system-fonts"],
                capture_output=True, text=True,
            )
            self.assertEqual(result.returncode, 0, result.stdout + result.stderr)
            svg = ET.parse(target).getroot()
            self.assertEqual(svg.attrib["role"], "img")
            self.assertNotIn("@import", target.read_text(encoding="utf-8"))
            source.write_text(source.read_text(encoding="utf-8").replace('<body>', '<body><script>alert(1)</script>'), encoding="utf-8")
            unsafe = subprocess.run(checker, capture_output=True, text=True)
            self.assertNotEqual(unsafe.returncode, 0, "Executable diagram must fail the static gate")
