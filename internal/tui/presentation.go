package tui

import (
	"fmt"
	"strings"
	"unicode"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/AlejoPalavecino/codex-engineering-harness/internal/integrations"
	"github.com/AlejoPalavecino/codex-engineering-harness/internal/manager"
	"github.com/charmbracelet/x/ansi"
)

var (
	crtGreen    = lipgloss.NewStyle().Foreground(lipgloss.Color("#9BE79B")).Bold(true)
	crtNeon     = lipgloss.NewStyle().Foreground(lipgloss.Color("#5CE63C")).Bold(true)
	crtShadow   = lipgloss.NewStyle().Foreground(lipgloss.Color("#2D6E36"))
	crtAmber    = lipgloss.NewStyle().Foreground(lipgloss.Color("#F4C97A")).Bold(true)
	crtCyan     = lipgloss.NewStyle().Foreground(lipgloss.Color("#79CFFF"))
	crtBeige    = lipgloss.NewStyle().Foreground(lipgloss.Color("#E6D8B8"))
	crtCaseSide = lipgloss.NewStyle().Foreground(lipgloss.Color("#B89768"))
	crtWhite    = lipgloss.NewStyle().Foreground(lipgloss.Color("#F2F4EB")).Background(lipgloss.Color("#101713"))
)

func colorBlockShadow(line string) string {
	var out strings.Builder
	var run strings.Builder
	mode := byte(0) // 0=space, 1=neon, 2=shadow
	flush := func() {
		if run.Len() == 0 {
			return
		}
		switch mode {
		case 1:
			out.WriteString(crtNeon.Render(run.String()))
		case 2:
			out.WriteString(crtShadow.Render(run.String()))
		default:
			out.WriteString(run.String())
		}
		run.Reset()
	}
	for _, r := range line {
		next := byte(1)
		if r == ' ' {
			next = 0
		} else if r == '░' {
			next = 2
		}
		if next != mode {
			flush()
			mode = next
		}
		run.WriteRune(r)
	}
	flush()
	return out.String()
}

func renderClientTabs(activeID string) string {
	var tabs []string
	for idx, c := range manager.Clients() {
		label := fmt.Sprintf("%d:%s", idx+1, c.Short)
		if c.ID == activeID {
			tabs = append(tabs, crtNeon.Render("[● "+label+"]"))
		} else {
			tabs = append(tabs, crtBeige.Render("[○ "+label+"]"))
		}
	}
	return strings.Join(tabs, "  ")
}

