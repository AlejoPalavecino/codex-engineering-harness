package manager

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"time"
)

const maxModelCatalog = 8 << 20

func validModelID(id string) bool {
	return len(id) <= 80 && component.MatchString(id)
}

func readModelFile(path string) ([]byte, error) {
	info, err := os.Lstat(path)
	if err != nil {
		return nil, err
	}
	if !info.Mode().IsRegular() || info.Size() > maxModelCatalog {
		return nil, errors.New("catálogo de modelos inválido")
	}
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	b, err := io.ReadAll(io.LimitReader(f, maxModelCatalog+1))
	if len(b) > maxModelCatalog {
		return nil, errors.New("catálogo de modelos demasiado grande")
	}
	return b, err
}

type modelOutput struct{ buf bytes.Buffer }

func (b *modelOutput) Write(p []byte) (int, error) {
	if len(p) > maxModelCatalog-b.buf.Len() {
		return 0, errors.New("lista de modelos demasiado grande")
	}
	return b.buf.Write(p)
}

func modelCommand(ctx context.Context, name string, args ...string) ([]byte, error) {
	cmd := exec.CommandContext(ctx, name, args...)
	cmd.WaitDelay = time.Second
	var out modelOutput
	cmd.Stdout = &out
	if err := cmd.Run(); err != nil {
		return nil, err
	}
	return out.buf.Bytes(), nil
}

func discoverCodexModels(home string) []string {
	for _, file := range []string{"models_cache.json", "models_cache.codex_work_desktop.json"} {
		raw, err := readModelFile(filepath.Join(home, file))
		var catalog struct {
			Models []struct{ Slug, Visibility string }
		}
		if err != nil || json.Unmarshal(raw, &catalog) != nil {
			continue
		}
		var ids []string
		seen := make(map[string]bool)
		for _, entry := range catalog.Models {
			if entry.Visibility == "list" && validModelID(entry.Slug) && !seen[entry.Slug] {
				ids = append(ids, entry.Slug)
				seen[entry.Slug] = true
			}
		}
		if len(ids) > 0 {
			return ids
		}
	}
	return nil
}

// AvailableModels reads the client's catalog without changing role assignments.
func (m Manager) AvailableModels(ctx context.Context, spec ClientSpec) ([]string, string) {
	ids, source, _ := m.AvailableModelCatalog(ctx, spec)
	return ids, source
}

// AvailableModelCatalog retains AGY labels alongside exact IDs for settings.json.
func (m Manager) AvailableModelCatalog(ctx context.Context, spec ClientSpec) ([]string, string, map[string]string) {
	home := m.Home
	if home == "" || spec.ID != m.effectiveTarget() {
		home, _ = ResolveClientHome(spec.ID)
	}
	switch spec.ID {
	case "codex":
		if home != "" {
			if ids := discoverCodexModels(home); len(ids) > 0 {
				return ids, "caché de Codex", nil
			}
		}
	case "claude":
		if home != "" {
			if ids := discoverClaudeModels(home); len(ids) > 0 {
				if !slices.Contains(ids, "inherit") {
					ids = append(ids, "inherit")
				}
				return ids, "caché de Claude Code", nil
			}
		}
	case "agy":
		if ids, labels := discoverAGYCatalogUsing(ctx, modelCommand); len(ids) > 0 {
			return ids, "AGY CLI", labels
		}
	}
	return spec.Models, "catálogo incluido en Chuli", nil
}
