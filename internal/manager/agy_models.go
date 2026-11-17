package manager

import (
	"context"
	"encoding/json"
	"fmt"
	"time"
	"unicode"
)

func discoverAGYModels(ctx context.Context) []string {
	return discoverAGYModelsUsing(ctx, modelCommand)
}

func discoverAGYModelsUsing(ctx context.Context, run func(context.Context, string, ...string) ([]byte, error)) []string {
	ids, _ := discoverAGYCatalogUsing(ctx, run)
	return ids
}

func validAGYModelLabel(label string) bool {
	if label == "" || len(label) > 160 {
		return false
	}
	for _, r := range label {
		if !unicode.IsLetter(r) && !unicode.IsNumber(r) && r != ' ' && r != '-' && r != '_' && r != '.' && r != '(' && r != ')' && r != '/' && r != '+' && r != ':' {
			return false
		}
	}
	return true
}

func validAGYSpecialistModel(id string) bool {
	switch id {
	case "pro", "flash", "flash_lite", "inherit", "orchestrator-argon", "orchestrator-carbon", "gemini-4-carbon", "gemini-4-argon", "gemini-4-hydrogen":
		return true
	}
	return false
}

func validateAGYLabelInstaller(version string, roles map[string]RoleModel) error {
	if roles["orchestrator"].Label == "" {
		return nil
	}
	cmp, err := Compare(version, "2.0.2")
	if err != nil || cmp < 0 {
		return fmt.Errorf("la etiqueta exacta AGY requiere actualizar el perfil/instalador a v2.0.2 o posterior antes de aplicar modelos")
	}
	return nil
}

func discoverAGYCatalogUsing(ctx context.Context, run func(context.Context, string, ...string) ([]byte, error)) ([]string, map[string]string) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	raw, err := run(ctx, "agy", "--output-format", "json", "models")
	if err != nil || ctx.Err() != nil {
		return nil, nil
	}
	// AGY's machine-readable response wraps IDs in command.data.models:
	// https://github.com/google-antigravity/antigravity-cli/issues/777
	var response struct {
		Status  string `json:"status"`
		Command struct {
			Name string `json:"name"`
			Data struct {
				Models []struct {
					ID    string `json:"id"`
					Label string `json:"label"`
				} `json:"models"`
			} `json:"data"`
		} `json:"command"`
	}
	if json.Unmarshal(raw, &response) != nil || response.Status != "SUCCESS" || response.Command.Name != "models" {
		return nil, nil
	}
	var ids []string
	seen := make(map[string]bool)
	labels := make(map[string]string)
	for _, model := range response.Command.Data.Models {
		if validModelID(model.ID) && validAGYModelLabel(model.Label) && !seen[model.ID] {
			ids = append(ids, model.ID)
			seen[model.ID] = true
			labels[model.ID] = model.Label
		}
	}
	return ids, labels
}

// resolveAGYRoleLabel snapshots the exact CLI label before preflight. Apply uses
// the resulting role JSON, so a changed catalog cannot change the reviewed plan.
func resolveAGYRoleLabel(ctx context.Context, roles map[string]RoleModel) (map[string]RoleModel, error) {
	orch, ok := roles["orchestrator"]
	if !ok {
		return roles, nil
	}
	if orch.Label != "" {
		if !validAGYModelLabel(orch.Label) {
			return nil, fmt.Errorf("etiqueta AGY inválida para %s", orch.Model)
		}
		return roles, nil
	}
	// Bundled legacy IDs retain their established effort-to-label mapping offline.
	spec, _ := FindClient("agy")
	for _, legacy := range spec.Models {
		if orch.Model == legacy {
			return roles, nil
		}
	}
	_, labels := discoverAGYCatalogUsing(ctx, modelCommand)
	label := labels[orch.Model]
	if label == "" {
		return nil, fmt.Errorf("no se pudo resolver la etiqueta AGY para %s; actualizá el catálogo antes de aplicar", orch.Model)
	}
	cloned := make(map[string]RoleModel, len(roles))
	for role, model := range roles {
		cloned[role] = model
	}
	orch.Label = label
	cloned["orchestrator"] = orch
	return cloned, nil
}
