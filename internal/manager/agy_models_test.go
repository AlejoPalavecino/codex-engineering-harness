package manager

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"runtime"
	"strings"
	"testing"
	"time"
)

func TestDiscoverAGYModels(t *testing.T) {
	tests := []struct {
		name string
		raw  string
		err  error
		want []string
	}{
		{"official envelope", `{"status":"SUCCESS","command":{"name":"models","data":{"models":[{"id":"gemini-3.7-flash-high","label":"Gemini 3.7 Flash (High)"},{"id":"gemini-3.7-flash-high","label":"Duplicate"},{"id":"gemini-4-carbon","label":"Gemini 4 Carbon"}]}}}`, nil, []string{"gemini-3.7-flash-high", "gemini-4-carbon"}},
		{"unsafe IDs filtered", `{"status":"SUCCESS","command":{"name":"models","data":{"models":[{"id":"../../escape","label":"Escape"},{"id":"$(bad)","label":"Bad"},{"id":"model\ncommand","label":"Newline"},{"id":"safe-model","label":"Safe Model"}]}}}`, nil, []string{"safe-model"}},
		{"missing labels filtered", `{"status":"SUCCESS","command":{"name":"models","data":{"models":[{"id":"missing"},{"id":"unsafe","label":"Bad\nInstruction"}]}}}`, nil, nil},
		{"invalid JSON", `not json`, nil, nil},
		{"trailing JSON", `{"status":"SUCCESS","command":{"name":"models","data":{"models":[{"id":"safe"}]}}} {}`, nil, nil},
		{"wrong schema", `["gemini-4-carbon"]`, nil, nil},
		{"wrong item type", `{"status":"SUCCESS","command":{"name":"models","data":{"models":["safe"]}}}`, nil, nil},
		{"failure status", `{"status":"ERROR","command":{"name":"models","data":{"models":[{"id":"safe"}]}}}`, nil, nil},
		{"other command", `{"status":"SUCCESS","command":{"name":"chat","data":{"models":[{"id":"safe"}]}}}`, nil, nil},
		{"empty response", `{}`, nil, nil},
		{"command failure", ``, errors.New("exit 1"), nil},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			run := func(ctx context.Context, name string, args ...string) ([]byte, error) {
				if name != "agy" || !reflect.DeepEqual(args, []string{"--output-format", "json", "models"}) {
					t.Fatalf("unexpected invocation: %q %q", name, args)
				}
				deadline, ok := ctx.Deadline()
				if !ok || time.Until(deadline) > 5*time.Second {
					t.Fatal("discovery must have a bounded context")
				}
				return []byte(tt.raw), tt.err
			}
			if got := discoverAGYModelsUsing(t.Context(), run); !reflect.DeepEqual(got, tt.want) {
				t.Fatalf("got %q, want %q", got, tt.want)
			}
		})
	}
}

func TestDiscoverAGYModelsCancelled(t *testing.T) {
	ctx, cancel := context.WithCancel(t.Context())
	cancel()
	got := discoverAGYModelsUsing(ctx, func(ctx context.Context, _ string, _ ...string) ([]byte, error) {
		return []byte(`{"status":"SUCCESS","command":{"name":"models","data":{"models":[{"id":"safe"}]}}}`), nil
	})
	if got != nil {
		t.Fatalf("cancelled discovery returned %q", got)
	}
}

