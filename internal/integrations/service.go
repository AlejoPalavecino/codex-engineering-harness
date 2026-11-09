package integrations

import (
	"bytes"
	"context"
	_ "embed"
	"encoding/json"
	"errors"
	"fmt"
	"os/exec"
	"path/filepath"
	"strings"
)

//go:embed config.py
var configProgram string

type Service struct{ Home string }

type State struct {
	Revision string `json:"revision"`
	Exists   bool   `json:"exists"`
	Matched  bool   `json:"matched"`
	Managed  bool   `json:"managed"`
	Message  string `json:"message"`
	Error    string `json:"error"`
}

// Plan holds the exact configuration revision reviewed before a mutation.
type Plan struct {
	ID, Action, Revision, Preview string
}

func (s Service) config(ctx context.Context, id, action, expected string) (State, error) {
	item, ok := Find(id)
	if !ok || id == "egolite" {
		return State{}, errors.New("esta opción usa instalación guiada, no configuración MCP")
	}
	if !filepath.IsAbs(s.Home) {
		return State{}, errors.New("CODEX_HOME debe ser absoluto")
	}
	python := ""
	for _, candidate := range []string{"python", "python3"} {
		if path, err := exec.LookPath(candidate); err == nil {
			if exec.CommandContext(ctx, path, "-c", "import sys; sys.exit(0 if sys.version_info >= (3, 11) else 1)").Run() == nil {
				python = path
				break
			}
		}
	}
	if python == "" {
		return State{}, errors.New("necesitás Python 3.11+ para revisar o configurar integraciones")
	}
	request, err := json.Marshal(map[string]any{"home": s.Home, "id": id, "entry": item.entry(), "action": action, "expected": expected})
	if err != nil {
		return State{}, err
	}
	cmd := exec.CommandContext(ctx, python, "-X", "utf8", "-B", "-c", configProgram)
	cmd.Stdin = bytes.NewReader(request)
	output, runErr := cmd.Output() // stderr can contain local values; never display it.
	var state State
	if json.Unmarshal(output, &state) != nil {
		return state, errors.New("no se pudo revisar el perfil; comprobá Python 3.11+, TOML y permisos")
	}
	if state.Error != "" {
		return state, errors.New(state.Error)
	}
	if runErr != nil {
		return state, errors.New("no se pudo completar la operación de configuración")
	}
	return state, nil
}

func (s Service) Inspect(ctx context.Context, id string) (State, error) {
	return s.config(ctx, id, "inspect", "")
}

func (s Service) Prepare(ctx context.Context, id, action string) (Plan, error) {
	if action != "configure" && action != "undo" {
		return Plan{}, errors.New("acción desconocida")
	}
	item, ok := Find(id)
	if !ok {
		return Plan{}, errors.New("integración desconocida")
	}
	state, err := s.Inspect(ctx, id)
	if err != nil {
		return Plan{}, err
	}
	if action == "undo" && !state.Managed {
		return Plan{}, errors.New("no hay una conexión propia intacta para deshacer; no se tocaron conexiones existentes")
	}
	if action == "configure" && !state.Exists && item.Command != "" {
		if _, err := exec.LookPath(item.Command); err != nil {
			return Plan{}, fmt.Errorf("falta %s en PATH; abrí la guía de descarga y reabrí el Manager", item.Command)
		}
	}
	preview := "Destino: " + s.Home + "\nIntegración: " + item.Name + "\n"
	if action == "undo" {
		preview += "Quitar sólo el bloque intacto agregado por Chuli. No borra programas, credenciales ni memorias.\n"
	} else if state.Exists {
		preview += "Ya existe una conexión o plugin relacionado. Se conservará sin cambios; no se duplicará.\n"
	} else {
		preview += "Agregar una conexión en config.toml y su registro de reversión en chuli-integrations/.\n"
		if item.URL != "" {
			preview += "Servidor: " + item.URL + "\n"
		} else {
			preview += "Comando local: " + item.Command + " " + strings.Join(item.Args, " ") + "\n"
		}
		if item.TokenEnv != "" {
			preview += "Credencial: referencia a " + item.TokenEnv + "; nunca su valor.\n"
		}
	}
	preview += "Cerrá otros editores de configuración antes de confirmar. No ejecuta instaladores ni prueba la conexión todavía."
	return Plan{ID: id, Action: action, Revision: state.Revision, Preview: preview}, nil
}

func (s Service) Apply(ctx context.Context, plan Plan) (string, error) {
	if plan.Revision == "" || (plan.Action != "configure" && plan.Action != "undo") {
		return "", errors.New("primero revisá una vista previa")
	}
	state, err := s.config(ctx, plan.ID, plan.Action, plan.Revision)
	return state.Message, err
}

func (s Service) Describe(ctx context.Context, id string) (string, error) {
	i, ok := Find(id)
	if !ok {
		return "", errors.New("integración desconocida")
	}
	status := "Sin configurar"
	if id == "egolite" {
		status = "Falta ego-browser en PATH; seguí la guía de instalación"
		if _, err := exec.LookPath(i.Command); err == nil {
			status = "Comando disponible; falta comprobar onboarding desde la guía instalada"
		}
	} else {
		state, err := s.Inspect(ctx, id)
		if err != nil {
			return "", err
		}
		if state.Exists {
			status = "Conexión o plugin existente; se preserva"
		}
		if state.Managed {
			status = "Configurada por Chuli; pendiente de comprobación en esta sesión"
		}
	}
	return i.Name + " — " + i.Purpose + "\nEstado: " + status + "\n\n" + i.Instructions + "\n\nGuía / descarga oficial:\n" + i.GuideURL, nil
}
