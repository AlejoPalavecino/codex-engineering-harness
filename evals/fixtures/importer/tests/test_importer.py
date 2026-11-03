import os
from pathlib import Path
import tempfile
import unittest

from importer import import_inventory


class ImporterTests(unittest.TestCase):
    def test_small_csv(self):
        source = Path(__file__).resolve().parents[1] / "sample.csv"
        with tempfile.TemporaryDirectory() as folder:
            count = import_inventory(str(source), os.path.join(folder, "example.sqlite"))
            self.assertEqual(count, 3)


if __name__ == "__main__":
    unittest.main()
