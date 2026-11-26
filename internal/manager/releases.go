package manager

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"
)

const (
	githubReleasesAPI  = "https://api.github.com/repos/" + Repository + "/releases?per_page="
	githubLatestWebURL = "https://github.com/" + Repository + "/releases/latest"
)

type Release struct {
	Tag         string
	URL         string
	Name        string
	Body        string
	PublishedAt string
}

type githubReleaseJSON struct {
	Tag         string `json:"tag_name"`
	URL         string `json:"html_url"`
	Name        string `json:"name"`
	Body        string `json:"body"`
	PublishedAt string `json:"published_at"`
	Draft       bool   `json:"draft"`
	Prerelease  bool   `json:"prerelease"`
}

func toRelease(v githubReleaseJSON) (Release, error) {
	if !releaseTag.MatchString(v.Tag) || v.Draft || v.Prerelease {
		return Release{}, errors.New("release no estable o tag no permitido")
	}
	return Release{
		Tag:         v.Tag,
		URL:         v.URL,
		Name:        strings.TrimSpace(v.Name),
		Body:        strings.ReplaceAll(v.Body, "\r\n", "\n"),
		PublishedAt: strings.TrimSpace(v.PublishedAt),
	}, nil
}

type githubHTTPStatusError struct {
	StatusCode  int
	RateLimited bool
	ResetAt     time.Time
}

func (e *githubHTTPStatusError) Error() string {
	if e.RateLimited {
		when := "en unos minutos"
		if !e.ResetAt.IsZero() && e.ResetAt.After(time.Now()) {
			when = "a las " + e.ResetAt.Local().Format("15:04")
		}
		return fmt.Sprintf("GitHub API limitó las consultas sin autenticar (HTTP %d); probá de nuevo %s o configurá GITHUB_TOKEN", e.StatusCode, when)
	}
	return fmt.Sprintf("GitHub Releases: HTTP %d", e.StatusCode)
}

func githubAPIToken() string {
	for _, env := range []string{"GITHUB_TOKEN", "GH_TOKEN"} {
		if token := strings.TrimSpace(os.Getenv(env)); token != "" {
			return token
		}
	}
	return ""
}

func (m Manager) fetchJSON(ctx context.Context, url string, limit int64, dst any) error {
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return err
	}
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("User-Agent", "Chuli-Manager")
	if req.URL.Scheme == "https" && req.URL.Hostname() == "api.github.com" {
		if token := githubAPIToken(); token != "" {
			req.Header.Set("Authorization", "Bearer "+token)
		}
	}
	resp, err := m.httpClient().Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 4<<10))
		rateLimited := resp.StatusCode == http.StatusTooManyRequests
		if resp.StatusCode == http.StatusForbidden {
			if resp.Header.Get("X-RateLimit-Remaining") == "0" || strings.Contains(strings.ToLower(string(body)), "rate limit") {
				rateLimited = true
			}
		}
		var resetAt time.Time
		if rawReset := strings.TrimSpace(resp.Header.Get("X-RateLimit-Reset")); rawReset != "" {
			if epoch, convErr := strconv.ParseInt(rawReset, 10, 64); convErr == nil && epoch > 0 {
				resetAt = time.Unix(epoch, 0)
			}
		}
		return &githubHTTPStatusError{
			StatusCode:  resp.StatusCode,
			RateLimited: rateLimited,
			ResetAt:     resetAt,
		}
	}
	return json.NewDecoder(io.LimitReader(resp.Body, limit)).Decode(dst)
}

func (m Manager) latestViaWebRedirect(ctx context.Context) (Release, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", githubLatestWebURL, nil)
	if err != nil {
		return Release{}, err
	}
	req.Header.Set("User-Agent", "Chuli-Manager")
	client := *m.httpClient()
	client.CheckRedirect = func(*http.Request, []*http.Request) error {
		return http.ErrUseLastResponse
	}
	resp, err := client.Do(req)
	if err != nil {
		return Release{}, err
	}
	defer resp.Body.Close()
	switch resp.StatusCode {
	case http.StatusMovedPermanently, http.StatusFound, http.StatusSeeOther, http.StatusTemporaryRedirect, http.StatusPermanentRedirect:
	default:
		return Release{}, fmt.Errorf("GitHub latest redirect: HTTP %d", resp.StatusCode)
	}
	loc, err := resp.Location()
	if err != nil {
		return Release{}, err
	}
	if loc.Scheme != "https" || loc.Hostname() != "github.com" {
		return Release{}, errors.New("redirección de release fuera de github.com")
	}
	prefix := "/" + Repository + "/releases/tag/"
	if !strings.HasPrefix(loc.Path, prefix) {
		return Release{}, errors.New("ruta de release inesperada")
	}
	tag := strings.TrimPrefix(loc.Path, prefix)
	if strings.Contains(tag, "/") || !releaseTag.MatchString(tag) {
		return Release{}, fmt.Errorf("tag de release inválido en redirección: %q", tag)
	}
	return Release{
		Tag:  tag,
		Name: tag,
		URL:  "https://github.com/" + Repository + "/releases/tag/" + tag,
	}, nil
}

