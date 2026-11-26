package manager

import (
	"context"
	"os"
	"path/filepath"
	"testing"
)

func TestHasGentleRoutingConflict(t *testing.T) {
	t.Run("detects real marker per client", func(t *testing.T) {
		cases := []struct {
			client string
			rel    string
		}{
			{client: "codex", rel: "AGENTS.md"},
			{client: "claude", rel: "CLAUDE.md"},
			{client: "agy", rel: filepath.Join("plugins", "chuli", "rules", "AGENTS.md")},
		}
		for _, tc := range cases {
			home := t.TempDir()
			target := filepath.Join(home, tc.rel)
			if err := os.MkdirAll(filepath.Dir(target), 0700); err != nil {
				t.Fatal(err)
			}
			content := "\xef\xbb\xbfPersonal\n<!-- gentle-ai:orchestrator -->\nRouting\n<!-- /gentle-ai:orchestrator -->\n"
			if err := os.WriteFile(target, []byte(content), 0600); err != nil {
				t.Fatal(err)
			}
			if !HasGentleRoutingConflict(tc.client, home) {
				t.Fatalf("expected conflict detected for %s", tc.client)
			}
		}
	})

	t.Run("ignores marker inside fenced code block", func(t *testing.T) {
		home := t.TempDir()
		content := "```markdown\n<!-- gentle-ai:agent-routing -->\nEjemplo\n<!-- /gentle-ai:agent-routing -->\n```\n"
		if err := os.WriteFile(filepath.Join(home, "AGENTS.md"), []byte(content), 0600); err != nil {
			t.Fatal(err)
		}
		if HasGentleRoutingConflict("codex", home) {
			t.Fatal("fenced code example must not trigger Gentle conflict")
		}
	})

	t.Run("returns false for clean or missing files", func(t *testing.T) {
		home := t.TempDir()
		if HasGentleRoutingConflict("codex", home) {
			t.Fatal("missing file must not report conflict")
		}
		if err := os.WriteFile(filepath.Join(home, "CLAUDE.md"), []byte("# Mis instrucciones\n"), 0600); err != nil {
			t.Fatal(err)
		}
		if HasGentleRoutingConflict("claude", home) {
			t.Fatal("clean file must not report conflict")
		}
	})
}

func TestCheckRuntimePreflightEmptyPath(t *testing.T) {
	t.Setenv("PATH", "")
	res := CheckRuntimePreflight(context.Background())
	if res.PowerShellOK || res.PythonOK {
		t.Fatalf("expected missing runtimes when PATH is empty: %+v", res)
	}
	if res.Detail == "" {
		t.Fatal("expected descriptive detail when runtimes are missing")
	}
}

func TestPresetDisplayName(t *testing.T) {
	if got := PresetDisplayName("codex", "custom"); got != "Personalizado por rol" {
		t.Fatalf("unexpected custom display name: %q", got)
	}
	if got := PresetDisplayName("codex", "codex-balanced"); got == "" || got == "codex-balanced" {
		t.Fatalf("expected friendly preset name, got %q", got)
	}
	if got := PresetDisplayName("codex", "unknown-preset"); got != "unknown-preset" {
		t.Fatalf("expected fallback to preset ID, got %q", got)
	}
}
