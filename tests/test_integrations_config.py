"""Behavior contracts for the Manager's embedded configuration helper."""
import importlib.util
import json
from pathlib import Path
import tempfile
import unittest
from types import SimpleNamespace
from unittest.mock import patch
import stat

HELPER = Path(__file__).resolve().parents[1] / "internal/integrations/config.py"


class IntegrationsConfig(unittest.TestCase):
    def setUp(self):
        spec = importlib.util.spec_from_file_location("integrations_config", HELPER)
        self.helper = importlib.util.module_from_spec(spec)
        spec.loader.exec_module(self.helper)
        self.temp = tempfile.TemporaryDirectory()
        self.addCleanup(self.temp.cleanup)
        self.home = Path(self.temp.name).resolve()
        self.config = self.home / "config.toml"
        self.entry = {"url": "https://mcp.exa.ai/mcp"}

    def request(self, action="inspect", **extra):
        return self.helper.execute(dict(home=str(self.home), id="exa", entry=self.entry, action=action, **extra))

    def test_apply_idempotence_and_undo_preserve_other_bytes(self):
        original = b'# personal\r\nmodel = "private-value"\r\n[mcp_servers.other]\r\nurl="https://other.test"\r\n'
        self.config.write_bytes(original)
        preview = self.request()
        self.assertEqual(self.config.read_bytes(), original)
        self.request("configure", expected=preview["revision"])
        self.assertTrue(self.request()["managed"])
        applied = self.config.read_bytes()
        self.request("configure", expected=self.request()["revision"])
        self.assertEqual(self.config.read_bytes(), applied)
        self.config.write_bytes(applied + b'\n[features]\nexample=true\n')
        self.request("undo", expected=self.request()["revision"])
        self.assertEqual(self.config.read_bytes(), original + b'\n[features]\nexample=true\n')
        self.assertNotIn("private-value", json.dumps(self.request()))

    def test_duplicate_alias_plugin_and_same_name_are_preserved(self):
        for content in ('[mcp_servers.search]\nurl="https://mcp.exa.ai/mcp/"\n',
                        '[mcp_servers.exa]\ncommand="custom"\n',
                        '[plugins."exa@marketplace"]\nenabled=true\n'):
            with self.subTest(content=content):
                self.config.write_text(content, encoding="utf-8")
                self.assertTrue(self.request()["exists"])
                self.request("configure", expected=self.request()["revision"])
                self.assertEqual(self.config.read_text(encoding="utf-8"), content)
                with self.assertRaises(ValueError):
                    self.request("undo", expected=self.request()["revision"])

    def test_stale_preview_invalid_toml_and_edited_connection_are_rejected(self):
        before = self.request()["revision"]
        self.config.write_text('model="other"\n', encoding="utf-8")
        with self.assertRaises(ValueError):
            self.request("configure", expected=before)
        self.request("configure", expected=self.request()["revision"])
        self.config.write_text(self.config.read_text().replace('https://mcp.exa.ai/mcp', 'https://edited.test'), encoding="utf-8")
        with self.assertRaises(ValueError):
            self.request("undo", expected=self.request()["revision"])
        self.config.write_text('[broken secret="never print me"', encoding="utf-8")
        with self.assertRaisesRegex(ValueError, '^config.toml inválido'):
            self.request()

    def test_fresh_home_undo_and_lock(self):
        self.request("configure", expected=self.request()["revision"])
        self.request("undo", expected=self.request()["revision"])
        self.assertFalse(self.config.exists())
        (self.home / ".chuli-integrations.lock").mkdir()
        with self.assertRaises(ValueError):
            self.request("configure", expected=self.request()["revision"])

    def test_symlink_config_and_modified_subtable_cannot_be_overwritten(self):
        target = self.home / "elsewhere.toml"
        target.write_text('model="outside"', encoding="utf-8")
        try:
            self.config.symlink_to(target)
        except OSError:
            self.skipTest("symlink unavailable")
        with self.assertRaises(ValueError):
            self.request()
        self.assertEqual(target.read_text(), 'model="outside"')

    def test_added_subtable_blocks_undo(self):
        self.request("configure", expected=self.request()["revision"])
        with self.config.open("a", encoding="utf-8") as stream:
            stream.write('\n[mcp_servers.exa.tools.search]\nenabled=false\n')
        with self.assertRaises(ValueError):
            self.request("undo", expected=self.request()["revision"])

    def test_corrupt_journal_cannot_expand_rollback_boundary(self):
        self.config.write_text('model="personal"\n', encoding="utf-8")
        self.request("configure", expected=self.request()["revision"])
        journal = self.home / "chuli-integrations/exa.json"
        record = json.loads(journal.read_text(encoding="utf-8"))
        record["block"] = self.config.read_text(encoding="utf-8")
        journal.write_text(json.dumps(record), encoding="utf-8")
        with self.assertRaises(ValueError):
            self.request()
        self.assertIn('model="personal"', self.config.read_text(encoding="utf-8"))

    def test_windows_reparse_flag_is_rejected_without_is_junction_api(self):
        info = SimpleNamespace(st_mode=stat.S_IFDIR, st_file_attributes=0x400)
        with patch.object(Path, "lstat", return_value=info):
            with self.assertRaisesRegex(ValueError, "Ruta simbólica"):
                self.helper.safe_path(self.config)
