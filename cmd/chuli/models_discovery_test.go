package main

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func discoveryCLI(t *testing.T) string {
	t.Helper()
	if testing.Short() {
		t.Skip("builds and executes the CLI")
	}
	binary := filepath.Join(t.TempDir(), "chuli.exe")
	if output, err := exec.Command("go", "build", "-o", binary, ".").CombinedOutput(); err != nil {
		t.Fatalf("build: %v\n%s", err, output)
	}
	// Isolate both the client profile and the Manager's private cache.
	t.Setenv("APPDATA", t.TempDir())
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	return binary
}

func runDiscoveryCLI(t *testing.T, binary string, args ...string) string {
	t.Helper()
	output, err := exec.Command(binary, args...).CombinedOutput()
	if err != nil {
		t.Fatalf("chuli %v: %v\n%s", args, err, output)
	}
	return string(output)
}

func writeDiscoveryCache(t *testing.T, home, name, content string) {
	t.Helper()
	if err := os.WriteFile(filepath.Join(home, name), []byte(content), 0600); err != nil {
		t.Fatal(err)
	}
}

func listedDiscoveryModels(t *testing.T, output string) string {
	t.Helper()
	for _, line := range strings.Split(output, "\n") {
		if strings.HasPrefix(line, "Modelos: ") {
			return strings.TrimPrefix(line, "Modelos: ")
		}
	}
	t.Fatalf("missing model list: %s", output)
	return ""
}

func TestCLIModelsCodexDiscoveryAndRefresh(t *testing.T) {
	binary := discoveryCLI(t)
	home := t.TempDir()
	t.Setenv("CODEX_HOME", home)
	writeDiscoveryCache(t, home, "models_cache.json", `{"models":[
		{"slug":"synthetic-visible","visibility":"list"},
		{"slug":"synthetic-hidden","visibility":"hide"},
		{"slug":"synthetic-visible","visibility":"list"},
		{"slug":"synthetic-second","visibility":"list"}]}`)
	output := runDiscoveryCLI(t, binary, "models", "list")
	if got := listedDiscoveryModels(t, output); got != "synthetic-visible, synthetic-second" {
		t.Fatalf("visible models must be unique and omit hidden entries: %q", got)
	}
	if !strings.Contains(output, "Fuente: caché de Codex") {
		t.Fatalf("missing discovered source: %s", output)
	}
	writeDiscoveryCache(t, home, "models_cache.json", `{"models":[{"slug":"synthetic-refreshed","visibility":"list"}]}`)
	if got := listedDiscoveryModels(t, runDiscoveryCLI(t, binary, "models")); got != "synthetic-refreshed" {
		t.Fatalf("cache changes must be visible on the next invocation: %q", got)
	}

	t.Run("alternate home and desktop cache", func(t *testing.T) {
		alternate := t.TempDir()
		t.Setenv("CODEX_HOME", alternate)
		writeDiscoveryCache(t, alternate, "models_cache.codex_work_desktop.json", `{"models":[{"slug":"synthetic-desktop","visibility":"list"}]}`)
		if got := listedDiscoveryModels(t, runDiscoveryCLI(t, binary, "models")); got != "synthetic-desktop" {
			t.Fatalf("must use the selected config home and desktop cache: %q", got)
		}
	})
	t.Run("invalid cache uses bundled fallback", func(t *testing.T) {
		writeDiscoveryCache(t, home, "models_cache.json", `{"models":`)
		output := runDiscoveryCLI(t, binary, "models")
		models := listedDiscoveryModels(t, output)
		if !strings.Contains(output, "Fuente: catálogo incluido en Chuli") ||
			!strings.Contains(models, "gpt-6.2-sol") || !strings.Contains(models, "gpt-6.2-astra") {
			t.Fatalf("invalid cache must fall back to a catalog including 6.2: %s", output)
		}
	})
}

func TestCLIInstallDiscoveredCodexModel(t *testing.T) {
	if _, err := exec.LookPath("pwsh"); err != nil {
		if _, err := exec.LookPath("powershell.exe"); err != nil {
			t.Skip("requires PowerShell")
		}
	}
	binary := discoveryCLI(t)
	home := t.TempDir()
	t.Setenv("CODEX_HOME", home)
	cache := `{"models":[{"slug":"synthetic-discovered-model","visibility":"list"}]}`
	writeDiscoveryCache(t, home, "models_cache.json", cache)
	if got := listedDiscoveryModels(t, runDiscoveryCLI(t, binary, "models")); got != "synthetic-discovered-model" {
		t.Fatalf("synthetic model was not discovered: %q", got)
	}
	runDiscoveryCLI(t, binary, "--source", "../..", "install", "--apply",
		"--role", "orchestrator=synthetic-discovered-model:high",
		"--role", "chuli_worker=synthetic-discovered-model:medium")
	for _, tc := range []struct{ path, want string }{
		{"config.toml", `model = "synthetic-discovered-model"`},
		{"agents/chuli_worker.toml", `model = "synthetic-discovered-model"`},
	} {
		raw, err := os.ReadFile(filepath.Join(home, filepath.FromSlash(tc.path)))
		if err != nil || !strings.Contains(string(raw), tc.want) {
			t.Fatalf("%s must contain installed discovered model: %v\n%s", tc.path, err, raw)
		}
	}
	raw, err := os.ReadFile(filepath.Join(home, "harness-install", "state.json"))
	if err != nil {
		t.Fatal(err)
	}
	var state struct {
		RoleModels map[string]struct{ Model, Effort string } `json:"role_models"`
	}
	if err := json.Unmarshal(raw, &state); err != nil {
		t.Fatal(err)
	}
	for role, effort := range map[string]string{"orchestrator": "high", "chuli_worker": "medium"} {
		got := state.RoleModels[role]
		if got.Model != "synthetic-discovered-model" || got.Effort != effort {
			t.Fatalf("manifest lost selected model/effort for %s: %+v", role, got)
		}
	}
	raw, err = os.ReadFile(filepath.Join(home, "models_cache.json"))
	if err != nil || string(raw) != cache {
		t.Fatalf("installation changed the client model cache: %v\n%s", err, raw)
	}
}