func renderChuliLogo(clientID, version string) string {
	ver := crtText(version)
	if ver == "" {
		ver = "2.0.0"
	}
	if lipgloss.Width(ver) > 11 {
		ver = ansi.Truncate(ver, 11, "")
	}
	// Each PC row is 20 visible columns wide to align with the 20-col chassis and
	// leave room for the 37-col CHULI block wordmark inside a 60-col terminal.
	pc := []string{
		"  " + crtBeige.Render("▄█▀▀▀▀▀▀▀▀▀▀▀▀█▄") + crtCaseSide.Render("░") + " ",
		"  " + crtBeige.Render("██") + "  " + crtNeon.Render("▀▄") + "        " + crtBeige.Render("██") + crtCaseSide.Render("░") + " ",
		"  " + crtBeige.Render("██") + "  " + crtNeon.Render("▄▀ ▄▄▄") + "    " + crtBeige.Render("██") + crtCaseSide.Render("░") + " ",
		"  " + crtBeige.Render("▀█▄▄▄▄▄▄▄▄▄▄▄▄█▀") + crtCaseSide.Render("░") + " ",
		" " + crtBeige.Render("▄▄▄▄▄▄▄████▄▄▄▄▄▄▄") + crtCaseSide.Render("░"),
		" " + crtBeige.Render("███▄▄▄██████▀▀▀███") + crtCaseSide.Render("░"),
		" " + crtBeige.Render("▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀") + " ",
	}
	right := []string{
		crtNeon.Render("●") + " " + crtAmber.Render("MULTI-CLIENT HARNESS") + crtBeige.Render(" · ") + crtCyan.Render("v"+ver),
		colorBlockShadow("▄█████▄ ██░ ██░ ██░ ██░ ██░    █████░"),
		colorBlockShadow("██░  ▀▀ ██░ ██░ ██░ ██░ ██░     ▐█▌░ "),
		colorBlockShadow("██░     ██████░ ██░ ██░ ██░     ▐█▌░ "),
		colorBlockShadow("██░  ▄▄ ██░ ██░ ██░ ██░ ██░     ▐█▌░ "),
		colorBlockShadow("▀█████▀ ██░ ██░ ▀████▀░ █████░ █████░"),
		renderClientTabs(clientID),
	}
	var out strings.Builder
	for i := range pc {
		out.WriteString(pc[i] + "  " + right[i])
		if i < len(pc)-1 {
			out.WriteByte('\n')
		}
	}
	return out.String()
}

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
	clientSpec, ok := manager.FindClient(m.mgr.Target)
	if !ok {
		clientSpec, _ = manager.FindClient("codex")
	}
	presetLabel := m.mgr.Preset
	if presetLabel == "" {
		presetLabel = clientSpec.DefaultPreset
	}
	var b strings.Builder
	if m.screen == home && width >= 60 && (m.height == 0 || m.height >= 24) {
		b.WriteString(renderChuliLogo(clientSpec.ID, m.managerVersion) + "\n")
	} else if m.screen == home {
		b.WriteString(crtNeon.Render("▐▛▀▜▌") + " " + crtGreen.Render("[ ● CHULI ]") + "\n")
	}
	headerLine := crtGreen.Render("CHULI") + " / " + crtBeige.Render("Manager: "+crtText(m.managerVersion))
	if width >= 68 {
		headerLine += " · " + crtAmber.Render("Cliente: "+clientSpec.Short) + " (" + crtCyan.Render(presetLabel) + ")"
	}
	b.WriteString(headerLine + "\n")
	b.WriteString("Perfil: " + installed + "\n")
	homeLabel := m.mgr.Home
	keep := max(1, width-12)
	if w := lipgloss.Width(homeLabel); w > keep {
		if keep <= 1 {
			homeLabel = "…"
		} else {
			homeLabel = ansi.TruncateLeft(homeLabel, w-keep+1, "…")
		}
	}
	b.WriteString("CODEX_HOME: " + homeLabel + "\n")
	b.WriteString(crtBeige.Render(strings.Repeat("━", min(width, 76))) + "\n")
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
		keys := "↑↓ / j k navegar · Enter elegir · Tab cliente · M modelos · q salir"
		if width < 68 {
			keys = "↑↓ / j k navegar · Enter elegir · q salir"
		}
		return m.crtHome(width), keys
	case modelsScreen:
		return m.crtModels(width)
	case busy:
		return crtGreen.Render("EN CURSO") + "\n" + m.title + "\n\nEjecutando comprobaciones.\nNo cierres la terminal durante una escritura.", ""
	case preview:
		return crtGreen.Render("REVISIÓN PREVIA / "+m.prepared.Version) + "\n\n" + page, "S / Y confirmar · N cancelar · Esc volver · q salir\n↑↓ / j k / PgUp / PgDn desplazar"
	case conflict:
		return crtBeige.Render("CONFLICTO / Gentle controla el enrutamiento") + "\n\nA · Usar Chuli como orquestador y respaldar Gentle\nPrimero verás el plan. La confirmación se solicita después.\n\n" + page, "A revisar plan · Esc volver · q salir\n↑↓ / j k / PgUp / PgDn desplazar"
	case result:
		return m.crtOutcome() + "\n" + m.title + "\n\n" + page, "Enter / Esc volver · q salir\n↑↓ / j k / PgUp / PgDn desplazar"
	case about:
		return crtGreen.Render("Acerca de Chuli") + "\n\nChuli es el harness de ingeniería para Codex, Claude Code y AGY CLI.\nChuli Manager instala, personaliza modelos por agente y actualiza el perfil.\nDespués de instalar, abrí tu cliente habitual (codex, claude o agy).\n\nEl paquete oficial se verifica con SHA-256 y se conserva para recuperación.\nRequisitos: Python 3.11+ y PowerShell (pwsh en Linux y macOS).", "Esc volver · q salir"
	}
	return "", "Esc volver · q salir"
}

func effortTag(effort string) string {
	short := strings.ToUpper(effort)
	switch short {
	case "MEDIUM":
		short = "MED"
	case "XHIGH":
		short = "XHIGH"
	}
	tag := "[" + short + "]"
	switch effort {
	case "high", "xhigh":
		return crtAmber.Render(tag)
	case "medium":
		return crtCyan.Render(tag)
	default:
		return crtBeige.Render(tag)
	}
}

