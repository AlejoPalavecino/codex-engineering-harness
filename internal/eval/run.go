package eval

import (
	"errors"
	"fmt"
	"io"
	"sort"
	"strings"
)

const (
	Pass      = "PASS"
	Fail      = "FAIL"
	Partial   = "PARTIAL"
	NotTested = "NOT_TESTED"
)

type Host struct {
	Client        string `json:"client"`
	ClientVersion string `json:"client_version"`
	OperatingOS   string `json:"os"`
	ConfigSHA256  string `json:"config_sha256"`
}

type CriterionResult struct {
	ID       string   `json:"id"`
	Status   string   `json:"status"`
	Notes    string   `json:"notes,omitempty"`
	Evidence []string `json:"evidence,omitempty"`
}

type ModelObservation struct {
	Role     string `json:"role"`
	Model    string `json:"model"`
	Evidence string `json:"evidence"`
}

type Measurements struct {
	Delegations  *int64 `json:"delegations,omitempty"`
	Retries      *int64 `json:"retries,omitempty"`
	InputTokens  *int64 `json:"input_tokens,omitempty"`
	OutputTokens *int64 `json:"output_tokens,omitempty"`
	Source       string `json:"source,omitempty"`
}

type CaseResult struct {
	CaseID       string             `json:"case_id"`
	Criteria     []CriterionResult  `json:"criteria"`
	Observations string             `json:"observations,omitempty"`
	Models       []ModelObservation `json:"models,omitempty"`
	Measurements Measurements       `json:"measurements,omitempty"`
}

type Run struct {
	Schema             int          `json:"schema"`
	Suite              string       `json:"suite"`
	RunID              string       `json:"run_id"`
	HarnessVersion     string       `json:"harness_version"`
	BaselineTag        string       `json:"baseline_tag"`
	BaselineSHA        string       `json:"baseline_sha"`
	CatalogFingerprint string       `json:"catalog_fingerprint"`
	Host               Host         `json:"host"`
	Results            []CaseResult `json:"results"`
}

func isStatus(s string) bool {
	return s == Pass || s == Fail || s == Partial || s == NotTested
}

func ReadRun(path string, catalog Catalog) (Run, error) {
	var run Run
	if err := decodeJSON(path, &run); err != nil {
		return Run{}, err
	}
	if err := ValidateRun(run, catalog); err != nil {
		return Run{}, fmt.Errorf("%s: %w", path, err)
	}
	return run, nil
}

func ValidateRun(run Run, catalog Catalog) error {
	if run.Schema != Schema || run.Suite != catalog.Manifest.Suite ||
		run.BaselineTag != catalog.Manifest.BaselineTag || run.BaselineSHA != catalog.Manifest.BaselineSHA ||
		run.CatalogFingerprint != catalog.Fingerprint {
		return errors.New("run schema, suite, pinned baseline or catalog fingerprint mismatch")
	}
	if strings.TrimSpace(run.RunID) == "" || strings.TrimSpace(run.HarnessVersion) == "" ||
		run.Host.Client == "" || run.Host.ClientVersion == "" || run.Host.OperatingOS == "" {
		return errors.New("run ID, harness version and host environment are required")
	}
	known := make(map[string]Case)
	for _, c := range catalog.Cases {
		known[c.ID] = c
	}
	seen := make(map[string]bool)
	for _, item := range run.Results {
		c, ok := known[item.CaseID]
		if !ok || seen[item.CaseID] {
			return fmt.Errorf("unknown or duplicate case result: %q", item.CaseID)
		}
		seen[item.CaseID] = true
		criteria := make(map[string]bool)
		for _, criterion := range c.Criteria {
			criteria[criterion.ID] = true
		}
		observed := make(map[string]bool)
		for _, cr := range item.Criteria {
			if !criteria[cr.ID] || observed[cr.ID] || !isStatus(cr.Status) {
				return fmt.Errorf("%s: invalid/duplicate criterion result %q", item.CaseID, cr.ID)
			}
			observed[cr.ID] = true
			if cr.Status == Pass || cr.Status == Fail || cr.Status == Partial {
				if len(cr.Evidence) == 0 || strings.TrimSpace(cr.Notes) == "" {
					return fmt.Errorf("%s/%s: %s needs a note and evidence reference",
						item.CaseID, cr.ID, cr.Status)
				}
			}
			for _, ref := range cr.Evidence {
				if strings.TrimSpace(ref) == "" || len(ref) > 800 || strings.ContainsAny(ref, "\r\n") {
					return fmt.Errorf("%s/%s: invalid evidence reference", item.CaseID, cr.ID)
				}
			}
		}
		for _, observedModel := range item.Models {
			if observedModel.Model == "" || observedModel.Role == "" ||
				strings.TrimSpace(observedModel.Evidence) == "" {
				return fmt.Errorf("%s: model claim requires role, model and verifiable provenance", item.CaseID)
			}
		}
		values := []*int64{item.Measurements.Delegations, item.Measurements.Retries,
			item.Measurements.InputTokens, item.Measurements.OutputTokens}
		for _, value := range values {
			if value != nil {
				if *value < 0 || strings.TrimSpace(item.Measurements.Source) == "" {
					return fmt.Errorf("%s: measurements need nonnegative values and a source", item.CaseID)
				}
			}
		}
	}
	return nil
}

