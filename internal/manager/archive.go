// The release archive is the unit of trust. Its SHA-256 comes from the same
// official GitHub Release; this checks integrity, not a publisher signature.
package manager

import (
	"archive/zip"
	"bytes"
	"context"
	"crypto/sha256"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
	"time"
)

const (
	maxArchive  = 16 << 20
	maxExpanded = 32 << 20
	maxEntry    = 4 << 20
	maxEntries  = 2048
)

var digestName = regexp.MustCompile("^[a-f0-9]{64}$")
var component = regexp.MustCompile("^[A-Za-z0-9_-][A-Za-z0-9_.-]*$")
var reserved = regexp.MustCompile("(?i)^(CON|PRN|AUX|NUL|COM[0-9]|LPT[0-9])([.]|$)")

type profileArchive struct {
	tag    string
	digest string
	data   []byte
}

func (m Manager) httpClient() *http.Client {
	if m.Client != nil {
		return m.Client
	}
	return &http.Client{Timeout: 45 * time.Second, CheckRedirect: func(req *http.Request, via []*http.Request) error {
		host := req.URL.Hostname()
		if len(via) > 5 || req.URL.Scheme != "https" || (host != "github.com" && host != "api.github.com" && !strings.HasSuffix(host, ".githubusercontent.com")) {
			return errors.New("redirección fuera de GitHub")
		}
		return nil
	}}
}

func bounded(r io.Reader, limit int64) ([]byte, error) {
	b, err := io.ReadAll(io.LimitReader(r, limit+1))
	if int64(len(b)) > limit {
		return nil, errors.New("se superó el límite de tamaño")
	}
	return b, err
}

func (m Manager) asset(ctx context.Context, tag, name string, limit int64) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", releaseBase+tag+"/"+name, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "Chuli-Manager")
	resp, err := m.httpClient().Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("no se pudo obtener %s de %s (HTTP %d)", name, tag, resp.StatusCode)
	}
	return bounded(resp.Body, limit)
}

func (m Manager) download(ctx context.Context, tag string) (*profileArchive, error) {
	if !releaseTag.MatchString(tag) {
		return nil, errors.New("tag de release inválido")
	}
	name := "chuli-profile-" + tag + ".zip"
	sums, err := m.asset(ctx, tag, "SHA256SUMS.txt", 64<<10)
	if err != nil {
		return nil, err
	}
	expected := ""
	for _, line := range strings.Split(string(sums), "\n") {
		fields := strings.Fields(line)
		if len(fields) != 2 || strings.TrimPrefix(fields[1], "*") != name {
			continue
		}
		hash := strings.ToLower(fields[0])
		if expected != "" || !digestName.MatchString(hash) {
			return nil, errors.New("entrada SHA-256 duplicada o inválida")
		}
		expected = hash
	}
	if expected == "" {
		return nil, fmt.Errorf("SHA256SUMS.txt no incluye %s", name)
	}
	payload, err := m.asset(ctx, tag, name, maxArchive)
	if err != nil {
		return nil, err
	}
	return verified(tag, expected, payload)
}

func verified(tag, digest string, data []byte) (*profileArchive, error) {
	if !releaseTag.MatchString(tag) || !digestName.MatchString(digest) || fmt.Sprintf("%x", sha256.Sum256(data)) != digest {
		return nil, errors.New("falló la comprobación SHA-256 o tag del paquete")
	}
	return &profileArchive{tag, digest, data}, nil
}

func safeName(name string) bool {
	if len(name) > 200 || name == "" {
		return false
	}
	for _, part := range strings.Split(name, "/") {
		if !component.MatchString(part) || strings.HasSuffix(part, ".") || reserved.MatchString(part) {
			return false
		}
	}
	return true
}

func (a *profileArchive) unpack() (Prepared, error) {
	z, err := zip.NewReader(bytes.NewReader(a.data), int64(len(a.data)))
	if err != nil {
		return Prepared{}, err
	}
	if len(z.File) > maxEntries {
		return Prepared{}, errors.New("demasiadas entradas en el paquete")
	}
	dir, err := os.MkdirTemp("", "chuli-profile-")
	if err != nil {
		return Prepared{}, err
	}
	plan := Prepared{Root: dir, Temporary: true, Version: strings.TrimPrefix(a.tag, "v"), archive: a}
	err = extract(z, dir)
	if err == nil {
		var v string
		v, err = ValidateSource(dir)
		if err == nil && v != plan.Version {
			err = fmt.Errorf("VERSION %q no coincide con %s", v, a.tag)
		}
	}
	if err != nil {
		plan.Cleanup()
		return Prepared{}, err
	}
	return plan, nil
}