func renderTeamCard(clientSpec manager.ClientSpec, presetID string, roles map[string]manager.RoleModel, cardWidth int, compact bool) string {
	if compact {
		return "\n\n" + crtAmber.Render("◆ Equipo activo: "+clientSpec.Name) + "\n" +
			crtBeige.Render("Preset: "+presetID) + "\n" +
			fmt.Sprintf("• Orquestador: %s (%s)\n", roles["orchestrator"].Model, roles["orchestrator"].Effort) +
			fmt.Sprintf("• Arquitecto:  %s (%s)\n", roles["solution_architect"].Model, roles["solution_architect"].Effort) +
			fmt.Sprintf("• Worker:      %s (%s)\n", roles["chuli_worker"].Model, roles["chuli_worker"].Effort) +
			fmt.Sprintf("• Explorer:    %s (%s)", roles["chuli_explorer"].Model, roles["chuli_explorer"].Effort)
	}
	outer := min(50, max(44, cardWidth))
	inner := outer - 4
	title := "◆ Equipo activo · " + clientSpec.Short + " "
	fill := max(1, inner+2-lipgloss.Width(title)-1)
	top := crtBeige.Render("╭─ ") + crtAmber.Render(title) + crtBeige.Render(strings.Repeat("─", fill)+"╮")
	bottom := crtBeige.Render("╰" + strings.Repeat("─", outer-2) + "╯")

	roleRows := []struct {
		id    string
		label string
	}{
		{"orchestrator", "Orquestador"},
		{"solution_architect", "Arquitecto"},
		{"quality_reviewer", "Revisor QA"},
		{"chuli_worker", "Worker"},
		{"test_engineer", "Tester"},
		{"product_designer", "Diseñador"},
		{"chuli_explorer", "Explorer"},
	}

	var b strings.Builder
	b.WriteString("\n\n" + top + "\n")
	presetLine := "Preset: " + presetID
	padPreset := max(0, inner-lipgloss.Width(presetLine))
	b.WriteString(crtBeige.Render("│ ") + crtCyan.Render(presetLine) + strings.Repeat(" ", padPreset) + crtBeige.Render(" │") + "\n")
	for _, rr := range roleRows {
		rm := roles[rr.id]
		tag := effortTag(rm.Effort)
		tagW := lipgloss.Width(ansi.Strip(tag))
		leftPart := fmt.Sprintf("• %-11s %-19s", rr.label, ansi.Truncate(rm.Model, 19, "…"))
		gap := max(1, inner-lipgloss.Width(leftPart)-tagW)
		b.WriteString(crtBeige.Render("│ ") + leftPart + strings.Repeat(" ", gap) + tag + crtBeige.Render(" │") + "\n")
	}
	b.WriteString(bottom)
	return b.String()
}

func (m model) crtHome(width int) string {
	var menu strings.Builder
	items := choices(m.mgr)
	compactHeight := m.height > 0 && m.height <= 25
	gap := "\n\n"
	if width < 60 || compactHeight {
		gap = "\n"
	}
	menu.WriteString(crtGreen.Render("Seleccioná una acción") + gap)
	for index, item := range items {
		name := item.label
		if width < 60 {
			name = map[string]string{
				"latest":       "Instalar o actualizar",
				"local":        "Instalar desde --source",
				"models":       "Modelos y presets",
				"integrations": "Integraciones",
				"profile":      "Modelos, agentes y skills",
				"doctor":       "Comprobar perfil",
				"rollback":     "Restaurar instalación",
				"about":        "Acerca de Chuli",
			}[item.action]
		}
		label := fmt.Sprintf("%d. %s", index+1, name)
		if index == m.cursor {
			marker := "› "
			if width >= 100 {
				marker = "▎› "
			}
			menu.WriteString(crtNeon.Render(marker+label) + "\n")
		} else {
			prefix := "  "
			if width >= 100 {
				prefix = "   "
			}
			menu.WriteString(prefix + label + "\n")
		}
	}
	explanations := map[string]string{
		"latest":       "Instalá el perfil y seguí trabajando en Codex.\nRevisás el plan antes de escribir.",
		"local":        "Instalá desde el checkout elegido con --source.\nRevisás el plan antes de escribir.",
		"models":       "Elegí el cliente (Codex, Claude Code o AGY CLI), un preset rápido\no el modelo y esfuerzo de cada uno de los 7 agentes.",
		"integrations": "Conectá documentación, memoria y herramientas.\nElegís cada conexión y confirmás su configuración.",
		"profile":      "Consultá los modelos, agentes y skills registrados.\nLos datos vienen del manifiesto de instalación.",
		"doctor":       "Comprobá el perfil instalado y sus recursos.\nEl resultado distingue fallos y límites.",
		"rollback":     "Revisá el plan para restaurar el perfil anterior.\nLa reversión necesita tu confirmación.",
		"about":        "Conocé Chuli Manager y sus requisitos.\nEl trabajo de ingeniería se realiza en Codex.",
	}
	explanation := explanations[items[max(0, min(m.cursor, len(items)-1))].action]
	if width >= 100 {
		clientSpec, ok := manager.FindClient(m.mgr.Target)
		if !ok {
			clientSpec, _ = manager.FindClient("codex")
		}
		presetID := m.mgr.Preset
		if presetID == "" {
			presetID = clientSpec.DefaultPreset
		}
		roles := m.mgr.RoleModels
		if len(roles) == 0 {
			_, roles, _ = manager.DefaultRolesForPreset(clientSpec.ID, presetID)
		}
		leftWidth := 52
		rightWidth := width - leftWidth - 2
		card := renderTeamCard(clientSpec, presetID, roles, rightWidth, compactHeight)
		left := lipgloss.NewStyle().Width(leftWidth).Render(menu.String())
		right := lipgloss.NewStyle().Width(rightWidth).Render(gap + explanation + card)
		return lipgloss.JoinHorizontal(lipgloss.Top, left, "  ", right)
	}
	separator := "\n"
	if compactHeight && len(items) > 7 {
		separator = ""
	}
	return menu.String() + separator + explanation
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
