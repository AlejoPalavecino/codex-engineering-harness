package manager

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func TestDiscoverClaudeModels(t *testing.T) {
	const catalog = `{"version":2,"catalog":{"surface":"cc","config":{"id":"cc","models":[{"id":"claude-sonnet-5-5"},{"id":"claude-opus-5-5","disabled":true},{"id":"claude-sonnet-5-5"},{"id":"bad id"},{"id":"claude-haiku-5-5"}]}}}`
	for _, tt := range []struct {
		name  string
		files map[string]string
		want  []string
	}{
		{"missing", nil, nil},
		{"version two catalog", map[string]string{"org-account-cc.json": catalog}, []string{"claude-sonnet-5-5", "claude-haiku-5-5"}},
		{"invalid JSON", map[string]string{"org-account-cc.json": `{`}, nil},
		{"unknown version", map[string]string{"org-account-cc.json": `{"version":3,"catalog":{"surface":"cc","config":{"id":"cc","models":[{"id":"claude-sonnet-5-5"}]}}}`}, nil},
		{"wrong surface", map[string]string{"org-account-cc.json": `{"version":2,"catalog":{"surface":"web","config":{"id":"cc","models":[{"id":"claude-sonnet-5-5"}]}}}`}, nil},
		{"wrong config", map[string]string{"org-account-cc.json": `{"version":2,"catalog":{"surface":"cc","config":{"id":"web","models":[{"id":"claude-sonnet-5-5"}]}}}`}, nil},
		{"multiple organizations", map[string]string{"org1-account-cc.json": catalog, "org2-account-cc.json": catalog}, nil},
		{"invalid second catalog remains ambiguous", map[string]string{"org1-account-cc.json": catalog, "org2-account-cc.json": `{`}, nil},
		{"unrelated file ignored", map[string]string{"org-account-cc.json": catalog, "published-floor.json": `{`, "org-account-web.json": catalog}, []string{"claude-sonnet-5-5", "claude-haiku-5-5"}},
	} {
		t.Run(tt.name, func(t *testing.T) {
			home := t.TempDir()
			dir := filepath.Join(home, "cache", "model-catalog")
			if err := os.MkdirAll(dir, 0o700); err != nil {
				t.Fatal(err)
			}
			for name, data := range tt.files {
				if err := os.WriteFile(filepath.Join(dir, name), []byte(data), 0o600); err != nil {
					t.Fatal(err)
				}
			}
			if got := discoverClaudeModels(home); !reflect.DeepEqual(got, tt.want) {
				t.Fatalf("discoverClaudeModels() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestDiscoverClaudeModelsRejectsNonRegularCatalog(t *testing.T) {
	home := t.TempDir()
	if err := os.MkdirAll(filepath.Join(home, "cache", "model-catalog", "org-account-cc.json"), 0o700); err != nil {
		t.Fatal(err)
	}
	if got := discoverClaudeModels(home); got != nil {
		t.Fatalf("directory returned models: %v", got)
	}
}
