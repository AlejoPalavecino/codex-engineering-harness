from contextlib import redirect_stderr, redirect_stdout
import io
import json
from pathlib import Path
import tempfile
import unittest

from test_taskboard import ROOT, module


class QualityGateTests(unittest.TestCase):
    def setUp(self):
        sandbox = tempfile.TemporaryDirectory()
        self.addCleanup(sandbox.cleanup)
        self.root = Path(sandbox.name)
        self.board = self.root / "board.md"
        self.source = (ROOT / "tests/fixtures/taskboard-quality.md").read_text(encoding="utf-8")
        self.board.write_text(self.source, encoding="utf-8")
        self.app = self.root / "app.py"
        self.app.write_text("candidate\n", encoding="utf-8")
        (self.root / "tests").mkdir()
        (self.root / "tests/test_app.py").write_text("checks\n", encoding="utf-8")

    def invoke(self, command, *args):
        output = io.StringIO()
        with redirect_stdout(output), redirect_stderr(output):
            code = module.main([command, str(self.board), *args, "--root", str(self.root)])
        return code, output.getvalue()

    def record(self, ident="TASK-001"):
        code, candidate = self.invoke("fingerprint", ident)
        self.assertEqual(code, 0, candidate)
        run = {"status": "passed", "command": "check", "result": "observed pass",
               "version": "runner 1.0", "observed_at": "2026-11-19T12:00:00Z"}
        return {"candidate": candidate.strip(), "runs": {
            case: run.copy() for case in ("concurrencia", "dueño", "aislamiento", "lint")}}

    def test_reports_four_states_and_refuses_incomplete_closure(self):
        record = self.record()
        record["runs"]["dueño"]["status"] = "failed"
        record["runs"]["aislamiento"]["status"] = "incomplete"
        del record["runs"]["lint"]
        evidence = self.root / "observed.json"
        evidence.write_text(json.dumps(record), encoding="utf-8")
        original = self.board.read_bytes()
        code, status = self.invoke("quality", "TASK-001", "--evidence-file", str(evidence))
        self.assertEqual(code, 1, status)
        for case, expected in (("concurrencia", "passed"), ("dueño", "failed"),
                               ("aislamiento", "incomplete"), ("lint", "not-run")):
            self.assertIn(f"{case}: {expected}", status)
        result = self.invoke("set", "TASK-001", "DONE", "--evidence", json.dumps(record))
        self.assertEqual(result[0], 2, result)
        self.assertEqual(self.board.read_bytes(), original)

    def test_all_pass_current_and_command_is_never_executed(self):
        record = self.record()
        marker = self.root / "must-not-exist"
        record["runs"]["lint"]["command"] = f'echo unsafe > "{marker}"'
        evidence = self.root / "observed.json"
        evidence.write_text(json.dumps(record, indent=2), encoding="utf-8-sig")
        result = self.invoke("set", "TASK-001", "DONE", "--evidence-file", str(evidence))
        self.assertEqual(result[0], 0, result)
        code, quality = self.invoke("quality", "TASK-001")
        self.assertEqual(code, 0, quality)
        self.assertIn("Quality gate: passed", quality)
        self.assertIn("Approval: approved", quality)
        self.assertIn("Autorización de entrega:", quality)
        self.assertEqual(self.invoke("check", "--current")[0], 0)
        self.assertFalse(marker.exists())

    def test_pass_does_not_grant_approval_or_move_a_task(self):
        record = self.record()
        self.board.write_text(self.source.replace("Approval:** approved", "Approval:** pending")
                             .replace("State:** VERIFYING", "State:** PLANNED"), encoding="utf-8")
        evidence = self.root / "observed.json"
        evidence.write_text(json.dumps(record), encoding="utf-8")
        original = self.board.read_bytes()
        code, quality = self.invoke("quality", "TASK-001", "--evidence-file", str(evidence))
        self.assertEqual(code, 0, quality)
        self.assertIn("Approval: pending", quality)
        self.assertEqual(self.invoke("set", "TASK-001", "READY")[0], 2)
        self.assertEqual(self.board.read_bytes(), original)

    def test_invalidates_changes_to_files_requirements_acceptance_tests_and_risk(self):
        record = self.record()
        result = self.invoke("set", "TASK-001", "DONE", "--evidence", json.dumps(record))
        self.assertEqual(result[0], 0, result)
        done = self.board.read_text(encoding="utf-8")
        for changed in (
            done.replace("sola reserva", "sola confirmación"),
            done.replace("un turno libre", "un turno cancelado"),
            done.replace("test_reserva_unica", "test_otro_caso"),
            done.replace("Risk:** critical", "Risk:** standard"),
        ):
            self.board.write_text(changed, encoding="utf-8")
            with self.subTest(changed=changed):
                self.assertEqual(self.invoke("check")[0], 0)
                self.assertEqual(self.invoke("check", "--current")[0], 2)
                code, quality = self.invoke("quality", "TASK-001")
                self.assertEqual(code, 1, quality)
                self.assertIn("stale", quality)
        self.board.write_text(done, encoding="utf-8")
        self.app.write_text("drift\n", encoding="utf-8")
        self.assertEqual(self.invoke("check", "--current")[0], 2)

    def test_invalid_risk_policy_or_evidence_does_not_qualify(self):
        self.board.write_text(self.source.replace('"kind":"negative"', '"kind":"test"'), encoding="utf-8")
        self.assertIn("negative", self.invoke("check")[1])
        self.board.write_text(self.source, encoding="utf-8")
        for changes in ({"version": ""}, {"observed_at": "pending"}, {"status": "ok"}):
            record = self.record()
            record["runs"]["lint"].update(changes)
            self.assertEqual(self.invoke("set", "TASK-001", "DONE", "--evidence", json.dumps(record))[0], 2)

    def test_revalidated_dependency_invalidates_the_previous_dependent_result(self):
        task = self.source[self.source.index("### TASK-001"):]
        dependent = (task.replace("TASK-001", "TASK-002").replace("Depends on:** none", "Depends on:** TASK-001")
                     .replace("State:** VERIFYING", "State:** PLANNED").replace('"app.py"', '"other.py"'))
        (self.root / "other.py").write_text("dependent", encoding="utf-8")
        self.board.write_text(self.source + "\n" + dependent, encoding="utf-8")
        prior = self.record()
        self.assertEqual(self.invoke("set", "TASK-001", "DONE", "--evidence", json.dumps(prior))[0], 0)
        for state in ("READY", "IN_PROGRESS", "VERIFYING"):
            self.assertEqual(self.invoke("set", "TASK-002", state)[0], 0)
        original = self.record("TASK-002")
        self.assertEqual(self.invoke("set", "TASK-002", "DONE", "--evidence", json.dumps(original))[0], 0)
        self.assertEqual(self.invoke("check", "--current")[0], 0)

        self.app.write_text("changed dependency", encoding="utf-8")
        refreshed = self.record()
        source = self.board.read_text(encoding="utf-8")
        self.board.write_text(source.replace(prior["candidate"], refreshed["candidate"], 1), encoding="utf-8")
        code, report = self.invoke("check", "--current")
        self.assertEqual(code, 2, report)
        self.assertIn("TASK-002: candidato o criterio cambió", report)
        self.assertNotIn("TASK-001: candidato o criterio cambió", report)
        self.assertEqual(self.invoke("quality", "TASK-002")[0], 1)

        self.assertEqual(self.invoke("set", "TASK-002", "READY", "--reason", "Cambió su dependencia")[0], 0)
        for state in ("IN_PROGRESS", "VERIFYING"):
            self.assertEqual(self.invoke("set", "TASK-002", state)[0], 0)
        repeated = self.record("TASK-002")
        self.assertEqual(self.invoke("set", "TASK-002", "DONE", "--evidence", json.dumps(repeated))[0], 0)
        self.assertEqual(self.invoke("check", "--current")[0], 0)

    def test_impact_identifies_regression_cases_without_running_them(self):
        task = self.source[self.source.index("### TASK-001"):]
        dependent = (task.replace("TASK-001", "TASK-002").replace("Depends on:** none", "Depends on:** TASK-001")
                     .replace("State:** VERIFYING", "State:** PLANNED").replace('"app.py"', '"other.py"')
                     .replace('"concurrencia"', '"integracion"'))
        self.board.write_text(self.source + "\n" + dependent, encoding="utf-8")
        code, report = self.invoke("impact", "app.py")
        self.assertEqual(code, 0, report)
        self.assertIn("concurrencia", report)
        self.assertIn("test_reserva_ajena", report)
        self.assertIn("critical", report)
        self.assertIn("Regresión (dependencia de tarea", report)
        self.assertIn("integracion [test]", report)
