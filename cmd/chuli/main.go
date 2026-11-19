package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"runtime/debug"
	"strings"
	"time"

	"github.com/AlejoPalavecino/codex-engineering-harness/v2/internal/eval"
	"github.com/AlejoPalavecino/codex-engineering-harness/v2/internal/integrations"
	"github.com/AlejoPalavecino/codex-engineering-harness/v2/internal/manager"
	"github.com/AlejoPalavecino/codex-engineering-harness/v2/internal/tui"
)

var version = "2.1.1-dev"

func currentVersion() string {
	info, ok := debug.ReadBuildInfo()
	return effectiveVersion(version, info, ok)
}

func effectiveVersion(configured string, info *debug.BuildInfo, ok bool) string {
	if !strings.HasSuffix(configured, "-dev") {
		return configured
	}
	if ok && info != nil {
		tag := strings.TrimPrefix(strings.TrimSpace(info.Main.Version), "v")
		if tag != "" && !strings.ContainsAny(tag, "-+") {
			if _, err := manager.Compare(tag, tag); err == nil {
				return tag
			}
		}
	}
	return configured
}

func main() {
	source := flag.String("source", "", "Checkout local para instalar, verificar o restaurar")
	client := flag.String("client", "codex", "Cliente objetivo: codex, claude o agy")
	showVersion := flag.Bool("version", false, "Mostrar versión del gestor")
	flag.Usage = func() {
		fmt.Fprintln(flag.CommandLine.Output(), "Uso: chuli [--client codex|claude|agy] [--source RUTA] [tui|install|update|models|status|doctor|rollback|version|eval|integrations]\ninstall, update y rollback muestran un plan. Agregá --apply al subcomando para aplicarlo. install y update admiten --preset, --role ROL=MODELO[:ESFUERZO] y --adopt-gentle-routing.")
		flag.PrintDefaults()
	}
	flag.Parse()
	v := currentVersion()
	if *showVersion {
		fmt.Println("Chuli Manager", v)
		return
	}
	m, err := manager.NewForClient(*source, *client)
	if err != nil {
		fail(err)
	}
	switch flag.Arg(0) {
	case "", "tui":
		if err := tui.Run(m, v); err != nil {
			fail(err)
		}
	case "version":
		fmt.Println("Chuli Manager", v)
	case "status":
		summary, err := m.Profile()
		if err != nil {
			fail(err)
		}
		spec, _ := manager.FindClient(m.Target)
		fmt.Printf("Chuli Manager %s\n%s: %s\n%s\n", v, spec.EnvHome, m.Home, summary)
	case "models":
		if err := modelsCLI(m, flag.Args()[1:], os.Stdout); err != nil {
			fail(err)
		}
	case "eval":
		if err := eval.CLI(flag.Args()[1:], *source, os.Stdout); err != nil {
			fail(err)
		}
	case "integrations":
		if err := (integrations.Service{Home: m.Home, Client: m.Target}).CLI(flag.Args()[1:], os.Stdout); err != nil {
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

type roleFlag []string

func (r *roleFlag) String() string { return strings.Join(*r, ",") }
func (r *roleFlag) Set(value string) error {
	*r = append(*r, value)
	return nil
}

func modelsCLI(m manager.Manager, args []string, out io.Writer) error {
	if len(args) > 0 && (args[0] == "list" || args[0] == "presets") {
		args = args[1:]
	}
	if len(args) > 0 && args[0] == "set" {
		return manage(m, append([]string{"install"}, args[1:]...), out)
	}
	flags := flag.NewFlagSet("models", flag.ContinueOnError)
	flags.SetOutput(out)
	subClient := flags.String("client", "", "Cliente objetivo: codex, claude o agy")
	preset := flags.String("preset", "", "Preset de modelos para el cliente")
	var roleSpecs roleFlag
	flags.Var(&roleSpecs, "role", "Asignar modelo a un rol: ROL=MODELO[:ESFUERZO]")
	apply := flags.Bool("apply", false, "Aplicar el plan")
	adopt := flags.Bool("adopt-gentle-routing", false, "Respaldar el enrutamiento reconocido de Gentle y usar Chuli")
	if err := flags.Parse(args); err != nil {
		return err
	}
	if *preset != "" || len(roleSpecs) > 0 || *apply || *adopt {
		return manage(m, append([]string{"install"}, args...), out)
	}
	target := m.Target
	if *subClient != "" {
		target = *subClient
	}
	c, ok := manager.FindClient(target)
	if !ok {
		return fmt.Errorf("cliente inválido %q (esperado codex, claude o agy)", target)
	}
	models, source := m.AvailableModels(context.Background(), c)
	fmt.Fprintf(out, "Cliente: %s (%s)\nFuente: %s\nModelos: %s\nEsfuerzos: %s\n\nPresets:\n", c.Name, c.ID, source, strings.Join(models, ", "), strings.Join(c.Efforts, ", "))
	if c.ID == "agy" {
		fmt.Fprintln(out, "Los IDs listados corresponden al orquestador. Especialistas: pro, flash, flash_lite, inherit (los presets anteriores conservan sus equivalencias).")
		fmt.Fprintln(out, "En modelos descubiertos de AGY, el esfuerzo lo define la variante/etiqueta elegida.")
	}
	for _, p := range c.Presets {
		fmt.Fprintf(out, "  %s — %s\n    %s\n", p.ID, p.Name, p.Description)
		for _, r := range manager.Roles() {
			rm := p.Roles[r.ID]
			fmt.Fprintf(out, "    - %-18s %s (%s)\n", r.ID, rm.Model, rm.Effort)
		}
	}
	return nil
}

func manage(m manager.Manager, args []string, out io.Writer) error {
	flags := flag.NewFlagSet(args[0], flag.ContinueOnError)
	flags.SetOutput(out)
	subClient := flags.String("client", "", "Cliente objetivo: codex, claude o agy")
	apply := flags.Bool("apply", false, "Aplicar el plan")
	adopt := flags.Bool("adopt-gentle-routing", false, "Respaldar el enrutamiento reconocido de Gentle y usar Chuli")
	preset := flags.String("preset", "", "Preset de modelos para el cliente")
	var roleSpecs roleFlag
	flags.Var(&roleSpecs, "role", "Asignar modelo a un rol: ROL=MODELO[:ESFUERZO]")
	if err := flags.Parse(args[1:]); err != nil {
		return err
	}
	if flags.NArg() != 0 {
		return errors.New("argumento inesperado")
	}
	if *subClient != "" && *subClient != m.Target {
		next, err := manager.NewForClient(m.Source, *subClient)
		if err != nil {
			return err
		}
		m = next
	} else if *subClient != "" {
		if _, ok := manager.FindClient(*subClient); !ok {
			return fmt.Errorf("cliente inválido %q (esperado codex, claude o agy)", *subClient)
		}
	}
	if *adopt && args[0] == "rollback" {
		return errors.New("rollback no admite --adopt-gentle-routing")
	}
	if (*preset != "" || len(roleSpecs) > 0) && args[0] == "rollback" {
		return errors.New("rollback no admite --preset ni --role")
	}
	if *preset != "" {
		if _, _, err := manager.DefaultRolesForPreset(m.Target, *preset); err != nil {
			return err
		}
		m.Preset = *preset
	}
	if len(roleSpecs) > 0 {
		roles, err := manager.ParseRoleModelSpecs(m.Target, *preset, roleSpecs)
		if err != nil {
			return err
		}
		m.RoleModels = roles
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
