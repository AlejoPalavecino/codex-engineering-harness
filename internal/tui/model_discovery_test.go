package tui

import (
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/AlejoPalavecino/codex-engineering-harness/v2/internal/manager"
)

func TestPickerLoadsNewModelsWhileOpen(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "models_cache.json")
	if err := os.WriteFile(path, []byte(`{"models":[{"slug":"future-model","visibility":"list"}]}`), 0600); err != nil {
		t.Fatal(err)
	}
	m := model{mgr: manager.Manager{Target: "codex", Home: dir}, screen: home, width: 80, height: 24}
	m, refresh := sendKey(m, "m")
	if refresh == nil {
		t.Fatal("model editor did not request a catalog")
	}
	m, _ = sendKey(m, "enter")
	updated, _ := m.Update(refresh())
	m = updated.(model)
	if !strings.Contains(m.View().Content, "future-model") || !slices.Equal(m.pickerChoices, []string{"future-model", "gpt-6.2-astra"}) {
		t.Fatalf("expected discovered model and current assignment: %v", m.pickerChoices)
	}
	if m.mgr.RoleModels["orchestrator"].Model != "gpt-6.2-astra" {
		t.Fatal("refresh changed assigned model")
	}
	m, _ = sendKey(m, "up")
	m, _ = sendKey(m, "enter")
	if m.mgr.RoleModels["orchestrator"].Model != "future-model" {
		t.Fatal("selected model was lost")
	}
	updated, _ = m.Update(modelsLoadedMsg{request: m.modelRequest - 1, target: "codex", ids: []string{"stale"}})
	if slices.Contains(updated.(model).modelChoices, "stale") {
		t.Fatal("accepted outdated query")
	}
}

func TestAGYModelChoicesRespectAgentTiers(t *testing.T) {
	spec, _ := manager.FindClient("agy")
	m := model{modelPickerState: modelPickerState{modelChoices: []string{"new-agy-model"}}}
	if !slices.Contains(m.choicesForRole(spec, "orchestrator"), "new-agy-model") {
		t.Fatal("main agent cannot select discovered ID")
	}
	if !slices.Equal(m.choicesForRole(spec, "chuli_worker"), []string{"inherit", "pro", "flash", "flash_lite"}) {
		t.Fatal("expected plugin agent tiers")
	}
}

func TestAGYSelectionKeepsExactLabel(t *testing.T) {
	m := model{mgr: manager.Manager{Target: "agy"}, screen: modelsScreen, width: 80, height: 24}
	m.ensureModelState()
	updated, _ := m.Update(modelsLoadedMsg{request: m.modelRequest, target: "agy", ids: []string{"gemini-next-high"}, labels: map[string]string{"gemini-next-high": "Gemini Next (High)"}})
	m = updated.(model)
	m.openRolePicker()
	m, _ = sendKey(m, "up")
	m, _ = sendKey(m, "enter")
	selected := m.mgr.RoleModels["orchestrator"]
	if selected.Model != "gemini-next-high" || selected.Label != "Gemini Next (High)" {
		t.Fatalf("selection lost ID/label pair: %+v", selected)
	}
	m.cycleRoleEffort()
	if m.mgr.RoleModels["orchestrator"].Effort != selected.Effort {
		t.Fatal("dynamic AGY variant must not expose independent effort changes")
	}
	// A subsequent refresh may change available options, never the selected pair.
	updated, _ = m.Update(modelsLoadedMsg{request: m.modelRequest, target: "agy", ids: []string{"gemini-next-high"}, labels: map[string]string{"gemini-next-high": "Different Label"}})
	if updated.(model).mgr.RoleModels["orchestrator"].Label != selected.Label {
		t.Fatal("refresh changed selected label")
	}
}