func TestAGYLabelInstallationSnapshot(t *testing.T) {
	if _, err := shell(); err != nil {
		t.Skip("requires PowerShell and Python")
	}
	repo, err := filepath.Abs("../..")
	if err != nil {
		t.Fatal(err)
	}
	bin := t.TempDir()
	fixture := filepath.Join(bin, "catalog.json")
	writeCatalog := func(label string) {
		t.Helper()
		raw := `{"status":"SUCCESS","command":{"name":"models","data":{"models":[{"id":"gemini-next-high","label":"` + label + `"}]}}}`
		if err := os.WriteFile(fixture, []byte(raw), 0600); err != nil {
			t.Fatal(err)
		}
	}
	writeCatalog("Gemini Next (High)")
	// A real executable validates argv and reads only this test's explicit fixture.
	source := filepath.Join(bin, "main.go")
	program := `package main
import ("os"; "reflect")
func main() {
 if !reflect.DeepEqual(os.Args[1:], []string{"--output-format","json","models"}) { os.Exit(2) }
 b, err := os.ReadFile(os.Getenv("CHULI_TEST_AGY_CATALOG")); if err != nil { os.Exit(3) }; os.Stdout.Write(b)
}`
	if err := os.WriteFile(source, []byte(program), 0600); err != nil {
		t.Fatal(err)
	}
	name := "agy"
	if runtime.GOOS == "windows" {
		name += ".exe"
	}
	if out, err := exec.Command("go", "build", "-o", filepath.Join(bin, name), source).CombinedOutput(); err != nil {
		t.Fatalf("build fake CLI: %v\n%s", err, out)
	}
	t.Setenv("PATH", bin+string(os.PathListSeparator)+os.Getenv("PATH"))
	t.Setenv("CHULI_TEST_AGY_CATALOG", fixture)
	for _, selectedLabel := range []string{"", "Gemini Chosen (High)"} {
		t.Run(selectedLabel, func(t *testing.T) {
			writeCatalog("Gemini Next (High)")
			home := t.TempDir()
			t.Setenv("AGY_HOME", home)
			_, roles, err := DefaultRolesForPreset("agy", "agy-balanced")
			if err != nil {
				t.Fatal(err)
			}
			roles["orchestrator"] = RoleModel{Model: "gemini-next-high", Effort: "low", Label: selectedLabel}
			m := Manager{Source: repo, Home: home, Target: "agy", RoleModels: roles, cacheDir: t.TempDir()}
			p, err := m.PrepareWithOptions(t.Context(), PrepareOptions{Action: "local", ConfigureModels: true})
			if err != nil {
				t.Fatal(err)
			}
			defer p.Cleanup()
			want := selectedLabel
			if want == "" {
				want = "Gemini Next (High)"
			}
			if !strings.Contains(p.roleModelsJSON, `"label":"`+want+`"`) {
				t.Fatalf("prepared plan lost label: %s", p.roleModelsJSON)
			}
			writeCatalog("Changed After Preflight")
			if _, err := m.Apply(t.Context(), p); err != nil {
				t.Fatal(err)
			}
			raw, err := os.ReadFile(filepath.Join(home, "settings.json"))
			if err != nil {
				t.Fatal(err)
			}
			var settings map[string]any
			if err := json.Unmarshal(raw, &settings); err != nil {
				t.Fatal(err)
			}
			if settings["model"] != want {
				t.Fatalf("model = %v, want %q", settings["model"], want)
			}
			var state struct {
				RoleModels map[string]RoleModel `json:"role_models"`
			}
			raw, err = os.ReadFile(filepath.Join(home, "harness-install", "state.json"))
			if err != nil {
				t.Fatal(err)
			}
			if err := json.Unmarshal(raw, &state); err != nil {
				t.Fatal(err)
			}
			if got := state.RoleModels["orchestrator"]; got.Model != "gemini-next-high" || got.Label != want {
				t.Fatalf("installation state lost pair: %+v", got)
			}
		})
	}
	writeCatalog("")
	_, roles, _ := DefaultRolesForPreset("agy", "agy-balanced")
	roles["orchestrator"] = RoleModel{Model: "gemini-next-high", Effort: "high"}
	if _, err := resolveAGYRoleLabel(t.Context(), roles); err == nil {
		t.Fatal("unknown model without usable label must fail")
	}
	if _, err := ParseRoleModelSpecs("agy", "", []string{"chuli_worker=gemini-next-high:high"}); err == nil {
		t.Fatal("specialist exact ID must not silently inherit")
	}
}

func TestAGYCatalogLabelValidation(t *testing.T) {
	_, labels := discoverAGYCatalogUsing(t.Context(), func(context.Context, string, ...string) ([]byte, error) {
		return []byte(`{"status":"SUCCESS","command":{"name":"models","data":{"models":[{"id":"safe","label":"Gemini Next (High)"},{"id":"unsafe","label":"Bad\nInstruction"}]}}}`), nil
	})
	if labels["safe"] != "Gemini Next (High)" || labels["unsafe"] != "" {
		t.Fatalf("labels = %v", labels)
	}
}

func TestAGYLabelRequiresCompatibleInstaller(t *testing.T) {
	roles := map[string]RoleModel{"orchestrator": {Model: "new-model", Effort: "high", Label: "Gemini Next (High)"}}
	for _, version := range []string{"2.0.1", "1.1.0", "invalid"} {
		if err := validateAGYLabelInstaller(version, roles); err == nil || !strings.Contains(err.Error(), "v2.0.2") {
			t.Fatalf("installer %s must require upgrade, got %v", version, err)
		}
	}
	if err := validateAGYLabelInstaller("2.0.2", roles); err != nil {
		t.Fatal(err)
	}
	roles["orchestrator"] = RoleModel{Model: "orchestrator-argon", Effort: "medium"}
	if err := validateAGYLabelInstaller("2.0.1", roles); err != nil {
		t.Fatal("legacy selection must work:", err)
	}
}
