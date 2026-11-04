package eval

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func createFixtureCatalog(t *testing.T) Catalog {
	t.Helper()
	root := t.TempDir()
	for _, dir := range []string{"cases", "fixtures/demo", "rubrics"} {
		if err := os.MkdirAll(filepath.Join(root, dir), 0700); err != nil {
			t.Fatal(err)
		}
	}
	m := Manifest{
		Schema: Schema, Suite: "chuli-test-v1", BaselineTag: "v0.1.3",
		BaselineSHA: "68b756bd5e39539b4e277bb490321f8c06870755",
		Cases:       []string{"demo"},
	}
	c := Case{
		ID: "demo", Title: "Prueba sintética", Route: "R0", Fixture: "fixtures/demo",
		Approval: "not_required", Setup: "Abrir una copia", Prompt: "Analizar sin escribir",
		Criteria: []Criterion{
			{ID: "no-write", Description: "Sin cambios", Severity: "critical", Method: "snapshot"},
			{ID: "facts", Description: "Sin inventos", Severity: "important", Method: "review"},
		},
	}
	for name, data := range map[string]any{"manifest.json": m, "cases/demo.json": c} {
		payload, err := json.MarshalIndent(data, "", "  ")
		if err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(root, name), payload, 0600); err != nil {
			t.Fatal(err)
		}
	}
	for path, data := range map[string]string{
		"fixtures/demo/demo.txt": "fixture sintético",
		"rubrics/readme.md":      "criterios de revisión",
	} {
		if err := os.WriteFile(filepath.Join(root, path), []byte(data), 0600); err != nil {
			t.Fatal(err)
		}
	}
	catalog, err := Load(root)
	if err != nil {
		t.Fatal(err)
	}
	return catalog
}

func TestCatalogLoadAndFingerprint(t *testing.T) {
	cat := createFixtureCatalog(t)
	if len(cat.Cases) != 1 || cat.Manifest.BaselineTag != "v0.1.3" || len(cat.Fingerprint) != 64 {
		t.Fatalf("invalid catalog: %+v", cat)
	}
	next, err := Fingerprint(cat.Root)
	if err != nil || next != cat.Fingerprint {
		t.Fatalf("non-deterministic fingerprint: %s %v", next, err)
	}
	file := filepath.Join(cat.Root, "fixtures/demo/demo.txt")
	if err := os.WriteFile(file, []byte("changed"), 0600); err != nil {
		t.Fatal(err)
	}
	updated, err := Fingerprint(cat.Root)
	if err != nil || updated == cat.Fingerprint {
		t.Fatal("changing a fixture must invalidate the reproducibility key")
	}
}

func TestCatalogRejectsUnknownOrUnsafeCases(t *testing.T) {
	cat := createFixtureCatalog(t)
	var c Case
	if err := decodeJSON(filepath.Join(cat.Root, "cases/demo.json"), &c); err != nil {
		t.Fatal(err)
	}
	c.Fixture = "fixtures/../../other"
	data, _ := json.Marshal(c)
	_ = os.WriteFile(filepath.Join(cat.Root, "cases/demo.json"), data, 0600)
	if _, err := Load(cat.Root); err == nil {
		t.Fatal("accepted unsafe fixture path")
	}
}

func TestCatalogRejectsSymlinks(t *testing.T) {
	cat := createFixtureCatalog(t)
	target := filepath.Join(cat.Root, "fixtures/demo/demo.txt")
	link := filepath.Join(cat.Root, "fixtures/demo/external-link")
	if err := os.Symlink(target, link); err != nil {
		t.Skip("symlinks unavailable on this OS")
	}
	if _, err := Load(cat.Root); err == nil {
		t.Fatal("symlink must be rejected")
	}
}

func TestSnapshotNoMutationAndDiff(t *testing.T) {
	dir := t.TempDir()
	_ = os.WriteFile(filepath.Join(dir, "module.txt"), []byte("old"), 0600)
	_ = os.MkdirAll(filepath.Join(dir, ".git"), 0700)
	_ = os.WriteFile(filepath.Join(dir, ".git/HEAD"), []byte("git metadata"), 0600)
	before, err := SnapshotWorkspace(dir)
	if err != nil {
		t.Fatal(err)
	}
	if len(before.Files) != 1 {
		t.Fatalf("git metadata must be excluded: %+v", before.Files)
	}
	if _, err := SaveSnapshot(dir, filepath.Join(dir, "snapshot.json")); err == nil {
		t.Fatal("snapshot may not write into measured directory")
	}
	out := filepath.Join(t.TempDir(), "before.json")
	if _, err := SaveSnapshot(dir, out); err != nil {
		t.Fatal(err)
	}
	loaded, err := ReadSnapshot(out)
	if err != nil || len(CompareSnapshots(before, loaded)) != 0 {
		t.Fatalf("snapshot mismatch: %+v %v", loaded, err)
	}
	_ = os.WriteFile(filepath.Join(dir, "module.txt"), []byte("new"), 0600)
	_ = os.WriteFile(filepath.Join(dir, "added.txt"), []byte("new"), 0600)
	after, err := SnapshotWorkspace(dir)
	if err != nil {
		t.Fatal(err)
	}
	change := CompareSnapshots(before, after)
	if len(change) != 2 || change[0].Status != "ADDED" || change[1].Status != "MODIFIED" {
		t.Fatalf("unexpected changes: %+v", change)
	}
	if _, err := SaveSnapshot(dir, out); err == nil {
		t.Fatal("snapshots should not be overwritten")
	}
}

