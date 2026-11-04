from pathlib import Path
import sqlite3
import tempfile
import unittest
from unittest.mock import patch

from importer import import_inventory


class ConnectionTests(unittest.TestCase):
    def test_connection_is_closed_after_success_or_error(self):
        source = Path(__file__).resolve().parents[1] / "sample.csv"
        for failing in (False, True):
            with self.subTest(failing=failing), tempfile.TemporaryDirectory() as folder:
                conn = sqlite3.connect(Path(folder) / "example.sqlite")
                if failing:
                    conn.execute("CREATE TABLE inventory (unexpected TEXT)")
                with patch("importer.sqlite3.connect", return_value=conn):
                    if failing:
                        with self.assertRaises(sqlite3.OperationalError):
                            import_inventory(str(source), "unused")
                    else:
                        self.assertEqual(import_inventory(str(source), "unused"), 3)
                try:
                    with self.assertRaises(sqlite3.ProgrammingError):
                        conn.execute("SELECT 1")
                finally:
                    conn.close()
