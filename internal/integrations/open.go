package integrations

import (
	"context"
	"errors"
	"os/exec"
	"runtime"
)

// OpenGuide opens only a catalog-owned HTTPS URL after an explicit UI action.
func OpenGuide(ctx context.Context, id string) error {
	i, ok := Find(id)
	if !ok {
		return errors.New("integración desconocida")
	}
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "windows":
		cmd = exec.CommandContext(ctx, "rundll32.exe", "url.dll,FileProtocolHandler", i.GuideURL)
	case "darwin":
		cmd = exec.CommandContext(ctx, "open", i.GuideURL)
	default:
		cmd = exec.CommandContext(ctx, "xdg-open", i.GuideURL)
	}
	if cmd.Run() != nil {
		return errors.New("no se pudo abrir el navegador; copiá el enlace de la guía")
	}
	return nil
}