// Status is mechanically derived from all rubric criteria, not declared by the agent.
// Missing criteria count as NOT_TESTED. Evidence is *referenced*, not authenticated.
func Status(c Case, result *CaseResult) string {
	if result == nil {
		return NotTested
	}
	observed := make(map[string]string)
	for _, cr := range result.Criteria {
		observed[cr.ID] = cr.Status
	}
	passed := 0
	failed := false
	partial := false
	for _, criterion := range c.Criteria {
		switch observed[criterion.ID] {
		case Pass:
			passed++
		case Fail:
			failed = true
		case Partial:
			partial = true
		}
	}
	switch {
	case failed:
		return Fail
	case passed == len(c.Criteria):
		return Pass
	case passed > 0 || partial:
		return Partial
	default:
		return NotTested
	}
}

func resultsByID(r Run) map[string]*CaseResult {
	out := make(map[string]*CaseResult)
	for idx := range r.Results {
		out[r.Results[idx].CaseID] = &r.Results[idx]
	}
	return out
}

func verdictChange(old, next string) string {
	switch {
	case old == NotTested || next == NotTested:
		return "NO COMPARABLE"
	case old == next:
		return "SIN CAMBIO"
	case old == Pass && next != Pass:
		return "REGRESIÓN"
	case (old == Fail || old == Partial) && next == Pass:
		return "MEJORA OBSERVADA"
	case old == Partial && next == Fail:
		return "REGRESIÓN"
	case old == Fail && next == Partial:
		return "MEJORA OBSERVADA"
	default:
		return "REVISAR"
	}
}

func criterionStatus(result *CaseResult, id string) string {
	if result != nil {
		for _, cr := range result.Criteria {
			if cr.ID == id {
				return cr.Status
			}
		}
	}
	return NotTested
}

func mdCell(s string) string {
	s = strings.ReplaceAll(s, "|", "\\|")
	s = strings.ReplaceAll(s, "\r", " ")
	s = strings.ReplaceAll(s, "\n", " ")
	return strings.TrimSpace(s)
}

