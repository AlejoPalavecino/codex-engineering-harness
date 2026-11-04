// Package eval defines reproducible, offline Chuli behavior evaluations.
// The evaluator never starts Codex, delegates agents or edits their configuration.
package eval

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
)

const Schema = 1

var caseID = regexp.MustCompile("^[a-z][a-z0-9-]*$")
var commitSHA = regexp.MustCompile("^[0-9a-f]{40}$")
var versionTag = regexp.MustCompile("^v[0-9]+[.][0-9]+[.][0-9]+$")

type Manifest struct {
	Schema      int      `json:"schema"`
	Suite       string   `json:"suite"`
	BaselineTag string   `json:"baseline_tag"`
	BaselineSHA string   `json:"baseline_sha"`
	Cases       []string `json:"cases"`
}

type Criterion struct {
	ID          string `json:"id"`
	Description string `json:"description"`
	Severity    string `json:"severity"` // critical | important
	Method      string `json:"method"`   // snapshot | artifact | review
}

type Case struct {
	ID       string      `json:"id"`
	Title    string      `json:"title"`
	Route    string      `json:"route"`
	Fixture  string      `json:"fixture"`
	Approval string      `json:"approval"` // required | not_required | already_approved
	Setup    string      `json:"setup"`
	Prompt   string      `json:"prompt"`
	Criteria []Criterion `json:"criteria"`
}

type Catalog struct {
	Root        string
	Manifest    Manifest
	Cases       []Case
	Fingerprint string
}

func decodeJSON(path string, destination any) error {
	f, err := os.Open(path)
	if err != nil {
		return fmt.Errorf("%s: %w", filepath.Base(path), err)
	}
	defer f.Close()
	d := json.NewDecoder(io.LimitReader(f, 4<<20))
	d.DisallowUnknownFields()
	if err := d.Decode(destination); err != nil {
		return fmt.Errorf("%s: %w", filepath.Base(path), err)
	}
	var extra any
	if err := d.Decode(&extra); !errors.Is(err, io.EOF) {
		return fmt.Errorf("%s: trailing JSON data", filepath.Base(path))
	}
	return nil
}

func validPath(path string) bool {
	return path != "" && path == filepath.ToSlash(filepath.Clean(path)) &&
		!filepath.IsAbs(path) && !strings.HasPrefix(path, "../") &&
		path != ".." && !strings.Contains(path, ":") && !strings.Contains(path, "`") &&
		!strings.Contains(path, "\\") && !strings.Contains(path, "//")
}

