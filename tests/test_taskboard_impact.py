"""Exercise file impact through the installed CLI's entry point."""
from contextlib import redirect_stdout, redirect_stderr
import io
from pathlib import Path
import tempfile
import unittest

from test_taskboard import BOARD, ROOT, module


class ImpactTests(unittest.TestCase):
    def query(self, source, filename):
        with tempfile.TemporaryDirectory() as directory:
            board = Path(directory) / "board.md"
            board.write_text(source, encoding="utf-8")
            before = board.read_bytes()
            output = io.StringIO()
            with redirect_stdout(output), redirect_stderr(output):
                code = module.main(["impact", str(board), filename])
            self.assertEqual(board.read_bytes(), before)
            return code, output.getvalue()

    def test_shared_file_reports_only_linked_tasks_requirements_and_tests(self):
        source = (ROOT / "templates/feature-tasks.md").read_text(encoding="utf-8")
        source = source.replace('"src/reservas.py"', '"other/reservas.py"', 1)
        code, result = self.query(source, r"src\reservas.py")
        self.assertEqual(code, 0)
        for expected in ("TASK-002", "REQ-002", "test_cancelar_reserva_propia",
                         "TASK-003", "REQ-003", "test_reservas_concurrentes"):
            self.assertIn(expected, result)
        for unrelated in ("TASK-001", "REQ-001", "test_reservar_turno"):
            self.assertNotIn(unrelated, result)

    def test_missing_scope_and_legacy_evidence(self):
        source = BOARD.replace("- **Evidence:** none",
                               '- **Scope:** ["app.py"]\n- **Evidence:** prueba anterior', 1)
        code, result = self.query(source, "app.py")
        self.assertEqual(code, 0)
        for expected in ("TASK-001", "prueba anterior", "sin Scope", "1"):
            self.assertIn(expected, result)
        self.assertNotIn("REQ-", result)
        code, result = self.query(source, "otro.py")
        self.assertEqual(code, 0)
        self.assertIn("Sin coincidencias", result)
        self.assertIn("sin Scope", result)

    def test_rejects_non_project_paths(self):
        for name in ("../app.py", "/app.py", r"C:\app.py"):
            with self.subTest(name=name):
                self.assertEqual(self.query(BOARD, name)[0], 2)
