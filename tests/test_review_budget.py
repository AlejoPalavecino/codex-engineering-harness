"""Run the review-size shell step with the GitHub API boundary stubbed."""
import os
from pathlib import Path
import shutil
import subprocess
import textwrap
import unittest

ROOT = Path(__file__).resolve().parents[1]
BASH = (shutil.which("bash") if os.name != "nt" else
        shutil.which("bash", path=str(Path(os.environ["ProgramFiles"]) / "Git/bin")))


@unittest.skipUnless(BASH, "requires Bash")
class ReviewBudgetTests(unittest.TestCase):
    def run_step(self, response):
        workflow = (ROOT / ".github/workflows/validate.yml").read_text(encoding="utf-8")
        job = workflow.split("  review-size:", 1)[1].split("\n  structure:", 1)[0]
        script = textwrap.dedent(job.split("run: |\n", 1)[1])
        return subprocess.run(
            [BASH, "-c", 'gh() { echo "$FIXTURE"; }\n' + script],
            env={**os.environ, "SHA": "current", "PR": "1", "GITHUB_REPOSITORY": "test/repo",
                 "FIXTURE": response}, capture_output=True, text=True)

    def test_boundary_and_excess(self):
        for size in (400, 401):
            with self.subTest(size=size):
                result = self.run_step(f"current\t{size}")
                self.assertEqual(result.returncode, 0, result.stderr)
                self.assertEqual("::warning" in result.stdout, size > 400)

    def test_stale_or_incomplete_response_fails(self):
        for response in ("old\t401", "current", "current\tnull"):
            with self.subTest(response=response):
                self.assertNotEqual(self.run_step(response).returncode, 0)
