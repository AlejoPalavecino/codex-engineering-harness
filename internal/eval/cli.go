package eval

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

func usage() string {
	return "Uso: chuli [--source RUTA_REPO] eval " +
		"<check|list|fingerprint|snapshot|diff|assert-unchanged|report> [archivos]\n" +
		"  check                         Valida casos, rúbricas y fixtures (sin ejecutar agentes)\n" +
		"  list                          Lista los escenarios EWP\n" +
		"  fingerprint                   Identificador reproducible del catálogo\n" +
		"  snapshot WORKSPACE OUT.json   Captura hashes de archivos fuera de .git\n" +
		"  diff BEFORE.json AFTER.json   Enumera cambios sin inferir calidad\n" +
		"  assert-unchanged B.json A.json  Falla si el workspace cambió\n" +
		"  report RUN.json [BASE.json]   Informe Markdown de criterios observados\n"
}

// CLI is deliberately offline: it neither calls Codex nor writes any configuration.
// Its only write operation is an explicit workspace snapshot outside that workspace.
func CLI(args []string, source string, out io.Writer) error {
	if len(args) == 0 || args[0] == "help" {
		_, _ = io.WriteString(out, usage())
		return nil
	}
	if source == "" {
		var err error
		source, err = os.Getwd()
		if err != nil {
			return err
		}
	}
	switch args[0] {
	case "snapshot":
		if len(args) != 3 {
			return errors.New("uso: chuli eval snapshot WORKSPACE OUT.json")
		}
		s, err := SaveSnapshot(args[1], args[2])
		if err != nil {
			return err
		}
		fmt.Fprintf(out, "Snapshot: %d archivos; SHA-256 %s\n", len(s.Files), s.TreeHash)
		return nil
	case "diff", "assert-unchanged":
		if len(args) != 3 {
			return errors.New("uso: chuli eval diff|assert-unchanged BEFORE.json AFTER.json")
		}
		before, err := ReadSnapshot(args[1])
		if err != nil {
			return err
		}
		after, err := ReadSnapshot(args[2])
		if err != nil {
			return err
		}
		changes := CompareSnapshots(before, after)
		for _, change := range changes {
			fmt.Fprintf(out, "%s %s\n", change.Status, change.Path)
		}
		fmt.Fprintf(out, "Cambios de workspace (sin .git): %d\n", len(changes))
		if args[0] == "assert-unchanged" && len(changes) > 0 {
			return fmt.Errorf("invariante R0 violado: %d archivos cambiaron", len(changes))
		}
		return nil
	}
	catalog, err := Load(filepath.Join(source, "evals"))
	if err != nil {
		return err
	}
	switch args[0] {
	case "check":
		if len(args) != 1 {
			return errors.New("uso: chuli eval check")
		}
		fmt.Fprintf(out, "OK: %d escenarios reproducibles; baseline %s / %s\nCatálogo SHA-256: %s\n",
			len(catalog.Cases), catalog.Manifest.BaselineTag, catalog.Manifest.BaselineSHA,
			catalog.Fingerprint)
	case "list":
		if len(args) != 1 {
			return errors.New("uso: chuli eval list")
		}
		for _, c := range catalog.Cases {
			fmt.Fprintf(out, "%s | %s | %s | approval=%s\n", c.ID, c.Route, c.Title, c.Approval)
		}
	case "fingerprint":
		if len(args) != 1 {
			return errors.New("uso: chuli eval fingerprint")
		}
		fmt.Fprintln(out, catalog.Fingerprint)
	case "report":
		if len(args) != 2 && len(args) != 3 {
			return errors.New("uso: chuli eval report RUN.json [BASELINE_RUN.json]")
		}
		run, err := ReadRun(args[1], catalog)
		if err != nil {
			return err
		}
		if len(args) == 3 {
			base, err := ReadRun(args[2], catalog)
			if err != nil {
				return err
			}
			return PrintReport(out, catalog, run, &base)
		}
		return PrintReport(out, catalog, run, nil)
	default:
		return fmt.Errorf("subcomando desconocido %q; %s", args[0], usage())
	}
	return nil
}
