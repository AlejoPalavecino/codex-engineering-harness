package manager

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/AlejoPalavecino/codex-engineering-harness/v2/config"
)

type RoleModel struct {
	Model  string `json:"model"`
	Effort string `json:"effort"`
	Label  string `json:"label,omitempty"`
}

type RoleSpec struct {
	ID      string `json:"id"`
	Name    string `json:"name"`
	Summary string `json:"summary"`
}

type PresetSpec struct {
	ID          string
	Name        string               `json:"name"`
	Description string               `json:"description"`
	Roles       map[string]RoleModel `json:"roles"`
}

type ClientSpec struct {
	ID            string
	Name          string
	Short         string
	DefaultHome   string
	EnvHome       string
	DefaultPreset string
	Models        []string
	Efforts       []string
	Presets       []PresetSpec
}

type orderedEntry struct {
	Key   string
	Value json.RawMessage
}

func decodeOrderedObject(raw []byte) ([]orderedEntry, error) {
	dec := json.NewDecoder(bytes.NewReader(raw))
	tok, err := dec.Token()
	if err != nil {
		return nil, err
	}
	delim, ok := tok.(json.Delim)
	if !ok || delim != '{' {
		return nil, errors.New("expected JSON object")
	}
	var entries []orderedEntry
	seen := make(map[string]bool)
	for dec.More() {
		keyTok, err := dec.Token()
		if err != nil {
			return nil, err
		}
		key, ok := keyTok.(string)
		if !ok {
			return nil, errors.New("expected JSON object key")
		}
		if seen[key] {
			return nil, fmt.Errorf("duplicate JSON key %q", key)
		}
		seen[key] = true
		var val json.RawMessage
		if err := dec.Decode(&val); err != nil {
			return nil, err
		}
		entries = append(entries, orderedEntry{Key: key, Value: val})
	}
	endTok, err := dec.Token()
	if err != nil {
		return nil, err
	}
	if d, ok := endTok.(json.Delim); !ok || d != '}' {
		return nil, errors.New("expected closing brace")
	}
	if _, err := dec.Token(); !errors.Is(err, io.EOF) {
		return nil, errors.New("trailing JSON data")
	}
	return entries, nil
}

