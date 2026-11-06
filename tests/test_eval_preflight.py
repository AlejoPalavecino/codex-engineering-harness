import contextlib
import copy
import importlib.util
import io
import json
from pathlib import Path
import tempfile
import unittest

ROOT = Path(__file__).resolve().parents[1]
spec = importlib.util.spec_from_file_location("eval_preflight", ROOT / "scripts/eval_preflight.py")
module = importlib.util.module_from_spec(spec)
spec.loader.exec_module(module)


def evidence(permission="read-only"):
    return {
        "schema": 1,
        "client": {"name": "codex-cli", "version": "1.2.3"},
        "harness_commit": "a" * 40,
        "catalog_fingerprint": "b" * 64,
        "requested_permission": permission,
        "observed_permission": permission,
        "probes": [
            {"capability": cap, "exit_code": 0, "marker_expected": "OK_" + cap,
             "stdout": "OK_" + cap}
            for cap in (["read"] if permission == "read-only" else ["read", "write", "tests"])
        ],
    }


class PreflightTests(unittest.TestCase):
    def run_input(self, payload, raw=False):
        with tempfile.TemporaryDirectory() as folder:
            path = Path(folder) / "input.json"
            path.write_text(payload if raw else json.dumps(payload), encoding="utf-8")
            output = io.StringIO()
            with contextlib.redirect_stdout(output):
                code = module.main([str(path)])
            return code, json.loads(output.getvalue())

    def test_profiles(self):
        for permission, expected in [("read-only", ["read"]),
                                     ("workspace-write", ["read", "write", "tests"])]:
            with self.subTest(permission=permission):
                code, result = self.run_input(evidence(permission))
                self.assertEqual((code, result["status"]), (0, "READY"))
                self.assertEqual(result["verified_capabilities"], expected)
                self.assertIn("EVIDENCE_NOT_AUTHENTICATED", result["reason_codes"])

    def test_mismatched_permissions_block(self):
        payload = evidence("workspace-write")
        payload["observed_permission"] = "read-only"
        self.assertEqual(self.run_input(payload)[0], 1)

    def test_successful_write_contradicts_read_only_permission(self):
        payload = evidence()
        payload["probes"].append({"capability": "write", "exit_code": 0,
                                  "marker_expected": "SYNTHETIC_WRITE_OK",
                                  "stdout": "SYNTHETIC_WRITE_OK"})
        code, result = self.run_input(payload)
        self.assertEqual((code, result["status"]), (1, "BLOCKED"))
        self.assertIn("PERMISSION_EVIDENCE_CONFLICT", result["reason_codes"])

    def test_marker_in_prose_does_not_prove_execution(self):
        payload = evidence()
        payload["probes"][0].update(marker_expected="SYNTHETIC_READ_OK",
                                    stdout="Expected SYNTHETIC_READ_OK but no tool ran")
        code, result = self.run_input(payload)
        self.assertEqual((code, result["status"]), (1, "BLOCKED"))
        self.assertEqual(result["verified_capabilities"], [])

    def test_marker_as_whole_line_allows_surrounding_whitespace(self):
        payload = evidence()
        payload["probes"][0].update(marker_expected="SYNTHETIC_READ_OK",
                                    stdout="diagnostic\r\n  SYNTHETIC_READ_OK\t\r\nfinished")
        self.assertEqual(self.run_input(payload)[0], 0)

    def test_missing_provenance_is_not_verifiable(self):
        for key in ["client", "harness_commit", "catalog_fingerprint", "observed_permission"]:
            payload = evidence()
            del payload[key]
            with self.subTest(key=key):
                code, result = self.run_input(payload)
                self.assertEqual((code, result["status"]), (2, "NOT_VERIFIABLE"))
        payload = evidence()
        payload["observed_permission"] = None
        self.assertEqual(self.run_input(payload)[0], 2)

    def test_completed_turn_without_probes_does_not_pass(self):
        payload = evidence()
        payload.update(probes=[], events=[{"type": "turn.completed", "exit_code": 0}])
        self.assertEqual(self.run_input(payload)[0], 1)

    def test_failed_or_missing_required_probes_block(self):
        for change in [{"exit_code": 3}, {"marker_expected": ""},
                       {"stdout": "model says tools work"}]:
            payload = evidence()
            payload["probes"][0].update(change)
            self.assertEqual(self.run_input(payload)[0], 1)
        for missing in ["read", "write", "tests"]:
            payload = evidence("workspace-write")
            payload["probes"] = [p for p in payload["probes"] if p["capability"] != missing]
            self.assertEqual(self.run_input(payload)[0], 1)

    def test_failed_optional_or_duplicate_probe_blocks(self):
        payload = evidence()
        probe = copy.deepcopy(payload["probes"][0])
        probe["exit_code"] = 1
        payload["probes"].append(probe)
        self.assertEqual(self.run_input(payload)[0], 1)

    def test_event_errors_block(self):
        for event in [{"type": "turn.failed", "error": {"message": "private"}},
                      {"type": "error", "message": "private"},
                      {"type": "item.completed", "item": {"type": "command_execution", "exit_code": 1}},
                      {"type": "item.completed", "item": {"type": "mcp_tool_call", "status": "failed"}},
                      {"type": "tool.result", "success": False},
                      {"type": "tool.result", "isError": True}]:
            payload = evidence()
            payload["events"] = [event]
            with self.subTest(event=event):
                self.assertEqual(self.run_input(payload)[0], 1)

    def test_invalid_input(self):
        mutations = [lambda p: p.update(schema=True), lambda p: p.update(schema=2),
                     lambda p: p.update(extra="private"), lambda p: p["client"].update(extra=1),
                     lambda p: p.update(requested_permission="all"),
                     lambda p: p.update(catalog_fingerprint="not a hash"),
                     lambda p: p["probes"][0].update(exit_code=True),
                     lambda p: p["probes"][0].update(extra=1),
                     lambda p: p["probes"][0].update(capability="model_claim"),
                     lambda p: p.update(events=["invalid"])]
        for mutation in mutations:
            payload = evidence()
            mutation(payload)
            self.assertEqual(self.run_input(payload)[0], 2)
        for raw in ["{", "[]", "null", '{"schema":1,"schema":1}']:
            self.assertEqual(self.run_input(raw, raw=True)[0], 2)

    def test_no_private_output(self):
        payload = evidence()
        secret = "secret-C:/private/token-super-sensitive"
        payload["probes"][0]["stdout"] += secret
        payload["events"] = [{"type": "error", "message": secret}]
        code, result = self.run_input(payload)
        self.assertEqual(code, 1)
        self.assertNotIn(secret, json.dumps(result))
        self.assertEqual(set(result), {"schema", "status", "reason_codes", "verified_capabilities",
                                      "client", "harness_commit", "catalog_fingerprint"})
        payload["client"]["name"] = secret
        code, result = self.run_input(payload)
        self.assertEqual(code, 2)
        self.assertNotIn(secret, json.dumps(result))

    def test_unreadable_file_and_usage_are_sanitized(self):
        for args in [[], ["C:/nonexistent/private-token.json"], ["one", "two"]]:
            output = io.StringIO()
            with contextlib.redirect_stdout(output):
                self.assertEqual(module.main(args), 2)
            self.assertNotIn("private-token", output.getvalue())
            self.assertEqual(json.loads(output.getvalue())["status"], "NOT_VERIFIABLE")


if __name__ == "__main__":
    unittest.main()
