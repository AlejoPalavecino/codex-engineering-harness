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
