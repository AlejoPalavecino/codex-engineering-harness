// Package manager invokes the existing reviewed installer without rewriting its rules.
package manager

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"sort"
	"strconv"
	"strings"
)

const Repository = "AlejoPalavecino/codex-engineering-harness"
const releaseBase = "https://github.com/" + Repository + "/releases/download/"
const githubAPI = "https://api.github.com/repos/" + Repository + "/releases/latest"

// ErrGentleConflict identifies a recognized, safely adoptable routing block.
var ErrGentleConflict = errors.New("se detectó un orquestador de Gentle que puede respaldarse")

var releaseTag = regexp.MustCompile("^v[0-9]+\\.[0-9]+\\.[0-9]+$")
var version = regexp.MustCompile("^v?([0-9]+)\\.([0-9]+)\\.([0-9]+)$")

type Manager struct {
	Home     string
	Source   string
	Client   *http.Client
	cacheDir string // resolved by New; private user configuration, outside the working directory
}

type Release struct {
	Tag string
	URL string
}

type Prepared struct {
	Root               string
	Temporary          bool
	Version            string
	Action             string
	Preview            string
	archive            *profileArchive
	stateID            [sha256.Size]byte
	adoptGentleRouting bool
}

func New(source string) (Manager, error) {
	home := os.Getenv("CODEX_HOME")
	if home == "" {
		user, err := os.UserHomeDir()
		if err != nil {
			return Manager{}, err
		}
		home = filepath.Join(user, ".codex")
	}
	if !filepath.IsAbs(home) {
		return Manager{}, errors.New("CODEX_HOME debe ser una ruta absoluta")
	}
	full, err := filepath.Abs(home)
	if err != nil {
		return Manager{}, err
	}
	config, err := os.UserConfigDir()
	if err != nil {
		return Manager{}, err
	}
	if !filepath.IsAbs(config) {
		return Manager{}, errors.New("el directorio de configuración debe ser absoluto")
	}
	if source != "" {
		source, err = filepath.Abs(source)
		if err != nil {
			return Manager{}, err
		}
	}
	return Manager{Home: full, Source: source, cacheDir: filepath.Join(config, "chuli", "profiles")}, nil
}

func Compare(a, b string) (int, error) {
	aa, bb := version.FindStringSubmatch(strings.TrimSpace(a)), version.FindStringSubmatch(strings.TrimSpace(b))
	if aa == nil || bb == nil {
		return 0, fmt.Errorf("versión no reconocida: %q / %q", a, b)
	}
	for i := 1; i <= 3; i++ {
		x, ex := strconv.Atoi(aa[i])
		y, ey := strconv.Atoi(bb[i])
		if ex != nil || ey != nil {
			return 0, errors.New("componente de versión fuera de rango")
		}
		if x < y {
			return -1, nil
		}
		if x > y {
			return 1, nil
		}
	}
	return 0, nil
}

type installationState struct {
	id                    [sha256.Size]byte
	Version               string                     `json:"version"`
	ManagedFiles          map[string]string          `json:"managed_files"`
	ManagedConfig         map[string]json.RawMessage `json:"managed_config"`
	SchemaVersion         int                        `json:"schema_version"`
	RollbackSchemaVersion int                        `json:"rollback_schema_version"`
}

func (m Manager) readInstallationState() (installationState, error) {
	path := filepath.Join(m.Home, "harness-install", "state.json")
	info, err := os.Lstat(path)
	if errors.Is(err, os.ErrNotExist) {
		return installationState{}, nil
	}
	if err != nil {
		return installationState{}, err
	}
	if !info.Mode().IsRegular() {
		return installationState{}, errors.New("state.json debe ser un archivo regular")
	}
	f, err := os.Open(path)
	if err != nil {
		return installationState{}, err
	}
	defer f.Close()
	raw, err := bounded(f, maxEntry)
	if err != nil {
		return installationState{}, err
	}
	var state installationState
	if err := json.Unmarshal(raw, &state); err != nil {
		return installationState{}, fmt.Errorf("state.json inválido: %w", err)
	}
	if state.Version == "" {
		return installationState{}, errors.New("state.json sin versión")
	}
	if _, err := Compare(state.Version, state.Version); err != nil {
		return installationState{}, err
	}
	if state.SchemaVersion == 0 {
		state.SchemaVersion = 1 // v0.2.0 and earlier manifests predate this field.
	}
	if state.SchemaVersion < 1 || state.SchemaVersion > 2 {
		return installationState{}, fmt.Errorf("state schema %d no soportado por este Manager", state.SchemaVersion)
	}
	if state.RollbackSchemaVersion == 0 {
		state.RollbackSchemaVersion = 1 // v0.2.0 and earlier manifests predate this field.
	}
	if state.RollbackSchemaVersion < 1 || state.RollbackSchemaVersion > 2 {
		return installationState{}, fmt.Errorf("rollback schema %d no soportado por este Manager", state.RollbackSchemaVersion)
	}
	state.id = sha256.Sum256(raw)
	return state, nil
}

