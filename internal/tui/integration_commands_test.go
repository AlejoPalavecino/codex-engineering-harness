package tui

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/AlejoPalavecino/codex-engineering-harness/internal/manager"
)

func TestIntegrationCommandPreparesWithoutWriting(t *testing.T) {
	m := model{mgr: manager.Manager{Home: canonicalTUITestHome(t)}, integrationID: "exa"}
	message := m.integrationCmd("configure")().(integrationMsg)
	if message.err != nil || message.screen != integrationPreview || message.plan.Revision == "" {
		t.Fatal(message)
	}
	if _, err := os.Stat(filepath.Join(m.mgr.Home, "config.toml")); !os.IsNotExist(err) {
		t.Fatal("preparation wrote a configuration")
	}
}

func canonicalTUITestHome(t *testing.T) string {
	t.Helper()
	home, err := filepath.EvalSymlinks(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	return home
}
