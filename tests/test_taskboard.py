import importlib.util
import json
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

    def test_candidate_evidence_and_drift(self):
        with tempfile.TemporaryDirectory() as directory:
            root = Path(directory)
            board = root / "tasks.md"
            artifact = root / "app.py"
            artifact.write_bytes(b"print('ok')\r\n")
            source = BOARD.replace("- **Approval:** approved", "- **Approval:** approved\n- **Evidence mode:** candidate-v1")
            source = source.replace("- **Evidence:** none", '- **Scope:** ["app.py"]\n- **Evidence:** none')
            source = source.replace("State:** READY", "State:** VERIFYING")
            board.write_text(source, encoding="utf-8")
            candidate = module.fingerprint(root, board, ["app.py"])
            artifact.write_bytes(b"print('ok')\n")
            self.assertEqual(candidate, module.fingerprint(root, board, ["app.py"]))
            with self.assertRaisesRegex(ValueError, "JSON"):
                module.transition(source, "TASK-001", "DONE", "VERIFYING", "tests passed", None, None)
            evidence = json.dumps({"command": "python -m unittest", "result": "7 tests OK", "candidate": candidate})
            done = module.transition(source, "TASK-001", "DONE", "VERIFYING", evidence, None, None)
            board.write_text(done, encoding="utf-8")
            self.assertEqual(module.main(["check", str(board), "--current", "--root", str(root)]), 0)
            artifact.write_text("changed", encoding="utf-8")
            self.assertEqual(module.main(["check", str(board), "--root", str(root)]), 0)
            self.assertEqual(module.main(["check", str(board), "--current", "--root", str(root)]), 2)

    def test_candidate_scope_rejects_escape_and_board(self):
        with tempfile.TemporaryDirectory() as directory:
            root = Path(directory)
            board = root / "tasks.md"
            board.write_text(BOARD, encoding="utf-8")
            for scope in (["../outside"], [str(board.resolve())], ["tasks.md"], [], ["missing"], ["a\\b"]):
                with self.subTest(scope=scope), self.assertRaises(ValueError):
                    module.fingerprint(root, board, scope)

    def test_candidate_done_requires_command_result_and_hash(self):
        source = BOARD.replace("Approval:** approved", "Approval:** approved\n- **Evidence mode:** candidate-v1")
        source = source.replace("State:** READY", "State:** DONE")
        source = source.replace("- **Evidence:** none", '- **Scope:** ["app.py"]\n- **Evidence:** RECORD')
        for record in ({}, {"command": "test", "candidate": "sha256:" + "a" * 64},
                       {"command": "test", "result": "OK", "candidate": "HEAD"}):
            with self.subTest(record=record):
                self.assertTrue(module.errors_for(source.replace("RECORD", json.dumps(record))))

    def test_cli_done_checks_candidate_before_write_and_never_executes_evidence(self):
        with tempfile.TemporaryDirectory() as directory:
            root = Path(directory)
            board = root / "tasks.md"
            artifact = root / "app.py"
            artifact.write_text("original", encoding="utf-8")
            source = BOARD.replace("Approval:** approved", "Approval:** approved\n- **Evidence mode:** candidate-v1")
            source = source.replace("State:** READY", "State:** VERIFYING")
            source = source.replace("- **Evidence:** none", '- **Scope:** ["app.py"]\n- **Evidence:** none')
            board.write_text(source, encoding="utf-8")
            record = json.dumps({"command": "touch MUST_NOT_EXIST", "result": "observed OK", "candidate": module.fingerprint(root, board, ["app.py"])})
            artifact.write_text("changed", encoding="utf-8")
            args = ["set", str(board), "TASK-001", "DONE", "--expect", "VERIFYING", "--evidence", record, "--root", str(root)]
            self.assertEqual(module.main(args), 2)
            self.assertEqual(board.read_text(encoding="utf-8"), source)
            artifact.write_text("original", encoding="utf-8")
            self.assertEqual(module.main(args), 0)
            self.assertFalse((root / "MUST_NOT_EXIST").exists())

    def test_scope_symlink_escape(self):
        with tempfile.TemporaryDirectory() as directory:
            base = Path(directory)
            root = base / "project"
            root.mkdir()
            board = root / "tasks.md"
            board.write_text(BOARD, encoding="utf-8")
            outside = base / "outside.py"
            outside.write_text("outside", encoding="utf-8")
            try:
                (root / "link.py").symlink_to(outside)
            except OSError:
                self.skipTest("El sistema no permite crear symlinks en esta sesión")
            with self.assertRaisesRegex(ValueError, "fuera"):
                module.fingerprint(root, board, ["link.py"])

    def test_done_rejects_stale_transitive_dependency_but_ignores_unrelated_task(self):
        with tempfile.TemporaryDirectory() as directory:
            root = Path(directory)
            board = root / "tasks.md"
            board.write_text("placeholder", encoding="utf-8")
            source = "# Feature\n- **Spec:** spec\n- **Design:** design\n- **Approval:** approved\n- **Evidence mode:** candidate-v1\n"
            records = {}
            for number, state, dependency in ((1, "DONE", "none"), (2, "DONE", "TASK-001"),
                                               (3, "VERIFYING", "TASK-002"), (4, "DONE", "none")):
                name = f"file-{number}.py"
                (root / name).write_text("original", encoding="utf-8")
                records[number] = json.dumps({"command": "test", "result": "OK", "candidate": module.fingerprint(root, board, [name])})
                evidence = "none" if state == "VERIFYING" else records[number]
                source += f"\n### TASK-{number:03d} — Resultado {number}\n- **State:** {state}\n- **Depends on:** {dependency}\n- **Owner:** worker\n- **Acceptance:** Resultado observable\n- **Scope:** [\"{name}\"]\n- **Evidence:** {evidence}\n- **Blocker:** none\n"
            board.write_text(source, encoding="utf-8")
            original = board.read_bytes()
            (root / "file-1.py").write_text("changed dependency", encoding="utf-8")
            (root / "file-4.py").write_text("changed unrelated", encoding="utf-8")
            args = ["set", str(board), "TASK-003", "DONE", "--expect", "VERIFYING", "--evidence", records[3], "--root", str(root)]
            self.assertEqual(module.main(args), 2)
            self.assertEqual(board.read_bytes(), original)
            (root / "file-1.py").write_text("original", encoding="utf-8")
            self.assertEqual(module.main(args), 0)


if __name__ == "__main__":
    unittest.main()
