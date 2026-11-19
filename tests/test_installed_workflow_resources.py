"""Installed workflow resources must work independently of the other clients."""
import sys
import unittest
from pathlib import Path

ROOT = Path(__file__).resolve().parents[1]
sys.path.insert(0, str(ROOT / "scripts"))
from profile_resources import inventory


class InstalledWorkflowResources(unittest.TestCase):
    def test_resources_identify_each_client_home(self):
        for client, prefix in (("codex", ""), ("claude", ""), ("agy", "plugins/chuli/")):
            resources = inventory(ROOT, client)
            for target in ("harness/docs/protocols/TTP.md", prefix + "skills/harness-specification/SKILL.md"):
                content = (ROOT / resources[target]).read_text(encoding="utf-8")
                for variable, default in (("CODEX_HOME", "~/.codex"), ("CLAUDE_CONFIG_DIR", "~/.claude"),
                                          ("AGY_HOME", "~/.gemini/config")):
                    with self.subTest(client=client, target=target, variable=variable):
                        self.assertIn(variable, content)
                        self.assertIn(default, content)

    def test_agy_loads_complete_diagram_entry_and_keeps_upstream_guide(self):
        resources = inventory(ROOT, "agy")
        base = "plugins/chuli/skills/diagram-design/"
        entry = (ROOT / resources[base + "SKILL.md"]).read_bytes()
        self.assertLessEqual(len(entry), 24000)
        self.assertIn(b"(GUIDE.md)", entry)
        self.assertEqual(resources[base + "GUIDE.md"], "skills/diagram-design/SKILL.md")
        self.assertFalse(any(k.startswith(base) for k in inventory(ROOT, "agy", {"diagram-design"})))
