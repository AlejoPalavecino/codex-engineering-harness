package tui

import (
	"fmt"
	"strings"

	tea "charm.land/bubbletea/v2"
	"github.com/AlejoPalavecino/codex-engineering-harness/v2/internal/manager"
)

const (
	skillsScreen    = "skills-view"
	skillsAddScreen = "skills-add"
)

type skillsLoadedMsg struct {
	items []manager.SkillInfo
	err   error
}

type customSkillActionMsg struct {
	msg string
	err error
}

func loadSkillsCmd(mgr manager.Manager) tea.Cmd {
	return func() tea.Msg {
		items, err := mgr.ListSkills()
		return skillsLoadedMsg{items: items, err: err}
	}
}

func addCustomSkillCmd(mgr manager.Manager, path string) tea.Cmd {
	return func() tea.Msg {
		info, err := mgr.AddCustomSkill(path)
		if err != nil {
			return customSkillActionMsg{err: err}
		}
		return customSkillActionMsg{msg: "Skill personalizada agregada: " + info.Name}
	}
}

func removeCustomSkillCmd(mgr manager.Manager, name string) tea.Cmd {
	return func() tea.Msg {
		err := mgr.RemoveCustomSkill(name)
		if err != nil {
			return customSkillActionMsg{err: err}
		}
		return customSkillActionMsg{msg: "Skill personalizada eliminada: " + name}
	}
}

func (m *model) syncSkillDraftFromList(items []manager.SkillInfo) {
	m.skillsList = items
	m.skillsDraftDisabled = map[string]bool{}
	for _, s := range items {
		if !s.Enabled && s.Source == "chuli" {
			m.skillsDraftDisabled[s.Name] = true
		}
	}
	m.skillsDirty = false
	if m.skillsCursor >= len(items) {
		m.skillsCursor = max(0, len(items)-1)
	}
}

func (m model) disabledSkillsSlice() []string {
	var out []string
	for _, s := range m.skillsList {
		if s.Source == "chuli" && m.skillsDraftDisabled[s.Name] {
			out = append(out, s.Name)
		}
	}
	return out
}

func (m model) updateSkills(message tea.Msg) (model, tea.Cmd, bool) {
	switch msg := message.(type) {
	case skillsLoadedMsg:
		if msg.err != nil {
			m.fault = msg.err.Error()
		} else {
			m.fault = ""
			m.syncSkillDraftFromList(msg.items)
		}
		return m, nil, true
	case customSkillActionMsg:
		if msg.err != nil {
			m.fault = msg.err.Error()
			m.screen = skillsScreen
			return m, nil, true
		}
		m.fault = ""
		m.skillNotice = msg.msg
		m.screen = skillsScreen
		return m, loadSkillsCmd(m.mgr), true
	}

	if m.screen != skillsScreen && m.screen != skillsAddScreen {
		return m, nil, false
	}
	msg, ok := message.(tea.KeyPressMsg)
	if !ok {
		return m, nil, false
	}
	key := strings.ToLower(msg.String())

	if m.screen == skillsAddScreen {
		switch key {
		case "ctrl+c":
			return m, tea.Quit, true
		case "esc":
			m.screen = skillsScreen
			m.skillPathInput = ""
			m.fault = ""
			return m, nil, true
		case "backspace":
			if len(m.skillPathInput) > 0 {
				runes := []rune(m.skillPathInput)
				m.skillPathInput = string(runes[:len(runes)-1])
			}
			return m, nil, true
		case "enter":
			path := strings.TrimSpace(m.skillPathInput)
			if path == "" {
				m.fault = "Pegá o escribí la ruta a la carpeta de la skill."
				return m, nil, true
			}
			m.skillPathInput = ""
			return m, addCustomSkillCmd(m.mgr, path), true
		default:
			if len(msg.Text) > 0 {
				m.skillPathInput += msg.Text
			}
			return m, nil, true
		}
	}

	switch key {
	case "q", "ctrl+c":
		return m, tea.Quit, true
	case "esc":
		m.screen = home
		m.fault = ""
		m.skillNotice = ""
		return m, nil, true
	case "tab":
		m.nextClient()
		m.skillNotice = ""
		return m, loadSkillsCmd(m.mgr), true
	case "1":
		m.switchClient("codex")
		m.skillNotice = ""
		return m, loadSkillsCmd(m.mgr), true
	case "2":
		m.switchClient("claude")
		m.skillNotice = ""
		return m, loadSkillsCmd(m.mgr), true
	case "3":
		m.switchClient("agy")
		m.skillNotice = ""
		return m, loadSkillsCmd(m.mgr), true
	case "up", "k":
		m.skillsCursor = max(0, m.skillsCursor-1)
	case "down", "j":
		m.skillsCursor = min(max(0, len(m.skillsList)-1), m.skillsCursor+1)
	case " ", "space":
		if len(m.skillsList) == 0 {
			return m, nil, true
		}
		item := m.skillsList[m.skillsCursor]
		if item.Source != "chuli" {
			m.fault = "Las skills propias se quitan con X (no se desactivan por manifiesto)."
			return m, nil, true
		}
		m.fault = ""
		m.skillsDraftDisabled[item.Name] = !m.skillsDraftDisabled[item.Name]
		m.skillsDirty = true
	case "a":
		m.screen = skillsAddScreen
		m.skillPathInput = ""
		m.fault = ""
	case "x":
		if len(m.skillsList) == 0 {
			return m, nil, true
		}
		item := m.skillsList[m.skillsCursor]
		if item.Source != "custom" {
			m.fault = "Para ocultar una skill oficial usá Espacio para desactivarla."
			return m, nil, true
		}
		return m, removeCustomSkillCmd(m.mgr, item.Name), true
	case "enter", "s":
		if !m.skillsDirty {
			return m, nil, true
		}
		if m.installed == "" && m.mgr.Source == "" {
			m.fault = "Primero instalá Chuli en esta plataforma para guardar cambios de skills."
			return m, nil, true
		}
		applyMgr := m.mgr
		applyMgr.DisabledSkills = m.disabledSkillsSlice()
		// Keep currently installed preset/models if present
		for _, st := range m.statuses {
			if st.Spec.ID == applyMgr.Target {
				if st.Preset != "" {
					applyMgr.Preset = st.Preset
				}
				if len(st.RoleModels) > 0 {
					applyMgr.RoleModels = st.RoleModels
				}
			}
		}
		m.screen = busy
		m.title = "Actualizando skills"
		m.progressStep, m.progressTotal = 1, 3
		m.progressLabel = "Aplicando selección de skills..."
		tr := &progressTracker{ch: make(chan progressEvent, 16)}
		m.activeTracker = tr
		return m, tea.Batch(
			runSingleActionCmd(applyMgr, "install", manager.PrepareOptions{Action: "apply-local"}, "Skills actualizadas", tr),
			waitProgressCmd(tr),
			animTickCmd(),
		), true
	}
	return m, nil, true
}

