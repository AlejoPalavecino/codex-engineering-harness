package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/AlejoPalavecino/codex-engineering-harness/internal/manager"
	"github.com/AlejoPalavecino/codex-engineering-harness/internal/tui"
)

var version = "0.1.3-dev"

func main() {
	source := flag.String("source", "", "Checkout local para instalar, verificar o restaurar")
	flag.Parse()
	m, err := manager.New(*source)
	if err != nil {
		fail(err)
	}
	switch flag.Arg(0) {
	case "", "tui":
		if err := tui.Run(m, version); err != nil {
			fail(err)
		}
	case "version", "--version":
		fmt.Println("Chuli Manager", version)
	case "status":
		v, err := m.Installed()
		if err != nil {
			fail(err)
		}
		if v == "" {
			v = "no instalado"
		}
		fmt.Printf("Chuli Manager %s\nHarness: %s\nCODEX_HOME: %s\n", version, v, m.Home)
	case "doctor":
		root, err := m.Local()
		if err != nil {
			fail(err)
		}
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
		defer cancel()
		output, err := m.Script(ctx, root, "verify.ps1", false)
		fmt.Print(output)
		if err != nil {
			fail(err)
		}
	default:
		fmt.Fprintln(os.Stderr, "Uso: chuli [--source RUTA] [tui|status|doctor|version]")
		os.Exit(2)
	}
}
func fail(err error) {
	fmt.Fprintln(os.Stderr, "Error:", err)
	os.Exit(1)
}
