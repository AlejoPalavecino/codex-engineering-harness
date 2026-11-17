package manager

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
)

// discoverClaudeModels reads the v2 Claude Code catalog in its configuration
// directory. With multiple account catalogs, identity cannot be established
// without reading account configuration, so discovery deliberately abstains.
// Cache schema evidence: https://github.com/yetone/magpie/issues/6418.
func discoverClaudeModels(home string) []string {
	dir := filepath.Join(home, "cache", "model-catalog")
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil
	}
	var candidate string
	for _, entry := range entries {
		if !strings.HasSuffix(entry.Name(), "-cc.json") {
			continue
		}
		if candidate != "" {
			return nil
		}
		candidate = filepath.Join(dir, entry.Name())
	}
	if candidate == "" {
		return nil
	}
	data, err := readModelFile(candidate)
	if err != nil {
		return nil
	}
	var cache struct {
		Version int `json:"version"`
		Catalog struct {
			Surface string `json:"surface"`
			Config  struct {
				ID     string `json:"id"`
				Models []struct {
					ID       string `json:"id"`
					Disabled bool   `json:"disabled"`
				} `json:"models"`
			} `json:"config"`
		} `json:"catalog"`
	}
	if json.Unmarshal(data, &cache) != nil || cache.Version != 2 || cache.Catalog.Surface != "cc" || cache.Catalog.Config.ID != "cc" {
		return nil
	}
	var models []string
	seen := make(map[string]bool)
	for _, model := range cache.Catalog.Config.Models {
		if model.Disabled || !validModelID(model.ID) || seen[model.ID] {
			continue
		}
		seen[model.ID] = true
		models = append(models, model.ID)
	}
	return models
}
