"""Fitness checks for the synthetic inventory decision ADR-001."""
from contextlib import closing
from pathlib import Path
import sqlite3
import tempfile
import unittest

from importer import import_inventory


class InventoryDecisionTests(unittest.TestCase):
    def test_reimport_preserves_one_current_description_after_reopening(self):
        with tempfile.TemporaryDirectory() as folder:
            csv_path = Path(folder) / "input.csv"
            database = Path(folder) / "inventory.sqlite"
            for label in ("primera", "revisada"):
                csv_path.write_text(f"sku,description\nA,{label}\n", encoding="utf-8")
                self.assertEqual(import_inventory(csv_path, database), 1)
            with closing(sqlite3.connect(database)) as connection:
                self.assertEqual(connection.execute("select * from inventory").fetchall(), [("A", "revisada")])

    def test_invalid_row_rolls_back_the_whole_import(self):
        with tempfile.TemporaryDirectory() as folder:
            csv_path = Path(folder) / "input.csv"
            database = Path(folder) / "inventory.sqlite"
            csv_path.write_text("sku,description\nA,original\n", encoding="utf-8")
            import_inventory(csv_path, database)
            csv_path.write_text("sku,description\nA,reemplazo\nB\n", encoding="utf-8")
            with self.assertRaises(sqlite3.IntegrityError):
                import_inventory(csv_path, database)
            with closing(sqlite3.connect(database)) as connection:
                self.assertEqual(connection.execute("select * from inventory").fetchall(), [("A", "original")])
