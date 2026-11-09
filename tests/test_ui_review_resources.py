"""Portable design/review references survive the installed profile layout."""
import importlib.util
from pathlib import Path
import re
import shutil
import tempfile
import unittest

ROOT = Path(__file__).resolve().parents[1]
spec = importlib.util.spec_from_file_location("ui_review_profile", ROOT / "scripts/profile_resources.py")
profile = importlib.util.module_from_spec(spec)
spec.loader.exec_module(profile)


class UiReviewResources(unittest.TestCase):
    def test_references_and_links_survive_install_layout(self):
        resources = profile.inventory(ROOT)
        references = (
            "skills/harness-ux-ui/references/task-design.md",
            "skills/harness-tdd-qa/references/browser-review.md",
        )
        for relative in references:
            self.assertIn(relative, resources, f"Missing distributed resource: {relative}")
        with tempfile.TemporaryDirectory(prefix="chuli-ui-resources-") as directory:
            home = Path(directory)
            for destination, source in resources.items():
                target = home / destination
                target.parent.mkdir(parents=True, exist_ok=True)
                shutil.copyfile(ROOT / source, target)
            documents = (*references, "skills/harness-ux-ui/SKILL.md", "skills/harness-tdd-qa/SKILL.md")
            for relative in documents:
                path = home / relative
                links = re.findall(r"\[[^\]]*\]\(([^)]+)\)", path.read_text(encoding="utf-8"))
                self.assertTrue(links, f"No portable references in {relative}")
                for link in links:
                    if "://" not in link and not link.startswith("#"):
                        target = (path.parent / link.split("#")[0]).resolve()
                        self.assertTrue(target.is_relative_to(home.resolve()))
                        self.assertTrue(target.is_file(), f"Broken installed link: {relative}: {link}")
