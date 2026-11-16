package integrations

import (
	"bytes"
	"context"
	_ "embed"
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
)

//go:embed config.py
var configProgram string

var (
	safeMCPID  = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9_-]{0,48}$`)
	safeEnvVar = regexp.MustCompile(`^[A-Z][A-Z0-9_]{1,60}$`)
)

type Service struct {
	Home   string
	Client string
}

type State struct {
	Revision string             `json:"revision"`
	Exists   bool               `json:"exists"`
	Matched  bool               `json:"matched"`
	Managed  bool               `json:"managed"`
	Message  string             `json:"message"`
	Error    string             `json:"error"`
	Servers  []ConfiguredServer `json:"servers"`
}

type ConfiguredServer struct {
	ID        string `json:"id"`
	Transport string `json:"transport"`
	Target    string `json:"target"`
	Managed   bool   `json:"managed"`
	TokenEnv  string `json:"token_env"`
}

type CustomMCP struct {
	ID       string
	URL      string
	Command  string
	Args     []string
	TokenEnv string
}

// Plan holds the exact configuration revision reviewed before a mutation.
type Plan struct {
	ID, Action, Revision, Preview string
	Entry                         map[string]any
}

func (s Service) effectiveClient() string {
	switch s.Client {
	case "claude", "agy":
		return s.Client
	default:
		return "codex"
	}
}

func (s Service) configTargetLabel() string {
	switch s.effectiveClient() {
	case "agy":
		return "mcp_config.json"
	case "claude":
		return ".claude.json"
	default:
		return "config.toml"
	}
}

func findPython(ctx context.Context) (string, error) {
	for _, candidate := range []string{"python", "python3"} {
		if path, err := exec.LookPath(candidate); err == nil {
			if exec.CommandContext(ctx, path, "-c", "import sys; sys.exit(0 if sys.version_info >= (3, 11) else 1)").Run() == nil {
				return path, nil
			}
		}
	}
	return "", errors.New("necesitás Python 3.11+ para revisar o configurar integraciones MCP")
}

func ValidateCustomMCP(in CustomMCP) (map[string]any, error) {
	id := strings.TrimSpace(in.ID)
	if !safeMCPID.MatchString(id) {
		return nil, errors.New("nombre de MCP inválido (usá letras, números, - o _)")
	}
	rawURL := strings.TrimSpace(in.URL)
	cmd := strings.TrimSpace(in.Command)
	tokenEnv := strings.TrimSpace(in.TokenEnv)
	if rawURL != "" && cmd != "" {
		return nil, errors.New("elegí URL remota o comando local, no ambos")
	}
	if rawURL == "" && cmd == "" {
		return nil, errors.New("indicá una URL https:// o un comando local")
	}
	if rawURL != "" {
		parsed, err := url.Parse(rawURL)
		if err != nil || parsed.Host == "" {
			return nil, errors.New("URL de servidor MCP inválida")
		}
		host := strings.ToLower(parsed.Hostname())
		isLocal := host == "localhost" || host == "127.0.0.1"
		if parsed.Scheme != "https" && !(parsed.Scheme == "http" && isLocal) {
			return nil, errors.New("la URL debe usar https:// (o http://localhost / 127.0.0.1)")
		}
		entry := map[string]any{"url": rawURL}
		if tokenEnv != "" {
			if !safeEnvVar.MatchString(tokenEnv) {
				return nil, errors.New("la variable de entorno debe ir en MAYÚSCULAS (ej. MI_API_KEY)")
			}
			entry["bearer_token_env_var"] = tokenEnv
		}
		return entry, nil
	}
	if tokenEnv != "" {
		return nil, errors.New("la variable bearer solo aplica a servidores HTTP remotos")
	}
	if strings.ContainsAny(cmd, "\r\n\x00") {
		return nil, errors.New("comando inválido")
	}
	args := make([]string, 0, len(in.Args))
	for _, a := range in.Args {
		trimmed := strings.TrimSpace(a)
		if trimmed == "" {
			continue
		}
		if strings.ContainsAny(trimmed, "\r\n\x00") {
			return nil, errors.New("argumento inválido")
		}
		args = append(args, trimmed)
	}
	return map[string]any{"command": cmd, "args": args}, nil
}

func (s Service) runConfig(ctx context.Context, payload map[string]any) (State, error) {
	if !filepath.IsAbs(s.Home) {
		return State{}, errors.New("la ruta del perfil debe ser absoluta")
	}
	python, err := findPython(ctx)
	if err != nil {
		return State{}, err
	}
	payload["home"] = s.Home
	payload["client"] = s.effectiveClient()
	request, err := json.Marshal(payload)
	if err != nil {
		return State{}, err
	}
	cmd := exec.CommandContext(ctx, python, "-X", "utf8", "-B", "-c", configProgram)
	cmd.Stdin = bytes.NewReader(request)
	output, runErr := cmd.Output() // stderr can contain local values; never display it.
	var state State
	if json.Unmarshal(output, &state) != nil {
		return state, errors.New("no se pudo revisar el perfil; comprobá Python 3.11+, sintaxis y permisos")
	}
	if state.Error != "" {
		return state, errors.New(state.Error)
	}
	if runErr != nil {
		return state, errors.New("no se pudo completar la operación de configuración")
	}
	return state, nil
}

func (s Service) config(ctx context.Context, id, action, expected string) (State, error) {
	item, ok := Find(id)
	if !ok || id == "egolite" {
		return State{}, errors.New("esta opción usa instalación guiada, no configuración MCP")
	}
	return s.runConfig(ctx, map[string]any{
		"id":       id,
		"entry":    item.entry(),
		"action":   action,
		"expected": expected,
	})
}

func (s Service) Inspect(ctx context.Context, id string) (State, error) {
	return s.config(ctx, id, "inspect", "")
}

func (s Service) InspectEntry(ctx context.Context, id string, entry map[string]any) (State, error) {
	return s.runConfig(ctx, map[string]any{
		"id":     id,
		"entry":  entry,
		"action": "inspect",
	})
}

func (s Service) ListConfigured(ctx context.Context) ([]ConfiguredServer, error) {
	st, err := s.runConfig(ctx, map[string]any{"action": "list"})
	if err != nil {
		return nil, err
	}
	return st.Servers, nil
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
	targetFile := s.configTargetLabel()
	preview := "Destino: " + s.Home + " (" + targetFile + ")\nMCP: " + item.Name + "\n"
	if action == "undo" {
		preview += "Quitar sólo el bloque intacto agregado por Chuli. No borra programas, credenciales ni memorias.\n"
	} else if state.Exists {
		preview += "Ya existe una conexión o plugin relacionado. Se conservará sin cambios; no se duplicará.\n"
	} else {
		preview += "Agregar una conexión en " + targetFile + " y su registro de reversión en chuli-integrations/.\n"
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
	return Plan{ID: id, Action: action, Revision: state.Revision, Preview: preview, Entry: item.entry()}, nil
}

func (s Service) PrepareCustom(ctx context.Context, in CustomMCP) (Plan, error) {
	entry, err := ValidateCustomMCP(in)
	if err != nil {
		return Plan{}, err
	}
	id := strings.TrimSpace(in.ID)
	state, err := s.InspectEntry(ctx, id, entry)
	if err != nil {
		return Plan{}, err
	}
	targetFile := s.configTargetLabel()
	preview := "Destino: " + s.Home + " (" + targetFile + ")\nMCP personalizado: " + id + "\n"
	if state.Exists {
		preview += "Ya existe un servidor con ese nombre o endpoint. Se conservará sin cambios.\n"
	} else {
		preview += "Agregar conexión en " + targetFile + " y su registro de reversión.\n"
		if u, ok := entry["url"].(string); ok {
			preview += "Servidor: " + u + "\n"
		} else if c, ok := entry["command"].(string); ok {
			preview += "Comando local: " + c + "\n"
		}
		if env, ok := entry["bearer_token_env_var"].(string); ok {
			preview += "Credencial: referencia a " + env + "; nunca su valor.\n"
		}
	}
	return Plan{ID: id, Action: "configure", Revision: state.Revision, Preview: preview, Entry: entry}, nil
}

func (s Service) PrepareRemoveCustom(ctx context.Context, srv ConfiguredServer) (Plan, error) {
	if item, ok := Find(srv.ID); ok && srv.ID != "egolite" {
		return s.Prepare(ctx, item.ID, "undo")
	}
	// Read exact entry from journal on disk if present; fallback to target summary.
	var entry map[string]any
	if srv.Transport == "http" {
		entry = map[string]any{"url": srv.Target}
		if srv.TokenEnv != "" {
			entry["bearer_token_env_var"] = srv.TokenEnv
		}
	} else {
		parts := strings.Fields(srv.Target)
		if len(parts) == 0 {
			return Plan{}, errors.New("no se pudo reconstruir el comando del servidor MCP")
		}
		entry = map[string]any{"command": parts[0], "args": parts[1:]}
	}
	state, err := s.InspectEntry(ctx, srv.ID, entry)
	if err != nil {
		return Plan{}, err
	}
	if !state.Managed {
		return Plan{}, errors.New("solo podés quitar desde aquí conexiones agregadas por Chuli (las externas se preservan)")
	}
	preview := fmt.Sprintf("Destino: %s (%s)\nQuitar conexión MCP: %s\nNo elimina credenciales ni programas.", s.Home, s.configTargetLabel(), srv.ID)
	return Plan{ID: srv.ID, Action: "undo", Revision: state.Revision, Preview: preview, Entry: entry}, nil
}

func (s Service) Apply(ctx context.Context, plan Plan) (string, error) {
	if plan.Revision == "" || (plan.Action != "configure" && plan.Action != "undo") {
		return "", errors.New("primero revisá una vista previa")
	}
	entry := plan.Entry
	if entry == nil {
		item, ok := Find(plan.ID)
		if !ok || plan.ID == "egolite" {
			return "", errors.New("integración desconocida")
		}
		entry = item.entry()
	}
	state, err := s.runConfig(ctx, map[string]any{
		"id":       plan.ID,
		"entry":    entry,
		"action":   plan.Action,
		"expected": plan.Revision,
	})
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
			status = "Conexión o plugin existente en " + s.configTargetLabel() + "; se preserva"
		}
		if state.Managed {
			status = "Configurada por Chuli en " + s.configTargetLabel()
		}
	}
	return i.Name + " — " + i.Purpose + "\nEstado (" + s.effectiveClient() + "): " + status + "\n\n" + i.Instructions + "\n\nGuía / descarga oficial:\n" + i.GuideURL, nil
}
