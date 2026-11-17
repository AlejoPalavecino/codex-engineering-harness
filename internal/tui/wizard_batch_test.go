package tui

import (
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/AlejoPalavecino/codex-engineering-harness/v2/internal/manager"
	"github.com/charmbracelet/x/ansi"
)

func batchTestManager(t *testing.T) manager.Manager {
	t.Helper()
	codex := canonicalTUITestHome(t)
	t.Setenv("CODEX_HOME", codex)
	t.Setenv("CLAUDE_CONFIG_DIR", canonicalTUITestHome(t))
	t.Setenv("AGY_HOME", canonicalTUITestHome(t))
	return manager.Manager{Home: codex, Target: "codex"}
}

func TestBatchFailureReportsAllPlatformsAndClearsStaleResult(t *testing.T) {
	mgr := batchTestManager(t)
	mgr.Source = t.TempDir() // Not a valid installer: fails before applying anything.
	tr := &progressTracker{ch: make(chan progressEvent, 16)}
	msg := runInstallBatchCmd(mgr, []string{"codex", "claude", "agy"}, nil, "local", true, false, false, tr)().(wizardBatchResultMsg)
	if msg.err == nil || len(msg.statuses) != 3 {
		t.Fatalf("missing failure/detection: %+v", msg)
	}
	for _, want := range []string{"Codex: fallido", "Claude: pendiente", "AGY: pendiente"} {
		if !strings.Contains(msg.summary, want) {
			t.Fatalf("missing %q: %s", want, msg.summary)
		}
	}
	m := model{mgr: mgr, screen: busy, width: 120, height: 50, resultSummary: "STALE SUCCESS", resultNext: "STALE NEXT"}
	next, _ := m.Update(msg)
	m = next.(model)
	view := ansi.Strip(m.renderCRT().Content)
	if !strings.Contains(view, "Claude: pendiente") || strings.Contains(view, "STALE") {
		t.Fatal(view)
	}
}