func parseCatalog(raw []byte) ([]RoleSpec, []ClientSpec, error) {
	var doc struct {
		SchemaVersion int             `json:"schema_version"`
		Roles         []RoleSpec      `json:"roles"`
		Clients       json.RawMessage `json:"clients"`
	}
	dec := json.NewDecoder(bytes.NewReader(raw))
	dec.DisallowUnknownFields()
	if err := dec.Decode(&doc); err != nil {
		return nil, nil, fmt.Errorf("invalid models catalog: %w", err)
	}
	if _, err := dec.Token(); !errors.Is(err, io.EOF) {
		return nil, nil, errors.New("invalid models catalog: trailing data")
	}
	if doc.SchemaVersion != 1 {
		return nil, nil, fmt.Errorf("unsupported models catalog schema_version %d", doc.SchemaVersion)
	}
	if len(doc.Roles) == 0 {
		return nil, nil, errors.New("models catalog has no roles")
	}
	roleIDs := make(map[string]bool, len(doc.Roles))
	for _, r := range doc.Roles {
		if r.ID == "" || roleIDs[r.ID] {
			return nil, nil, fmt.Errorf("invalid or duplicate role %q", r.ID)
		}
		roleIDs[r.ID] = true
	}
	clientEntries, err := decodeOrderedObject(doc.Clients)
	if err != nil {
		return nil, nil, fmt.Errorf("invalid clients map: %w", err)
	}
	if len(clientEntries) == 0 {
		return nil, nil, errors.New("models catalog has no clients")
	}
	clients := make([]ClientSpec, 0, len(clientEntries))
	for _, ce := range clientEntries {
		var rawClient struct {
			ID            string          `json:"id"`
			Name          string          `json:"name"`
			Short         string          `json:"short"`
			DefaultHome   string          `json:"default_home"`
			EnvHome       string          `json:"env_home"`
			DefaultPreset string          `json:"default_preset"`
			Models        []string        `json:"models"`
			Efforts       []string        `json:"efforts"`
			Presets       json.RawMessage `json:"presets"`
		}
		cdec := json.NewDecoder(bytes.NewReader(ce.Value))
		cdec.DisallowUnknownFields()
		if err := cdec.Decode(&rawClient); err != nil {
			return nil, nil, fmt.Errorf("client %q: %w", ce.Key, err)
		}
		if rawClient.ID != ce.Key || rawClient.Name == "" || rawClient.DefaultHome == "" || rawClient.EnvHome == "" {
			return nil, nil, fmt.Errorf("client %q: invalid identity or home fields", ce.Key)
		}
		presetEntries, err := decodeOrderedObject(rawClient.Presets)
		if err != nil {
			return nil, nil, fmt.Errorf("client %q presets: %w", ce.Key, err)
		}
		if len(presetEntries) == 0 {
			return nil, nil, fmt.Errorf("client %q has no presets", ce.Key)
		}
		presets := make([]PresetSpec, 0, len(presetEntries))
		hasDefault := false
		for _, pe := range presetEntries {
			var preset PresetSpec
			pdec := json.NewDecoder(bytes.NewReader(pe.Value))
			pdec.DisallowUnknownFields()
			if err := pdec.Decode(&preset); err != nil {
				return nil, nil, fmt.Errorf("client %q preset %q: %w", ce.Key, pe.Key, err)
			}
			preset.ID = pe.Key
			if preset.ID == rawClient.DefaultPreset {
				hasDefault = true
			}
			for roleID := range roleIDs {
				if rm, ok := preset.Roles[roleID]; !ok || rm.Model == "" || rm.Effort == "" {
					return nil, nil, fmt.Errorf("client %q preset %q missing role %q", ce.Key, pe.Key, roleID)
				}
			}
			presets = append(presets, preset)
		}
		if !hasDefault {
			return nil, nil, fmt.Errorf("client %q default_preset %q not found", ce.Key, rawClient.DefaultPreset)
		}
		clients = append(clients, ClientSpec{
			ID:            rawClient.ID,
			Name:          rawClient.Name,
			Short:         rawClient.Short,
			DefaultHome:   rawClient.DefaultHome,
			EnvHome:       rawClient.EnvHome,
			DefaultPreset: rawClient.DefaultPreset,
			Models:        rawClient.Models,
			Efforts:       rawClient.Efforts,
			Presets:       presets,
		})
	}
	return doc.Roles, clients, nil
}

var (
	roleCatalog   []RoleSpec
	clientCatalog []ClientSpec
)

func init() {
	roles, clients, err := parseCatalog(config.ModelsJSON)
	if err != nil {
		panic("chuli: embedded config/models.json invalid: " + err.Error())
	}
	roleCatalog = roles
	clientCatalog = clients
}

func cloneClientSpec(c ClientSpec) ClientSpec {
	out := c
	out.Models = append([]string(nil), c.Models...)
	out.Efforts = append([]string(nil), c.Efforts...)
	out.Presets = make([]PresetSpec, len(c.Presets))
	for i, p := range c.Presets {
		roles := make(map[string]RoleModel, len(p.Roles))
		for k, v := range p.Roles {
			roles[k] = v
		}
		p.Roles = roles
		out.Presets[i] = p
	}
	return out
}

func Roles() []RoleSpec {
	out := make([]RoleSpec, len(roleCatalog))
	copy(out, roleCatalog)
	return out
}

func Clients() []ClientSpec {
	out := make([]ClientSpec, len(clientCatalog))
	for i, c := range clientCatalog {
		out[i] = cloneClientSpec(c)
	}
	return out
}

