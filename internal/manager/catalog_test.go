package manager

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func TestCatalogMatchesConfigModelsJSON(t *testing.T) {
	raw, err := os.ReadFile(filepath.Join("..", "..", "config", "models.json"))
	if err != nil {
		t.Fatalf("read config/models.json: %v", err)
	}
	wantRoles, wantClients, err := parseCatalog(raw)
	if err != nil {
		t.Fatalf("parse config/models.json: %v", err)
	}
	if !reflect.DeepEqual(Roles(), wantRoles) {
		t.Fatalf("Roles() diverged from config/models.json\ngot:  %+v\nwant: %+v", Roles(), wantRoles)
	}
	if !reflect.DeepEqual(Clients(), wantClients) {
		t.Fatalf("Clients() diverged from config/models.json\ngot:  %+v\nwant: %+v", Clients(), wantClients)
	}
	gotIDs := make([]string, 0, len(wantClients))
	for _, c := range wantClients {
		gotIDs = append(gotIDs, c.ID)
	}
	if !reflect.DeepEqual(gotIDs, []string{"codex", "claude", "agy"}) {
		t.Fatalf("client order = %v, want [codex claude agy]", gotIDs)
	}
}

func TestParseCatalogRejectsInvalidPayload(t *testing.T) {
	for _, raw := range []string{
		`{}`,
		`{"schema_version":99,"roles":[],"clients":{}}`,
		`{"schema_version":1,"roles":[{"id":"orchestrator"}],"clients":{}}`,
		`{"schema_version":1,"roles":[{"id":"orchestrator"}],"clients":{"codex":{"id":"other"}}}`,
	} {
		if _, _, err := parseCatalog([]byte(raw)); err == nil {
			t.Fatalf("expected error for invalid catalog payload: %s", raw)
		}
	}
}
