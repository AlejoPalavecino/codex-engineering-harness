package integrations

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"time"
)

// Verify distinguishes configuration, executable availability, and a live MCP
// handshake. It never invokes tools/call or reads project data.
func (s Service) Verify(ctx context.Context, id string) (string, error) {
	i, ok := Find(id)
	if !ok {
		return "", errors.New("integración desconocida")
	}
	if i.Command != "" {
		if _, err := exec.LookPath(i.Command); err != nil {
			return "", fmt.Errorf("falta %s en PATH; seguí la guía de descarga y reabrí el Manager", i.Command)
		}
		return "Comando disponible: " + i.Command + ". Conexión efectiva no probada.\n" + i.Instructions, nil
	}
	state, err := s.Inspect(ctx, id)
	if err != nil {
		return "", err
	}
	if !state.Exists {
		return "", errors.New("primero configurá la conexión")
	}
	if !state.Matched {
		return "", errors.New("conexión o plugin personalizado preservado; comprobalo desde el panel MCP de Codex")
	}
	token := ""
	if i.TokenEnv != "" {
		token = os.Getenv(i.TokenEnv)
		if token == "" {
			return "", fmt.Errorf("falta %s en el entorno del Manager; seguí la guía y reabrí Manager y Codex", i.TokenEnv)
		}
	}
	ctx, cancel := context.WithTimeout(ctx, 15*time.Second)
	defer cancel()
	count, err := probe(ctx, http.DefaultClient, i.URL, token)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("Conexión MCP comprobada: initialize y tools/list respondieron (%d herramientas en la primera página).\nNo se ejecutaron herramientas. Esto no comprueba permisos para operar sobre tus datos. Abrí una sesión nueva de Codex y revisá su panel MCP.", count), nil
}