func exampleRun(cat Catalog) Run {
	return Run{
		Schema: Schema, Suite: cat.Manifest.Suite, RunID: "baseline-example",
		HarnessVersion: "0.1.3", BaselineTag: cat.Manifest.BaselineTag,
		BaselineSHA: cat.Manifest.BaselineSHA, CatalogFingerprint: cat.Fingerprint,
		Host: Host{Client: "Codex Desktop", ClientVersion: "OBSERVED-TEST", OperatingOS: "Windows"},
	}
}

func TestStatusDerivedFromAllCriteria(t *testing.T) {
	cat := createFixtureCatalog(t)
	c := cat.Cases[0]
	if Status(c, nil) != NotTested {
		t.Fatal("missing case should not be a pass")
	}
	passing := CaseResult{CaseID: c.ID, Criteria: []CriterionResult{
		{ID: "no-write", Status: Pass, Notes: "same digest", Evidence: []string{"before/after"}},
		{ID: "facts", Status: Pass, Notes: "sources checked", Evidence: []string{"review.md"}},
	}}
	run := exampleRun(cat)
	run.Results = []CaseResult{passing}
	if err := ValidateRun(run, cat); err != nil {
		t.Fatal(err)
	}
	if Status(c, &passing) != Pass {
		t.Fatal("all criteria passed should be PASS")
	}
	passing.Criteria = passing.Criteria[:1]
	if Status(c, &passing) != Partial {
		t.Fatal("missing criterion must make case PARTIAL")
	}
	passing.Criteria[0].Status = Fail
	if Status(c, &passing) != Fail {
		t.Fatal("any failing criterion must fail the case")
	}
	passing.Criteria[0].Evidence = nil
	run.Results = []CaseResult{passing}
	if err := ValidateRun(run, cat); err == nil {
		t.Fatal("FAIL without evidence reference accepted")
	}
}

func TestRunRejectsUnobservedMetricsAndWrongFingerprint(t *testing.T) {
	cat := createFixtureCatalog(t)
	run := exampleRun(cat)
	run.CatalogFingerprint = "stale"
	if err := ValidateRun(run, cat); err == nil {
		t.Fatal("mismatched fixture catalog accepted")
	}
	run.CatalogFingerprint = cat.Fingerprint
	one := int64(1)
	run.Results = []CaseResult{{CaseID: "demo", Measurements: Measurements{InputTokens: &one}}}
	if err := ValidateRun(run, cat); err == nil {
		t.Fatal("tokens without provenance accepted")
	}
	run.Results[0].Measurements.Source = "observed client event"
	run.Results[0].Models = []ModelObservation{{Role: "explorer", Model: "model-x"}}
	if err := ValidateRun(run, cat); err == nil {
		t.Fatal("observed model without evidence accepted")
	}
}

func TestReportComparisonShowsRegressionWithoutInventedScore(t *testing.T) {
	cat := createFixtureCatalog(t)
	base := exampleRun(cat)
	base.Results = []CaseResult{{CaseID: "demo", Criteria: []CriterionResult{
		{ID: "no-write", Status: Pass, Notes: "no changes", Evidence: []string{"snapshot-A"}},
		{ID: "facts", Status: Pass, Notes: "correct", Evidence: []string{"review-A"}},
	}}}
	candidate := base
	candidate.RunID = "candidate-example"
	candidate.HarnessVersion = "0.2.0-dev"
	candidate.Results = []CaseResult{{CaseID: "demo", Criteria: []CriterionResult{
		{ID: "no-write", Status: Fail, Notes: "diff detected", Evidence: []string{"diff-B"}},
		{ID: "facts", Status: Pass, Notes: "correct", Evidence: []string{"review-B"}},
	}}}
	var output bytes.Buffer
	if err := PrintReport(&output, cat, candidate, &base); err != nil {
		t.Fatal(err)
	}
	result := output.String()
	if !strings.Contains(result, "REGRESIÓN") || !strings.Contains(result, "sin .git") && !strings.Contains(result, "Regresiones") {
		t.Fatalf("missing regression: %s", result)
	}
	if strings.Contains(result, "calificación 100") || strings.Contains(result, "API pricing") {
		t.Fatal("report fabricated scoring/cost")
	}
}

func TestCLIListAndNoMutation(t *testing.T) {
	cat := createFixtureCatalog(t)
	var output bytes.Buffer
	repo := filepath.Dir(cat.Root)
	if err := os.Rename(cat.Root, filepath.Join(repo, "evals")); err != nil {
		t.Fatal(err)
	}
	if err := CLI([]string{"check"}, repo, &output); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(output.String(), "baseline v0.1.3") {
		t.Fatal("CLI did not validate baseline")
	}
}

func TestActualRepositoryCatalog(t *testing.T) {
	catalog, err := Load(filepath.Join("..", "..", "evals"))
	if err != nil {
		t.Fatal(err)
	}
	if len(catalog.Cases) != 11 {
		t.Fatalf("expected 11 curated scenarios, got %d", len(catalog.Cases))
	}
	if catalog.Manifest.BaselineSHA != "68b756bd5e39539b4e277bb490321f8c06870755" {
		t.Fatal("immutable v0.1.3 reference changed")
	}
}
