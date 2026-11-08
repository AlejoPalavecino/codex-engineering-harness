"""Profile inventory, ownership boundaries and instruction merge contracts."""
import json
import os
from pathlib import Path
import sys
import tempfile
import unittest

sys.path.insert(0, str(Path(__file__).resolve().parents[1] / "scripts"))
import profile_resources as profile


class ProfileTests(unittest.TestCase):
    def setUp(self):
        temp = tempfile.TemporaryDirectory(prefix="chuli-profile-test-")
        self.addCleanup(temp.cleanup)
        self.home = Path(temp.name).resolve() / "home"
        self.home.mkdir()
        self.output = self.home.parent / "candidate"

    def test_inventory_covers_every_distributed_file(self):
        resources = profile.preflight(self.home, "install")
        for directory in ("agents", "skills", "docs/protocols"):
            expected = {p.relative_to(profile.ROOT).as_posix() for p in (profile.ROOT / directory).rglob("*") if p.is_file()}
            self.assertEqual(expected, {s for s in resources.values() if s.startswith(directory + "/")})
        self.assertIn("harness/THIRD_PARTY_NOTICES.md", resources)

    def test_active_gentle_routing_and_override_report_source_before_writes(self):
        instructions = self.home / "AGENTS.md"
        for marker in ("sdd-orchestrator", "agent-routing", "orchestrator"):
            instructions.write_text(f"<!-- gentle-ai:{marker} -->\nbody\n<!-- /gentle-ai:{marker} -->")
            with self.subTest(marker=marker), self.assertRaisesRegex(ValueError, r"Gentle AI instructions .*AGENTS.md"):
                profile.preflight(self.home, "install")
        instructions.write_text("<!-- gentle-ai:engram-protocol -->\nmemory\n<!-- /gentle-ai:engram-protocol -->")
        self.assertIn("AGENTS.md", profile.preflight(self.home, "install"))
        (self.home / "AGENTS.override.md").write_text("my own global instructions")
        with self.assertRaisesRegex(ValueError, "AGENTS.override.md"):
            profile.preflight(self.home, "install")

    def test_incomplete_or_duplicate_blocks_are_rejected(self):
        for data in (profile.END, profile.BEGIN, profile.END + b"\n" + profile.BEGIN,
                     profile.BEGIN + b"\n" + profile.BEGIN + b"\n" + profile.END):
            with self.subTest(data=data), self.assertRaises(ValueError):
                profile.block_span(data)

    def test_preserves_surrounding_bytes_and_customized_block_across_updates(self):
        path = self.home / "AGENTS.md"
        personal = profile.BOM + "Prepará el mate\r\n".encode()
        path.write_bytes(personal)
        report = profile.prepare_instructions(self.home, self.output, False)
        self.assertTrue(self.output.read_bytes().startswith(personal))
        installed = self.output.read_bytes().replace(profile.BEGIN + b"\n", profile.BEGIN + b"\nmy customized rules\n")
        path.write_bytes(installed + b"added personal note\n")
        state = self.home / "harness-install/state.json"
        state.parent.mkdir()
        for _ in range(2):
            state.write_text(json.dumps({"orchestrator_sha256": report["orchestrator_sha256"],
                                         "orchestrator_baseline_sha256": report["baseline_sha256"]}))
            report = profile.prepare_instructions(self.home, self.output, False)
            self.assertTrue(report["preserved"])
            self.assertEqual(self.output.read_bytes(), path.read_bytes())
        profile.prepare_instructions(self.home, self.output, True)
        self.assertNotIn(b"my customized rules", self.output.read_bytes())
        self.assertTrue(self.output.read_bytes().endswith(b"added personal note\n"))
        self.assertTrue(self.output.read_bytes().startswith(personal))

    def test_rejects_out_of_scope_manifest_targets_and_backups(self):
        manifest = self.home / "harness-install/state.json"
        manifest.parent.mkdir()
        for target in ("../auth.json", "auth.json", "skills/../auth.json", r"agents\..\auth.json",
                       "config.toml:stream", "/abs", "harness-install/state.json"):
            manifest.write_text(json.dumps({"version":"1.0.0", "files":[{"path":target, "existed":False}]}))
            with self.subTest(target=target), self.assertRaisesRegex(ValueError, "rollback target"):
                profile.preflight(self.home, "rollback")
        manifest.write_text(json.dumps({"version":"1.0.0","files":[],"backup_dir":str(self.home.parent)}))
        with self.assertRaisesRegex(ValueError, "backup belonging"):
            profile.preflight(self.home, "rollback")

    def test_rejects_linked_parent_leaf_and_absent_link_target(self):
        external = self.home.parent / "external"
        external.mkdir()
        link = self.home / "skills"
        try:
            link.symlink_to(external, target_is_directory=True)
        except OSError as exc:
            self.skipTest(f"Symlinks unavailable: {exc}")
        with self.assertRaisesRegex(ValueError, "link|reparse"):
            profile.preflight(self.home, "install")
        link.unlink()
        (self.home / "AGENTS.md").symlink_to(external / "missing")
        with self.assertRaisesRegex(ValueError, "link|reparse"):
            profile.preflight(self.home, "install")
        self.assertEqual(list(external.iterdir()), [])

    def test_rejects_hardlinked_config(self):
        external = self.home.parent / "private-config"
        external.write_bytes(b"private")
        os.link(external, self.home / "config.toml")
        with self.assertRaisesRegex(ValueError, "one link"):
            profile.preflight(self.home, "install")