func extract(z *zip.Reader, dir string) error {
	// A new private directory and os.Root contain all writes, including on Windows.
	root, err := os.OpenRoot(dir)
	if err != nil {
		return err
	}
	defer root.Close()
	seen := map[string]bool{}
	spelling := map[string]string{}
	var total uint64
	for _, f := range z.File {
		name := strings.TrimSuffix(f.Name, "/")
		mode := f.Mode()
		if !safeName(name) || seen[strings.ToLower(name)] || (!mode.IsDir() && !mode.IsRegular()) || (mode.IsDir() != (name != f.Name)) {
			return fmt.Errorf("entrada ZIP inválida o duplicada: %q", f.Name)
		}
		seen[strings.ToLower(name)] = true
		for prefix := name; prefix != ""; {
			key := strings.ToLower(prefix)
			if prev, ok := spelling[key]; ok && prev != prefix {
				return fmt.Errorf("colisión de rutas ZIP: %q y %q", prefix, prev)
			}
			spelling[key] = prefix
			i := strings.LastIndexByte(prefix, '/')
			if i < 0 {
				break
			}
			prefix = prefix[:i]
		}
		if f.UncompressedSize64 > maxEntry || (mode.IsDir() && f.UncompressedSize64 != 0) {
			return fmt.Errorf("tamaño inválido: %s", name)
		}
		total += f.UncompressedSize64
		if total > maxExpanded {
			return errors.New("paquete descomprimido demasiado grande")
		}
		local := filepath.FromSlash(name)
		if mode.IsDir() {
			err = root.MkdirAll(local, 0700)
		} else {
			if err = root.MkdirAll(filepath.Dir(local), 0700); err == nil {
				err = extractFile(root, local, f)
			}
		}
		if err != nil {
			return fmt.Errorf("extracción de %s: %w", name, err)
		}
	}
	return nil
}

func extractFile(root *os.Root, name string, f *zip.File) error {
	src, err := f.Open()
	if err != nil {
		return err
	}
	defer src.Close()
	dst, err := root.OpenFile(name, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0600)
	if err != nil {
		return err
	}
	n, err := io.Copy(dst, io.LimitReader(src, int64(f.UncompressedSize64)+1))
	closeErr := dst.Close()
	if err != nil {
		return err
	}
	if uint64(n) != f.UncompressedSize64 {
		return errors.New("tamaño ZIP inconsistente")
	}
	return closeErr
}

// Refuse linked or shared writable application directories. The parent of
// chuli/ is the platform's user configuration directory, selected by New.
func (m Manager) checkCache(create bool) error {
	if !filepath.IsAbs(m.cacheDir) {
		return errors.New("ubicación local inválida")
	}
	appDir := filepath.Dir(m.cacheDir)
	if create {
		if err := os.MkdirAll(filepath.Dir(appDir), 0700); err != nil {
			return err
		}
	}
	for _, dir := range []string{appDir, m.cacheDir} {
		if create {
			err := os.Mkdir(dir, 0700)
			if err != nil && !errors.Is(err, os.ErrExist) {
				return err
			}
		}
		if err := privateDir(dir); err != nil {
			return err
		}
	}
	return nil
}

func privateDir(path string) error {
	info, err := os.Lstat(path)
	if err != nil {
		return err
	}
	if !info.IsDir() || (runtime.GOOS != "windows" && info.Mode().Perm()&0022 != 0) {
		return fmt.Errorf("directorio inseguro para paquetes: %s", path)
	}
	return nil
}

// Publish a complete version directory with a single rename. A version is
// immutable: a second archive under the same tag cannot replace the first.
func (m Manager) remember(a *profileArchive) error {
	if err := m.checkCache(true); err != nil {
		return err
	}
	dir, err := os.MkdirTemp(m.cacheDir, ".incoming-")
	if err != nil {
		return err
	}
	defer os.RemoveAll(dir)
	if err = os.WriteFile(filepath.Join(dir, a.digest+".zip"), a.data, 0600); err != nil {
		return err
	}
	err = os.Rename(dir, filepath.Join(m.cacheDir, a.tag))
	if err == nil {
		return nil
	}
	existing, e := m.cached(a.tag)
	if e != nil {
		return fmt.Errorf("no se pudo conservar paquete (%v): %w", err, e)
	}
	if existing.digest != a.digest {
		return errors.New("el paquete local tiene otra huella para la misma versión")
	}
	return nil
}

func (m Manager) cached(tag string) (*profileArchive, error) {
	if m.cacheDir == "" || !releaseTag.MatchString(tag) {
		return nil, errors.New("ubicación o versión local inválida")
	}
	versionDir := filepath.Join(m.cacheDir, tag)
	if err := m.checkCache(false); err != nil {
		return nil, err
	}
	if err := privateDir(versionDir); err != nil {
		return nil, err
	}
	files, err := os.ReadDir(versionDir)
	if err != nil {
		return nil, err
	}
	if len(files) != 1 {
		return nil, errors.New("copia local incompleta o ambigua")
	}
	name := files[0].Name()
	hash, ok := strings.CutSuffix(name, ".zip")
	if !ok || !digestName.MatchString(hash) || !files[0].Type().IsRegular() {
		return nil, errors.New("archivo local inesperado")
	}
	root, err := os.OpenRoot(versionDir)
	if err != nil {
		return nil, err
	}
	defer root.Close()
	src, err := root.Open(name)
	if err != nil {
		return nil, err
	}
	defer src.Close()
	data, err := bounded(src, maxArchive)
	if err != nil {
		return nil, err
	}
	return verified(tag, hash, data)
}
