import importlib.util
from pathlib import Path
import sys
import tomllib
import unittest

ROOT = Path(__file__).resolve().parents[1]
spec = importlib.util.spec_from_file_location("toml_config", ROOT / "scripts" / "toml_config.py")
mod = importlib.util.module_from_spec(spec)
sys.modules["toml_config"] = mod
spec.loader.exec_module(mod)


class SafeTomlTests(unittest.TestCase):
    def test_comment_after_header_no_duplicate(self):
        source = '[tools.update_plan] # valid comment\n enabled = false\n[mcp_servers.engram]\ncommand = "engram"\n'
        updated, report = mod.prepare(source)
        parsed = tomllib.loads(updated)
        self.assertFalse(parsed["tools"]["update_plan"]["enabled"])
        self.assertEqual(parsed["mcp_servers"]["engram"]["command"], "engram")
        self.assertEqual(updated.count("[tools.update_plan]"), 1)
        self.assertIn("tools.update_plan.enabled", report["preserved"])

    def test_explicit_replacement_comment_header(self):
        source = '[tools.update_plan] # comment\n enabled = false\n'
        updated, report = mod.prepare(source, replace_managed=True)
        self.assertTrue(tomllib.loads(updated)["tools"]["update_plan"]["enabled"])
        self.assertEqual(updated.count("[tools.update_plan]"), 1)
        self.assertIn("tools.update_plan.enabled", report["changed"])

    def test_unknown_keys_and_comments_survive(self):
        source = '# personal note\nmodel = "other"\n[mcp_servers.github]\nurl="example"\n'
        updated, _ = mod.prepare(source)
        doc = tomllib.loads(updated)
        self.assertEqual(doc["model"], "other")
        self.assertEqual(doc["mcp_servers"]["github"]["url"], "example")
        self.assertIn("# personal note", updated)

    def test_implicit_parent_table_is_valid(self):
        source = '[agents.solution_architect]\ndescription = "x"\n'
        updated, _ = mod.prepare(source, configure_models=True, replace_managed=True)
        parsed = tomllib.loads(updated)
        self.assertEqual(parsed["agents"]["default_subagent_model"], "gpt-6.1-sol")
        self.assertEqual(parsed["agents"]["solution_architect"]["description"],
                         tomllib.loads((ROOT / "agents/solution_architect.toml").read_text())["description"])

    def test_new_profile_uses_defaults_but_individual_overrides_survive(self):
        current, report = mod.prepare('model = "personal"\n')
        parsed = tomllib.loads(current)
        self.assertEqual(parsed["model"], "personal")
        self.assertIn("model", report["preserved"])
        self.assertEqual(parsed["model_reasoning_effort"], "medium")
        self.assertEqual(parsed["agents"]["default_subagent_model"], "gpt-6.1-sol")
        self.assertEqual(set(parsed["agents"]) - {"default_subagent_model"},
                         {p.stem for p in (ROOT / "agents").glob("*.toml")})
        new, _ = mod.prepare("")
        self.assertEqual(tomllib.loads(new)["model"], "gpt-6.1-sol")

    def test_models_only_on_explicit_flag(self):
        source = 'model = "another-model"\n'
        updated, report = mod.prepare(source, configure_models=True)
        self.assertEqual(tomllib.loads(updated)["model"], "gpt-6.1-sol")
        self.assertIn("model", report["changed"])

    def test_reject_inline_table_without_writing(self):
        source = '[tools]\nupdate_plan = { enabled = false }\n'
        with self.assertRaisesRegex(ValueError, "Unsafe TOML layout"):
            mod.prepare(source, replace_managed=True)

    def test_reject_invalid_existing(self):
        with self.assertRaises(tomllib.TOMLDecodeError):
            mod.prepare('[tools.update_plan]\nenabled = false\n[tools.update_plan]\nenabled = true\n')

    def test_validate_managed_drift(self):
        current, report = mod.prepare("")
        state = {"managed_config": report["tracked"]}
        self.assertEqual(mod.validate(current, state)["managed_drift"], [])
        altered = current.replace('enabled = true', 'enabled = false')
        self.assertIn("tools.update_plan.enabled", mod.validate(altered, state)["managed_drift"])

    def test_skip_plan(self):
        current, report = mod.prepare("", skip_task_progress=True)
        self.assertNotIn("update_plan", current)
        self.assertNotIn("tools.update_plan.enabled", report["tracked"])


if __name__ == "__main__":
    unittest.main()
