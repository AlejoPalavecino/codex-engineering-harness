"""Compare supplied workflow observations offline; never executes an agent.

Exit 0: comparable pairs, 1: incompatible conditions, 2: invalid input.
Comparability is not a quality verdict. Evidence references are not authenticated.
"""

import json
import math
from pathlib import Path
import sys

SCENARIOS = ("r1-small", "r2-interruption", "critical-permissions")
CONDITIONS = {"fixture", "prompt", "client", "host", "permissions", "rubric",
              "procedure", "model", "tools"}
METRICS = ("repeated_questions", "reworked_units", "first_useful_test_seconds")
QUALITY = {"PASS", "FAIL", "PARTIAL", "NOT_TESTED"}


def require(condition):
    if not condition:
        raise ValueError("Invalid observation schema")


def nonempty(value):
    return isinstance(value, str) and bool(value.strip())


def validate(run):
    require(isinstance(run, dict) and set(run) == {"schema", "variant", "synthetic", "observations"})
    require(type(run["schema"]) is int and run["schema"] == 1)
    require(nonempty(run["variant"]) and type(run["synthetic"]) is bool)
    observations = run["observations"]
    require(isinstance(observations, list) and len(observations) == len(SCENARIOS))
    indexed = {}
    for row in observations:
        require(isinstance(row, dict) and set(row) ==
                {"scenario", "conditions", "metrics", "quality", "evidence"})
        name = row["scenario"]
        require(isinstance(name, str) and name in SCENARIOS and name not in indexed)
        conditions = row["conditions"]
        require(isinstance(conditions, dict) and set(conditions) == CONDITIONS)
        require(all(value is None or nonempty(value) for value in conditions.values()))
        metrics = row["metrics"]
        require(isinstance(metrics, dict) and set(metrics) == set(METRICS))
        for key, value in metrics.items():
            if value is None:
                continue
            require(type(value) in (int, float) and math.isfinite(value) and value >= 0)
            if key != "first_useful_test_seconds":
                require(type(value) is int)
        require(isinstance(row["quality"], str) and row["quality"] in QUALITY)
        evidence = row["evidence"]
        require(isinstance(evidence, list) and all(nonempty(item) for item in evidence))
        if any(value is not None for value in metrics.values()) or row["quality"] != "NOT_TESTED":
            require(bool(evidence))
        indexed[name] = row
    return indexed


def compare(baseline, candidate):
    before, after = validate(baseline), validate(candidate)
    result = {"schema": 1, "status": "COMPARABLE",
              "synthetic": baseline["synthetic"] or candidate["synthetic"],
              "evidence_authenticated": False, "scenarios": []}
    for name in SCENARIOS:
        left, right = before[name], after[name]
        mismatch = sorted(key for key in CONDITIONS
                          if left["conditions"][key] is None or right["conditions"][key] is None
                          or left["conditions"][key] != right["conditions"][key])
        if baseline["synthetic"] != candidate["synthetic"]:
            mismatch.append("synthetic")
        row = {"scenario": name, "status": "NOT_COMPARABLE" if mismatch else "COMPARABLE"}
        if mismatch:
            result["status"] = "NOT_COMPARABLE"
            row["incompatible_or_unknown_conditions"] = mismatch
        else:
            row["quality"] = {"baseline": left["quality"], "candidate": right["quality"]}
            row["quality_regression"] = (
                right["quality"] == "FAIL" and left["quality"] in {"PASS", "PARTIAL"}
                or left["quality"] == "PASS" and right["quality"] == "PARTIAL")
            row["quality_coverage_lost"] = (
                left["quality"] != "NOT_TESTED" and right["quality"] == "NOT_TESTED")
            row["metrics"] = {}
            for key in METRICS:
                a, b = left["metrics"][key], right["metrics"][key]
                known = a is not None and b is not None
                row["metrics"][key] = {"baseline": a, "candidate": b,
                                       "delta": b - a if known else None,
                                       "status": "MEASURED" if known else "UNKNOWN"}
        result["scenarios"].append(row)
    return result


def unique_object(pairs):
    result = {}
    for key, value in pairs:
        require(key not in result)
        result[key] = value
    return result


def main(argv=None):
    args = sys.argv[1:] if argv is None else argv
    try:
        require(len(args) == 2)
        runs = [json.loads(Path(path).read_text(encoding="utf-8"),
                           object_pairs_hook=unique_object) for path in args]
        result = compare(*runs)
    except (ValueError, TypeError, OSError, OverflowError, RecursionError):
        print(json.dumps({"status": "INVALID_INPUT"}))
        return 2
    print(json.dumps(result, indent=2, allow_nan=False))
    return 0 if result["status"] == "COMPARABLE" else 1


if __name__ == "__main__":
    sys.exit(main())
