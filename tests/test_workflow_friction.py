import copy
import contextlib
import importlib.util
import io
import json
from pathlib import Path
import subprocess
import sys
import tempfile
import unittest

ROOT = Path(__file__).resolve().parents[1]
FIXTURES = ROOT / "tests/fixtures/workflow-friction"
spec = importlib.util.spec_from_file_location("workflow_friction", ROOT / "scripts/workflow_friction.py")
module = importlib.util.module_from_spec(spec)
spec.loader.exec_module(module)


class WorkflowFrictionTests(unittest.TestCase):
    def setUp(self):
        self.before = json.loads((FIXTURES / "baseline.json").read_text(encoding="utf-8"))
        self.after = json.loads((FIXTURES / "candidate.json").read_text(encoding="utf-8"))

    def test_pilot_keeps_scenarios_separate(self):
        result = module.compare(self.before, self.after)
        self.assertEqual(result["status"], "COMPARABLE")
        self.assertNotIn("total", result)
        self.assertTrue(result["synthetic"])
        self.assertEqual(len(result["scenarios"]), 3)
        first = result["scenarios"][0]
        self.assertEqual(first["metrics"]["repeated_questions"]["delta"], -1)
        self.assertEqual(first["quality"], {"baseline": "PASS", "candidate": "PASS"})

    def test_unknown_is_not_zero(self):
        self.after["observations"][0]["metrics"]["repeated_questions"] = None
        row = module.compare(self.before, self.after)["scenarios"][0]
        self.assertIsNone(row["metrics"]["repeated_questions"]["delta"])
        self.assertEqual(row["metrics"]["repeated_questions"]["baseline"], 1)
        self.assertEqual(row["metrics"]["repeated_questions"]["status"], "UNKNOWN")

    def test_incompatible_never_computes_delta(self):
        for key in self.before["observations"][0]["conditions"]:
            with self.subTest(key=key):
                changed = copy.deepcopy(self.after)
                changed["observations"][0]["conditions"][key] = "different"
                result = module.compare(self.before, changed)
                self.assertEqual(result["status"], "NOT_COMPARABLE")
                self.assertNotIn("metrics", result["scenarios"][0])

    def test_invalid_data(self):
        for value in [-1, True, "0", float("nan"), float("inf"), 0.5]:
            with self.subTest(value=value):
                changed = copy.deepcopy(self.after)
                changed["observations"][0]["metrics"]["repeated_questions"] = value
                with self.assertRaises(ValueError):
                    module.compare(self.before, changed)
        self.after["observations"].pop()
        with self.assertRaises(ValueError):
            module.compare(self.before, self.after)

    def test_evidence_required_for_observed_values(self):
        self.after["observations"][0]["evidence"] = []
        with self.assertRaises(ValueError):
            module.compare(self.before, self.after)

    def test_duplicate_scenario_rejected(self):
        self.after["observations"][2] = self.after["observations"][0]
        with self.assertRaises(ValueError):
            module.compare(self.before, self.after)

    def test_quality_regression_is_visible(self):
        self.after["observations"][0]["quality"] = "FAIL"
        row = module.compare(self.before, self.after)["scenarios"][0]
        self.assertTrue(row["quality_regression"])

    def test_unknown_conditions_prevent_comparison(self):
        self.before["observations"][0]["conditions"]["model"] = None
        self.after["observations"][0]["conditions"]["model"] = None
        row = module.compare(self.before, self.after)["scenarios"][0]
        self.assertEqual(row["status"], "NOT_COMPARABLE")
        self.assertNotIn("metrics", row)

    def test_synthetic_and_real_cannot_be_compared(self):
        self.after["synthetic"] = False
        result = module.compare(self.before, self.after)
        self.assertTrue(all(row["status"] == "NOT_COMPARABLE" for row in result["scenarios"]))

    def test_quality_coverage_loss_is_visible(self):
        self.after["observations"][0]["quality"] = "NOT_TESTED"
        row = module.compare(self.before, self.after)["scenarios"][0]
        self.assertTrue(row["quality_coverage_lost"])
        self.assertFalse(row["quality_regression"])

    def test_cli_pilot(self):
        command = [sys.executable, str(ROOT / "scripts/workflow_friction.py"),
                   str(FIXTURES / "baseline.json"), str(FIXTURES / "candidate.json")]
        run = subprocess.run(command, capture_output=True, text=True, check=False)
        self.assertEqual(run.returncode, 0, run.stderr)
        self.assertEqual(json.loads(run.stdout)["status"], "COMPARABLE")

    def test_cli_rejects_invalid_and_duplicate_json_keys(self):
        for payload in ["{", '{"schema":1,"schema":1}', "[]"]:
            with self.subTest(payload=payload), tempfile.TemporaryDirectory() as folder:
                path = Path(folder) / "invalid.json"
                path.write_text(payload, encoding="utf-8")
                output = io.StringIO()
                with contextlib.redirect_stdout(output):
                    code = module.main([str(path), str(FIXTURES / "candidate.json")])
                self.assertEqual(code, 2)
                self.assertEqual(json.loads(output.getvalue())["status"], "INVALID_INPUT")

    def test_cli_incompatible_exit_code(self):
        self.after["observations"][0]["conditions"]["fixture"] = "other"
        with tempfile.TemporaryDirectory() as folder:
            path = Path(folder) / "different.json"
            path.write_text(json.dumps(self.after), encoding="utf-8")
            output = io.StringIO()
            with contextlib.redirect_stdout(output):
                code = module.main([str(FIXTURES / "baseline.json"), str(path)])
            self.assertEqual(code, 1)
            self.assertEqual(json.loads(output.getvalue())["status"], "NOT_COMPARABLE")


if __name__ == "__main__":
    unittest.main()
