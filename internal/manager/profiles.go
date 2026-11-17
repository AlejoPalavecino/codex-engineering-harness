package manager

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"time"
)

var safeProfileName = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9 _.-]{0,48}$`)

type SavedProfile struct {
	Client     string               `json:"client"`
	Name       string               `json:"name"`
	RoleModels map[string]RoleModel `json:"role_models"`
	UpdatedAt  string               `json:"updated_at"`
}

func (m Manager) profilesFilePath() (string, error) {
	if err := m.checkCache(true); err != nil {
		return "", err
	}
	return filepath.Join(filepath.Dir(m.cacheDir), "ai-profiles.json"), nil
}

func ValidateRoleModels(clientID string, roles map[string]RoleModel) error {
	spec, ok := FindClient(clientID)
	if !ok {
		return fmt.Errorf("cliente no soportado: %q", clientID)
	}
	if len(roles) == 0 {
		return errors.New("se requiere la configuración de roles")
	}
	validEfforts := map[string]bool{}
	for _, e := range spec.Efforts {
		validEfforts[e] = true
	}
	for _, r := range Roles() {
		rm, ok := roles[r.ID]
		if !ok {
			return fmt.Errorf("falta el rol %s", r.ID)
		}
		if rm.Model == "" || !component.MatchString(rm.Model) || len(rm.Model) > 80 {
			return fmt.Errorf("modelo inválido para %s: %q", r.ID, rm.Model)
		}
		if !validEfforts[rm.Effort] {
			return fmt.Errorf("esfuerzo %q no soportado para %s", rm.Effort, spec.Name)
		}
		if clientID == "agy" && rm.Label != "" && !validAGYModelLabel(rm.Label) {
			return fmt.Errorf("etiqueta AGY inválida para %s", r.ID)
		}
	}
	return nil
}

func (m Manager) loadAllSavedProfiles() ([]SavedProfile, error) {
	path, err := m.profilesFilePath()
	if err != nil {
		return nil, err
	}
	info, err := os.Lstat(path)
	if errors.Is(err, os.ErrNotExist) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	if !info.Mode().IsRegular() {
		return nil, errors.New("ai-profiles.json debe ser un archivo regular")
	}
	raw, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	if len(strings.TrimSpace(string(raw))) == 0 {
		return nil, nil
	}
	var all []SavedProfile
	if err := json.Unmarshal(raw, &all); err != nil {
		return nil, fmt.Errorf("ai-profiles.json inválido: %w", err)
	}
	return all, nil
}

func (m Manager) writeAllSavedProfiles(all []SavedProfile) error {
	path, err := m.profilesFilePath()
	if err != nil {
		return err
	}
	data, err := json.MarshalIndent(all, "", "  ")
	if err != nil {
		return err
	}
	data = append(data, '\n')
	dir := filepath.Dir(path)
	tmp, err := os.CreateTemp(dir, ".ai-profiles-*.tmp")
	if err != nil {
		return err
	}
	tmpPath := tmp.Name()
	cleanup := true
	defer func() {
		if cleanup {
			_ = os.Remove(tmpPath)
		}
	}()
	if _, err := tmp.Write(data); err != nil {
		_ = tmp.Close()
		return err
	}
	if err := tmp.Close(); err != nil {
		return err
	}
	if err := os.Chmod(tmpPath, 0600); err != nil && os.IsPermission(err) {
		return err
	}
	if err := os.Rename(tmpPath, path); err != nil {
		return err
	}
	cleanup = false
	return nil
}

func (m Manager) ListSavedProfiles(clientID string) ([]SavedProfile, error) {
	if _, ok := FindClient(clientID); !ok {
		return nil, fmt.Errorf("cliente no soportado: %q", clientID)
	}
	all, err := m.loadAllSavedProfiles()
	if err != nil {
		return nil, err
	}
	var out []SavedProfile
	for _, p := range all {
		if p.Client == clientID {
			out = append(out, p)
		}
	}
	sort.Slice(out, func(i, j int) bool {
		return strings.ToLower(out[i].Name) < strings.ToLower(out[j].Name)
	})
	return out, nil
}

func (m Manager) SaveProfile(clientID, name string, roles map[string]RoleModel) error {
	name = strings.TrimSpace(name)
	if !safeProfileName.MatchString(name) {
		return errors.New("nombre de perfil inválido (usá letras, números, espacios, -, _ o .)")
	}
	if err := ValidateRoleModels(clientID, roles); err != nil {
		return err
	}
	all, err := m.loadAllSavedProfiles()
	if err != nil {
		return err
	}
	cloned := make(map[string]RoleModel, len(roles))
	for k, v := range roles {
		cloned[k] = v
	}
	entry := SavedProfile{
		Client:     clientID,
		Name:       name,
		RoleModels: cloned,
		UpdatedAt:  time.Now().UTC().Format(time.RFC3339),
	}
	replaced := false
	for i := range all {
		if all[i].Client == clientID && strings.EqualFold(all[i].Name, name) {
			all[i] = entry
			replaced = true
			break
		}
	}
	if !replaced {
		all = append(all, entry)
	}
	return m.writeAllSavedProfiles(all)
}

func (m Manager) DeleteSavedProfile(clientID, name string) error {
	name = strings.TrimSpace(name)
	all, err := m.loadAllSavedProfiles()
	if err != nil {
		return err
	}
	filtered := make([]SavedProfile, 0, len(all))
	removed := false
	for _, p := range all {
		if p.Client == clientID && strings.EqualFold(p.Name, name) {
			removed = true
			continue
		}
		filtered = append(filtered, p)
	}
	if !removed {
		return fmt.Errorf("no existe el perfil %q para %s", name, clientID)
	}
	return m.writeAllSavedProfiles(filtered)
}
