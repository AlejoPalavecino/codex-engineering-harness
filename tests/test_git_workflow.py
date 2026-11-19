"""Contract of the Git workflow delivered to each client."""
from pathlib import Path
import sys
import unittest

ROOT = Path(__file__).resolve().parents[1]
sys.path.insert(0, str(ROOT / "scripts"))
import client_targets
import profile_resources


class GitWorkflowTests(unittest.TestCase):
    def test_each_client_can_resolve_the_shared_scenarios(self):
        destination = "harness/docs/protocols/GIT-GITHUB.md"
        scenarios = ("Consulta R0", "Arreglo R1", "Funcionalidad R2",
                     "Presupuesto excedido", "Sin acceso remoto", "Retomar")
        for client in ("codex", "claude", "agy"):
            with self.subTest(client=client):
                instructions = client_targets.render_orchestrator_body(client).decode()
                self.assertIn(destination, instructions)
                resources = profile_resources.inventory(ROOT, client)
                skill_root = "plugins/chuli/skills" if client == "agy" else "skills"
                for skill in ("issue-creation", "work-unit-commits", "branch-pr", "chained-pr"):
                    self.assertIn(f"{skill_root}/{skill}/SKILL.md", resources)
                source = resources[destination]
                protocol = (ROOT / source).read_text(encoding="utf-8")
                for scenario in scenarios:
                    self.assertIn(scenario, protocol)

    def test_chain_guidance_uses_authored_lines_and_immediate_bases(self):
        for client in ("codex", "claude", "agy"):
            with self.subTest(client=client):
                resources = profile_resources.inventory(ROOT, client)
                skill_root = "plugins/chuli/skills" if client == "agy" else "skills"
                skill_path = f"{skill_root}/chained-pr"
                skill = (ROOT / resources[f"{skill_path}/SKILL.md"]).read_text(encoding="utf-8")
                details_path = f"{skill_path}/references/chaining-details.md"
                details = (ROOT / resources[details_path]).read_text(encoding="utf-8")

                self.assertIn("about 400 authored changed lines", skill)
                self.assertIn("generated, vendored, lockfile, and binary changes separately", skill)
                self.assertIn("target its immediate parent", details)
                self.assertIn("only when the destination policy requires it", details)
                self.assertNotIn("exceeds 400 changed lines without", details)
