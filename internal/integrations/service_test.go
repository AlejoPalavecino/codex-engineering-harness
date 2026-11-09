package integrations

import (
	"context"
	"path/filepath"
	"testing"
)

func TestServicePreviewDoesNotWriteAndCanUndo(t *testing.T) {
	s := Service{Home: canonicalTestHome(t)}
	ctx := context.Background()
	p, err := s.Prepare(ctx, "exa", "configure")
	if err != nil {
		t.Fatal(err)
	}
	state, err := s.Inspect(ctx, "exa")
	if err != nil || state.Exists {
		t.Fatal(state, err)
	}
	if _, err = s.Apply(ctx, p); err != nil {
		t.Fatal(err)
	}
	p, err = s.Prepare(ctx, "exa", "undo")
	if err != nil {
		t.Fatal(err)
	}
	if _, err = s.Apply(ctx, p); err != nil {
		t.Fatal(err)
	}
	if _, err = s.Prepare(ctx, "egolite", "configure"); err == nil {
		t.Fatal("registered browser as MCP")
	}
}

func canonicalTestHome(t *testing.T) string {
	t.Helper()
	home, err := filepath.EvalSymlinks(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	return home
}
