"""Static structure and secret-hygiene checks. Python 3.11+; no dependencies."""
from pathlib import Path
import hashlib
import json
import re
import tomllib

root = Path(__file__).resolve().parents[1]
roles = {"solution_architect", "product_designer", "test_engineer", "quality_reviewer", "chuli_worker", "chuli_explorer"}
skills = {"harness-discovery", "harness-architecture", "harness-ux-ui",
          "harness-specification", "harness-tdd-qa"}

skills.update({"branch-pr", "chained-pr", "cognitive-doc-design", "comment-writer",
               "gentle-ai-bench", "go-testing", "hermes-ephemeral-delegation", "issue-creation",
               "judgment-day", "rdd-defect-workflow", "skill-creator", "skill-improver",
               "skill-registry", "systemic-issue-triage", "work-unit-commits",
               "gentle-ai-collab-perfect", "issue-root-resolution", "rdd-advisory-transport", "diagram-design"})
assert {p.stem for p in (root / "agents").glob("*.toml")} == roles
assert {p.parent.name for p in (root / "skills").glob("*/SKILL.md")} == skills

for role in roles:
    path = root / "agents" / f"{role}.toml"
    assert path.exists(), str(path)
    data = tomllib.loads(path.read_text(encoding="utf-8"))
    assert data["name"] == role
    assert data["description"].strip()
    assert data["developer_instructions"].strip()
    assert data["model"].startswith("gpt-6")

for skill in skills:
    path = root / "skills" / skill / "SKILL.md"
    assert path.exists(), str(path)
    content = path.read_text(encoding="utf-8")
    assert content.startswith("---\n")
    assert re.search(r"^name: " + re.escape(skill) + r"$", content, re.MULTILINE)
    assert re.search(r"^description: ", content, re.MULTILINE)

instructions = (root / "config/AGENTS.md").read_text(encoding="utf-8")
for required in ("R0", "R1", "R2", "Criticidad", "Human-in-the-Loop", "Engram", "CES v1.0"):
    assert required in instructions, required

for required_path in ("docs/protocols/TTP.md", "templates/feature-tasks.md", "scripts/taskboard.py", "scripts/legacy-managed-hashes.json", "tests/test_taskboard.py", "VERSION"):
    assert (root / required_path).is_file(), required_path
assert (root / "VERSION").read_text(encoding="utf-8").strip() == "1.1.1"

legacy_hashes = json.loads((root / "scripts/legacy-managed-hashes.json").read_text(encoding="utf-8"))
assert set(legacy_hashes) == {"0.1.3", "0.2.0"}
for version, entries in legacy_hashes.items():
    assert entries, version
    for relative, digest in entries.items():
        assert not Path(relative).is_absolute(), relative
        assert re.fullmatch(r"[a-f0-9]{64}", digest), f"{version}: {relative}"
legacy_fixture = root / "tests/fixtures/legacy-v0.2.0-harness-specification.md"
legacy_digest = hashlib.sha256(legacy_fixture.read_bytes()).hexdigest()
assert legacy_digest == legacy_hashes["0.2.0"]["skills/harness-specification/SKILL.md"]
assert "update_plan" in (root / "scripts/toml_config.py").read_text(encoding="utf-8")
assert "TTP v1.0" in instructions
for required_path in ("evals/manifest.json", "evals/rubrics/criteria.md", "docs/EVALUATION.md", "internal/eval/catalog.go", "internal/eval/snapshot.go", "internal/eval/run.go", "internal/eval/cli.go", "internal/eval/eval_test.go"):
    assert (root / required_path).is_file(), required_path

for script in ("install.ps1", "verify.ps1", "rollback.ps1"):
    assert (root / "scripts" / script).is_file(), script

# Avoid committing actual Windows homes, tokens, or user memories.
windows_home = re.compile(r"C:[\\/]+Users[\\/]+[^\\/\s]+[\\/]", re.IGNORECASE)
for path in root.rglob("*"):
    if not path.is_file() or ".git" in path.parts:
        continue
    relative = path.relative_to(root).as_posix().lower()
    assert path.name.lower() not in {"auth.json", "config.toml"} and not relative.endswith(".sqlite3")
    assert "/.engram/" not in "/" + relative
    if path.suffix.lower() not in (".md", ".toml", ".ps1", ".json", ".yml", ".py"):
        continue
    source = path.read_text(encoding="utf-8")
    assert not windows_home.search(source), relative
    for secret_prefix in ("gh" + "p_", "sk-" + "proj-"):
        assert secret_prefix not in source, relative

print(f"OK: {len(roles)} agents, {len(skills)} skills, EWP/CES and install scripts")
