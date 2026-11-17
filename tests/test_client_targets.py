"""Unit tests for multi-client catalog, agent rendering, and JSON settings (Chuli v2.0)."""
import json
from pathlib import Path
import sys
import tempfile
import tomllib
import unittest

sys.path.insert(0, str(Path(__file__).resolve().parents[1] / "scripts"))
import client_targets as ct
import profile_resources as profile


class ClientTargetsTests(unittest.TestCase):
    def test_agy_exact_label_survives_resolution_and_settings(self):
        _, roles = ct.resolve_role_models("agy", overrides={"orchestrator": {
            "model": "gemini-next-high", "effort": "low", "label": "Gemini Next (High)"
        }})
        self.assertEqual(roles["orchestrator"]["model"], "gemini-next-high")
        raw, _ = ct.prepare_json_settings(b'{"theme":"dark"}', "agy", roles, configure_models=True)
        self.assertEqual(json.loads(raw)["model"], "Gemini Next (High)")
        self.assertEqual(json.loads(raw)["theme"], "dark")
        roles["orchestrator"].pop("label")
        with self.assertRaisesRegex(ValueError, "exact CLI label"):
            ct.prepare_json_settings(b"{}", "agy", roles, configure_models=True)

    def test_agy_unknown_specialist_never_silently_inherits(self):
        with self.assertRaisesRegex(ValueError, "specialist"):
            ct.parse_role_spec(["chuli_worker=new-model:high"], "agy")
        _, legacy = ct.resolve_role_models("agy", overrides={"chuli_worker": {"model": "new-model", "effort": "high"}})
        self.assertEqual(ct.agy_agent_tier(legacy["chuli_worker"]["model"], "medium"), "inherit")
        self.assertEqual(ct.agy_agent_tier("gemini-4-argon", "medium"), "pro")

    def test_agy_rejects_unsafe_labels(self):
        for label in ("Gemini\nInjected", "Gemini `command`", "x" * 161):
            with self.subTest(label=label), self.assertRaisesRegex(ValueError, "label"):
                ct.resolve_role_models("agy", overrides={"orchestrator": {
                    "model": "new-model", "effort": "high", "label": label
                }})

    def test_catalog_contains_all_clients_roles_and_presets(self):
        cat = ct.load_catalog()
        roles = ct.role_ids(cat)
        self.assertEqual(
            roles,
            [
                "orchestrator",
                "chuli_explorer",
                "chuli_worker",
                "solution_architect",
                "product_designer",
                "test_engineer",
                "quality_reviewer",
            ],
        )
        self.assertEqual(set(cat["clients"]), {"codex", "claude", "agy"})
        for client, spec in cat["clients"].items():
            self.assertGreaterEqual(len(spec["presets"]), 3, client)
            for preset_name, preset in spec["presets"].items():
                self.assertEqual(set(preset["roles"]), set(roles), f"{client}/{preset_name}")

    def test_resolve_and_custom_role_overrides(self):
        preset, roles = ct.resolve_role_models("claude", "claude-balanced")
        self.assertEqual(preset, "claude-balanced")
        self.assertEqual(roles["orchestrator"]["model"], "claude-opus-5-5")
        self.assertEqual(roles["solution_architect"]["model"], "claude-fable-5-2")
        self.assertEqual(roles["chuli_explorer"]["model"], "claude-haiku-5-5")

        overrides = ct.parse_role_spec(["orchestrator=claude-fable-5-2:high", "chuli_worker=claude-opus-5-5:high"], "claude")
        preset, custom = ct.resolve_role_models("claude", "", overrides)
        self.assertEqual(preset, "custom")
        self.assertEqual(custom["orchestrator"], {"model": "claude-fable-5-2", "effort": "high"})
        self.assertEqual(custom["chuli_worker"], {"model": "claude-opus-5-5", "effort": "high"})

        with self.assertRaises(ValueError):
            ct.resolve_role_models("claude", "", {"unknown_role": {"model": "claude-opus-5-5", "effort": "medium"}})
        with self.assertRaises(ValueError):
            ct.resolve_role_models("codex", "", {"orchestrator": {"model": "gpt-6.1-sol", "effort": "invalid"}})

    def test_agy_model_display_formatting(self):
        self.assertEqual(ct.format_agy_model("orchestrator-argon", "medium"), "Orchestrator (Argon)")
        self.assertEqual(ct.format_agy_model("orchestrator-carbon", "high"), "Orchestrator (Carbon)")
        self.assertEqual(ct.format_agy_model("gemini-4-carbon", "high"), "Gemini 4 Carbon (High)")
        self.assertEqual(ct.format_agy_model("gemini-4-argon", "medium"), "Gemini 4 Argon (Medium)")
        self.assertEqual(ct.format_agy_model("gemini-4-hydrogen", "xhigh"), "Gemini 4 Hydrogen (Extra)")
        self.assertEqual(ct.agy_agent_tier("gemini-4-carbon", "high"), "pro")
        self.assertEqual(ct.agy_agent_tier("gemini-4-argon", "medium"), "pro")
        self.assertEqual(ct.agy_agent_tier("gemini-4-hydrogen", "high"), "flash")
        self.assertEqual(ct.agy_agent_tier("gemini-4-hydrogen", "low"), "flash_lite")
        self.assertEqual(ct.agy_agent_tier("inherit", "medium"), "inherit")

    def test_render_agents_for_codex_claude_and_agy(self):
        agents = ct.load_source_agents()
        codex_bytes = ct.render_codex_agent(agents["chuli_explorer"], "gpt-6.1-sol", "high")
        parsed_codex = tomllib.loads(codex_bytes.decode("utf-8"))
        self.assertEqual(parsed_codex["model"], "gpt-6.1-sol")
        self.assertEqual(parsed_codex["model_reasoning_effort"], "high")
        self.assertEqual(parsed_codex["sandbox_mode"], "read-only")

        claude_explorer = ct.render_claude_agent(agents["chuli_explorer"], "claude-haiku-5-5", "low").decode("utf-8")
        self.assertIn("name: chuli_explorer", claude_explorer)
        self.assertIn("model: claude-haiku-5-5", claude_explorer)
        self.assertIn("disallowedTools: Write, Edit", claude_explorer)

        claude_worker = ct.render_claude_agent(agents["chuli_worker"], "claude-sonnet-5-5", "medium").decode("utf-8")
        self.assertNotIn("disallowedTools", claude_worker)

        agy_arch = ct.render_agy_agent(agents["solution_architect"], "gemini-4-carbon", "high").decode("utf-8")
        self.assertIn("subagent: true", agy_arch)
        self.assertIn("mainAgent: false", agy_arch)
        self.assertIn("model: pro", agy_arch)
        self.assertNotIn("Gemini 4 Carbon", agy_arch)
        self.assertIn("tools: [read_file, run_command, manage_task, search_web]", agy_arch)

        agy_explorer = ct.render_agy_agent(agents["chuli_explorer"], "gemini-4-hydrogen", "low").decode("utf-8")
        self.assertIn("model: flash_lite", agy_explorer)

    def test_json_settings_preserves_bom_and_unrelated_keys(self):
        _, claude_roles = ct.resolve_role_models("claude", "claude-balanced")
        existing = ct.BOM + json.dumps(
            {
                "outputStyle": "concise",
                "permissions": {"allow": ["Bash(git status)"]},
                "model": "custom-user-model",
            }
        ).encode("utf-8")
        updated, report = ct.prepare_json_settings(existing, "claude", claude_roles, configure_models=False)
        self.assertTrue(updated.startswith(ct.BOM))
        self.assertIn("model", report["preserved"])
        data = json.loads(updated.decode("utf-8-sig"))
        self.assertEqual(data["model"], "custom-user-model")
        self.assertEqual(data["outputStyle"], "concise")
        self.assertEqual(data["env"]["CLAUDE_CODE_EFFORT_LEVEL"], "medium")
        self.assertEqual(data["env"]["CLAUDE_CODE_SUBAGENT_MODEL"], "claude-sonnet-5-5")

        replaced, report2 = ct.prepare_json_settings(existing, "claude", claude_roles, configure_models=True)
        data2 = json.loads(replaced.decode("utf-8-sig"))
        self.assertEqual(data2["model"], "claude-opus-5-5")
        self.assertEqual(report2["preserved"], [])

        _, agy_roles = ct.resolve_role_models("agy", "agy-balanced")
        agy_out, agy_report = ct.prepare_json_settings(b'{"colorScheme":"dark"}', "agy", agy_roles, configure_models=True)
        agy_data = json.loads(agy_out.decode("utf-8"))
        self.assertEqual(agy_data["model"], "Orchestrator (Argon)")
        self.assertEqual(agy_data["colorScheme"], "dark")
        self.assertEqual(
            ct.validate_json_settings(agy_out, {"managed_config": agy_report["tracked"]})["managed_drift"],
            [],
        )

    def test_profile_inventory_and_instructions_for_claude_and_agy(self):
        with tempfile.TemporaryDirectory(prefix="chuli-multiclient-") as tmp:
            base = Path(tmp).resolve()
            for client, expected_instr in (("claude", "CLAUDE.md"), ("agy", "plugins/chuli/rules/AGENTS.md")):
                home = base / client
                home.mkdir(parents=True, exist_ok=True)
                inv = profile.preflight(home, "install", False, client)
                self.assertIn(expected_instr, inv)
                out = base / f"{client}-instr.md"
                preset = f"{client}-balanced"
                report = profile.prepare_instructions(home, out, False, False, client, preset, "")
                self.assertFalse(report["preserved"])
                rendered = out.read_text(encoding="utf-8")
                if client == "claude":
                    self.assertIn("Claude Code", rendered)
                    self.assertIn("claude-fable-5-2", rendered)
                else:
                    self.assertIn("AGY CLI", rendered)
                    self.assertIn("Gemini 4 Carbon (High)", rendered)


if __name__ == "__main__":
    unittest.main()
