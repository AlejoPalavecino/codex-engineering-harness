"""Offline, untrusted-evidence gate; never runs probes or authenticates evidence.

Exit codes: 0 READY, 1 BLOCKED, 2 NOT_VERIFIABLE (including invalid input).
Only fixed reason codes and validated provenance are emitted. Event payloads and
probe output stay private. EVIDENCE_NOT_AUTHENTICATED applies to every result.
"""

import json
from pathlib import Path
import re
import sys

PERMISSIONS = {"read-only", "workspace-write"}
CAPABILITIES = ("read", "write", "tests")
FIELDS = {"schema", "client", "harness_commit", "catalog_fingerprint",
          "requested_permission", "observed_permission", "probes", "events"}


def _require(condition):
    if not condition:
        raise ValueError("Invalid input")


def _identifier(value):
    # Deliberately narrow: free text, paths and credential-like strings are not
    # safe provenance. This validates syntax, not authenticity or absence of secrets.
    return (isinstance(value, str) and len(value) <= 80
            and re.fullmatch(r"[A-Za-z0-9][A-Za-z0-9_.+\-]*", value) is not None)


def _unique_object(pairs):
    result = {}
    for key, value in pairs:
        _require(key not in result)
        result[key] = value
    return result


def _event_failed(value):
    if isinstance(value, list):
        return any(_event_failed(item) for item in value)
    if not isinstance(value, dict):
        return False
    if (value.get("type") in {"error", "turn.failed", "tool.failed"}
            or value.get("status") in {"failed", "error"}
            or value.get("success") is False or value.get("isError") is True
            or value.get("is_error") is True or bool(value.get("error"))):
        return True
    if "exit_code" in value:
        if type(value["exit_code"]) is not int or value["exit_code"] != 0:
            return True
    return any(_event_failed(item) for item in value.values())


def evaluate(data):
    _require(isinstance(data, dict) and not (set(data) - FIELDS))
    _require(type(data.get("schema")) is int and data["schema"] == 1)
    _require(data.get("requested_permission") in PERMISSIONS)
    _require(data.get("observed_permission") is None
             or data["observed_permission"] in PERMISSIONS)
    _require(isinstance(data.get("probes"), list))
    _require(isinstance(data.get("events", []), list))
    _require(all(isinstance(event, dict) for event in data.get("events", [])))
    result = {"schema": 1, "status": "NOT_VERIFIABLE",
              "reason_codes": ["EVIDENCE_NOT_AUTHENTICATED"], "verified_capabilities": []}
    incomplete = False
    client = data.get("client")
    if client is None:
        incomplete = True
    else:
        _require(isinstance(client, dict) and not (set(client) - {"name", "version"}))
        for key in ("name", "version"):
            if key in client:
                _require(_identifier(client[key]))
            else:
                incomplete = True
        if set(client) == {"name", "version"}:
            result["client"] = client.copy()
    for key, length in (("harness_commit", 40), ("catalog_fingerprint", 64)):
        if key not in data or data[key] is None:
            incomplete = True
        else:
            _require(isinstance(data[key], str)
                     and re.fullmatch(r"[a-fA-F0-9]{%d}" % length, data[key]) is not None)
            result[key] = data[key]
    incomplete |= data.get("observed_permission") is None
    passed = set()
    failed = False
    for probe in data["probes"]:
        _require(isinstance(probe, dict) and set(probe) == {
            "capability", "exit_code", "marker_expected", "stdout"})
        _require(probe["capability"] in CAPABILITIES)
        _require(type(probe["exit_code"]) is int)
        _require(isinstance(probe["marker_expected"], str) and isinstance(probe["stdout"], str))
        if (probe["exit_code"] == 0 and probe["marker_expected"]
                and any(probe["marker_expected"] == line.strip()
                        for line in probe["stdout"].splitlines())):
            passed.add(probe["capability"])
        else:
            failed = True
    result["verified_capabilities"] = [cap for cap in CAPABILITIES if cap in passed]
    reasons = result["reason_codes"]
    blocked = False
    if data.get("observed_permission") == "read-only" and "write" in passed:
        blocked = True
        reasons.append("PERMISSION_EVIDENCE_CONFLICT")
    if (data.get("observed_permission") is not None
            and data["requested_permission"] != data["observed_permission"]):
        blocked = True
        reasons.append("PERMISSION_MISMATCH")
    required = {"read"} if data["requested_permission"] == "read-only" else set(CAPABILITIES)
    if required - passed:
        blocked = True
        reasons.append("REQUIRED_PROBE_MISSING_OR_FAILED")
    if failed:
        blocked = True
        reasons.append("PROBE_FAILED")
    if any(_event_failed(event) for event in data.get("events", [])):
        blocked = True
        reasons.append("EVENT_ERROR")
    if incomplete:
        reasons.append("PROVENANCE_INCOMPLETE")
    result["status"] = "BLOCKED" if blocked else "NOT_VERIFIABLE" if incomplete else "READY"
    return result


def main(argv=None):
    args = sys.argv[1:] if argv is None else argv
    try:
        _require(len(args) == 1)
        data = json.loads(Path(args[0]).read_text(encoding="utf-8"),
                          object_pairs_hook=_unique_object,
                          parse_constant=lambda value: _require(False))
        result = evaluate(data)
    except (OSError, ValueError, TypeError, RecursionError):
        result = {"schema": 1, "status": "NOT_VERIFIABLE",
                  "reason_codes": ["EVIDENCE_NOT_AUTHENTICATED", "INVALID_INPUT"],
                  "verified_capabilities": []}
    print(json.dumps(result, sort_keys=True, separators=(",", ":")))
    return {"READY": 0, "BLOCKED": 1, "NOT_VERIFIABLE": 2}[result["status"]]


if __name__ == "__main__":
    sys.exit(main())
