package tui

import (
	"charm.land/lipgloss/v2"
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/AlejoPalavecino/codex-engineering-harness/internal/integrations"
	"github.com/AlejoPalavecino/codex-engineering-harness/internal/manager"
	"github.com/charmbracelet/x/ansi"
)

func integrationKey(m model, key string) (model, tea.Cmd) {
	code := rune(key[0])
	if key == "enter" {
		code = tea.KeyEnter
	}
	if key == "esc" {
		code = tea.KeyEscape
	}
	next, cmd := m.Update(tea.KeyPressMsg{Code: code, Text: func() string {
		if len(key) == 1 {
			return key
		}
		return ""
	}()})
	return next.(model), cmd
}

func TestIntegrationNavigationAndCancelDoNotWrite(t *testing.T) {
	m := model{mgr: manager.Manager{Home: canonicalTUITestHome(t)}, screen: home, height: 30}
	for index, item := range choices(m.mgr) {
		if item.action == "integrations" {
			m.cursor = index
		}
	}
	m, cmd := integrationKey(m, "enter")
	if cmd != nil || m.screen != integrationList {
		t.Fatal("catalog must open without action")
	}
	for _, name := range []string{"Context7", "Engram", "Exa", "GitHub", "Egolite"} {
		if !strings.Contains(m.View().Content, name) {
			t.Fatal(name)
		}
	}
	m, cmd = integrationKey(m, "enter")
	if cmd == nil || m.screen != busy {
		t.Fatal("detail should inspect asynchronously")
	}
	next, _ := m.Update(cmd())
	m = next.(model)
	if m.screen != integrationDetail {
		t.Fatal(m.screen, m.fault)
	}
	m, cmd = integrationKey(m, "c")
	if cmd == nil {
		t.Fatal("missing preview")
	}
	next, _ = m.Update(cmd())
	m = next.(model)
	if m.screen != integrationPreview {
		t.Fatal(m.screen, m.fault)
	}
	m, cmd = integrationKey(m, "n")
	if cmd != nil || m.screen != integrationDetail {
		t.Fatal("cancel applied")
	}
	if _, err := os.Stat(filepath.Join(m.mgr.Home, "config.toml")); !os.IsNotExist(err) {
		t.Fatal("preview/cancel wrote configuration")
	}
	m, _ = integrationKey(m, "esc")
	if m.screen != integrationList {
		t.Fatal("Esc should return one level")
	}
	m, _ = integrationKey(m, "esc")
	if m.screen != home {
		t.Fatal("Esc should return home")
	}
}

func TestIntegrationConfirmApplyAndUndo(t *testing.T) {
	homeDir := canonicalTUITestHome(t)
	service := integrations.Service{Home: homeDir}
	plan, err := service.Prepare(context.Background(), "exa", "configure")
	if err != nil {
		t.Fatal(err)
	}
	m := model{mgr: manager.Manager{Home: homeDir}, screen: integrationPreview, integrationID: "exa", integrationPlan: plan}
	m, cmd := integrationKey(m, "s")
	if cmd == nil || m.screen != busy {
		t.Fatal("explicit confirmation did not apply")
	}
	next, _ := m.Update(cmd())
	m = next.(model)
	if m.screen != integrationResult || m.fault != "" {
		t.Fatal(m.screen, m.fault)
	}
	state, err := service.Inspect(context.Background(), "exa")
	if err != nil || !state.Managed {
		t.Fatal(state, err)
	}
	m, cmd = integrationKey(m, "enter")
	next, _ = m.Update(cmd())
	m = next.(model)
	m, cmd = integrationKey(m, "d")
	next, _ = m.Update(cmd())
	m = next.(model)
	if m.screen != integrationPreview {
		t.Fatal("undo must have preview")
	}
	m, cmd = integrationKey(m, "s")
	next, _ = m.Update(cmd())
	m = next.(model)
	if m.fault != "" {
		t.Fatal(m.fault)
	}
	state, err = service.Inspect(context.Background(), "exa")
	if err != nil || state.Exists {
		t.Fatal(state, err)
	}
}

func TestIntegrationTextRemainsScrollableInNarrowTerminal(t *testing.T) {
	m := model{screen: integrationDetail, integrationID: "github", width: 42, height: 24, log: strings.Repeat("contenido legible ", 50)}
	for _, line := range m.lines() {
		if lipgloss.Width(line) > 40 {
			t.Fatal("text exceeds content width", line)
		}
	}
	before := m.page()
	m, cmd := integrationKey(m, "j")
	if cmd != nil || m.offset != 1 || m.page() == before {
		t.Fatal("long text cannot be navigated")
	}
	view := ansi.Strip(m.View().Content)
	rows := strings.Split(view, "\n")
	for i := range rows {
		rows[i] = strings.TrimRight(rows[i], " ")
	}
	view = strings.Join(rows, "\n")
	if !strings.Contains(view, "O abrir sitio oficial\nEnter revisar ficha\nEsc volver") {
		t.Fatalf("controls should wrap for narrow terminal: %q", view)
	}
}
