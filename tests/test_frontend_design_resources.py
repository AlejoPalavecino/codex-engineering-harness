"""Pinned upstream resources and multi-client packaging checks for frontend-design."""
from pathlib import Path
import hashlib
import importlib.util
import json
import re
import unittest

ROOT = Path(__file__).resolve().parents[1]
SKILL = ROOT / "skills/frontend-design"
spec = importlib.util.spec_from_file_location("frontend_profile", ROOT / "scripts/profile_resources.py")
profile = importlib.util.module_from_spec(spec)
spec.loader.exec_module(profile)


class FrontendDesignResources(unittest.TestCase):
    def test_pinned_bytes_and_multiclient_inventory(self):
        manifest = json.loads((SKILL / "UPSTREAM.json").read_text(encoding="utf-8"))
        self.assertEqual(manifest["repository"], "https://github.com/anthropics/skills")
        self.assertEqual(manifest["commit"], "e4bb5d960d87126ff3b3a8af1a104b7db67560ba")
        self.assertEqual(manifest["license"], "Apache-2.0")
        self.assertEqual(set(manifest["files"]), {"SKILL.md", "LICENSE.txt"})
        self.assertEqual(
            hashlib.sha256((SKILL / "LICENSE.txt").read_bytes()).hexdigest(),
            manifest["license_sha256"],
        )

        for client, prefix in (
            ("codex", "skills/frontend-design/"),
            ("claude", "skills/frontend-design/"),
            ("agy", "plugins/chuli/skills/frontend-design/"),
        ):
            resources = profile.inventory(ROOT, client)
            for relative, digest in manifest["files"].items():
                with self.subTest(client=client, resource=relative):
                    self.assertEqual(
                        hashlib.sha256((SKILL / relative).read_bytes()).hexdigest(),
                        digest,
                    )
                    self.assertIn(prefix + relative, resources)

    def test_harness_ux_ui_links_frontend_design_portably(self):
        ux_skill = (ROOT / "skills/harness-ux-ui/SKILL.md").read_text(encoding="utf-8")
        self.assertIn("../frontend-design/SKILL.md", ux_skill)
        links = re.findall(r"\[[^\]]*\]\(([^)]+)\)", ux_skill)
        self.assertIn("../frontend-design/SKILL.md", links)
