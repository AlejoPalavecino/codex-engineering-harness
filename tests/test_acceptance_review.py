import io
from contextlib import redirect_stdout
import unittest

from test_taskboard import module


class AcceptanceReviewTests(unittest.TestCase):
    def review(self, *criteria):
        output = io.StringIO()
        with redirect_stdout(output):
            status = module.main(["review-criteria", *criteria])
        return status, output.getvalue()

    def test_reports_decision_next_to_each_ambiguous_fragment(self):
        status, output = self.review("Con 200 pedidos, la búsqueda debe responder rápidamente.",
                                     "El formulario debe ser fácil de usar y manejar errores correctamente.")
        self.assertEqual(status, 1)
        for expected in ("Criterio 1", "rápidamente", "tiempo máximo", "Criterio 2",
                         "fácil de usar", "qué tarea", "correctamente", "resultado observable"):
            self.assertIn(expected, output)

    def test_concrete_outcomes_and_literal_ui_copy(self):
        status, output = self.review("Con 200 pedidos, la búsqueda responde en menos de 500 ms.",
                                     "Al guardar, muestra el texto «rápidamente» y conserva el pedido.",
                                     "Sin permiso de edición, devuelve 403 y conserva el registro.")
        self.assertEqual(status, 0)
        self.assertNotIn("¿", output)

    def test_unresolved_rule_and_blank_criterion(self):
        status, output = self.review("El sistema pide confirmación cuando corresponda.", "  ")
        self.assertEqual(status, 1)
        self.assertIn("condición", output)
        self.assertIn("Criterio 2", output)
        self.assertIn("resultado observable", output)