func (m Manager) Latest(ctx context.Context) (Release, error) {
	var raw githubReleaseJSON
	if err := m.fetchJSON(ctx, githubAPI, 256<<10, &raw); err != nil {
		var statusErr *githubHTTPStatusError
		if errors.As(err, &statusErr) && (statusErr.StatusCode == http.StatusForbidden || statusErr.StatusCode == http.StatusTooManyRequests) {
			if fallback, fbErr := m.latestViaWebRedirect(ctx); fbErr == nil {
				return fallback, nil
			}
		}
		return Release{}, err
	}
	return toRelease(raw)
}

func (m Manager) ListReleases(ctx context.Context, limit int) ([]Release, error) {
	if limit <= 0 {
		limit = 10
	}
	if limit > 30 {
		limit = 30
	}
	var raw []githubReleaseJSON
	if err := m.fetchJSON(ctx, githubReleasesAPI+strconv.Itoa(limit), 512<<10, &raw); err != nil {
		return nil, err
	}
	var releases []Release
	for _, entry := range raw {
		rel, err := toRelease(entry)
		if err != nil {
			continue
		}
		if cmp, err := Compare(rel.Tag, "0.2.1"); err != nil || cmp < 0 {
			continue
		}
		releases = append(releases, rel)
	}
	sort.SliceStable(releases, func(i, j int) bool {
		cmp, err := Compare(releases[i].Tag, releases[j].Tag)
		return err == nil && cmp > 0
	})
	return releases, nil
}

func (m Manager) CachedTags() ([]string, error) {
	if m.cacheDir == "" {
		return nil, nil
	}
	entries, err := os.ReadDir(m.cacheDir)
	if errors.Is(err, os.ErrNotExist) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	var tags []string
	for _, entry := range entries {
		if !entry.IsDir() || !releaseTag.MatchString(entry.Name()) {
			continue
		}
		if _, err := m.cached(entry.Name()); err == nil {
			tags = append(tags, entry.Name())
		}
	}
	sort.SliceStable(tags, func(i, j int) bool {
		cmp, err := Compare(tags[i], tags[j])
		return err == nil && cmp > 0
	})
	return tags, nil
}

var mdLink = regexp.MustCompile(`\[([^\]]+)\]\([^)]+\)`)
var htmlTag = regexp.MustCompile(`<[^>]+>`)

func SummarizeReleaseNotes(body string, maxLines int) string {
	if maxLines <= 0 {
		maxLines = 12
	}
	normalized := strings.ReplaceAll(body, "\r\n", "\n")
	var lines []string
	for _, raw := range strings.Split(normalized, "\n") {
		line := strings.TrimSpace(raw)
		if line == "" || strings.HasPrefix(line, "<!--") || strings.HasPrefix(line, "---") {
			continue
		}
		line = htmlTag.ReplaceAllString(line, "")
		line = mdLink.ReplaceAllString(line, "$1")
		line = strings.ReplaceAll(line, "**", "")
		line = strings.ReplaceAll(line, "__", "")
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		if strings.HasPrefix(line, "#") {
			line = strings.TrimSpace(strings.TrimLeft(line, "#"))
			if line != "" {
				lines = append(lines, "• "+line)
			}
			continue
		}
		if strings.HasPrefix(line, "- ") || strings.HasPrefix(line, "* ") {
			line = "• " + strings.TrimSpace(line[2:])
		}
		lines = append(lines, line)
		if len(lines) >= maxLines {
			break
		}
	}
	if len(lines) == 0 {
		return "Esta versión incluye mejoras y correcciones del harness (sin notas detalladas)."
	}
	return strings.Join(lines, "\n")
}

