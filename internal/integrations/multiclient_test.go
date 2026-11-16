package integrations

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestMultiClientCustomMCPListApplyAndUndo(t *testing.T) {
	ctx := context.Background()
	for _, client := range []string{"codex", "agy", "claude"} {
		t.Run(client, func(t *testing.T) {
			home := canonicalTestHome(t)
			s := Service{Home: home, Client: client}

			plan, err := s.PrepareCustom(ctx, CustomMCP{
				ID:       "mi-api",
				URL:      "https://mcp.ejemplo.test/mcp",
				TokenEnv: "MI_API_KEY",
			})
			if err != nil {
				t.Fatal(err)
			}
			if _, err := s.Apply(ctx, plan); err != nil {
				t.Fatal(err)
			}

			servers, err := s.ListConfigured(ctx)
			if err != nil {
				t.Fatal(err)
			}
			if len(servers) != 1 || servers[0].ID != "mi-api" || !servers[0].Managed || servers[0].TokenEnv != "MI_API_KEY" {
				t.Fatalf("unexpected listed servers for %s: %+v", client, servers)
			}

			undoPlan, err := s.PrepareRemoveCustom(ctx, servers[0])
			if err != nil {
				t.Fatal(err)
			}
			if _, err := s.Apply(ctx, undoPlan); err != nil {
				t.Fatal(err)
			}
			servers, err = s.ListConfigured(ctx)
			if err != nil {
				t.Fatal(err)
			}
			if len(servers) != 0 {
				t.Fatalf("expected 0 servers after undo, got %+v", servers)
			}
		})
	}

	// Verify AGY preserves existing servers inside mcp_config.json
	agyHome := canonicalTestHome(t)
	existing := map[string]any{"mcpServers": map[string]any{"pi-bridge": map[string]any{"serverUrl": "http://127.0.0.1:64512/mcp"}}}
	raw, _ := json.Marshal(existing)
	if err := os.WriteFile(filepath.Join(agyHome, "mcp_config.json"), raw, 0600); err != nil {
		t.Fatal(err)
	}
	agySvc := Service{Home: agyHome, Client: "agy"}
	p, err := agySvc.Prepare(ctx, "context7", "configure")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := agySvc.Apply(ctx, p); err != nil {
		t.Fatal(err)
	}
	after, _ := os.ReadFile(filepath.Join(agyHome, "mcp_config.json"))
	if !strings.Contains(string(after), "pi-bridge") || !strings.Contains(string(after), "context7") {
		t.Fatalf("expected preserved pi-bridge and added context7, got:\n%s", after)
	}
}

func TestValidateCustomMCPRejectsUnsafeInputs(t *testing.T) {
	for _, bad := range []CustomMCP{
		{ID: "../bad", URL: "https://mcp.test/mcp"},
		{ID: "ok", URL: "http://remote.test/mcp"},
		{ID: "ok", URL: "https://mcp.test/mcp", TokenEnv: "sk-secret-value"},
		{ID: "ok", URL: "https://mcp.test/mcp", Command: "npx"},
		{ID: "ok"},
	} {
		if _, err := ValidateCustomMCP(bad); err == nil {
			t.Errorf("expected validation error for %+v", bad)
		}
	}
}