func PrintReport(out io.Writer, catalog Catalog, run Run, baseline *Run) error {
	if err := ValidateRun(run, catalog); err != nil {
		return err
	}
	if baseline != nil {
		if err := ValidateRun(*baseline, catalog); err != nil {
			return fmt.Errorf("invalid baseline run: %w", err)
		}
		if baseline.HarnessVersion == run.HarnessVersion && baseline.RunID == run.RunID {
			return errors.New("candidate and baseline must be distinct runs")
		}
	}
	results := resultsByID(run)
	var baseResults map[string]*CaseResult
	if baseline != nil {
		baseResults = resultsByID(*baseline)
	}
	fmt.Fprintf(out, "# Chuli evaluation — %s\n\n", mdCell(run.RunID))
	fmt.Fprintf(out, "- Suite: %s; catálogo SHA-256: %s\n", mdCell(run.Suite), run.CatalogFingerprint)
	fmt.Fprintf(out, "- Versión evaluada: %s; baseline anclado: %s (%s)\n", mdCell(run.HarnessVersion),
		run.BaselineTag, run.BaselineSHA)
	fmt.Fprintf(out, "- Entorno: %s %s / %s\n", mdCell(run.Host.Client),
		mdCell(run.Host.ClientVersion), mdCell(run.Host.OperatingOS))
	fmt.Fprintln(out, "- **Resultados registrados por observador:** la herramienta valida estructura y evidencia declarada; no ejecuta ni certifica agentes.")
	if baseline != nil {
		fmt.Fprintf(out, "- Comparación con: %s (%s)\n", mdCell(baseline.RunID), mdCell(baseline.HarnessVersion))
		if baseline.Host.Client != run.Host.Client || baseline.Host.ClientVersion != run.Host.ClientVersion ||
			baseline.Host.OperatingOS != run.Host.OperatingOS {
			fmt.Fprintln(out, "- **ADVERTENCIA:** clientes/versiones/SO distintos; comparación condicionada.")
		}
		if baseline.Host.ConfigSHA256 == "" || run.Host.ConfigSHA256 == "" {
			fmt.Fprintln(out, "- Configuración no identificada en ambos registros; equivalencia de configuración NO VERIFICABLE.")
		} else if baseline.Host.ConfigSHA256 != run.Host.ConfigSHA256 {
			fmt.Fprintln(out, "- ADVERTENCIA: hashes de configuración distintos; la atribución causal requiere comprobar qué variable cambió.")
		}
	}
	fmt.Fprintln(out, "\n| Caso | Ruta | Resultado |"+func() string {
		if baseline != nil {
			return " Baseline | Cambio |"
		}
		return ""
	}())
	if baseline != nil {
		fmt.Fprintln(out, "| --- | --- | --- | --- | --- |")
	} else {
		fmt.Fprintln(out, "| --- | --- | --- |")
	}
	counts := make(map[string]int)
	var regressions []string
	for _, c := range catalog.Cases {
		current := Status(c, results[c.ID])
		counts[current]++
		if baseline == nil {
			fmt.Fprintf(out, "| %s | %s | %s |\n", c.ID, c.Route, current)
		} else {
			old := Status(c, baseResults[c.ID])
			change := verdictChange(old, current)
			previousRegressions := len(regressions)
			// Aggregate FAIL can conceal a newly failing critical criterion.
			for _, cr := range c.Criteria {
				if verdictChange(criterionStatus(baseResults[c.ID], cr.ID),
					criterionStatus(results[c.ID], cr.ID)) == "REGRESIÓN" {
					regressions = append(regressions, c.ID+"/"+cr.ID)
					change = "REGRESIÓN"
				}
			}
			if change == "REGRESIÓN" && len(regressions) == previousRegressions {
				regressions = append(regressions, c.ID)
			}
			fmt.Fprintf(out, "| %s | %s | %s | %s | %s |\n", c.ID, c.Route, current, old, change)
		}
	}
	fmt.Fprintf(out, "\nCasos: PASS %d, FAIL %d, PARTIAL %d, NOT_TESTED %d. **No equivale a una puntuación general de calidad.**\n",
		counts[Pass], counts[Fail], counts[Partial], counts[NotTested])
	if baseline != nil {
		if len(regressions) > 0 {
			sort.Strings(regressions)
			fmt.Fprintf(out, "\n**Regresiones registradas que requieren revisión:** %s.\n", strings.Join(regressions, ", "))
		} else {
			fmt.Fprintln(out, "\nNo hay regresiones entre los resultados comparables registrados; esto no prueba ausencia de fallos.")
		}
	}
	fmt.Fprint(out, "\n## Evidencia y observaciones registradas\n\n")
	for _, c := range catalog.Cases {
		item := results[c.ID]
		if item == nil {
			continue
		}
		fmt.Fprintf(out, "### %s — %s\n", c.ID, mdCell(c.Title))
		rubric := make(map[string]CriterionResult)
		for _, cr := range item.Criteria {
			rubric[cr.ID] = cr
		}
		for _, cr := range c.Criteria {
			observed, found := rubric[cr.ID]
			if !found {
				fmt.Fprintf(out, "- %s: NOT_TESTED\n", cr.ID)
				continue
			}
			fmt.Fprintf(out, "- %s: %s — %s", cr.ID, observed.Status, mdCell(observed.Notes))
			if len(observed.Evidence) > 0 {
				fmt.Fprintf(out, " (referencias: %s)", mdCell(strings.Join(observed.Evidence, "; ")))
			}
			fmt.Fprintln(out)
		}
		for _, model := range item.Models {
			fmt.Fprintf(out, "- Modelo observado: %s → %s; fuente: %s\n", mdCell(model.Role),
				mdCell(model.Model), mdCell(model.Evidence))
		}
		if strings.TrimSpace(item.Observations) != "" {
			fmt.Fprintf(out, "- Notas: %s\n", mdCell(item.Observations))
		}
		fmt.Fprintln(out)
	}
	return nil
}
