package tui

import (
	"fmt"
	"strings"
	"unicode"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/AlejoPalavecino/codex-engineering-harness/internal/integrations"
	"github.com/charmbracelet/x/ansi"
)

var (
	crtGreen = lipgloss.NewStyle().Foreground(lipgloss.Color("#9BE79B")).Bold(true)
	crtBeige = lipgloss.NewStyle().Foreground(lipgloss.Color("#E6D8B8"))
	crtWhite = lipgloss.NewStyle().Foreground(lipgloss.Color("#F2F4EB")).Background(lipgloss.Color("#101713"))
)

// crtText removes terminal commands before untrusted content enters the layout.
// Newlines remain meaningful for logs; other control characters cannot move the
// cursor, replace displayed text or modify the terminal clipboard/title.
func crtText(s string) string {
	s = ansi.Strip(strings.ReplaceAll(s, "\r\n", "\n"))
	return strings.Map(func(r rune) rune {
		if r == '\n' {
			return r
		}
		if r == '\t' {
			return ' '
		}
		if unicode.IsControl(r) {
			return -1
		}
		return r
	}, s)
}

// renderCRT presents the existing state machine without starting operations or
// changing its navigation. Width zero uses the ordinary 80-column layout.
func (m model) renderCRT() tea.View {
	page := ""
	if m.crtPaged() {
		page = m.crtPage()
	}
	view := tea.NewView(m.crtFrame(page))
	view.AltScreen = true
	view.WindowTitle = "Chuli Manager"
	return view
}

// crtFrame accepts page content separately so pagination can measure the actual
// wrapped header, state and shortcuts without recursively rendering a page.
func (m model) crtFrame(page string) string {
	width := m.width
	if width <= 0 {
		width = 80
	}
	m.log, m.fault, m.title = crtText(m.log), crtText(m.fault), crtText(m.title)
	m.mgr.Home = crtText(m.mgr.Home)
	m.prepared.Version = crtText(m.prepared.Version)
	installed := crtText(m.installed)
	if installed == "" {
		installed = "No instalado"
	}
	var b strings.Builder
	if m.screen == home && width >= 60 && (m.height == 0 || m.height >= 24) {
		b.WriteString(crtBeige.Render("  .------------.\n  |   CHULI    |\n  |____________|\n    /________\\") + "\n")
	} else if m.screen == home {
		b.WriteString(crtGreen.Render("[ CHULI ]") + "\n")
	}
	b.WriteString(crtGreen.Render("CHULI") + " / " + crtBeige.Render("Manager: "+crtText(m.managerVersion)) + "\n")
	b.WriteString("Perfil: " + installed + "\n")
	homeLabel := m.mgr.Home
	if lipgloss.Width(homeLabel) > width-12 {
		homeLabel = ansi.TruncateLeft(homeLabel, max(1, width-12), "…")
	}
	b.WriteString("CODEX_HOME: " + homeLabel + "\n")
	b.WriteString(crtBeige.Render(strings.Repeat("─", min(width, 76))) + "\n")
	body, keys := m.crtBody(width, page)
	b.WriteString(body)
	if keys != "" {
		b.WriteString("\n\n" + crtBeige.Render(keys))
	}
	// Wrap rather than truncate: destinations, changes and shortcuts stay readable.
	content := ansi.Wrap(b.String(), width, "")
	return crtWhite.Width(width).Render(content)
}

func (m model) crtBody(width int, page string) (string, string) {
	if strings.HasPrefix(m.screen, "integration-") {
		return m.crtIntegration(width, page)
	}
	switch m.screen {
	case home:
		return m.crtHome(width), "↑↓ / j k navegar · Enter elegir · q salir"
	case busy:
		return crtGreen.Render("EN CURSO") + "\n" + m.title + "\n\nEjecutando comprobaciones.\nNo cierres la terminal durante una escritura.", ""
	case preview:
		return crtGreen.Render("REVISIÓN PREVIA / "+m.prepared.Version) + "\n\n" + page, "S / Y confirmar · N cancelar · Esc volver · q salir\n↑↓ / j k / PgUp / PgDn desplazar"
	case conflict:
		return crtBeige.Render("CONFLICTO / Gentle controla el enrutamiento") + "\n\nA · Usar Chuli como orquestador y respaldar Gentle\nPrimero verás el plan. La confirmación se solicita después.\n\n" + page, "A revisar plan · Esc volver · q salir\n↑↓ / j k / PgUp / PgDn desplazar"
	case result:
		return m.crtOutcome() + "\n" + m.title + "\n\n" + page, "Enter / Esc volver · q salir\n↑↓ / j k / PgUp / PgDn desplazar"
	case about:
		return crtGreen.Render("Acerca de Chuli") + "\n\nChuli es el perfil de ingeniería para Codex.\nChuli Manager instala y actualiza ese perfil en CODEX_HOME.\nDespués de instalar, abrí Codex como siempre.\n\nEl paquete oficial se verifica con SHA-256 y se conserva para recuperación.\nRequisitos: Python 3.11+ y PowerShell (pwsh en Linux y macOS).", "Esc volver · q salir"
	}
	return "", "Esc volver · q salir"
}

