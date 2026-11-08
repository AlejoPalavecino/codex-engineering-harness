package manager

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestCompare(t *testing.T) {
	cases := []struct {
		a, b string
		want int
	}{
		{"v0.1.3", "0.1.2", 1},
		{"v1.0.0", "1.0.0", 0},
		{"0.2.0", "v0.1.12", 1},
		{"0.1.2", "0.1.11", -1},
	}
	for _, c := range cases {
		got, err := Compare(c.a, c.b)
		if err != nil || got != c.want {
			t.Errorf("Compare(%s,%s)=%d %v; want %d", c.a, c.b, got, err, c.want)
		}
	}
	for _, bad := range []string{"main", "v0.1.3-rc", "../../main", "v0.1.2;whoami"} {
		if _, err := Compare(bad, "v0.1.2"); err == nil {
			t.Errorf("accepted unsafe version %q", bad)
		}
	}
}

type httpFunc func(*http.Request) (*http.Response, error)

func (f httpFunc) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }

func testClient(h http.HandlerFunc) *http.Client {
	return &http.Client{Transport: httpFunc(func(r *http.Request) (*http.Response, error) {
		w := httptest.NewRecorder()
		h(w, r)
		return w.Result(), nil
	})}
}

func TestLatestValidatesMetadata(t *testing.T) {
	for _, tc := range []struct {
		json  string
		valid bool
	}{
		{`{"tag_name":"v1.0.0"}`, true}, {`{"tag_name":"v1.0.0","draft":true}`, false},
		{`{"tag_name":"v1.0.0","prerelease":true}`, false}, {`{"tag_name":"../../main"}`, false},
	} {
		m := Manager{Client: testClient(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.String() != githubAPI || r.Header.Get("User-Agent") == "" {
				t.Fatal("unexpected request", r.URL)
			}
			_, _ = w.Write([]byte(tc.json))
		})}
		_, err := m.Latest(t.Context())
		if (err == nil) != tc.valid {
			t.Errorf("Latest(%s): %v", tc.json, err)
		}
	}
}
func TestReadInstalled(t *testing.T) {
	m := Manager{Home: t.TempDir()}
	if v, err := m.Installed(); v != "" || err != nil {
		t.Fatalf("not installed: %q %v", v, err)
	}
	dir := filepath.Join(m.Home, "harness-install")
	if err := os.MkdirAll(dir, 0700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "state.json"), []byte("{\"version\":\"0.1.2\"}"), 0600); err != nil {
		t.Fatal(err)
	}
	if v, err := m.Installed(); v != "0.1.2" || err != nil {
		t.Fatalf("installed: %q %v", v, err)
	}
	if err := os.WriteFile(filepath.Join(dir, "state.json"), []byte("bad"), 0600); err != nil {
		t.Fatal(err)
	}
	if _, err := m.Installed(); err == nil {
		t.Fatal("accepted broken manifest")
	}
	unsupported := []byte(`{"version":"0.2.1","schema_version":3,"rollback_schema_version":2}`)
	if err := os.WriteFile(filepath.Join(dir, "state.json"), unsupported, 0600); err != nil {
		t.Fatal(err)
	}
	if _, err := m.Installed(); err == nil {
		t.Fatal("accepted unsupported state schema")
	}
}
func TestLocalSource(t *testing.T) {
	root := t.TempDir()
	m := Manager{Source: root}
	if _, err := m.Local(); err == nil {
		t.Fatal("accepted incomplete checkout")
	}
	if err := os.MkdirAll(filepath.Join(root, "scripts"), 0700); err != nil {
		t.Fatal(err)
	}
	for _, name := range []string{"install.ps1", "verify.ps1", "rollback.ps1"} {
		if err := os.WriteFile(filepath.Join(root, "scripts", name), []byte("# script"), 0600); err != nil {
			t.Fatal(err)
		}
	}
	if err := os.WriteFile(filepath.Join(root, "VERSION"), []byte("0.1.3\n"), 0600); err != nil {
		t.Fatal(err)
	}
	if path, err := m.Local(); err != nil || path != root {
		t.Fatalf("local %s %v", path, err)
	}
}

