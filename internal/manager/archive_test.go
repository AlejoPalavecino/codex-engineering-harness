package manager

import (
	"archive/zip"
	"bytes"
	"crypto/sha256"
	"errors"
	"fmt"
	"io/fs"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

type zipEntry struct {
	name, body string
	mode       fs.FileMode
}

func zipBytes(t *testing.T, entries []zipEntry) []byte {
	t.Helper()
	var buf bytes.Buffer
	w := zip.NewWriter(&buf)
	for _, e := range entries {
		h := &zip.FileHeader{Name: e.name, Method: zip.Deflate}
		h.SetMode(e.mode)
		f, err := w.CreateHeader(h)
		if err != nil {
			t.Fatal(err)
		}
		if _, err := f.Write([]byte(e.body)); err != nil {
			t.Fatal(err)
		}
	}
	if err := w.Close(); err != nil {
		t.Fatal(err)
	}
	return buf.Bytes()
}

func TestRejectDangerousZip(t *testing.T) {
	for _, entries := range [][]zipEntry{
		{{"../outside", "", 0600}}, {{"/etc/passwd", "", 0600}}, {{"C:/test", "", 0600}}, {{`..\test`, "", 0600}},
		{{"scripts/CON.txt", "", 0600}}, {{"test:ads", "", 0600}}, {{"name.", "", 0600}}, {{"link", "../../escape", fs.ModeSymlink | 0777}},
		{{"x", "", 0600}, {"x", "", 0600}}, {{"Scripts/x", "", 0600}, {"scripts/y", "", 0600}},
		{{"x", "", 0600}, {"x/y", "", 0600}}, {{"huge", strings.Repeat("x", maxEntry+1), 0600}},
	} {
		data := zipBytes(t, entries)
		reader, err := zip.NewReader(bytes.NewReader(data), int64(len(data)))
		if err != nil {
			continue
		} // archive/zip may itself reject insecure names
		if err = extract(reader, t.TempDir()); err == nil {
			t.Errorf("accepted unsafe ZIP entry %q", entries[len(entries)-1].name)
		}
	}
}
func TestCacheAndEmbeddedVersion(t *testing.T) {
	entries := []zipEntry{{"VERSION", "1.0.0", 0600}, {"scripts/verify.ps1", "", 0600}, {"scripts/install.ps1", "", 0600}, {"scripts/rollback.ps1", "", 0600}}
	payload := zipBytes(t, entries)
	digest := fmt.Sprintf("%x", sha256.Sum256(payload))
	m := Manager{cacheDir: t.TempDir()}
	a, err := verified("v1.0.0", digest, payload)
	if err != nil {
		t.Fatal(err)
	}
	if err = m.remember(a); err != nil {
		t.Fatal(err)
	}
	if err = m.remember(a); err != nil {
		t.Fatal(err)
	}
	cached, err := m.cached(a.tag)
	if err != nil {
		t.Fatal(err)
	}
	cached.tag = "v1.0.1"
	if p, err := cached.unpack(); err == nil {
		p.Cleanup()
		t.Fatal("accepted mismatched VERSION")
	}
	differentData := zipBytes(t, append(entries, zipEntry{"extra", "new", 0600}))
	different, _ := verified(a.tag, fmt.Sprintf("%x", sha256.Sum256(differentData)), differentData)
	if err = m.remember(different); err == nil {
		t.Fatal("changed a pinned tag")
	}
	if err = os.WriteFile(filepath.Join(m.cacheDir, a.tag, digest+".zip"), []byte("alterado"), 0600); err != nil {
		t.Fatal(err)
	}
	if _, err = m.cached(a.tag); err == nil {
		t.Fatal("accepted altered cache")
	}
}

func TestDownloadRejectsInvalidSums(t *testing.T) {
	name := "chuli-profile-v1.0.0.zip"
	line := strings.Repeat("0", 64) + "  " + name
	for _, sum := range []string{"", "invalid  " + name, line + "\n" + line, line, strings.Repeat("0", 64) + "  ../" + name} {
		m := Manager{Client: testClient(func(w http.ResponseWriter, r *http.Request) {
			if strings.HasSuffix(r.URL.Path, "SHA256SUMS.txt") {
				fmt.Fprintln(w, sum)
			} else {
				fmt.Fprint(w, "wrong")
			}
		})}
		if _, err := m.download(t.Context(), "v1.0.0"); err == nil {
			t.Fatal("accepted invalid checksums")
		}
	}
}

func realProfile(t *testing.T, repo, version string) []byte {
	t.Helper()
	notice, err := os.ReadFile(filepath.Join(repo, "THIRD_PARTY_NOTICES.md"))
	if err != nil {
		t.Fatal(err)
	}
	entries := []zipEntry{{"VERSION", version + "\n", 0600}, {"THIRD_PARTY_NOTICES.md", string(notice), 0600}}
	for _, prefix := range []string{"config", "agents", "skills", "scripts", "templates", "docs/protocols"} {
		base := filepath.Join(repo, filepath.FromSlash(prefix))
		err := filepath.WalkDir(base, func(path string, d fs.DirEntry, err error) error {
			if err != nil {
				return err
			}
			if d.IsDir() {
				if d.Name() == "__pycache__" {
					return fs.SkipDir
				}
				return nil
			}
			rel, e := filepath.Rel(repo, path)
			if e != nil {
				return e
			}
			body, e := os.ReadFile(path)
			if e != nil {
				return e
			}
			entries = append(entries, zipEntry{filepath.ToSlash(rel), string(body), 0600})
			return nil
		})
		if err != nil {
			t.Fatal(err)
		}
	}
	return zipBytes(t, entries)
}

// Exercise the actual scripts from release-shaped archives. No real GitHub,
// personal CODEX_HOME, Git checkout as cwd, or network during recovery.
func TestOfficialProfileLifecycle(t *testing.T) {
	if _, err := shell(); err != nil {
		t.Skip("requires PowerShell and Python 3.11+")
	}
	repo, err := filepath.Abs("../..")
	if err != nil {
		t.Fatal(err)
	}
	pkg := map[string][]byte{}
	for _, v := range []string{"1.0.0", "1.0.1"} {
		pkg["v"+v] = realProfile(t, repo, v)
	}
	latest := "v1.0.0"
	t.Chdir(t.TempDir())
	homeDir, err := filepath.EvalSymlinks(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	m := Manager{Home: homeDir, cacheDir: t.TempDir()}
	m.Client = testClient(func(w http.ResponseWriter, r *http.Request) {
		if latest == "" {
			t.Fatal("network during offline recovery")
		}
		name := "chuli-profile-" + latest + ".zip"
		base := releaseBase + latest + "/"
		switch r.URL.String() {
		case githubAPI:
			fmt.Fprintf(w, `{"tag_name":"%s"}`, latest)
		case base + "SHA256SUMS.txt":
			fmt.Fprintf(w, "%x  %s\n", sha256.Sum256(pkg[latest]), name)
		case base + name:
			w.Write(pkg[latest])
		default:
			t.Fatalf("unexpected URL: %s", r.URL)
		}
	})
	do := func(action, want string) {
		t.Helper()
		p, e := m.Prepare(t.Context(), action, false)
		if e != nil {
			t.Fatal(e)
		}
		defer p.Cleanup()
		if want == "1.0.0" && action == "latest" {
			if entries, _ := os.ReadDir(m.Home); len(entries) != 0 {
				t.Fatal("preview modified profile")
			}
			if entries, _ := os.ReadDir(m.cacheDir); len(entries) != 0 {
				t.Fatal("preview modified cache")
			}
			// Apply must re-extract the verified bytes.
			if e = os.WriteFile(filepath.Join(p.Root, "scripts", "install.ps1"), []byte("throw 'untrusted preview'"), 0600); e != nil {
				t.Fatal(e)
			}
		}
		log, e := m.Apply(t.Context(), p)
		if e != nil {
			t.Fatalf("%s: %v\n%s", action, e, log)
		}
		if got, e := m.Installed(); got != want || e != nil {
			t.Fatalf("installed=%s want=%s err=%v", got, want, e)
		}
	}
	do("latest", "1.0.0")
	if _, err := m.Prepare(t.Context(), "latest", false); err == nil {
		t.Fatal("accepted same release")
	}
	// Gentle may have been added after Chuli was installed. The same official
	// version may reconcile that conflict only after a new opt-in preview.
	path := filepath.Join(m.Home, "AGENTS.md")
	original, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	conflicted := append(bytes.Clone(original), []byte("\n<!-- gentle-ai:orchestrator -->\nGentle routes here.\n<!-- /gentle-ai:orchestrator -->\n")...)
	if err := os.WriteFile(path, conflicted, 0600); err != nil {
		t.Fatal(err)
	}
	if _, err := m.Prepare(t.Context(), "latest", false); !errors.Is(err, ErrGentleConflict) {
		t.Fatalf("expected an adoptable conflict: %v", err)
	}
	adoption, err := m.Prepare(t.Context(), "latest", true)
	if err != nil {
		t.Fatal(err)
	}
	defer adoption.Cleanup()
	if b, e := os.ReadFile(path); e != nil || !bytes.Equal(b, conflicted) {
		t.Fatal("adoption preview changed instructions")
	}
	if log, e := m.Apply(t.Context(), adoption); e != nil {
		t.Fatalf("adoption apply: %v\n%s", e, log)
	}
	if b, e := os.ReadFile(path); e != nil || bytes.Contains(b, []byte("<!-- gentle-ai:orchestrator -->")) {
		t.Fatal("routing was not adopted")
	}
	do("rollback", "1.0.0")
	if b, e := os.ReadFile(path); e != nil || !bytes.Equal(b, conflicted) {
		t.Fatal("rollback did not restore Gentle")
	}
	if err := os.WriteFile(path, original, 0600); err != nil {
		t.Fatal(err)
	}
	latest = "v1.0.1"
	do("latest", "1.0.1")
	latest = ""
	if log, err := m.Doctor(t.Context()); err != nil {
		t.Fatalf("doctor: %v\n%s", err, log)
	}
	do("rollback", "1.0.0")
	do("rollback", "")
}

func TestCacheRefusesDirectoryLink(t *testing.T) {
	base := t.TempDir()
	m := Manager{cacheDir: filepath.Join(base, "chuli", "profiles")}
	target := t.TempDir()
	if err := os.Symlink(target, filepath.Dir(m.cacheDir)); err != nil {
		t.Skip("directory link unavailable")
	}
	if err := m.checkCache(true); err == nil {
		t.Fatal("accepted cache symlink")
	}
	if _, err := os.Stat(filepath.Join(target, "profiles")); !os.IsNotExist(err) {
		t.Fatal("wrote through cache symlink")
	}
}
