"""Replay the Git/GitHub delivery scenarios locally with three installed profiles.

The fixtures assert local state and draft contents. They do not launch the
clients, resolve credentials, or contact GitHub.
"""

import argparse
import hashlib
import json
import os
from pathlib import Path
import subprocess
import tempfile


PROTOCOL = "harness/docs/protocols/GIT-GITHUB.md"
GIT_ENV = {
    name: os.environ[name]
    for name in ("PATH", "SYSTEMROOT", "WINDIR", "COMSPEC", "PATHEXT", "TEMP", "TMP")
    if name in os.environ
}
GIT_ENV.update(
    GIT_CONFIG_NOSYSTEM="1",
    GIT_CONFIG_GLOBAL=os.devnull,
    GIT_TERMINAL_PROMPT="0",
)


class Fixture:
    def __init__(self, root, client, name):
        self.cwd = root / client / name
        self.cwd.mkdir(parents=True)
        self.git("init", "-b", "trunk")
        self.git("config", "user.name", "Synthetic Fixture")
        self.git("config", "user.email", "fixture@example.invalid")
        self.git("config", "core.autocrlf", "false")
        self.write("widget.txt", "syntax errror\n")
        self.write("notes.txt", "original fixture\n")
        self.write("CONTRIBUTING.md", "Branch base: trunk\nIssue optional for text fixes\nFeatures: topic/NNN-name branch; TASK NNN commit prefix\n")
        self.write("package-lock.json", "")
        self.git("add", ".")
        self.git("commit", "-m", "Fixture base")
        self.base = self.git("rev-parse", "HEAD")

    def git(self, *args):
        return subprocess.check_output(["git", *args], cwd=self.cwd, env=GIT_ENV, text=True, encoding="utf-8", stderr=subprocess.STDOUT).strip()

    def write(self, path, content):
        dest = self.cwd / path
        dest.parent.mkdir(exist_ok=True, parents=True)
        dest.write_text(content, encoding="utf-8", newline="\n")

    def commit(self, msg):
        self.git("add", ".")
        self.git("commit", "-m", msg)
        return self.git("rev-parse", "HEAD")

    def state(self):
        return (self.git("status", "--porcelain", "-b"), self.git("diff", "--binary"), self.git("rev-parse", "HEAD"), self.git("for-each-ref"), self.git("remote", "-v"))

def r0(f):
    f.write("notes.txt", "edit inherited from user\n")
    before = f.state()
    diagnosis = f.git("show", "HEAD:widget.txt")
    assert "errror" in diagnosis and " M notes.txt" in before[0]
    assert f.state() == before
    return "diagnosis, branch/ref/head/diff unchanged, no remote"

def r1(f):
    policy = (f.cwd / "CONTRIBUTING.md").read_text()
    assert "Issue optional for text fixes" in policy
    f.git("switch", "-c", "fix/widget-copy", "trunk")
    f.write("widget.txt", "syntax error\n")
    assert (f.cwd / "widget.txt").read_text().strip() == "syntax error"
    assert f.git("diff", "--name-only", "trunk") == "widget.txt"
    assert f.git("merge-base", "trunk", "HEAD") == f.base
    assert not (f.cwd / "odd").exists() and not f.git("remote", "-v")
    return "fix/widget-copy from trunk, one-file diff, focused test, no issue or board"

def r2(f):
    f.write("issues-local.json", json.dumps([{"number": 3, "state": "closed", "duplicate_of": 7}, {"number": 7, "state": "open", "approved": True}]))
    f.git("add", "issues-local.json")
    f.git("commit", "-m", "Add synthetic issue snapshot")
    issues = json.loads((f.cwd / "issues-local.json").read_text())
    assert issues[0]["duplicate_of"] == issues[1]["number"] and issues[1]["approved"]
    f.git("switch", "-c", "topic/007-widget", "trunk")
    f.write("src/widget.txt", "behavior enabled\n")
    f.write("tests/widget.txt", "behavior enabled\n")
    f.write("docs/widget.md", "Widget behavior is enabled.\n")
    assert (f.cwd / "src/widget.txt").read_bytes() == (f.cwd / "tests/widget.txt").read_bytes()
    f.commit("TASK 007 Enable widget")
    assert set(f.git("show", "--format=", "--name-only", "HEAD").splitlines()) == {"src/widget.txt", "tests/widget.txt", "docs/widget.md"}
    assert f.git("log", "-1", "--format=%s") == "TASK 007 Enable widget"
    assert not f.git("remote", "-v")
    return "reuse #7 via local snapshot; topic/007-widget; TASK 007 embeds code/test/doc"

