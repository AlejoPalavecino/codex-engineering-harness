"""Portable contracts for the canonical TTP resources (Python 3.11+, stdlib)."""
from pathlib import Path
import re
import subprocess
import sys
import tempfile
import unittest


ROOT = Path(__file__).resolve().parents[1]


class TtpResourceContracts(unittest.TestCase):
    def test_installed_instructions_resolve_all_resources_and_require_check(self):
        for relative in ("config/AGENTS.md", "skills/harness-specification/SKILL.md"):
            with self.subTest(document=relative):
                text = (ROOT / relative).read_text(encoding="utf-8")
                for resource in (
                    "harness/templates/feature-tasks.md",
                    "harness/docs/protocols/TTP.md",
                    "harness/bin/taskboard.py",
                ):
                    self.assertTrue(resource in text, f"{relative}: falta ruta instalada {resource}")
                self.assertRegex(
                    text, r"taskboard\.py[`\"']?\s+check\s+[^\n]+",
                    f"{relative}: falta comando ejecutable check antes de DONE",
                )
                self.assertIn("DONE", text)
                self.assertRegex(text, r"R0\s*(?:/|o|y)\s*(?:cambios triviales\s+)?R1")

    def test_protocol_template_link_resolves_to_canonical_source(self):
        protocol = ROOT / "docs/protocols/TTP.md"
        links = re.findall(r"\[[^\]]*\]\(([^)]+)\)", protocol.read_text(encoding="utf-8"))
        targets = [(protocol.parent / link).resolve() for link in links if not "://" in link]
        self.assertIn((ROOT / "templates/feature-tasks.md").resolve(), targets)

    def test_canonical_template_is_valid_before_and_after_explicit_approval(self):
        source = (ROOT / "templates/feature-tasks.md").read_text(encoding="utf-8")
        self.assertIn("- **Approval:** pending", source, "La plantilla no autoriza ejecución")
        self.assertNotIn("- **State:** DONE", source)
        with tempfile.TemporaryDirectory(prefix="codex-ttp-contract-") as directory:
            board = Path(directory) / "feature.md"
            for approved in (False, True):
                with self.subTest(approved=approved):
                    board.write_text(
                        source.replace("- **Approval:** pending", "- **Approval:** approved")
                        if approved else source,
                        encoding="utf-8",
                    )
                    result = subprocess.run(
                        [sys.executable, "-X", "utf8", str(ROOT / "scripts/taskboard.py"), "check", str(board)],
                        capture_output=True, text=True, encoding="utf-8",
                    )
                    self.assertEqual(result.returncode, 0, result.stdout + result.stderr)


if __name__ == "__main__":
    unittest.main()