func TestRollbackCompatibility(t *testing.T) {
	cases := []struct {
		source  string
		schema  int
		wantErr bool
	}{
		{source: "0.1.2", schema: 1, wantErr: true},
		{source: "0.2.0", schema: 1, wantErr: true},
		{source: "0.2.1", schema: 1, wantErr: false},
		{source: "0.2.1", schema: 2, wantErr: false},
		{source: "0.3.0", schema: 2, wantErr: false},
		{source: "0.3.0", schema: 3, wantErr: true},
	}
	for _, tc := range cases {
		err := validateRollbackCompatibility(tc.source, tc.schema)
		if (err != nil) != tc.wantErr {
			t.Errorf("validateRollbackCompatibility(%q, %d) error = %v; wantErr %v", tc.source, tc.schema, err, tc.wantErr)
		}
	}
}

func TestRollbackRejectsOldCheckoutBeforeRunningPowerShell(t *testing.T) {
	home := t.TempDir()
	stateDir := filepath.Join(home, "harness-install")
	if err := os.MkdirAll(stateDir, 0700); err != nil {
		t.Fatal(err)
	}
	state := []byte(`{"version":"0.2.0","files":[]}`)
	if err := os.WriteFile(filepath.Join(stateDir, "state.json"), state, 0600); err != nil {
		t.Fatal(err)
	}

	source := t.TempDir()
	if err := os.MkdirAll(filepath.Join(source, "scripts"), 0700); err != nil {
		t.Fatal(err)
	}
	for _, name := range []string{"install.ps1", "verify.ps1", "rollback.ps1"} {
		if err := os.WriteFile(filepath.Join(source, "scripts", name), []byte("# script"), 0600); err != nil {
			t.Fatal(err)
		}
	}
	if err := os.WriteFile(filepath.Join(source, "VERSION"), []byte("0.1.2\n"), 0600); err != nil {
		t.Fatal(err)
	}

	m := Manager{Home: home, Source: source}
	_, err := m.Prepare(context.Background(), "rollback", false)
	if err == nil || !strings.Contains(err.Error(), "rollback requiere el ejecutor seguro v0.2.1") {
		t.Fatalf("old rollback executor was not rejected before PowerShell: %v", err)
	}
}

func TestNeverExecutesImplicitCheckout(t *testing.T) {
	dir, err := filepath.Abs("../..")
	if err != nil {
		t.Fatal(err)
	}
	t.Chdir(dir)
	m := Manager{Home: t.TempDir(), cacheDir: t.TempDir()}
	if _, err := m.Local(); err == nil {
		t.Fatal("implicit checkout accepted")
	}
	if _, err := m.Doctor(t.Context()); err == nil {
		t.Fatal("implicit diagnostics accepted")
	}
}

func TestProfileReportsRecordedInventory(t *testing.T) {
	m := Manager{Home: t.TempDir()}
	dir := filepath.Join(m.Home, "harness-install")
	if err := os.Mkdir(dir, 0700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "state.json"), []byte(`{"version":"1.0.0","managed_files":{"skills/example/SKILL.md":"hash","skills/_shared/extra.md":"hash","agents/review.toml":"hash"},"managed_config":{"model":"private-model","model_reasoning_effort":"medium","secret":"sensitive"}}`), 0600); err != nil {
		t.Fatal(err)
	}
	s, err := m.Profile()
	if err != nil || !strings.Contains(s, "private-model") || !strings.Contains(s, "Skills registradas (1)\nexample") || !strings.Contains(s, "Agentes registrados (1)\nreview") || strings.Contains(s, "sensitive") {
		t.Fatalf("profile: %q, %v", s, err)
	}
}

func TestApplyRejectsChangedInstallation(t *testing.T) {
	m := Manager{Home: t.TempDir()}
	dir := filepath.Join(m.Home, "harness-install")
	if err := os.Mkdir(dir, 0700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "state.json"), []byte(`{"version":"1.0.0"}`), 0600); err != nil {
		t.Fatal(err)
	}
	if _, err := m.Apply(t.Context(), Prepared{Action: "install"}); err == nil || !strings.Contains(err.Error(), "cambió") {
		t.Fatal(err)
	}
}
