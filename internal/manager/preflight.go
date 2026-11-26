package manager

import (
	"bytes"
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
	"time"
)

var (
	gentleHintPattern = regexp.MustCompile(`(?i)<!--\s*/?\s*gentle-ai\s*:\s*(?:sdd-orchestrator|orchestrator|agent-routing)\b`)
	markdownFenceLine = regexp.MustCompile("^ {0,3}(`{3,}|~{3,})")
)

// RuntimePreflight summarizes non-destructive runtime prerequisite checks.
type RuntimePreflight struct {
	PowerShellOK bool
	PythonOK     bool
	Detail       string
}

func instructionFilePath(clientID, home string) string {
	switch clientID {
	case "claude":
		return filepath.Join(home, "CLAUDE.md")
	case "agy":
		return filepath.Join(home, "plugins", "chuli", "rules", "AGENTS.md")
	default:
		return filepath.Join(home, "AGENTS.md")
	}
}

// HasGentleRoutingConflict checks whether an existing client instruction file
// contains Gentle AI routing delimiters outside fenced Markdown code blocks.
func HasGentleRoutingConflict(clientID, home string) bool {
	if strings.TrimSpace(home) == "" {
		return false
	}
	path := instructionFilePath(clientID, home)
	info, err := os.Lstat(path)
	if err != nil || !info.Mode().IsRegular() {
		return false
	}
	f, err := os.Open(path)
	if err != nil {
		return false
	}
	defer f.Close()
	raw, err := bounded(f, maxEntry)
	if err != nil {
		return false
	}
	raw = bytes.TrimPrefix(raw, []byte("\xef\xbb\xbf"))
	text := strings.ReplaceAll(string(raw), "\r\n", "\n")

	fence := ""
	for _, line := range strings.Split(text, "\n") {
		if fence != "" {
			trimmed := strings.TrimSpace(line)
			if strings.HasPrefix(trimmed, fence) && strings.Trim(trimmed, string(fence[0])) == "" {
				fence = ""
			}
			continue
		}
		if match := markdownFenceLine.FindStringSubmatch(line); len(match) == 2 {
			fence = match[1]
			continue
		}
		if gentleHintPattern.MatchString(line) {
			return true
		}
	}
	return false
}

func hasPython311(ctx context.Context) bool {
	candidates := []struct {
		bin  string
		args []string
	}{
		{bin: "python"},
		{bin: "python3"},
	}
	if runtime.GOOS == "windows" {
		candidates = append(candidates, struct {
			bin  string
			args []string
		}{bin: "py", args: []string{"-3.11"}})
	}
	for _, cand := range candidates {
		path, err := exec.LookPath(cand.bin)
		if err != nil {
			continue
		}
		checkCtx, cancel := context.WithTimeout(ctx, 2*time.Second)
		args := append(append([]string(nil), cand.args...), "-c", "import tomllib, sys; sys.exit(0 if sys.version_info >= (3, 11) else 1)")
		err = exec.CommandContext(checkCtx, path, args...).Run()
		cancel()
		if err == nil {
			return true
		}
	}
	return false
}

// CheckRuntimePreflight verifies in read-only mode whether PowerShell and Python 3.11+
// are available in PATH before starting installation.
func CheckRuntimePreflight(ctx context.Context) RuntimePreflight {
	if ctx == nil {
		ctx = context.Background()
	}
	_, psErr := shell()
	psOK := psErr == nil
	pyOK := hasPython311(ctx)

	var detail string
	switch {
	case !psOK && !pyOK:
		detail = "Faltan PowerShell y Python 3.11+ en PATH (requeridos para instalar)."
	case !psOK:
		detail = "Falta PowerShell (pwsh) en PATH (requerido para instalar)."
	case !pyOK:
		detail = "Falta Python 3.11+ en PATH (requerido para validar configuración)."
	}
	return RuntimePreflight{
		PowerShellOK: psOK,
		PythonOK:     pyOK,
		Detail:       detail,
	}
}

// PresetDisplayName returns a human-readable preset name for the client.
func PresetDisplayName(clientID, presetID string) string {
	if presetID == "custom" {
		return "Personalizado por rol"
	}
	spec, ok := FindClient(clientID)
	if !ok {
		return presetID
	}
	for _, p := range spec.Presets {
		if p.ID == presetID {
			return p.Name
		}
	}
	return presetID
}
