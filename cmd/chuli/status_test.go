package main

import (
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestCLIStatusLabelsSelectedClientHome(t *testing.T) {
	binary := filepath.Join(t.TempDir(), "chuli.exe")
	if output, err := exec.Command("go", "build", "-o", binary, ".").CombinedOutput(); err != nil {
		t.Fatalf("build: %v\n%s", err, output)
	}
	for _, tc := range []struct{ client, env string }{
		{"codex", "CODEX_HOME"},
		{"claude", "CLAUDE_CONFIG_DIR"},
		{"agy", "AGY_HOME"},
	} {
		t.Run(tc.client, func(t *testing.T) {
			home := t.TempDir()
			t.Setenv(tc.env, home)
			output, err := exec.Command(binary, "--client", tc.client, "status").CombinedOutput()
			if err != nil {
				t.Fatalf("status: %v\n%s", err, output)
			}
			if !strings.Contains(string(output), tc.env+": "+home+"\n") {
				t.Fatalf("missing selected client home label %s: %s", tc.env, output)
			}
			if tc.client != "codex" && strings.Contains(string(output), "CODEX_HOME:") {
				t.Fatalf("incorrect Codex home label: %s", output)
			}
		})
	}
}
