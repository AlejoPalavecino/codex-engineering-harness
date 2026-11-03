"""Static structure and secret-hygiene checks. Python 3.11+; no dependencies."""
from pathlib import Path
import re
import tomllib

root = Path(__file__).resolve().parents[1]
roles = {"solution_architect", "product_designer", "test_engineer", "quality_reviewer"}
skills = {"harness-discovery", "harness-architecture", "harness-ux-ui",
          "harness-specification", "harness-tdd-qa"}

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

for required_path in ("docs/protocols/TTP.md", "templates/feature-tasks.md", "scripts/taskboard.py", "tests/test_taskboard.py", "VERSION"):
    assert (root / required_path).is_file(), required_path
assert (root / "VERSION").read_text(encoding="utf-8").strip() == "0.1.2"
assert "update_plan" in (root / "scripts/toml_config.py").read_text(encoding="utf-8")
assert "TTP v1.0" in instructions

for script in ("install.ps1", "verify.ps1", "rollback.ps1"):
    assert (root / "scripts" / script).is_file(), script

# Avoid committing actual Windows homes, tokens, or user memories.
windows_home = re.compile(r"C:[\\/]+Users[\\/]+[^\\/\s]+[\\/]", re.IGNORECASE)
for path in root.rglob("*"):
    if not path.is_file() or ".git" in path.parts:
        continue
    relative = path.relative_to(root).as_posix().lower()
    assert not relative.endswith(("/auth.json", "/config.toml", ".sqlite3"))
    assert "/.engram/" not in "/" + relative
    if path.suffix.lower() not in (".md", ".toml", ".ps1", ".json", ".yml", ".py"):
        continue
    source = path.read_text(encoding="utf-8")
    assert not windows_home.search(source), relative
    for secret_prefix in ("gh" + "p_", "sk-" + "proj-"):
        assert secret_prefix not in source, relative

print(f"OK: {len(roles)} agents, {len(skills)} skills, EWP/CES and install scripts")
