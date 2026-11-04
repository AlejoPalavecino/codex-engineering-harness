package eval

import (
	"bytes"
	"strings"
	"testing"
)

func TestUntestedRunsAreNotComparable(t *testing.T) {
	if got := verdictChange(NotTested, NotTested); got != "NO COMPARABLE" {
		t.Fatalf("unobserved results must not imply equivalence: %s", got)
	}
}

func TestReportDetectsCriterionRegressionWithinFailingCase(t *testing.T) {
	cat := createFixtureCatalog(t)
	base := exampleRun(cat)
	base.Results = []CaseResult{{CaseID: "demo", Criteria: []CriterionResult{
		{ID: "no-write", Status: Pass, Notes: "unchanged", Evidence: []string{"snapshot-A"}},
		{ID: "facts", Status: Fail, Notes: "incorrect", Evidence: []string{"review-A"}},
	}}}
	candidate := base
	candidate.RunID = "candidate"
	candidate.Results = []CaseResult{{CaseID: "demo", Criteria: []CriterionResult{
		{ID: "no-write", Status: Fail, Notes: "unauthorized write", Evidence: []string{"snapshot-B"}},
		{ID: "facts", Status: Pass, Notes: "correct", Evidence: []string{"review-B"}},
	}}}
	var output bytes.Buffer
	if err := PrintReport(&output, cat, candidate, &base); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(output.String(), "demo/no-write") ||
		!strings.Contains(output.String(), "REGRESIÓN") ||
		strings.Contains(output.String(), "No hay regresiones") {
		t.Fatalf("critical regression hidden by aggregate FAIL: %s", output.String())
	}
}

func TestReportDoesNotDenyAggregateCoverageRegression(t *testing.T) {
	cat := createFixtureCatalog(t)
	base := exampleRun(cat)
	base.Results = []CaseResult{{CaseID: "demo", Criteria: []CriterionResult{
		{ID: "no-write", Status: Pass, Notes: "unchanged", Evidence: []string{"snapshot-A"}},
		{ID: "facts", Status: Pass, Notes: "correct", Evidence: []string{"review-A"}},
	}}}
	candidate := base
	candidate.RunID = "candidate"
	candidate.Results = []CaseResult{{CaseID: "demo", Criteria: base.Results[0].Criteria[:1]}}
	var output bytes.Buffer
	if err := PrintReport(&output, cat, candidate, &base); err != nil {
		t.Fatal(err)
	}
	if strings.Contains(output.String(), "No hay regresiones") {
		t.Fatalf("summary contradicts PASS to PARTIAL regression: %s", output.String())
	}
}
