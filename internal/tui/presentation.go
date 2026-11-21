package tui

import (
	"fmt"
	"strings"
	"unicode"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/AlejoPalavecino/codex-engineering-harness/v2/internal/integrations"
	"github.com/AlejoPalavecino/codex-engineering-harness/v2/internal/manager"
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
		ver = "2.0.2"
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
	prefix := "Carpeta " + clientSpec.Short + ": "
	homeLabel := m.mgr.Home
	keep := max(1, width-lipgloss.Width(prefix)-1)
	if w := lipgloss.Width(homeLabel); w > keep {
		if keep <= 1 {
			homeLabel = "…"
		} else {
			homeLabel = ansi.TruncateLeft(homeLabel, w-keep+1, "…")
		}
	}
	b.WriteString(prefix + homeLabel + "\n")
	if m.screen == home && (m.updateReport.AnyProfileUpdate || m.updateReport.ManagerNewer) && m.updateReport.Latest.Tag != "" {
		notice := fmt.Sprintf("★ Nueva versión %s disponible — Entrá en Actualizar para ver novedades", crtText(m.updateReport.Latest.Tag))
		if width < 68 {
			notice = fmt.Sprintf("★ Hay actualización %s (ver Actualizar)", crtText(m.updateReport.Latest.Tag))
		}
		b.WriteString(crtAmber.Render(notice) + "\n")
	}
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

func renderProgressBar(step, total, width int) string {
	if total <= 0 {
		total = 1
	}
	if step < 0 {
		step = 0
	}
	if step > total {
		step = total
	}
	barWidth := min(24, max(10, width-18))
	filled := (step * barWidth) / total
	if filled == 0 && step > 0 {
		filled = 1
	}
	return "[" + crtNeon.Render(strings.Repeat("█", filled)) + crtCaseSide.Render(strings.Repeat("░", barWidth-filled)) + fmt.Sprintf("] %d/%d", step, total)
}

func (m model) crtBody(width int, page string) (string, string) {
	if strings.HasPrefix(m.screen, "integration-") {
		return m.crtIntegration(width, page)
	}
	switch m.screen {
	case installPlatformsScreen, installProfileScreen, installConfirmScreen, installGentleScreen:
		return m.crtWizard(width, page)
	case updateScreen, versionsScreen:
		return m.crtVersions(width, page)
	case skillsScreen, skillsAddScreen:
		return m.crtSkills(width)
	case home:
		keys := "↑↓ navegar · Enter elegir · Tab cliente · D verificar · ? ayuda · q salir"
		if width < 72 {
			keys = "↑↓ nav · Enter elegir · D verificar · q salir"
		}
		return m.crtHome(width), keys
	case modelsScreen, modelsSaveScreen, rolePickerScreen:
		return m.crtModels(width)
	case busy:
		spin := spinnerFrames[m.spinnerIdx%len(spinnerFrames)]
		label := crtText(m.progressLabel)
		if label == "" {
			label = "Preparando configuración..."
		}
		bar := renderProgressBar(m.progressStep, m.progressTotal, width)
		var out strings.Builder
		out.WriteString(crtGreen.Render("EN CURSO · "+spin) + " " + crtAmber.Render(m.title) + "\n\n")
		out.WriteString(bar + "\n\n")
		out.WriteString(crtBeige.Render(label) + "\n\n")
		keys := ""
		if m.busyCancelable {
			out.WriteString("Consultando GitHub Releases...")
			keys = "Esc cancelar · q salir"
		} else {
			out.WriteString("Estamos trabajando por vos. No cierres la terminal.")
		}
		return out.String(), keys
	case preview:
		return crtGreen.Render("REVISIÓN PREVIA / "+m.prepared.Version) + "\n\n" + page, "S / Y confirmar · N cancelar · Esc volver · q salir\n↑↓ / j k / PgUp / PgDn desplazar"
	case conflict:
		return crtBeige.Render("CONFLICTO / Gentle controla el enrutamiento") + "\n\nA · Usar Chuli como orquestador y respaldar Gentle\nPrimero verás el plan. La confirmación se solicita después.\n\n" + page, "A revisar plan · Esc volver · q salir\n↑↓ / j k / PgUp / PgDn desplazar"
	case result:
		if m.resultSummary != "" && !m.showTechLog {
			var res strings.Builder
			res.WriteString(m.crtOutcome() + "\n")
			if m.title != "" {
				res.WriteString(crtAmber.Render(m.title) + "\n\n")
			}
			if m.fault != "" {
				res.WriteString(crtAmber.Render("Motivo: ") + crtText(m.fault) + "\n\n")
			}
			res.WriteString(crtText(m.resultSummary))
			if m.resultNext != "" {
				res.WriteString("\n\n" + crtGreen.Render("Próximo paso:") + "\n" + crtText(m.resultNext))
			}
			keys := "Enter / Esc volver al inicio · q salir"
			if m.log != "" {
				keys = "V ver detalle técnico · " + keys
			}
			return res.String(), keys
		}
		keys := "Enter / Esc volver · q salir\n↑↓ / j k / PgUp / PgDn desplazar"
		if m.resultSummary != "" {
			keys = "V ocultar detalle técnico · " + keys
		}
		return m.crtOutcome() + "\n" + m.title + "\n\n" + page, keys
	case about:
		return crtGreen.Render("Acerca de Chuli") + "\n\nChuli es el harness de ingeniería para Codex, Claude Code y AGY CLI.\nChuli Manager instala, actualiza, versiona y configura modelos, MCPs y skills.\nDespués de instalar, abrí tu herramienta habitual (codex, claude o agy).\n\nEl paquete oficial se verifica con SHA-256 y crea respaldos automáticos.\nRequisitos: Python 3.11+ y PowerShell (pwsh en Linux y macOS).", "Esc volver · q salir"
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
	orchEffort := roles["orchestrator"].Effort
	if clientSpec.ID == "agy" && roles["orchestrator"].Label != "" {
		orchEffort = "variante AGY"
	}
	if compact {
		return "\n\n" + crtAmber.Render("◆ Equipo activo: "+clientSpec.Name) + "\n" +
			crtBeige.Render("Preset: "+presetID) + "\n" +
			fmt.Sprintf("• Orquestador: %s (%s)\n", roles["orchestrator"].Model, orchEffort) +
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
		if rr.id == "orchestrator" {
			tag = effortTag(orchEffort)
		}
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
		if item.action == "update" && m.updateReport.AnyProfileUpdate && m.updateReport.Latest.Tag != "" {
			if width >= 54 {
				name += " [NUEVA " + m.updateReport.Latest.Tag + "]"
			} else {
				name += " [★]"
			}
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
		"install":      "Elegí AGY, Claude o Codex, seleccioná un preset o personalizá\ntus modelos e instalá paso a paso con animación limpia.",
		"update":       "Detecta nuevas releases, muestra un resumen de qué incluyen\ny actualiza tus plataformas ya instaladas.",
		"versions":     "Mirá qué versión tenés en cada plataforma, cambiá a otra\nrelease o restaurá la instalación anterior (rollback).",
		"models":       "Configurá presets, modelos por rol o perfiles propios\nsegún el proveedor (Codex, Claude o AGY) y aplicalos directo.",
		"integrations": "Conectá, probá o agregá servidores MCP por plataforma\n(Context7, Engram, Exa, GitHub o MCPs personalizados).",
		"skills":       "Explorá qué hace cada skill, prendelas o apagalas por\nplataforma y sumá tus propias skills locales.",
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
	if compactHeight && len(items) > 6 {
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
	spec, _ := manager.FindClient(m.mgr.Target)
	var b strings.Builder
	b.WriteString(crtGreen.Render("MCPs e Integraciones") + "  " + renderClientTabs(spec.ID) + "\n\n")

	if m.screen == integrationCustomForm {
		tipo := "URL remota (HTTPS)"
		targetLabel := "URL del servidor (https://...)"
		if m.customMCPIsCmd {
			tipo = "Comando local (stdio)"
			targetLabel = "Comando y argumentos (ej. npx -y ...)"
		}
		fields := []struct{ label, val string }{
			{"Nombre corto", m.customMCPName},
			{"Tipo (←/→)", tipo},
			{targetLabel, m.customMCPTarget},
			{"Variable env token (opcional)", m.customMCPEnv},
		}
		b.WriteString(crtAmber.Render("Agregar MCP personalizado para "+spec.Short) + "\n\n")
		for idx, f := range fields {
			val := f.val
			if idx == m.customMCPField && idx != 1 {
				val += "_"
			}
			line := fmt.Sprintf("%-30s: %s", f.label, val)
			if width < 64 {
				line = f.label + ": " + val
			}
			if idx == m.customMCPField {
				b.WriteString(crtNeon.Render("› "+line) + "\n")
			} else {
				b.WriteString("  " + line + "\n")
			}
		}
		if m.fault != "" {
			b.WriteString("\n" + crtAmber.Render(m.fault))
		}
		return strings.TrimRight(b.String(), "\n"), "↑↓ / Tab campo · ←→ tipo · Enter continuar · Esc cancelar"
	}

	if m.screen == integrationList {
		items := m.mcpListItems()
		for index, item := range items {
			badge := "[ ]"
			if item.Active {
				badge = "[✓]"
			}
			label := fmt.Sprintf("%s %s · %s", badge, item.Name, item.Purpose)
			if width < 78 {
				label = fmt.Sprintf("%s %s", badge, item.Name)
			}
			if index == m.integrationCursor {
				b.WriteString(crtGreen.Render("› "+label) + "\n")
			} else {
				b.WriteString("  " + label + "\n")
			}
		}
		b.WriteString("\nElegí una opción o presioná N para sumar un MCP propio.")
		if m.fault != "" {
			b.WriteString("\n" + crtAmber.Render(m.fault))
		}
		return b.String(), "↑↓ navegar · Enter elegir · N nuevo MCP · Tab proveedor · Esc volver"
	}

	name := m.integrationID
	if item, ok := integrations.Find(m.integrationID); ok {
		name = item.Name
	}
	b.WriteString(name + "\n")
	if m.screen == integrationPreview {
		b.WriteString(crtGreen.Render("REVISIÓN PREVIA") + "\n")
	} else if m.screen == integrationResult || m.fault != "" {
		b.WriteString(m.crtOutcome() + "\n")
	}
	b.WriteString("\n" + page)
	if m.screen == integrationPreview {
		return b.String(), "S / Y confirmar · N cancelar · Esc volver · q salir\n↑↓ / j k / PgUp / PgDn desplazar"
	}
	_, isCatalog := integrations.Find(m.integrationID)
	keys := ""
	if !isCatalog {
		keys = "D quitar si es de Chuli · V comprobar\nEnter actualizar ficha · Esc volver · q salir"
		return b.String(), keys
	}
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
	case preview, conflict, result, updateScreen, integrationDetail, integrationPreview, integrationResult:
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
