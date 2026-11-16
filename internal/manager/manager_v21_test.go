package manager

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestDetectClientsAndForTarget(t *testing.T) {
	codexHome, _ := filepath.EvalSymlinks(t.TempDir())
	claudeHome, _ := filepath.EvalSymlinks(t.TempDir())
	tmpGemini, _ := filepath.EvalSymlinks(t.TempDir())
	geminiRoot := filepath.Join(tmpGemini, ".gemini")
	agyHome := filepath.Join(geminiRoot, "config")
	agyCli := filepath.Join(geminiRoot, "antigravity-cli")
	if err := os.MkdirAll(agyCli, 0700); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(codexHome, "harness-install"), 0700); err != nil {
		t.Fatal(err)
	}
	state, _ := json.Marshal(map[string]any{
		"version": "2.0.0",
		"client":  "codex",
		"preset":  "codex-balanced",
	})
	if err := os.WriteFile(filepath.Join(codexHome, "harness-install", "state.json"), state, 0600); err != nil {
		t.Fatal(err)
	}
	// claudeHome exists empty -> HomeExists true, InstalledVersion ""
	t.Setenv("CODEX_HOME", codexHome)
	t.Setenv("CLAUDE_CONFIG_DIR", claudeHome)
	t.Setenv("AGY_HOME", agyHome)

	m, err := NewForClient("", "codex")
	if err != nil {
		t.Fatal(err)
	}
	statuses := m.DetectClients()
	if len(statuses) != 3 {
		t.Fatalf("expected 3 client statuses, got %d", len(statuses))
	}
	byID := map[string]ClientStatus{}
	for _, st := range statuses {
		byID[st.Spec.ID] = st
	}
	if !byID["codex"].Detected || byID["codex"].InstalledVersion != "2.0.0" || byID["codex"].Preset != "codex-balanced" {
		t.Fatalf("unexpected codex status: %+v", byID["codex"])
	}
	if !byID["claude"].HomeExists || !byID["claude"].Detected || byID["claude"].InstalledVersion != "" {
		t.Fatalf("unexpected claude status: %+v", byID["claude"])
	}
	if !byID["agy"].HomeExists || !byID["agy"].Detected {
		t.Fatalf("expected agy detected via sibling antigravity-cli: %+v", byID["agy"])
	}
	agyMgr, err := m.ForTarget("agy")
	if err != nil || agyMgr.Target != "agy" || agyMgr.Home != agyHome {
		t.Fatalf("ForTarget(agy) = %+v, %v", agyMgr, err)
	}
}

func TestListReleasesAndSummarizeNotes(t *testing.T) {
	m := Manager{Client: testClient(func(w http.ResponseWriter, r *http.Request) {
		if !strings.Contains(r.URL.Path, "/releases") {
			t.Fatalf("unexpected URL: %s", r.URL)
		}
		_, _ = w.Write([]byte(`[
			{"tag_name":"v2.1.0","name":"Chuli v2.1.0","body":"## Novedades\n- Wizard simple\n- Soporte MCP multi-cliente","published_at":"2026-11-16T10:00:00Z","html_url":"https://github.com/example/v2.1.0"},
			{"tag_name":"v2.2.0-rc1","prerelease":true},
			{"tag_name":"v2.0.5","draft":true},
			{"tag_name":"v0.2.0","name":"Legacy"},
			{"tag_name":"v2.0.0","name":"Chuli v2.0.0","body":"Multi-cliente"}
		]`))
	})}
	releases, err := m.ListReleases(t.Context(), 10)
	if err != nil {
		t.Fatal(err)
	}
	if len(releases) != 2 || releases[0].Tag != "v2.1.0" || releases[1].Tag != "v2.0.0" {
		t.Fatalf("unexpected filtered releases: %+v", releases)
	}
	summary := SummarizeReleaseNotes(releases[0].Body, 6)
	if !strings.Contains(summary, "Wizard simple") || !strings.Contains(summary, "Soporte MCP multi-cliente") {
		t.Fatalf("unexpected release summary: %q", summary)
	}
}