def review(f):
    f.git("switch", "-c", "review/widget")
    f.write("src/widget.txt", "".join(f"author entry {i}\n" for i in range(120)))
    f.write("package-lock.json", "".join(f"generated entry {i}\n" for i in range(700)))
    f.commit("Review fixture diff")
    stats = dict((path, int(a) + int(d)) for a, d, path in (row.split("\t") for row in f.git("diff", "--numstat", "trunk", "HEAD").splitlines()))
    assert stats == {"package-lock.json": 700, "src/widget.txt": 120}
    assert sum(stats.values()) == 820 and stats["src/widget.txt"] < 400
    in_policy = "one cohesive PR; repo warning for total lines; explain generated lockfile"
    return f"{stats['src/widget.txt']} authored; {stats['package-lock.json']} generated; {in_policy}"

def chain(f):
    f.git("branch", "feature/007-widget", "trunk")
    f.git("switch", "-c", "feature/007-model", "feature/007-widget")
    f.write("src/model.txt", "".join(f"model {i}\n" for i in range(230)))
    f.write("tests/model.txt", "model 0\n")
    a = f.commit("TASK 007 Add model")
    f.git("switch", "-c", "feature/007-use", "feature/007-model")
    f.write("src/use.txt", "".join(f"use model {i}\n" for i in range(230)))
    f.write("tests/use.txt", "use model 0\n")
    b = f.commit("TASK 007 Use model")
    f.write("draft.md", "1. feature/007-model -> feature/007-widget (Refs #7)\n2. feature/007-use -> feature/007-model (Refs #7)\n3. feature/007-widget -> trunk after integration (Closes #7 if approved and completed)\nPublish pending; GitHub unavailable.\n")
    assert f.git("merge-base", "feature/007-widget", "feature/007-model") == f.base
    assert f.git("merge-base", "feature/007-model", "feature/007-use") == a and b != a
    assert set(f.git("diff", "--name-only", "feature/007-model", "feature/007-use").splitlines()) == {"src/use.txt", "tests/use.txt"}
    assert not f.git("remote", "-v")
    return "local draft with 2 ordered partial refs, immediate bases and conditional closing; no GitHub"

def resume(f):
    f.git("switch", "-c", "topic/007-widget")
    f.write("widget.txt", "syntax error\n")
    old = f.commit("TASK 007 Fix fixture")
    f.write("board.json", json.dumps({"issue": 7, "branch": "topic/007-widget", "pr": 9, "check": {"head": old, "status": "pass"}}))
    f.write("widget.txt", "syntax error fixed\n")
    f.git("add", "widget.txt")
    f.git("commit", "-m", "TASK 007 Update fixture")
    board = json.loads((f.cwd / "board.json").read_text())
    head = f.git("rev-parse", "HEAD")
    ci_current = board["check"]["head"] == head and board["check"]["status"] == "pass"
    assert f.git("branch", "--show-current") == board["branch"] and not ci_current
    return "local board reconciled with fresh head; earlier green SHA is stale; CI pending"

def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--codex-home", type=Path, required=True)
    parser.add_argument("--claude-home", type=Path, required=True)
    parser.add_argument("--agy-home", type=Path, required=True)
    options = parser.parse_args()
    profiles = {
        "codex": (options.codex_home, "AGENTS.md"),
        "claude": (options.claude_home, "CLAUDE.md"),
        "agy": (options.agy_home, "plugins/chuli/rules/AGENTS.md"),
    }
    scenarios = {
        "R0": r0,
        "R1": r1,
        "R2": r2,
        "budget": review,
        "offline-chain": chain,
        "resume": resume,
    }
    protocols = {}
    results = {}
    with tempfile.TemporaryDirectory(prefix="chuli-git-fixtures-") as workspace:
        for client, (home, entrypoint) in profiles.items():
            assert PROTOCOL in (home / entrypoint).read_text(encoding="utf-8-sig"), client
            protocols[client] = hashlib.sha256((home / PROTOCOL).read_bytes()).hexdigest()
            results[client] = {
                name: scenario(Fixture(Path(workspace), client, name))
                for name, scenario in scenarios.items()
            }
    assert len(set(protocols.values())) == 1, protocols
    assert results["codex"] == results["claude"] == results["agy"], results
    print(json.dumps({"protocols": protocols, "fixture_results": results}, indent=2))


if __name__ == "__main__":
    main()
