package manager

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

type SkillInfo struct {
	Name        string
	Description string
	Source      string // "chuli" | "custom"
	Enabled     bool
	Path        string
}

func parseSkillFrontmatter(raw string) (string, string) {
	norm := strings.ReplaceAll(raw, "\r\n", "\n")
	if !strings.HasPrefix(norm, "---\n") {
		return "", ""
	}
	rest := norm[4:]
	end := strings.Index(rest, "\n---\n")
	if end < 0 {
		end = strings.Index(rest, "\n---")
		if end < 0 {
			return "", ""
		}
	}
	header := rest[:end]
	var name, desc string
	var inDescBlock bool
	var descParts []string
	for _, line := range strings.Split(header, "\n") {
		if inDescBlock {
			if strings.HasPrefix(line, "  ") || strings.HasPrefix(line, "\t") {
				descParts = append(descParts, strings.TrimSpace(line))
				continue
			}
			inDescBlock = false
			desc = strings.Join(descParts, " ")
		}
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "name:") {
			name = strings.Trim(strings.TrimSpace(strings.TrimPrefix(trimmed, "name:")), `"'`)
		} else if strings.HasPrefix(trimmed, "description:") {
			val := strings.TrimSpace(strings.TrimPrefix(trimmed, "description:"))
			if val == ">-" || val == ">" || val == "|" || val == "|-" {
				inDescBlock = true
				descParts = nil
			} else {
				desc = strings.Trim(val, `"'`)
			}
		}
	}
	if inDescBlock && len(descParts) > 0 {
		desc = strings.Join(descParts, " ")
	}
	return strings.TrimSpace(name), strings.TrimSpace(desc)
}

func scanSkillsDir(dir, source string, disabledSet map[string]bool) map[string]SkillInfo {
	found := map[string]SkillInfo{}
	entries, err := os.ReadDir(dir)
	if err != nil {
		return found
	}
	for _, entry := range entries {
		if !entry.IsDir() || entry.Name() == "_shared" || !component.MatchString(entry.Name()) {
			continue
		}
		skillFile := filepath.Join(dir, entry.Name(), "SKILL.md")
		raw, err := os.ReadFile(skillFile)
		if err != nil {
			continue
		}
		name, desc := parseSkillFrontmatter(string(raw))
		if name == "" {
			name = entry.Name()
		}
		if desc == "" {
			desc = "Sin descripción"
		}
		found[entry.Name()] = SkillInfo{
			Name:        entry.Name(),
			Description: desc,
			Source:      source,
			Enabled:     !disabledSet[entry.Name()],
			Path:        filepath.Join(dir, entry.Name()),
		}
	}
	return found
}

func clientManagedSkillsDir(target, home string) string {
	if target == "agy" {
		return filepath.Join(home, "plugins", "chuli", "skills")
	}
	return filepath.Join(home, "skills")
}

func clientCustomSkillsDir(target, home string) string {
	return filepath.Join(home, "skills")
}

func (m Manager) chuliCatalogSkills(installed string) (map[string]SkillInfo, func(), error) {
	noop := func() {}
	if m.Source != "" {
		root, err := m.Local()
		if err != nil {
			return nil, noop, err
		}
		return scanSkillsDir(filepath.Join(root, "skills"), "chuli", map[string]bool{}), noop, nil
	}
	if installed != "" {
		a, err := m.cached("v" + strings.TrimPrefix(installed, "v"))
		if err == nil {
			p, err := a.unpack()
			if err == nil {
				skills := scanSkillsDir(filepath.Join(p.Root, "skills"), "chuli", map[string]bool{})
				return skills, p.Cleanup, nil
			}
		}
	}
	return map[string]SkillInfo{}, noop, nil
}

