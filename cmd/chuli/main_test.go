package main

import (
	"bytes"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"runtime/debug"
	"strings"
	"testing"

	"github.com/AlejoPalavecino/codex-engineering-harness/internal/manager"
)

func TestEffectiveVersion(t *testing.T) {
	cases := []struct {
		name       string
		configured string
		info       *debug.BuildInfo
		ok         bool
		want       string
	}{
		{
			name:       "go install tagged module overrides -dev",
			configured: "1.1.2-dev",
			info:       &debug.BuildInfo{Main: debug.Module{Version: "v1.1.2"}},
			ok:         true,
			want:       "1.1.2",
		},
		{
			name:       "ldflags release version wins over build info",
			configured: "1.1.2",
			info:       &debug.BuildInfo{Main: debug.Module{Version: "(devel)"}},
			ok:         true,
			want:       "1.1.2",
		},
		{
			name:       "local devel build keeps -dev",
			configured: "1.1.2-dev",
			info:       &debug.BuildInfo{Main: debug.Module{Version: "(devel)"}},
			ok:         true,
			want:       "1.1.2-dev",
		},
		{
			name:       "dirty module version keeps -dev",
			configured: "1.1.2-dev",
			info:       &debug.BuildInfo{Main: debug.Module{Version: "v1.1.2+dirty"}},
			ok:         true,
			want:       "1.1.2-dev",
		},
		{
			name:       "pseudo-version keeps -dev",
			configured: "1.1.2-dev",
			info:       &debug.BuildInfo{Main: debug.Module{Version: "v1.1.2-0.20261116000000-abcdef123456"}},
			ok:         true,
			want:       "1.1.2-dev",
		},
		{
			name:       "missing build info keeps configured",
			configured: "1.1.2-dev",
			info:       nil,
			ok:         false,
			want:       "1.1.2-dev",
		},
	}
	for _, tc := range cases {
		if got := effectiveVersion(tc.configured, tc.info, tc.ok); got != tc.want {
			t.Errorf("%s: effectiveVersion() = %q, want %q", tc.name, got, tc.want)
		}
	}
}

func TestCLIAdoptionRequiresExplicitApply(t *testing.T) {
	if _, err := exec.LookPath("pwsh"); err != nil {
		if _, err = exec.LookPath("powershell.exe"); err != nil {
			t.Skip("requires PowerShell")
		}
	}
	home, err := filepath.EvalSymlinks(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	t.Setenv("CODEX_HOME", home)
	m, err := manager.New("../..")
	if err != nil {
		t.Fatal(err)
	}
	original := "Personal\n<!-- gentle-ai:orchestrator -->\nRouting\n<!-- /gentle-ai:orchestrator -->\n"
	path := filepath.Join(home, "AGENTS.md")
	if err = os.WriteFile(path, []byte(original), 0600); err != nil {
		t.Fatal(err)
	}
	var out bytes.Buffer
	if err = manage(m, []string{"install", "--apply"}, &out); !errors.Is(err, manager.ErrGentleConflict) {
		t.Fatal("default must refuse conflict", err)
	}
	if err = manage(m, []string{"install", "--adopt-gentle-routing"}, &out); err != nil {
		t.Fatal(err)
	}
	if got, err := os.ReadFile(path); err != nil || string(got) != original {
		t.Fatal("preview changed original")
	}
	if err = manage(m, []string{"update", "--apply", "--adopt-gentle-routing"}, &out); err != nil {
		t.Fatalf("%v\n%s", err, &out)
	}
	if got, err := os.ReadFile(path); err != nil || strings.Contains(string(got), "<!-- gentle-ai:orchestrator -->") || !strings.Contains(string(got), "Personal") {
		t.Fatal("expected adopted routing and preserved user text")
	}
	if err = manage(m, []string{"rollback", "--adopt-gentle-routing", "--apply"}, &out); err == nil {
		t.Fatal("accepted adoption option for rollback")
	}
}