func Load(root string) (Catalog, error) {
	abs, err := filepath.Abs(root)
	if err != nil {
		return Catalog{}, err
	}
	var m Manifest
	if err := decodeJSON(filepath.Join(abs, "manifest.json"), &m); err != nil {
		return Catalog{}, err
	}
	if m.Schema != Schema || m.Suite == "" || !versionTag.MatchString(m.BaselineTag) ||
		!commitSHA.MatchString(m.BaselineSHA) || len(m.Cases) == 0 {
		return Catalog{}, errors.New("invalid evaluation manifest: schema, baseline tag/SHA, suite or cases")
	}
	cases := make([]Case, 0, len(m.Cases))
	seen := make(map[string]bool)
	for _, id := range m.Cases {
		if !caseID.MatchString(id) || seen[id] {
			return Catalog{}, fmt.Errorf("invalid or repeated case ID: %q", id)
		}
		seen[id] = true
		var c Case
		if err := decodeJSON(filepath.Join(abs, "cases", id+".json"), &c); err != nil {
			return Catalog{}, err
		}
		if c.ID != id || strings.TrimSpace(c.Title) == "" ||
			strings.TrimSpace(c.Prompt) == "" || strings.TrimSpace(c.Setup) == "" ||
			!validPath(c.Fixture) || !strings.HasPrefix(c.Fixture, "fixtures/") {
			return Catalog{}, fmt.Errorf("%s: invalid identity, prompt, setup or fixture", id)
		}
		if c.Route != "R0" && c.Route != "R1" && c.Route != "R2" && c.Route != "R2+C" {
			return Catalog{}, fmt.Errorf("%s: unsupported EWP route %q", id, c.Route)
		}
		if c.Approval != "required" && c.Approval != "not_required" && c.Approval != "already_approved" {
			return Catalog{}, fmt.Errorf("%s: invalid approval mode", id)
		}
		stat, err := os.Stat(filepath.Join(abs, filepath.FromSlash(c.Fixture)))
		if err != nil || !stat.IsDir() {
			return Catalog{}, fmt.Errorf("%s: missing fixture %q", id, c.Fixture)
		}
		if len(c.Criteria) < 2 {
			return Catalog{}, fmt.Errorf("%s: at least two independent criteria required", id)
		}
		criteria := make(map[string]bool)
		critical := false
		for _, cr := range c.Criteria {
			if !caseID.MatchString(cr.ID) || criteria[cr.ID] || strings.TrimSpace(cr.Description) == "" {
				return Catalog{}, fmt.Errorf("%s: invalid or repeated criterion %q", id, cr.ID)
			}
			criteria[cr.ID] = true
			if cr.Severity != "critical" && cr.Severity != "important" {
				return Catalog{}, fmt.Errorf("%s: criterion %s has invalid severity", id, cr.ID)
			}
			if cr.Severity == "critical" {
				critical = true
			}
			if cr.Method != "snapshot" && cr.Method != "artifact" && cr.Method != "review" {
				return Catalog{}, fmt.Errorf("%s: criterion %s has invalid method", id, cr.ID)
			}
		}
		if !critical {
			return Catalog{}, fmt.Errorf("%s: needs at least one critical criterion", id)
		}
		cases = append(cases, c)
	}
	// Unlisted case files must not silently escape the test suite.
	entries, err := os.ReadDir(filepath.Join(abs, "cases"))
	if err != nil {
		return Catalog{}, err
	}
	for _, entry := range entries {
		if entry.IsDir() {
			return Catalog{}, fmt.Errorf("unexpected case directory: %s", entry.Name())
		}
		if strings.HasSuffix(entry.Name(), ".json") &&
			!seen[strings.TrimSuffix(entry.Name(), ".json")] {
			return Catalog{}, fmt.Errorf("case not declared in manifest: %s", entry.Name())
		}
	}
	fp, err := Fingerprint(abs)
	if err != nil {
		return Catalog{}, err
	}
	return Catalog{Root: abs, Manifest: m, Cases: cases, Fingerprint: fp}, nil
}

func Fingerprint(root string) (string, error) {
	// Include only catalog definitions, rubrics and synthetic fixtures.
	// Reports, README and local outputs never affect the reproducibility key.
	var names []string
	for _, sub := range []string{"manifest.json", "cases", "fixtures", "rubrics"} {
		p := filepath.Join(root, sub)
		err := filepath.WalkDir(p, func(path string, entry os.DirEntry, err error) error {
			if err != nil {
				return err
			}
			if entry.Type()&os.ModeSymlink != 0 {
				return fmt.Errorf("symlink in evaluation inputs: %s", entry.Name())
			}
			if entry.IsDir() {
				return nil
			}
			if !entry.Type().IsRegular() {
				return fmt.Errorf("not a regular evaluation input: %s", entry.Name())
			}
			rel, err := filepath.Rel(root, path)
			if err != nil {
				return err
			}
			names = append(names, filepath.ToSlash(rel))
			return nil
		})
		if err != nil {
			return "", err
		}
	}
	sort.Strings(names)
	h := sha256.New()
	for _, name := range names {
		// Delimit fields to rule out concatenation collisions.
		_, _ = io.WriteString(h, name+"\x00")
		data, err := os.ReadFile(filepath.Join(root, filepath.FromSlash(name)))
		if err != nil {
			return "", err
		}
		_, _ = h.Write(data)
		_, _ = io.WriteString(h, "\x00")
	}
	return hex.EncodeToString(h.Sum(nil)), nil
}