func (m Manager) Installed() (string, error) {
	state, err := m.readInstallationState()
	if err != nil {
		return "", err
	}
	return state.Version, nil
}

// Profile describes the inventory recorded by the installer. It does not load
// arbitrary user configuration or inspect an active Codex session.
func (m Manager) Profile() (string, error) {
	state, err := m.readInstallationState()
	if err != nil {
		return "", err
	}
	if state.Version == "" {
		return "Perfil Chuli para Codex: sin instalar", nil
	}
	lines := []string{"Perfil Chuli para Codex " + state.Version, "Datos del manifiesto de instalación"}
	for _, item := range []struct{ key, label string }{{"model", "Modelo registrado"}, {"model_reasoning_effort", "Esfuerzo registrado"}} {
		var value string
		if raw := state.ManagedConfig[item.key]; raw != nil && json.Unmarshal(raw, &value) == nil && component.MatchString(value) && len(value) <= 100 {
			lines = append(lines, item.label+": "+value)
		}
	}
	if state.ManagedFiles != nil {
		var skills, agents []string
		for path := range state.ManagedFiles {
			parts := strings.Split(path, "/")
			if len(parts) == 3 && parts[0] == "skills" && parts[2] == "SKILL.md" && component.MatchString(parts[1]) {
				skills = append(skills, parts[1])
			}
			if len(parts) == 2 && parts[0] == "agents" && strings.HasSuffix(parts[1], ".toml") && component.MatchString(parts[1]) {
				agents = append(agents, strings.TrimSuffix(parts[1], ".toml"))
			}
		}
		sort.Strings(agents)
		sort.Strings(skills)
		lines = append(lines, fmt.Sprintf("Agentes registrados (%d)", len(agents)))
		lines = append(lines, agents...)
		lines = append(lines, fmt.Sprintf("Skills registradas (%d)", len(skills)))
		lines = append(lines, skills...)
	}
	return strings.Join(lines, "\n"), nil
}

func validateRollbackCompatibility(sourceVersion string, schema int) error {
	cmp, err := Compare(sourceVersion, "0.2.1")
	if err != nil {
		return err
	}
	if cmp < 0 {
		return fmt.Errorf("rollback requiere el ejecutor seguro v0.2.1 o posterior; checkout local %s rechazado", sourceVersion)
	}
	if schema < 1 || schema > 2 {
		return fmt.Errorf("rollback schema %d no soportado por este Manager", schema)
	}
	return nil
}

func (m Manager) Latest(ctx context.Context) (Release, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", githubAPI, nil)
	if err != nil {
		return Release{}, err
	}
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("User-Agent", "Chuli-Manager")
	resp, err := m.httpClient().Do(req)
	if err != nil {
		return Release{}, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return Release{}, fmt.Errorf("GitHub Releases: HTTP %d", resp.StatusCode)
	}
	var v struct {
		Tag        string `json:"tag_name"`
		URL        string `json:"html_url"`
		Draft      bool   `json:"draft"`
		Prerelease bool   `json:"prerelease"`
	}
	if err := json.NewDecoder(io.LimitReader(resp.Body, 65536)).Decode(&v); err != nil {
		return Release{}, err
	}
	if !releaseTag.MatchString(v.Tag) || v.Draft || v.Prerelease {
		return Release{}, errors.New("release no estable o tag no permitido")
	}
	return Release{v.Tag, v.URL}, nil
}

func ValidateSource(root string) (string, error) {
	for _, file := range []string{"scripts/install.ps1", "scripts/verify.ps1", "scripts/rollback.ps1", "VERSION"} {
		info, err := os.Stat(filepath.Join(root, filepath.FromSlash(file)))
		if err != nil {
			return "", fmt.Errorf("fuente incompleta (%s): %w", file, err)
		}
		if !info.Mode().IsRegular() {
			return "", fmt.Errorf("archivo inválido: %s", file)
		}
	}
	raw, err := os.ReadFile(filepath.Join(root, "VERSION"))
	if err != nil {
		return "", err
	}
	v := strings.TrimSpace(string(raw))
	if _, err := Compare(v, v); err != nil {
		return "", err
	}
	return v, nil
}