func TestCheckUpdatesDetectsOutdatedProfilesAndManager(t *testing.T) {
	codexHome, _ := filepath.EvalSymlinks(t.TempDir())
	claudeHome, _ := filepath.EvalSymlinks(t.TempDir())
	agyHome, _ := filepath.EvalSymlinks(t.TempDir())
	if err := os.MkdirAll(filepath.Join(codexHome, "harness-install"), 0700); err != nil {
		t.Fatal(err)
	}
	_ = os.WriteFile(filepath.Join(codexHome, "harness-install", "state.json"), []byte(`{"version":"2.0.0"}`), 0600)
	t.Setenv("CODEX_HOME", codexHome)
	t.Setenv("CLAUDE_CONFIG_DIR", claudeHome)
	t.Setenv("AGY_HOME", agyHome)

	m := Manager{
		Home:   codexHome,
		Target: "codex",
		Client: testClient(func(w http.ResponseWriter, r *http.Request) {
			_, _ = w.Write([]byte(`{"tag_name":"v2.1.0","name":"v2.1.0","body":"• Animación limpia\n• MCPs multi-cliente","html_url":"https://github.com/example/v2.1.0"}`))
		}),
	}
	report, err := m.CheckUpdates(t.Context(), "2.0.0-dev")
	if err != nil {
		t.Fatal(err)
	}
	if !report.ManagerNewer || !report.AnyProfileUpdate || report.OutdatedCount != 1 {
		t.Fatalf("unexpected update report: %+v", report)
	}
	if !strings.Contains(report.Summary, "Animación limpia") {
		t.Fatalf("missing release summary: %q", report.Summary)
	}
}

func TestPrepareApplyLocalAndSpecificTagDowngrade(t *testing.T) {
	if _, err := shell(); err != nil {
		t.Skip("requires PowerShell and Python 3.11+")
	}
	repo, err := filepath.Abs("../..")
	if err != nil {
		t.Fatal(err)
	}
	pkg := map[string][]byte{
		"v1.0.0": realProfile(t, repo, "1.0.0"),
		"v1.0.1": realProfile(t, repo, "1.0.1"),
	}
	t.Chdir(t.TempDir())
	homeDir, err := filepath.EvalSymlinks(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	networkAllowed := true
	m := Manager{Home: homeDir, Target: "codex", cacheDir: t.TempDir()}
	m.Client = testClient(func(w http.ResponseWriter, r *http.Request) {
		if !networkAllowed {
			t.Fatalf("unexpected network call during local apply/cached downgrade: %s", r.URL)
		}
		for tag, data := range pkg {
			name := "chuli-profile-" + tag + ".zip"
			base := releaseBase + tag + "/"
			switch r.URL.String() {
			case base + "SHA256SUMS.txt":
				fmt.Fprintf(w, "%x  %s\n", sha256.Sum256(data), name)
				return
			case base + name:
				w.Write(data)
				return
			}
		}
		w.WriteHeader(http.StatusNotFound)
	})

	// Install v1.0.0 first, then upgrade to v1.0.1 so both sit in cacheDir.
	for _, tag := range []string{"v1.0.0", "v1.0.1"} {
		p, err := m.PrepareWithOptions(t.Context(), PrepareOptions{Action: "tag:" + tag})
		if err != nil {
			t.Fatalf("prepare %s: %v", tag, err)
		}
		if _, err := m.Apply(t.Context(), p); err != nil {
			p.Cleanup()
			t.Fatalf("apply %s: %v", tag, err)
		}
		p.Cleanup()
	}

	// Disable network: apply-local and cached downgrade must work offline.
	networkAllowed = false
	m.Preset = "codex-fast"
	localPlan, err := m.PrepareWithOptions(t.Context(), PrepareOptions{Action: "apply-local", ConfigureModels: true})
	if err != nil {
		t.Fatalf("apply-local failed: %v", err)
	}
	if _, err := m.Apply(t.Context(), localPlan); err != nil {
		localPlan.Cleanup()
		t.Fatalf("apply-local apply failed: %v", err)
	}
	localPlan.Cleanup()
	cfg, _ := os.ReadFile(filepath.Join(homeDir, "config.toml"))
	if !strings.Contains(string(cfg), `model_reasoning_effort = "low"`) {
		t.Fatalf("expected codex-fast effort applied locally, got:\n%s", cfg)
	}

	// Downgrade to v1.0.0 without AllowDowngrade must fail.
	if _, err := m.PrepareWithOptions(t.Context(), PrepareOptions{Action: "tag:v1.0.0"}); err == nil {
		t.Fatal("downgrade without AllowDowngrade should fail")
	}
	// With AllowDowngrade: succeeds offline from verified cache.
	downPlan, err := m.PrepareWithOptions(t.Context(), PrepareOptions{Action: "tag:v1.0.0", AllowDowngrade: true})
	if err != nil {
		t.Fatalf("downgrade prepare failed: %v", err)
	}
	if _, err := m.Apply(t.Context(), downPlan); err != nil {
		downPlan.Cleanup()
		t.Fatalf("downgrade apply failed: %v", err)
	}
	downPlan.Cleanup()
	got, err := m.Installed()
	if err != nil || got != "1.0.0" {
		t.Fatalf("installed=%q err=%v; want 1.0.0", got, err)
	}
}
