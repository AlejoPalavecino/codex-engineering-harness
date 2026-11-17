package main

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/AlejoPalavecino/codex-engineering-harness/v2/internal/manager"
)

func TestManageUpdatePreservesInstalledModelSelection(t *testing.T) {
	if testing.Short() {
		t.Skip("runs installer against a temporary Claude profile")
	}
	home, err := filepath.EvalSymlinks(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	t.Setenv("CLAUDE_CONFIG_DIR", home)
	m, err := manager.NewForClient("../..", "claude")
	if err != nil {
		t.Fatal(err)
	}
	var output bytes.Buffer
	if err := manage(m, []string{"install", "--apply", "--preset", "claude-fast", "--role", "orchestrator=claude-fable-5-2:high", "--role", "chuli_worker=claude-fable-5-2:high"}, &output); err != nil {
		t.Fatalf("install: %v", err)
	}
	type modelState struct {
		Preset string                       `json:"preset"`
		Roles  map[string]manager.RoleModel `json:"role_models"`
	}
	readState := func() modelState {
		t.Helper()
		raw, err := os.ReadFile(filepath.Join(home, "harness-install", "state.json"))
		if err != nil {
			t.Fatal(err)
		}
		var state modelState
		if err := json.Unmarshal(raw, &state); err != nil {
			t.Fatal(err)
		}
		return state
	}
	before := readState()
	if before.Preset != "claude-fast" || before.Roles["orchestrator"].Model != "claude-fable-5-2" {
		t.Fatalf("unexpected custom installation: %+v", before)
	}
	settingsPath := filepath.Join(home, "settings.json")
	settings := []byte(`{"model":"claude-sonnet-5-5","env":{"PERSONAL":"keep"},"personal":true}`)
	if err := os.WriteFile(settingsPath, settings, 0600); err != nil {
		t.Fatal(err)
	}
	agentPath := filepath.Join(home, "agents", "chuli_worker.md")
	agentBefore, err := os.ReadFile(agentPath)
	if err != nil {
		t.Fatal(err)
	}
	stateBefore, err := os.Stat(filepath.Join(home, "harness-install", "state.json"))
	if err != nil {
		t.Fatal(err)
	}
	output.Reset()
	if err := manage(m, []string{"update", "--apply"}, &output); err != nil {
		t.Fatalf("update: %v", err)
	}
	after := readState()
	if !reflect.DeepEqual(before, after) {
		t.Fatalf("update without selection changed model assignment: before=%+v after=%+v", before, after)
	}
	stateAfter, err := os.Stat(filepath.Join(home, "harness-install", "state.json"))
	if err != nil || stateAfter.ModTime().Equal(stateBefore.ModTime()) {
		t.Fatal("update did not apply a new manifest")
	}
	checkPreserved := func() {
		t.Helper()
		agentAfter, err := os.ReadFile(agentPath)
		if err != nil || !bytes.Equal(agentBefore, agentAfter) {
			t.Fatal("update changed the selected specialist")
		}
		raw, err := os.ReadFile(settingsPath)
		if err != nil {
			t.Fatal(err)
		}
		var settings struct {
			Model    string
			Env      map[string]string
			Personal bool
		}
		if err := json.Unmarshal(raw, &settings); err != nil {
			t.Fatal(err)
		}
		if settings.Model != "claude-sonnet-5-5" || settings.Env["PERSONAL"] != "keep" || !settings.Personal {
			t.Fatalf("update overwrote personal settings: %s", raw)
		}
		if settings.Env["CLAUDE_CODE_EFFORT_LEVEL"] != "high" || settings.Env["CLAUDE_CODE_SUBAGENT_MODEL"] != "" {
			t.Fatalf("missing settings were not restored from installed roles: %s", raw)
		}
	}
	checkPreserved()
	output.Reset()
	if err := manage(m, []string{"update", "--apply"}, &output); err != nil {
		t.Fatalf("repeat update: %v", err)
	}
	if !reflect.DeepEqual(before, readState()) {
		t.Fatal("repeat update changed installed selection")
	}
	checkPreserved()
	output.Reset()
	if err := manage(m, []string{"update", "--apply", "--preset", "claude-balanced"}, &output); err != nil {
		t.Fatalf("explicit update: %v", err)
	}
	explicit := readState()
	if explicit.Preset != "claude-balanced" || reflect.DeepEqual(explicit.Roles, before.Roles) {
		t.Fatalf("explicit selection did not replace custom models: %+v", explicit)
	}
	output.Reset()
	if err := manage(m, []string{"rollback", "--apply"}, &output); err != nil {
		t.Fatalf("rollback: %v", err)
	}
	if !reflect.DeepEqual(before, readState()) {
		t.Fatal("rollback lost the previous model selection")
	}
	checkPreserved()
}

func TestManageDefaultCodexUpdateKeepsCanonicalAgents(t *testing.T) {
	if testing.Short() {
		t.Skip("runs installer against a temporary Codex profile")
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
	var output bytes.Buffer
	if err := manage(m, []string{"install", "--apply"}, &output); err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(home, "agents", "chuli_worker.toml")
	original, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	personalPath := filepath.Join(home, "agents", "quality_reviewer.toml")
	personal, err := os.ReadFile(personalPath)
	if err != nil {
		t.Fatal(err)
	}
	personal = append(personal, []byte("\n# Personal customization\n")...)
	if err := os.WriteFile(personalPath, personal, 0600); err != nil {
		t.Fatal(err)
	}
	output.Reset()
	if err := manage(m, []string{"update", "--apply"}, &output); err != nil {
		t.Fatal(err)
	}
	updated, err := os.ReadFile(path)
	if err != nil || !bytes.Equal(original, updated) {
		t.Fatal("default repeat update changed canonical agent bytes")
	}
	updatedPersonal, err := os.ReadFile(personalPath)
	if err != nil || !bytes.Equal(personal, updatedPersonal) {
		t.Fatal("repeat update overwrote customized agent")
	}
	if strings.Contains(output.String(), "UPDATE               agents/chuli_worker.toml") {
		t.Fatal("repeat update rewrote canonical agent")
	}
}