func (m Manager) Local() (string, error) {
	root := m.Source
	if root == "" {
		return "", errors.New("indicá un checkout de confianza con --source RUTA")
	}
	abs, err := filepath.Abs(root)
	if err != nil {
		return "", err
	}
	if _, err = ValidateSource(abs); err != nil {
		return "", fmt.Errorf("no se encontró un checkout válido; usá --source RUTA: %w", err)
	}
	return abs, nil
}

func shell() (string, error) {
	names := []string{"pwsh"}
	if runtime.GOOS == "windows" {
		names = append(names, "powershell.exe")
	}
	for _, name := range names {
		if path, err := exec.LookPath(name); err == nil {
			return path, nil
		}
	}
	return "", errors.New("se requiere PowerShell 7 (pwsh) o Windows PowerShell 5.1")
}

func trunc(b []byte) string {
	if len(b) > 65536 {
		return "[salida truncada]\n" + string(b[len(b)-65536:])
	}
	return string(b)
}

type scriptOptions struct{ apply, adoptGentleRouting bool }

func (m Manager) script(ctx context.Context, root, file string, opts scriptOptions) (string, error) {
	if file != "install.ps1" && file != "rollback.ps1" && file != "verify.ps1" {
		return "", errors.New("script no permitido")
	}
	psh, err := shell()
	if err != nil {
		return "", err
	}
	if _, err = ValidateSource(root); err != nil {
		return "", err
	}
	args := []string{"-NoProfile", "-NonInteractive", "-File", filepath.Join(root, "scripts", file), "-CodexHome", m.Home}
	if opts.adoptGentleRouting {
		if file != "install.ps1" {
			return "", errors.New("la adopción solo se permite durante instalación o actualización")
		}
		args = append(args, "-AdoptGentleRouting")
	}
	if opts.apply {
		if file == "verify.ps1" {
			return "", errors.New("verificador de solo lectura")
		}
		args = append(args, "-Apply")
	}
	cmd := exec.CommandContext(ctx, psh, args...)
	cmd.Dir = root
	cmd.Env = append(os.Environ(), "CODEX_HOME="+m.Home)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return trunc(output), fmt.Errorf("%s falló: %w", file, err)
	}
	return trunc(output), nil
}

// executor selects the explicitly trusted checkout or the verified archive of the
// installed profile. A legacy installation has no saved archive.
func (m Manager) executor(installed string) (Prepared, error) {
	if m.Source != "" {
		root, err := m.Local()
		if err != nil {
			return Prepared{}, err
		}
		v, err := ValidateSource(root)
		return Prepared{Root: root, Version: v}, err
	}
	if installed == "" {
		return Prepared{}, errors.New("no existe instalación registrada")
	}
	a, err := m.cached("v" + strings.TrimPrefix(installed, "v"))
	if errors.Is(err, os.ErrNotExist) {
		return Prepared{}, fmt.Errorf("perfil %s sin paquete local; para instalaciones anteriores usá --source RUTA con un checkout de confianza", installed)
	}
	if err != nil {
		return Prepared{}, err
	}
	return a.unpack()
}

func (m Manager) Doctor(ctx context.Context) (string, error) {
	v, err := m.Installed()
	if err != nil {
		return "", err
	}
	p, err := m.executor(v)
	if err != nil {
		return "", err
	}
	defer p.Cleanup()
	return m.script(ctx, p.Root, "verify.ps1", scriptOptions{})
}

