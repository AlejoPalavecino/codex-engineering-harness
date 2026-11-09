package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"time"

	"github.com/AlejoPalavecino/codex-engineering-harness/internal/eval"
	"github.com/AlejoPalavecino/codex-engineering-harness/internal/integrations"
	"github.com/AlejoPalavecino/codex-engineering-harness/internal/manager"
	"github.com/AlejoPalavecino/codex-engineering-harness/internal/tui"
)

var version = "1.1.1-dev"

func main() {
	source := flag.String("source", "", "Checkout local para instalar, verificar o restaurar")
	showVersion := flag.Bool("version", false, "Mostrar versión del gestor")
	flag.Usage = func() {
		fmt.Fprintln(flag.CommandLine.Output(), "Uso: chuli [--source RUTA] [tui|install|update|status|doctor|rollback|version|eval|integrations]\ninstall, update y rollback muestran un plan. Agregá --apply al subcomando para aplicarlo. install y update admiten --adopt-gentle-routing para respaldar el routing de Gentle y usar Chuli.")
		flag.PrintDefaults()
	}
	flag.Parse()
	if *showVersion {
		fmt.Println("Chuli Manager", version)
		return
	}
	m, err := manager.New(*source)
	if err != nil {
		fail(err)
	}
	switch flag.Arg(0) {
	case "", "tui":
		if err := tui.Run(m, version); err != nil {
			fail(err)
		}
	case "version":
		fmt.Println("Chuli Manager", version)
	case "status":
		summary, err := m.Profile()
		if err != nil {
			fail(err)
		}
		fmt.Printf("Chuli Manager %s\nCODEX_HOME: %s\n%s\n", version, m.Home, summary)
	case "eval":
		if err := eval.CLI(flag.Args()[1:], *source, os.Stdout); err != nil {
			fail(err)
		}
	case "integrations":
		if err := (integrations.Service{Home: m.Home}).CLI(flag.Args()[1:], os.Stdout); err != nil {
			fail(err)
		}
	case "doctor":
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
		defer cancel()
		output, err := m.Doctor(ctx)
		fmt.Print(output)
		if err != nil {
			fail(err)
		}
	case "install", "update", "rollback":
		if err := manage(m, flag.Args(), os.Stdout); err != nil {
			fail(err)
		}
	default:
		flag.Usage()
		os.Exit(2)
	}
}
func fail(err error) {
	fmt.Fprintln(os.Stderr, "Error:", err)
	os.Exit(1)
}

func manage(m manager.Manager, args []string, out io.Writer) error {
	flags := flag.NewFlagSet(args[0], flag.ContinueOnError)
	flags.SetOutput(out)
	apply := flags.Bool("apply", false, "Aplicar el plan")
	adopt := flags.Bool("adopt-gentle-routing", false, "Respaldar el enrutamiento reconocido de Gentle y usar Chuli")
	if err := flags.Parse(args[1:]); err != nil {
		return err
	}
	if flags.NArg() != 0 {
		return errors.New("argumento inesperado")
	}
	if *adopt && args[0] == "rollback" {
		return errors.New("rollback no admite --adopt-gentle-routing")
	}
	action := args[0]
	if action != "rollback" {
		action = "latest"
		if m.Source != "" {
			action = "local"
		}
	}
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
	defer cancel()
	p, err := m.Prepare(ctx, action, *adopt)
	if errors.Is(err, manager.ErrGentleConflict) {
		return fmt.Errorf("%w\nRevisá el plan de adopción con %s --adopt-gentle-routing.", err, args[0])
	}
	if err != nil {
		return err
	}
	defer p.Cleanup()
	fmt.Fprintln(out, p.Preview)
	if !*apply {
		return nil
	}
	log, err := m.Apply(context.Background(), p)
	fmt.Fprintln(out, log)
	return err
}