func FindClient(id string) (ClientSpec, bool) {
	if id == "" {
		id = "codex"
	}
	for _, c := range clientCatalog {
		if c.ID == id {
			return cloneClientSpec(c), true
		}
	}
	return ClientSpec{}, false
}

func DefaultRolesForPreset(clientID, presetID string) (PresetSpec, map[string]RoleModel, error) {
	c, ok := FindClient(clientID)
	if !ok {
		return PresetSpec{}, nil, fmt.Errorf("cliente no soportado: %q", clientID)
	}
	if presetID == "" {
		presetID = c.DefaultPreset
	}
	for _, p := range c.Presets {
		if p.ID == presetID {
			roles := make(map[string]RoleModel, len(p.Roles))
			for k, v := range p.Roles {
				roles[k] = v
			}
			return p, roles, nil
		}
	}
	return PresetSpec{}, nil, fmt.Errorf("preset %q no reconocido para %s", presetID, c.Name)
}

func ResolveClientHome(clientID string) (string, error) {
	c, ok := FindClient(clientID)
	if !ok {
		return "", fmt.Errorf("cliente no soportado: %q", clientID)
	}
	home := os.Getenv(c.EnvHome)
	if home == "" {
		user, err := os.UserHomeDir()
		if err != nil {
			return "", err
		}
		home = filepath.Join(user, filepath.FromSlash(c.DefaultHome))
	}
	if !filepath.IsAbs(home) {
		return "", fmt.Errorf("%s debe ser una ruta absoluta", c.EnvHome)
	}
	return filepath.Abs(home)
}

func ParseRoleModelSpecs(clientID, presetID string, specs []string) (map[string]RoleModel, error) {
	c, ok := FindClient(clientID)
	if !ok {
		return nil, fmt.Errorf("cliente no soportado: %q", clientID)
	}
	_, roles, err := DefaultRolesForPreset(clientID, presetID)
	if err != nil {
		return nil, err
	}
	validRoles := map[string]bool{}
	for _, r := range roleCatalog {
		validRoles[r.ID] = true
	}
	validEfforts := map[string]bool{}
	for _, e := range c.Efforts {
		validEfforts[e] = true
	}
	for _, raw := range specs {
		parts := strings.SplitN(strings.TrimSpace(raw), "=", 2)
		if len(parts) != 2 {
			return nil, fmt.Errorf("formato inválido %q; usá ROL=MODELO[:ESFUERZO]", raw)
		}
		role := strings.TrimSpace(parts[0])
		if !validRoles[role] {
			return nil, fmt.Errorf("rol desconocido: %q", role)
		}
		modelPart := strings.TrimSpace(parts[1])
		effort := roles[role].Effort
		if strings.Contains(modelPart, ":") {
			mp := strings.SplitN(modelPart, ":", 2)
			modelPart, effort = strings.TrimSpace(mp[0]), strings.ToLower(strings.TrimSpace(mp[1]))
		}
		if modelPart == "" || !component.MatchString(modelPart) || len(modelPart) > 80 {
			return nil, fmt.Errorf("modelo inválido para %s: %q", role, modelPart)
		}
		if !validEfforts[effort] {
			return nil, fmt.Errorf("esfuerzo %q no soportado en %s", effort, c.Name)
		}
		if clientID == "agy" && role != "orchestrator" && !validAGYSpecialistModel(modelPart) {
			return nil, fmt.Errorf("modelo AGY para %s requiere un nivel: pro, flash, flash_lite o inherit", role)
		}
		roles[role] = RoleModel{Model: modelPart, Effort: effort}
	}
	return roles, nil
}

func EncodeRoleModels(roles map[string]RoleModel) (string, error) {
	if len(roles) == 0 {
		return "", nil
	}
	b, err := json.Marshal(roles)
	if err != nil {
		return "", errors.New("no se pudo codificar la matriz de modelos")
	}
	return string(b), nil
}
