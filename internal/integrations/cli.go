package integrations

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"time"
)

// CLI exposes the same preview/apply boundary for terminals without a TUI.
func (s Service) CLI(args []string, out io.Writer) error {
	if len(args) == 0 || (len(args) == 1 && args[0] == "list") {
		for _, i := range Catalog() {
			fmt.Fprintf(out, "%s · %s · %s\n", i.ID, i.Name, i.Purpose)
		}
		fmt.Fprintln(out, "Uso: chuli integrations [status|guide|verify|configure|undo] ID [--apply]")
		return nil
	}
	if len(args) < 2 {
		return errors.New("elegí una integración del catálogo con chuli integrations list")
	}
	action, id := args[0], args[1]
	flags := flag.NewFlagSet("integrations", flag.ContinueOnError)
	flags.SetOutput(out)
	apply := flags.Bool("apply", false, "Confirmar configuración o reversión")
	if err := flags.Parse(args[2:]); err != nil {
		return err
	}
	if flags.NArg() != 0 || (*apply && action != "configure" && action != "undo") {
		return errors.New("argumentos inesperados")
	}
	i, ok := Find(id)
	if !ok {
		return errors.New("integración desconocida")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 25*time.Second)
	defer cancel()
	switch action {
	case "status":
		text, err := s.Describe(ctx, id)
		if err != nil {
			return err
		}
		fmt.Fprintln(out, text)
	case "guide":
		fmt.Fprintln(out, i.Instructions+"\n\n"+i.GuideURL)
	case "verify":
		text, err := s.Verify(ctx, id)
		if err != nil {
			return err
		}
		fmt.Fprintln(out, text)
	case "configure", "undo":
		plan, err := s.Prepare(ctx, id, action)
		if err != nil {
			return err
		}
		fmt.Fprintln(out, plan.Preview)
		if *apply {
			text, err := s.Apply(ctx, plan)
			if err != nil {
				return err
			}
			fmt.Fprintln(out, text)
		}
	default:
		return errors.New("acción desconocida; usá status, guide, verify, configure o undo")
	}
	return nil
}