func (m Manager) ListSkills() ([]SkillInfo, error) {
	target := m.effectiveTarget()
	state, err := m.readInstallationState()
	if err != nil {
		return nil, err
	}
	disabledSet := map[string]bool{}
	for _, s := range state.DisabledSkills {
		disabledSet[s] = true
	}
	catalog, cleanup, err := m.chuliCatalogSkills(state.Version)
	if err != nil {
		return nil, err
	}
	defer cleanup()

	managedDir := clientManagedSkillsDir(target, m.Home)
	managedFound := scanSkillsDir(managedDir, "chuli", disabledSet)
	customDir := clientCustomSkillsDir(target, m.Home)
	customFound := scanSkillsDir(customDir, "custom", map[string]bool{})

	merged := map[string]SkillInfo{}
	for name, info := range catalog {
		info.Enabled = !disabledSet[name]
		if _, ok := managedFound[name]; ok && !disabledSet[name] {
			info.Path = managedFound[name].Path
		}
		merged[name] = info
	}
	for name, info := range managedFound {
		if _, ok := merged[name]; !ok {
			if target == "agy" {
				// Any extra folder inside plugins/chuli/skills not in catalog/state
				if _, known := state.ManagedFiles["plugins/chuli/skills/"+name+"/SKILL.md"]; !known && len(catalog) > 0 {
					info.Source = "custom"
				}
			} else if _, known := state.ManagedFiles["skills/"+name+"/SKILL.md"]; !known && len(state.ManagedFiles) > 0 {
				info.Source = "custom"
			}
			merged[name] = info
		}
	}
	if customDir != managedDir {
		for name, info := range customFound {
			if _, ok := merged[name]; !ok {
				merged[name] = info
			}
		}
	}

	out := make([]SkillInfo, 0, len(merged))
	for _, item := range merged {
		out = append(out, item)
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].Source != out[j].Source {
			return out[i].Source == "chuli"
		}
		return out[i].Name < out[j].Name
	})
	return out, nil
}

func (m Manager) AddCustomSkill(sourceDir string) (SkillInfo, error) {
	abs, err := filepath.Abs(strings.TrimSpace(sourceDir))
	if err != nil {
		return SkillInfo{}, err
	}
	info, err := os.Lstat(abs)
	if err != nil || !info.IsDir() || info.Mode()&os.ModeSymlink != 0 {
		return SkillInfo{}, errors.New("indicá una carpeta regular con SKILL.md")
	}
	name := filepath.Base(abs)
	if name == "_shared" || !component.MatchString(name) || reserved.MatchString(name) {
		return SkillInfo{}, fmt.Errorf("nombre de skill inválido: %q", name)
	}
	skillFile := filepath.Join(abs, "SKILL.md")
	raw, err := os.ReadFile(skillFile)
	if err != nil {
		return SkillInfo{}, errors.New("la carpeta debe contener un archivo SKILL.md legible")
	}
	parsedName, desc := parseSkillFrontmatter(string(raw))
	if parsedName == "" || desc == "" {
		return SkillInfo{}, errors.New("SKILL.md debe tener frontmatter YAML con name y description")
	}
	if parsedName != name {
		return SkillInfo{}, fmt.Errorf("el frontmatter name (%q) debe coincidir con la carpeta (%q)", parsedName, name)
	}

	existing, err := m.ListSkills()
	if err != nil {
		return SkillInfo{}, err
	}
	for _, s := range existing {
		if strings.EqualFold(s.Name, name) {
			return SkillInfo{}, fmt.Errorf("ya existe una skill llamada %q; elegí otro nombre", name)
		}
	}

	destRoot := clientCustomSkillsDir(m.effectiveTarget(), m.Home)
	dest := filepath.Join(destRoot, name)
	if err := os.MkdirAll(dest, 0700); err != nil {
		return SkillInfo{}, err
	}
	err = filepath.WalkDir(abs, func(path string, d fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if d.Type()&os.ModeSymlink != 0 {
			return errors.New("no se permiten enlaces simbólicos dentro de la skill")
		}
		rel, err := filepath.Rel(abs, path)
		if err != nil {
			return err
		}
		targetPath := filepath.Join(dest, rel)
		if d.IsDir() {
			return os.MkdirAll(targetPath, 0700)
		}
		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		return os.WriteFile(targetPath, data, 0600)
	})
	if err != nil {
		_ = os.RemoveAll(dest)
		return SkillInfo{}, err
	}
	return SkillInfo{Name: name, Description: desc, Source: "custom", Enabled: true, Path: dest}, nil
}

func (m Manager) RemoveCustomSkill(name string) error {
	name = strings.TrimSpace(name)
	skills, err := m.ListSkills()
	if err != nil {
		return err
	}
	for _, s := range skills {
		if s.Name == name {
			if s.Source != "custom" {
				return errors.New("las skills oficiales de Chuli se desactivan desde la lista sin borrar el catálogo")
			}
			if !strings.HasPrefix(filepath.Clean(s.Path), filepath.Clean(m.Home)+string(os.PathSeparator)) {
				return errors.New("ruta fuera del perfil")
			}
			return os.RemoveAll(s.Path)
		}
	}
	return fmt.Errorf("no se encontró la skill personalizada %q", name)
}