func (m model) crtSkills(width int) (string, string) {
	spec, _ := manager.FindClient(m.mgr.Target)
	var b strings.Builder

	if m.screen == skillsAddScreen {
		b.WriteString(crtGreen.Render("Agregar skill propia") + " · " + spec.Short + "\n\n")
		b.WriteString("Indicá la ruta de la carpeta que contiene SKILL.md:\n\n")
		b.WriteString(crtNeon.Render("> "+m.skillPathInput+"_") + "\n\n")
		b.WriteString(crtBeige.Render("La skill debe tener frontmatter con name (igual a la carpeta) y description."))
		if m.fault != "" {
			b.WriteString("\n\n" + crtAmber.Render(m.fault))
		}
		return b.String(), "Enter agregar · Esc cancelar"
	}

	b.WriteString(crtGreen.Render("Skills") + "  " + renderClientTabs(spec.ID) + "\n")
	if m.skillsDirty {
		b.WriteString(crtAmber.Render("● Cambios sin aplicar — Presioná S o Enter para guardar") + "\n")
	} else if m.skillNotice != "" {
		b.WriteString(crtNeon.Render("✓ "+m.skillNotice) + "\n")
	}
	b.WriteString("\n")

	if len(m.skillsList) == 0 {
		b.WriteString("No se encontraron skills instaladas ni paquete local para " + spec.Short + ".\n")
		b.WriteString("Instalá Chuli primero en esta plataforma o agregá una skill propia con A.")
	} else {
		maxRows := 9
		if m.height > 26 {
			maxRows = min(14, m.height-14)
		}
		start := max(0, m.skillsCursor-maxRows/2)
		if start+maxRows > len(m.skillsList) {
			start = max(0, len(m.skillsList)-maxRows)
		}
		end := min(len(m.skillsList), start+maxRows)
		for i := start; i < end; i++ {
			s := m.skillsList[i]
			enabled := s.Enabled
			if s.Source == "chuli" {
				enabled = !m.skillsDraftDisabled[s.Name]
			}
			mark := "[✓]"
			if !enabled {
				mark = "[ ]"
			}
			tag := "Chuli"
			if s.Source == "custom" {
				tag = "Propia"
			}
			line := fmt.Sprintf("%s %-28s (%s)", mark, s.Name, tag)
			if width < 58 {
				line = fmt.Sprintf("%s %s", mark, s.Name)
			}
			if i == m.skillsCursor {
				b.WriteString(crtNeon.Render("› "+line) + "\n")
			} else {
				b.WriteString("  " + line + "\n")
			}
		}
		sel := m.skillsList[m.skillsCursor]
		b.WriteString("\n" + crtAmber.Render(sel.Name+": ") + sel.Description)
	}

	if m.fault != "" {
		b.WriteString("\n\n" + crtAmber.Render(m.fault))
	}
	keys := "↑↓ navegar · Espacio activar/desactivar · Enter/S aplicar · A agregar · X quitar propia · Tab cliente · Esc volver"
	if width < 72 {
		keys = "↑↓ nav · Espacio on/off · Enter aplicar\nA agregar · Tab cliente · Esc volver"
	}
	return strings.TrimRight(b.String(), "\n"), keys
}