func TestBatchGentleRetryResumesRemainingPlatforms(t *testing.T) {
	if testing.Short() {
		t.Skip("runs installer against temporary profiles")
	}
	if _, err := exec.LookPath("pwsh"); err != nil {
		if _, err := exec.LookPath("powershell.exe"); err != nil {
			t.Skip("requires PowerShell")
		}
	}
	mgr := batchTestManager(t)
	root, err := filepath.Abs(filepath.Join("..", ".."))
	if err != nil {
		t.Fatal(err)
	}
	mgr.Source = root
	claudeHome := os.Getenv("CLAUDE_CONFIG_DIR")
	conflict := "<!-- gentle-ai:orchestrator -->\nGentle routes here.\n<!-- /gentle-ai:orchestrator -->\n"
	if err := os.WriteFile(filepath.Join(claudeHome, "CLAUDE.md"), []byte(conflict), 0600); err != nil {
		t.Fatal(err)
	}
	agyInstructions := filepath.Join(os.Getenv("AGY_HOME"), "plugins", "chuli", "rules", "AGENTS.md")
	if err := os.MkdirAll(filepath.Dir(agyInstructions), 0700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(agyInstructions, []byte(conflict), 0600); err != nil {
		t.Fatal(err)
	}
	tr := &progressTracker{ch: make(chan progressEvent, 32)}
	msg := runInstallBatchCmd(mgr, []string{"codex", "claude", "agy"}, nil, "local", true, false, false, tr)().(wizardBatchResultMsg)
	if !errors.Is(msg.err, manager.ErrGentleConflict) || msg.retry == nil {
		t.Fatalf("expected resumable Gentle conflict: %+v", msg)
	}
	for _, want := range []string{"Codex: completado", "Claude: pendiente (conflicto Gentle", "AGY: pendiente (sin ejecutar)"} {
		if !strings.Contains(msg.summary, want) {
			t.Fatalf("missing %q: %s", want, msg.summary)
		}
	}
	codexState := filepath.Join(mgr.Home, "harness-install", "state.json")
	before, err := os.Stat(codexState)
	if err != nil {
		t.Fatal(err)
	}
	m := model{mgr: mgr, screen: busy, width: 120, height: 50}
	next, _ := m.Update(msg)
	m = next.(model)
	view := ansi.Strip(m.renderCRT().Content)
	if !strings.Contains(view, "Codex: completado") || !strings.Contains(view, "AGY: pendiente") || strings.Contains(view, "Chuli se detuvo antes de escribir") {
		t.Fatal(view)
	}
	m, cmd := sendKey(m, "enter")
	if m.screen != busy || cmd == nil {
		t.Fatal("Gentle confirmation did not resume")
	}
	batch := cmd().(tea.BatchMsg)
	resumed := batch[0]().(wizardBatchResultMsg)
	if !errors.Is(resumed.err, manager.ErrGentleConflict) || resumed.retry == nil || !strings.Contains(resumed.summary, "Claude: completado") {
		t.Fatalf("Claude authorization must not adopt AGY's conflict: %+v", resumed)
	}
	agyBefore, err := os.ReadFile(agyInstructions)
	if err != nil || string(agyBefore) != conflict {
		t.Fatal("AGY conflict changed without separate authorization")
	}
	resumed = resumed.retry(&progressTracker{ch: make(chan progressEvent, 32)})().(wizardBatchResultMsg)
	if resumed.err != nil {
		t.Fatalf("retry failed: %v\n%s", resumed.err, resumed.techLog)
	}
	for _, name := range []string{"Codex", "Claude", "AGY"} {
		if !strings.Contains(resumed.summary, name) {
			t.Fatalf("missing completed %s: %s", name, resumed.summary)
		}
	}
	after, err := os.Stat(codexState)
	if err != nil || !after.ModTime().Equal(before.ModTime()) {
		t.Fatal("retry rewrote already completed Codex")
	}
	if strings.Count(resumed.techLog, "=== OpenAI Codex ===") != 1 {
		t.Fatal("completed client's log missing or duplicated")
	}
}

func TestUpdateSearchPreservesOperationLog(t *testing.T) {
	for _, tc := range []struct {
		name         string
		screen       string
		returnScreen string
	}{
		{name: "installation result", screen: result},
		{name: "Gentle conflict", screen: installGentleScreen},
		{name: "installation running", screen: busy},
		{name: "update search", screen: busy, returnScreen: updateScreen},
		{name: "update view", screen: updateScreen},
	} {
		t.Run(tc.name, func(t *testing.T) {
			const operationLog = "INSTALLER_TECHNICAL_LOG"
			const notes = "RELEASE_NOTES"
			m := model{screen: tc.screen, busyReturnScreen: tc.returnScreen, log: operationLog, resultSummary: "Installation outcome", showTechLog: true, width: 100, height: 30}
			next, _ := m.Update(updatesLoadedMsg{report: manager.UpdateReport{Summary: notes}})
			m = next.(model)
			if m.log != operationLog {
				t.Fatalf("background update replaced operation log: %q", m.log)
			}
			if m.updateReport.Summary != notes || !m.updatesChecked {
				t.Fatal("update report was not refreshed")
			}
			if tc.returnScreen != "" && m.screen != updateScreen {
				t.Fatalf("search did not return to update view: %s", m.screen)
			}
			if m.screen == updateScreen {
				if got := strings.Join(m.lines(), "\n"); !strings.Contains(got, notes) || strings.Contains(got, operationLog) {
					t.Fatalf("update view lost its own notes: %s", got)
				}
			}
			if m.screen == result {
				if got := ansi.Strip(m.renderCRT().Content); !strings.Contains(got, operationLog) || strings.Contains(got, notes) {
					t.Fatalf("technical result view lost operation log: %s", got)
				}
			}
		})
	}
}