func (m model) crtHome(width int) string {
	var menu strings.Builder
	menu.WriteString(crtGreen.Render("Seleccioná una acción") + "\n\n")
	for index, item := range choices(m.mgr) {
		name := item.label
		if width < 60 {
			name = map[string]string{"latest": "Instalar o actualizar", "local": "Instalar desde --source", "integrations": "Integraciones", "profile": "Modelos, agentes y skills", "doctor": "Comprobar perfil", "rollback": "Restaurar instalación", "about": "Acerca de Chuli"}[item.action]
		}
		label := fmt.Sprintf("%d. %s", index+1, name)
		if index == m.cursor {
			menu.WriteString(crtGreen.Render("› "+label) + "\n")
		} else {
			menu.WriteString("  " + label + "\n")
		}
	}
	explanations := map[string]string{
		"latest":       "Instalá el perfil y seguí trabajando en Codex.\nRevisás el plan antes de escribir.",
		"local":        "Instalá desde el checkout elegido con --source.\nRevisás el plan antes de escribir.",
		"integrations": "Conectá documentación, memoria y herramientas.\nElegís cada conexión y confirmás su configuración.",
		"profile":      "Consultá los modelos, agentes y skills registrados.\nLos datos vienen del manifiesto de instalación.",
		"doctor":       "Comprobá el perfil instalado y sus recursos.\nEl resultado distingue fallos y límites.",
		"rollback":     "Revisá el plan para restaurar el perfil anterior.\nLa reversión necesita tu confirmación.",
		"about":        "Conocé Chuli Manager y sus requisitos.\nEl trabajo de ingeniería se realiza en Codex.",
	}
	items := choices(m.mgr)
	explanation := explanations[items[max(0, min(m.cursor, len(items)-1))].action]
	if width >= 100 {
		leftWidth := 58
		left := lipgloss.NewStyle().Width(leftWidth).Render(menu.String())
		right := lipgloss.NewStyle().Width(width - leftWidth - 3).Render("\n\n" + explanation)
		return lipgloss.JoinHorizontal(lipgloss.Top, left, "   ", right)
	}
	return menu.String() + "\n" + explanation
}

func (m model) crtOutcome() string {
	if m.fault != "" {
		return crtBeige.Render("ERROR / No se completó la operación")
	}
	return crtGreen.Render("COMPLETADO")
}

func (m model) crtPage() string {
	// page retains the same offset and line count used by Update. Recolor only
	// its generated pagination hint, after sanitizing external content above.
	return ansi.Strip(m.page())
}

func (m model) crtIntegration(width int, page string) (string, string) {
	var b strings.Builder
	b.WriteString(crtGreen.Render("Integraciones") + "\n\n")
	if m.screen == integrationList {
		for index, item := range integrations.Catalog() {
			label := item.Name
			if width >= 90 {
				label += " · " + item.Purpose
			}
			if index == m.integrationCursor {
				b.WriteString(crtGreen.Render("› "+label) + "\n")
			} else {
				b.WriteString("  " + label + "\n")
			}
		}
		b.WriteString("\nElegí una opción para revisar estado, requisitos y guía.")
		return b.String(), "↑↓ / j k navegar · Enter elegir · Esc volver · q salir"
	}
	item, _ := integrations.Find(m.integrationID)
	b.WriteString(item.Name + "\n")
	if m.screen == integrationPreview {
		b.WriteString(crtGreen.Render("REVISIÓN PREVIA") + "\n")
	} else if m.screen == integrationResult || m.fault != "" {
		b.WriteString(m.crtOutcome() + "\n")
	}
	b.WriteString("\n" + page)
	if m.screen == integrationPreview {
		return b.String(), "S / Y confirmar · N cancelar · Esc volver · q salir\n↑↓ / j k / PgUp / PgDn desplazar"
	}
	keys := ""
	if m.integrationID != "egolite" {
		keys = "C configurar · D deshacer\n"
	}
	keys += "V comprobar · G guía · O abrir sitio oficial\nEnter revisar ficha · Esc volver · q salir\n↑↓ / j k / PgUp / PgDn desplazar"
	if width < 70 {
		keys = strings.ReplaceAll(keys, " · O abrir", "\nO abrir")
		keys = strings.ReplaceAll(keys, " · Esc volver", "\nEsc volver")
	}
	return b.String(), keys
}

func (m model) crtPaged() bool {
	switch m.screen {
	case preview, conflict, result, integrationDetail, integrationPreview, integrationResult:
		return true
	}
	return false
}

func (m model) crtPageSize() int {
	width := m.width
	if width <= 0 {
		width = 80
	}
	overhead := lipgloss.Height(m.crtFrame(""))
	total := len(m.lines())
	hint := fmt.Sprintf("Líneas %d–%d de %d · ↑↓ / PgUp / PgDn", total, total, total)
	hintRows := lipgloss.Height(lipgloss.NewStyle().Width(width).Render(hint))
	return max(1, m.height-overhead-hintRows)
}
