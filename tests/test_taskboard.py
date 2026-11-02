import importlib.util
from pathlib import Path
import tempfile
import unittest

ROOT = Path(__file__).resolve().parents[1]
path = ROOT / "scripts" / "taskboard.py"
spec = importlib.util.spec_from_file_location("taskboard", path)
module = importlib.util.module_from_spec(spec)
import sys
sys.modules["taskboard"] = module
spec.loader.exec_module(module)

BOARD = """# Feature: reservas

- **Spec:** docs/specs/reservas.md
- **Design:** docs/design/reservas.md
- **Approval:** approved

## Tasks

### TASK-001 — Validar disponibilidad
- **State:** READY
- **Depends on:** none
- **Owner:** test_engineer
- **Acceptance:** Reservas simultáneas no se duplican.
- **Evidence:** none
- **Blocker:** none

### TASK-002 — Implementar reserva
- **State:** PLANNED
- **Depends on:** TASK-001
- **Owner:** worker
- **Acceptance:** Un turno disponible se reserva correctamente.
- **Evidence:** none
- **Blocker:** none
"""


class TaskboardTests(unittest.TestCase):
    def test_valid(self):
        self.assertEqual(module.errors_for(BOARD), [])

    def test_transition_and_dependencies(self):
        with self.assertRaisesRegex(ValueError, "dependencia"):
            module.transition(BOARD, "TASK-002", "READY", "PLANNED", None, None, None)
        first = module.transition(BOARD, "TASK-001", "IN_PROGRESS", "READY", None, None, None)
        first = module.transition(first, "TASK-001", "VERIFYING", "IN_PROGRESS", None, None, None)
        with self.assertRaisesRegex(ValueError, "evidence"):
            module.transition(first, "TASK-001", "DONE", "VERIFYING", None, None, None)
        first = module.transition(first, "TASK-001", "DONE", "VERIFYING", "pytest 4 passed", None, None)
        second = module.transition(first, "TASK-002", "READY", "PLANNED", None, None, None)
        self.assertEqual(module.errors_for(second), [])
        self.assertIn("- **Evidence:** pytest 4 passed", second)

    def test_approval_and_invalid_status(self):
        source = BOARD.replace("Approval:** approved", "Approval:** pending")
        self.assertTrue(any("Approval" in x for x in module.errors_for(source)))
        source = BOARD.replace("State:** READY", "State:** DONE")
        self.assertTrue(any("evidencia" in x for x in module.errors_for(source)))

    def test_cycles(self):
        source = BOARD.replace("TASK-001 — Validar disponibilidad", "TASK-001 — Validar disponibilidad")
        source = source.replace("- **Depends on:** none", "- **Depends on:** TASK-002", 1)
        self.assertTrue(any("Ciclo" in x for x in module.errors_for(source)))

    def test_blocker_and_reopen(self):
        blocked = module.transition(BOARD, "TASK-001", "BLOCKED", "READY", None, "Falta respuesta API", None)
        self.assertIn("Falta respuesta API", blocked)
        ready = module.transition(blocked, "TASK-001", "READY", "BLOCKED", None, None, None)
        self.assertIn("**Blocker:** none", ready)
        work = module.transition(ready, "TASK-001", "IN_PROGRESS", "READY", None, None, None)
        work = module.transition(work, "TASK-001", "VERIFYING", "IN_PROGRESS", None, None, None)
        done = module.transition(work, "TASK-001", "DONE", "VERIFYING", "test passed", None, None)
        with self.assertRaisesRegex(ValueError, "reason"):
            module.transition(done, "TASK-001", "READY", "DONE", None, None, None)
        reopened = module.transition(done, "TASK-001", "READY", "DONE", None, None, "Cambio en criterios")
        self.assertIn("**Reopen reason:** Cambio en criterios", reopened)
        self.assertIn("**Evidence:** none", reopened)

    def test_atomic_write_and_cli(self):
        with tempfile.TemporaryDirectory() as folder:
            board = Path(folder) / "tasks.md"
            board.write_text(BOARD, encoding="utf-8")
            self.assertEqual(module.main(["check", str(board)]), 0)
            self.assertEqual(module.main(["set", str(board), "TASK-001", "IN_PROGRESS", "--expect", "READY"]), 0)
            self.assertEqual(module.main(["show", str(board)]), 0)
            self.assertIn("**State:** IN_PROGRESS", board.read_text(encoding="utf-8"))

    def test_template_parses(self):
        template = (ROOT / "templates" / "feature-tasks.md").read_text(encoding="utf-8")
        self.assertEqual(module.errors_for(template), [])


if __name__ == "__main__":
    unittest.main()