func (m Manager) Prepare(ctx context.Context, action string, adoptGentleRouting bool) (Prepared, error) {
	state, err := m.readInstallationState()
	if err != nil {
		return Prepared{}, err
	}
	var p Prepared
	var sameVersion bool
	script := "install.ps1"
	switch action {
	case "local":
		if m.Source == "" {
			return Prepared{}, errors.New("instalación local requiere --source RUTA")
		}
		p, err = m.executor("")
	case "latest":
		rel, e := m.Latest(ctx)
		if e != nil {
			return Prepared{}, e
		}
		if state.Version != "" {
			cmp, e := Compare(rel.Tag, state.Version)
			if e != nil {
				return Prepared{}, e
			}
			sameVersion = cmp == 0
			if cmp < 0 {
				return Prepared{}, fmt.Errorf("sin actualización: instalada %s, última release %s", state.Version, rel.Tag)
			}
		}
		var a *profileArchive
		if sameVersion {
			a, e = m.cached(rel.Tag)
		}
		if !sameVersion || errors.Is(e, os.ErrNotExist) {
			a, e = m.download(ctx, rel.Tag)
		}
		if e != nil {
			return Prepared{}, e
		}
		p, err = a.unpack()
	case "rollback":
		if adoptGentleRouting {
			return Prepared{}, errors.New("rollback no admite adopción de Gentle")
		}
		if state.Version == "" {
			return Prepared{}, errors.New("no existe instalación registrada")
		}
		cmp, _ := Compare(state.Version, "0.1.2")
		if cmp < 0 {
			return Prepared{}, errors.New("rollback de esta versión bloqueado por fallo conocido; recuperá el respaldo manualmente")
		}
		p, err = m.executor(state.Version)
		script = "rollback.ps1"
	default:
		return Prepared{}, errors.New("acción no soportada")
	}
	if err != nil {
		return Prepared{}, err
	}
	p.stateID = state.id
	p.Action = "install"
	if action == "rollback" {
		p.Action = "rollback"
		err = validateRollbackCompatibility(p.Version, state.RollbackSchemaVersion)
	} else {
		cmp, e := Compare(p.Version, "0.2.1")
		if e != nil {
			err = e
		} else if cmp < 0 {
			err = errors.New("instalador anterior a v0.2.1 rechazado por riesgos conocidos")
		}
		if state.Version != "" && err == nil {
			cmp, err = Compare(p.Version, state.Version)
			if err == nil && cmp < 0 {
				err = errors.New("versión anterior a la instalada; utilizá rollback")
			}
		}
	}
	p.adoptGentleRouting = adoptGentleRouting
	if err == nil && sameVersion {
		// An unchanged release may be reapplied only to reconcile a recognized
		// conflict. Ask the installer, which owns the exact marker grammar.
		p.Preview, err = m.script(ctx, p.Root, script, scriptOptions{})
		switch {
		case err == nil:
			p.Cleanup()
			return Prepared{}, fmt.Errorf("sin actualización: la versión %s ya está instalada", p.Version)
		case adoptGentleRouting && strings.Contains(p.Preview, "CHULI_GENTLE_CONFLICT"):
			err = nil
		}
	}
	if err == nil {
		p.Preview, err = m.script(ctx, p.Root, script, scriptOptions{adoptGentleRouting: adoptGentleRouting})
	}
	if err != nil && p.Action == "install" && !adoptGentleRouting && strings.Contains(p.Preview, "CHULI_GENTLE_CONFLICT") {
		err = fmt.Errorf("%w: %w", ErrGentleConflict, err)
	}
	if err != nil {
		p.Cleanup()
		return Prepared{}, fmt.Errorf("preflight rechazado: %w\n%s", err, p.Preview)
	}
	return p, nil
}

func (m Manager) Apply(ctx context.Context, p Prepared) (string, error) {
	current, err := m.readInstallationState()
	if err != nil {
		return "", err
	}
	if p.stateID != current.id {
		return "", errors.New("la instalación cambió desde la revisión previa; revisá nuevamente el plan")
	}
	file := "install.ps1"
	if p.Action == "rollback" {
		file = "rollback.ps1"
	} else if p.Action != "install" {
		return "", errors.New("acción no soportada")
	}
	if p.archive != nil {
		if p.Action == "install" {
			if err := m.remember(p.archive); err != nil {
				return "", fmt.Errorf("no se pudo guardar el paquete para recuperación: %w", err)
			}
		}
		// Never execute an extracted file left over from the preview.
		fresh, err := p.archive.unpack()
		if err != nil {
			return "", err
		}
		defer fresh.Cleanup()
		p.Root = fresh.Root
	}
	log, err := m.script(ctx, p.Root, file, scriptOptions{apply: true, adoptGentleRouting: p.adoptGentleRouting})
	if err != nil {
		return log, err
	}
	if p.Action == "install" {
		out, e := m.script(ctx, p.Root, "verify.ps1", scriptOptions{})
		log += "\n--- Verificación ---\n" + out
		if e != nil {
			return log, fmt.Errorf("instalado pero verificación fallida: %w", e)
		}
	}
	return log, nil
}

func (p Prepared) Cleanup() {
	if p.Temporary && p.Root != "" {
		_ = os.RemoveAll(p.Root)
	}
}
