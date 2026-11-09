package tui

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/AlejoPalavecino/codex-engineering-harness/internal/manager"
)

func TestIntegrationCommandPreparesWithoutWriting(t *testing.T) {
	m := model{mgr: manager.Manager{Home: t.TempDir()}, integrationID: "exa"}
	message := m.integrationCmd("configure")().(integrationMsg)
	if message.err != nil || message.screen != integrationPreview || message.plan.Revision == "" {
		t.Fatal(message)
	}
	if _, err := os.Stat(filepath.Join(m.mgr.Home, "config.toml")); !os.IsNotExist(err) {
		t.Fatal("preparation wrote a configuration")
	}
}