type ClientStatus struct {
	Spec             ClientSpec
	Home             string
	HomeExists       bool
	CommandFound     bool
	CommandPath      string
	Detected         bool
	InstalledVersion string
	Preset           string
	RoleModels       map[string]RoleModel
	DisabledSkills   []string
	StateError       string
}

func (m Manager) ForTarget(target string) (Manager, error) {
	if target == "" {
		target = "codex"
	}
	home, err := ResolveClientHome(target)
	if err != nil {
		return Manager{}, err
	}
	next := m
	next.Target = target
	next.Home = home
	spec, _ := FindClient(target)
	next.Preset = spec.DefaultPreset
	_, roles, _ := DefaultRolesForPreset(spec.ID, spec.DefaultPreset)
	next.RoleModels = roles
	next.DisabledSkills = nil
	return next, nil
}

func clientHomeDetected(id, home string) bool {
	if _, err := os.Stat(home); err == nil {
		return true
	}
	switch id {
	case "agy":
		if filepath.Base(home) == "config" && filepath.Base(filepath.Dir(home)) == ".gemini" {
			if _, err := os.Stat(filepath.Join(filepath.Dir(home), "antigravity-cli")); err == nil {
				return true
			}
		}
	case "claude":
		if _, err := os.Stat(filepath.Join(filepath.Dir(home), ".claude.json")); err == nil {
			return true
		}
	}
	return false
}

func (m Manager) DetectClients() []ClientStatus {
	var out []ClientStatus
	for _, spec := range Clients() {
		st := ClientStatus{Spec: spec}
		home, err := ResolveClientHome(spec.ID)
		if err != nil {
			st.StateError = err.Error()
			out = append(out, st)
			continue
		}
		st.Home = home
		st.HomeExists = clientHomeDetected(spec.ID, home)
		if path, err := exec.LookPath(spec.ID); err == nil {
			st.CommandFound = true
			st.CommandPath = path
		}
		cm := Manager{Home: home, Source: m.Source, Client: m.Client, Target: spec.ID, cacheDir: m.cacheDir}
		state, err := cm.readInstallationState()
		if err != nil {
			st.StateError = err.Error()
		} else {
			st.InstalledVersion = state.Version
			st.Preset = state.Preset
			st.RoleModels = state.RoleModels
			st.DisabledSkills = state.DisabledSkills
		}
		st.Detected = st.HomeExists || st.CommandFound || st.InstalledVersion != ""
		out = append(out, st)
	}
	return out
}

type ClientUpdateStatus struct {
	ID           string
	Name         string
	Installed    string
	HasInstall   bool
	HasUpdate    bool
	Preset       string
	DisabledList []string
}

type UpdateReport struct {
	Latest           Release
	Summary          string
	ManagerCurrent   string
	ManagerNewer     bool
	Clients          []ClientUpdateStatus
	OutdatedCount    int
	AnyProfileUpdate bool
}

func baseSemver(v string) string {
	v = strings.TrimSpace(strings.TrimPrefix(v, "v"))
	if idx := strings.IndexAny(v, "-+"); idx > 0 {
		v = v[:idx]
	}
	return v
}

func (m Manager) CheckUpdates(ctx context.Context, managerVersion string) (UpdateReport, error) {
	latest, err := m.Latest(ctx)
	if err != nil {
		return UpdateReport{}, err
	}
	report := UpdateReport{
		Latest:         latest,
		Summary:        SummarizeReleaseNotes(latest.Body, 14),
		ManagerCurrent: managerVersion,
	}
	if cur := baseSemver(managerVersion); cur != "" {
		if cmp, err := Compare(latest.Tag, cur); err == nil && cmp > 0 {
			report.ManagerNewer = true
		}
	}
	for _, st := range m.DetectClients() {
		cu := ClientUpdateStatus{
			ID:           st.Spec.ID,
			Name:         st.Spec.Short,
			Installed:    st.InstalledVersion,
			HasInstall:   st.InstalledVersion != "",
			Preset:       st.Preset,
			DisabledList: st.DisabledSkills,
		}
		if cu.HasInstall {
			if cmp, err := Compare(latest.Tag, cu.Installed); err == nil && cmp > 0 {
				cu.HasUpdate = true
				report.OutdatedCount++
				report.AnyProfileUpdate = true
			}
		}
		report.Clients = append(report.Clients, cu)
	}
	return report, nil
}
