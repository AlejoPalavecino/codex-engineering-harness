package manager

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
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

var roleCatalog = []RoleSpec{
	{ID: "orchestrator", Name: "Orquestador", Summary: "Enrutamiento EWP, autoridad y continuidad"},
	{ID: "chuli_explorer", Name: "Exploración · chuli_explorer", Summary: "Mapa de código y contratos en lectura"},
	{ID: "chuli_worker", Name: "Implementación · chuli_worker", Summary: "Unidad autorizada con pruebas y verificación"},
	{ID: "solution_architect", Name: "Arquitectura · solution_architect", Summary: "Alternativas técnicas y contratos"},
	{ID: "product_designer", Name: "Producto · product_designer", Summary: "Flujos por tarea, estados UX/UI y accesibilidad"},
	{ID: "test_engineer", Name: "Pruebas · test_engineer", Summary: "Casos ejecutables derivados de aceptación y TDD"},
	{ID: "quality_reviewer", Name: "Revisión · quality_reviewer", Summary: "Verificación independiente de candidatos"},
}

var clientCatalog = []ClientSpec{
	{
		ID:            "codex",
		Name:          "OpenAI Codex",
		Short:         "Codex",
		DefaultHome:   ".codex",
		EnvHome:       "CODEX_HOME",
		DefaultPreset: "codex-balanced",
		Models:        []string{"gpt-6.2-sol", "gpt-6.2-astra", "gpt-6.1-sol", "gpt-6.1-terra", "gpt-6-luna", "gpt-6-astra"},
		Efforts:       []string{"low", "medium", "high"},
		Presets: []PresetSpec{
			{
				ID:          "codex-balanced",
				Name:        "Equilibrado (Sol / Terra / Luna)",
				Description: "Configuración recomendada para ingeniería diaria en Codex.",
				Roles: map[string]RoleModel{
					"orchestrator":       {Model: "gpt-6.1-sol", Effort: "medium"},
					"chuli_explorer":     {Model: "gpt-6-luna", Effort: "medium"},
					"chuli_worker":       {Model: "gpt-6.1-sol", Effort: "medium"},
					"solution_architect": {Model: "gpt-6.1-sol", Effort: "high"},
					"product_designer":   {Model: "gpt-6.1-terra", Effort: "medium"},
					"test_engineer":      {Model: "gpt-6.1-sol", Effort: "medium"},
					"quality_reviewer":   {Model: "gpt-6.1-sol", Effort: "high"},
				},
			},
			{
				ID:          "codex-astra-max",
				Name:        "Máxima Capacidad (Astra + Sol High)",
				Description: "Prioriza GPT-6 Astra en decisiones críticas, arquitectura y revisión.",
				Roles: map[string]RoleModel{
					"orchestrator":       {Model: "gpt-6-astra", Effort: "high"},
					"chuli_explorer":     {Model: "gpt-6.1-sol", Effort: "medium"},
					"chuli_worker":       {Model: "gpt-6.1-sol", Effort: "high"},
					"solution_architect": {Model: "gpt-6-astra", Effort: "high"},
					"product_designer":   {Model: "gpt-6.1-terra", Effort: "high"},
					"test_engineer":      {Model: "gpt-6.1-sol", Effort: "high"},
					"quality_reviewer":   {Model: "gpt-6-astra", Effort: "high"},
				},
			},
			{
				ID:          "codex-fast",
				Name:        "Rápido y Económico (Sol + Luna)",
				Description: "Menor latencia y consumo de cuota para cambios acotados.",
				Roles: map[string]RoleModel{
					"orchestrator":       {Model: "gpt-6.1-sol", Effort: "low"},
					"chuli_explorer":     {Model: "gpt-6-luna", Effort: "low"},
					"chuli_worker":       {Model: "gpt-6-luna", Effort: "medium"},
					"solution_architect": {Model: "gpt-6.1-sol", Effort: "medium"},
					"product_designer":   {Model: "gpt-6.1-terra", Effort: "low"},
					"test_engineer":      {Model: "gpt-6-luna", Effort: "medium"},
					"quality_reviewer":   {Model: "gpt-6.1-sol", Effort: "medium"},
				},
			},
		},
	},
	{
		ID:            "claude",
		Name:          "Claude Code",
		Short:         "Claude",
		DefaultHome:   ".claude",
		EnvHome:       "CLAUDE_CONFIG_DIR",
		DefaultPreset: "claude-balanced",
		Models:        []string{"claude-fable-5-2", "claude-opus-5-5", "claude-sonnet-5-5", "claude-haiku-5-5", "inherit", "opus", "sonnet", "haiku"},
		Efforts:       []string{"low", "medium", "high"},
		Presets: []PresetSpec{
			{
				ID:          "claude-balanced",
				Name:        "Equilibrado Pro/Max (Opus 5.5 + Fable 5.2 + Sonnet 5.5)",
				Description: "Opus 5.5 al mando, Fable 5.2 en arquitectura/revisión, Sonnet 5.5 en código y Haiku 5.5 en lectura.",
				Roles: map[string]RoleModel{
					"orchestrator":       {Model: "claude-opus-5-5", Effort: "medium"},
					"chuli_explorer":     {Model: "claude-haiku-5-5", Effort: "low"},
					"chuli_worker":       {Model: "claude-sonnet-5-5", Effort: "medium"},
					"solution_architect": {Model: "claude-fable-5-2", Effort: "high"},
					"product_designer":   {Model: "claude-sonnet-5-5", Effort: "medium"},
					"test_engineer":      {Model: "claude-sonnet-5-5", Effort: "medium"},
					"quality_reviewer":   {Model: "claude-fable-5-2", Effort: "high"},
				},
			},
			{
				ID:          "claude-fable-max",
				Name:        "Máxima Capacidad (Fable 5.2 + Opus 5.5)",
				Description: "Razonamiento profundo con Claude Fable 5.2 y Opus 5.5 en todo el flujo.",
				Roles: map[string]RoleModel{
					"orchestrator":       {Model: "claude-fable-5-2", Effort: "high"},
					"chuli_explorer":     {Model: "claude-sonnet-5-5", Effort: "medium"},
					"chuli_worker":       {Model: "claude-opus-5-5", Effort: "high"},
					"solution_architect": {Model: "claude-fable-5-2", Effort: "high"},
					"product_designer":   {Model: "claude-opus-5-5", Effort: "medium"},
					"test_engineer":      {Model: "claude-opus-5-5", Effort: "medium"},
					"quality_reviewer":   {Model: "claude-fable-5-2", Effort: "high"},
				},
			},
			{
				ID:          "claude-fast",
				Name:        "Rápido y Eficiente (Sonnet 5.5 + Haiku 5.5)",
				Description: "Optimiza velocidad y cuota con Sonnet 5.5 y exploración en Haiku 5.5.",
				Roles: map[string]RoleModel{
					"orchestrator":       {Model: "claude-sonnet-5-5", Effort: "medium"},
					"chuli_explorer":     {Model: "claude-haiku-5-5", Effort: "low"},
					"chuli_worker":       {Model: "claude-sonnet-5-5", Effort: "medium"},
					"solution_architect": {Model: "claude-opus-5-5", Effort: "medium"},
					"product_designer":   {Model: "claude-sonnet-5-5", Effort: "medium"},
					"test_engineer":      {Model: "claude-sonnet-5-5", Effort: "low"},
					"quality_reviewer":   {Model: "claude-opus-5-5", Effort: "high"},
				},
			},
		},
	},
	{
		ID:            "agy",
		Name:          "AGY CLI (Antigravity)",
		Short:         "AGY",
		DefaultHome:   ".gemini/config",
		EnvHome:       "AGY_HOME",
		DefaultPreset: "agy-balanced",
		Models:        []string{"orchestrator-argon", "orchestrator-carbon", "gemini-4-carbon", "gemini-4-argon", "gemini-4-hydrogen", "inherit"},
		Efforts:       []string{"low", "medium", "high", "xhigh"},
		Presets: []PresetSpec{
			{
				ID:          "agy-balanced",
				Name:        "Equilibrado (Orchestrator Argon + Carbon + Hydrogen)",
				Description: "Combina planificación en Argon, arquitectura/revisión en Carbon y ejecución en Hydrogen.",
				Roles: map[string]RoleModel{
					"orchestrator":       {Model: "orchestrator-argon", Effort: "medium"},
					"chuli_explorer":     {Model: "gemini-4-hydrogen", Effort: "medium"},
					"chuli_worker":       {Model: "gemini-4-hydrogen", Effort: "high"},
					"solution_architect": {Model: "gemini-4-carbon", Effort: "high"},
					"product_designer":   {Model: "gemini-4-argon", Effort: "medium"},
					"test_engineer":      {Model: "gemini-4-hydrogen", Effort: "high"},
					"quality_reviewer":   {Model: "gemini-4-carbon", Effort: "high"},
				},
			},
			{
				ID:          "agy-carbon-max",
				Name:        "Máxima Capacidad (Orchestrator Carbon + Carbon High)",
				Description: "Frontera Gemini 4 Carbon en orquestación, arquitectura y revisión crítica.",
				Roles: map[string]RoleModel{
					"orchestrator":       {Model: "orchestrator-carbon", Effort: "high"},
					"chuli_explorer":     {Model: "gemini-4-argon", Effort: "medium"},
					"chuli_worker":       {Model: "gemini-4-argon", Effort: "high"},
					"solution_architect": {Model: "gemini-4-carbon", Effort: "high"},
					"product_designer":   {Model: "gemini-4-carbon", Effort: "medium"},
					"test_engineer":      {Model: "gemini-4-argon", Effort: "high"},
					"quality_reviewer":   {Model: "gemini-4-carbon", Effort: "high"},
				},
			},
			{
				ID:          "agy-hydrogen-fast",
				Name:        "Alta Velocidad (Hydrogen High/Extra + Argon)",
				Description: "Ejecución rápida en Gemini 4 Hydrogen con soporte de Argon en decisiones.",
				Roles: map[string]RoleModel{
					"orchestrator":       {Model: "gemini-4-hydrogen", Effort: "xhigh"},
					"chuli_explorer":     {Model: "gemini-4-hydrogen", Effort: "low"},
					"chuli_worker":       {Model: "gemini-4-hydrogen", Effort: "high"},
					"solution_architect": {Model: "gemini-4-argon", Effort: "high"},
					"product_designer":   {Model: "gemini-4-hydrogen", Effort: "high"},
					"test_engineer":      {Model: "gemini-4-hydrogen", Effort: "medium"},
					"quality_reviewer":   {Model: "gemini-4-argon", Effort: "high"},
				},
			},
		},
	},
}

func Roles() []RoleSpec {
	out := make([]RoleSpec, len(roleCatalog))
	copy(out, roleCatalog)
	return out
}

func Clients() []ClientSpec {
	out := make([]ClientSpec, len(clientCatalog))
	copy(out, clientCatalog)
	return out
}

func FindClient(id string) (ClientSpec, bool) {
	if id == "" {
		id = "codex"
	}
	for _, c := range clientCatalog {
		if c.ID == id {
			return c, true
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
