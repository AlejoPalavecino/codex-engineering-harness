import io
from contextlib import redirect_stdout, redirect_stderr
from pathlib import Path
import tempfile
import unittest

from test_taskboard import BOARD, module


def traced_board():
    source = BOARD.replace("## Tasks", """- **Trace mode:** requirements-v1
- **REQ-001:** Reservar un turno
- **REQ-002:** Evitar reservas dobles

## Tasks""")
    return source.replace("- **Acceptance:**", """- **Requirements:** REQ-001
- **Verification:** tests/reservas.py::test_reservar
- **Acceptance:**""").replace("Requirements:** REQ-001", "Requirements:** REQ-001, REQ-002", 1)


class TraceTests(unittest.TestCase):
    def test_query_filters_and_preserves_board(self):
        with tempfile.TemporaryDirectory() as folder:
            path = Path(folder) / "board.md"
            path.write_text(traced_board(), encoding="utf8")
            original = path.read_bytes()
            output = io.StringIO()
            with redirect_stdout(output):
                self.assertEqual(module.main(["trace", str(path), "REQ-002"]), 0)
            for value in ("Evitar reservas dobles", "TASK-001", "READY", "Reservas simultáneas",
                          "tests/reservas.py::test_reservar", "Evidence: none"):
                self.assertIn(value, output.getvalue())
            self.assertNotIn("TASK-002", output.getvalue())
            with redirect_stderr(io.StringIO()):
                self.assertEqual(module.main(["trace", str(path), "REQ-999"]), 2)
            self.assertEqual(path.read_bytes(), original)

    def test_rejects_broken_links(self):
        source = traced_board()
        for old, new in [
            ("REQ-002:** Evitar", "REQ-001:** Evitar"),
            ("REQ-002:** Evitar reservas dobles", "REQ-002:** "),
            ("Requirements:** REQ-001, REQ-002", "Requirements:** REQ-001, REQ-999"),
            ("Requirements:** REQ-001, REQ-002", "Requirements:** REQ-001, REQ-001"),
            ("Requirements:** REQ-001, REQ-002", "Requirements:** REQ-001"),
            ("Verification:** tests/reservas.py::test_reservar", "Verification:** none"),
            ("Trace mode:** requirements-v1", "Trace mode:** typo"),
        ]:
            with self.subTest(new=new):
                self.assertTrue(module.errors_for(source.replace(old, new)))
        self.assertEqual(module.errors_for(source), [])

    def test_legacy_transition_and_traceable_transition(self):
        for source in (BOARD, traced_board()):
            updated = module.transition(source, "TASK-001", "IN_PROGRESS", "READY", None, None, None)
            self.assertEqual(updated, source.replace("State:** READY", "State:** IN_PROGRESS", 1))
