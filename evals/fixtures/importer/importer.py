"""Importador didáctico sin dependencias externas."""

import csv
import sqlite3


def import_inventory(csv_path: str, db_path: str) -> int:
    with open(csv_path, encoding="utf-8", newline="") as stream:
        rows = list(csv.DictReader(stream))
    with sqlite3.connect(db_path) as conn:
        conn.execute(
            "CREATE TABLE IF NOT EXISTS inventory (sku TEXT PRIMARY KEY, description TEXT NOT NULL)"
        )
        conn.executemany(
            "INSERT OR REPLACE INTO inventory (sku, description) VALUES (?, ?)",
            [(row["sku"], row["description"]) for row in rows],
        )
    return len(rows)
