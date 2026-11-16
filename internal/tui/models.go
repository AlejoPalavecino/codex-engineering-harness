package tui

import (
	"fmt"
	"strings"

	tea "charm.land/bubbletea/v2"
	"github.com/AlejoPalavecino/codex-engineering-harness/v2/internal/manager"
)

const modelsScreen = "models"

func (m *model) ensureModelState() {
	target := m.mgr.Target
	if target == "" {
		target = "codex"
		m.mgr.Target = target
	}
	spec, ok := manager.FindClient(target)
	if !ok {
		spec, _ = manager.FindClient("codex")
	}
	if m.mgr.Preset == "" {
		m.mgr.Preset = spec.DefaultPreset
	}
	if len(m.mgr.RoleModels) == 0 {
		_, roles, err := manager.DefaultRolesForPreset(spec.ID, m.mgr.Preset)
		if err != nil {
			_, roles, _ = manager.DefaultRolesForPreset(spec.ID, spec.DefaultPreset)
			m.mgr.Preset = spec.DefaultPreset
		}
		m.mgr.RoleModels = roles
	}
}

func (m *model) switchClient(target string) {
	spec, ok := manager.FindClient(target)
	if !ok {
		return
	}
	m.mgr.Target = spec.ID
	if home, err := manager.ResolveClientHome(spec.ID); err == nil {
		m.mgr.Home = home
	}
	m.mgr.Preset = spec.DefaultPreset
	_, roles, _ := manager.DefaultRolesForPreset(spec.ID, spec.DefaultPreset)
	m.mgr.RoleModels = roles
	if inst, err := m.mgr.Installed(); err == nil {
		m.installed = inst
	} else {
		m.installed = ""
	}
}

func (m *model) nextClient() {
	clients := manager.Clients()
	cur := m.mgr.Target
	if cur == "" {
		cur = "codex"
	}
	for i, c := range clients {
		if c.ID == cur {
			m.switchClient(clients[(i+1)%len(clients)].ID)
			return
		}
	}
	m.switchClient("codex")
}

func (m *model) cyclePreset() {
	m.ensureModelState()
	spec, ok := manager.FindClient(m.mgr.Target)
	if !ok || len(spec.Presets) == 0 {
		return
	}
	idx := 0
	for i, p := range spec.Presets {
		if p.ID == m.mgr.Preset {
			idx = (i + 1) % len(spec.Presets)
			break
		}
	}
	chosen := spec.Presets[idx]
	m.mgr.Preset = chosen.ID
	_, roles, _ := manager.DefaultRolesForPreset(spec.ID, chosen.ID)
	m.mgr.RoleModels = roles
}

func (m *model) cycleRoleModel(delta int) {
	m.ensureModelState()
	spec, ok := manager.FindClient(m.mgr.Target)
	if !ok || len(spec.Models) == 0 {
		return
	}
	roles := manager.Roles()
	if m.roleCursor < 0 || m.roleCursor >= len(roles) {
		m.roleCursor = 0
	}
	roleID := roles[m.roleCursor].ID
	current := m.mgr.RoleModels[roleID]
	idx := 0
	for i, mod := range spec.Models {
		if mod == current.Model {
			idx = (i + delta + len(spec.Models)) % len(spec.Models)
			break
		}
	}
	current.Model = spec.Models[idx]
	if current.Effort == "" && len(spec.Efforts) > 0 {
		current.Effort = "medium"
	}
	m.mgr.RoleModels[roleID] = current
	m.mgr.Preset = "custom"
}

func (m *model) cycleRoleEffort() {
	m.ensureModelState()
	spec, ok := manager.FindClient(m.mgr.Target)
	if !ok || len(spec.Efforts) == 0 {
		return
	}
	roles := manager.Roles()
	if m.roleCursor < 0 || m.roleCursor >= len(roles) {
		m.roleCursor = 0
	}
	roleID := roles[m.roleCursor].ID
	current := m.mgr.RoleModels[roleID]
	idx := 0
	for i, eff := range spec.Efforts {
		if eff == current.Effort {
			idx = (i + 1) % len(spec.Efforts)
			break
		}
	}
	current.Effort = spec.Efforts[idx]
	m.mgr.RoleModels[roleID] = current
	m.mgr.Preset = "custom"
}

func (m model) updateModels(message tea.Msg) (model, tea.Cmd, bool) {
	if m.screen != modelsScreen {
		return m, nil, false
	}
	msg, ok := message.(tea.KeyPressMsg)
	if !ok {
		return m, nil, false
	}
	m.ensureModelState()
	key := strings.ToLower(msg.String())
	roles := manager.Roles()
	switch key {
	case "q", "ctrl+c":
		return m, tea.Quit, true
	case "esc":
		m.screen = home
		return m, nil, true
	case "tab":
		m.nextClient()
		return m, nil, true
	case "1":
		m.switchClient("codex")
		return m, nil, true
	case "2":
		m.switchClient("claude")
		return m, nil, true
	case "3":
		m.switchClient("agy")
		return m, nil, true
	case "p":
		m.cyclePreset()
		return m, nil, true
	case "up", "k":
		m.roleCursor = max(0, m.roleCursor-1)
		return m, nil, true
	case "down", "j":
		m.roleCursor = min(len(roles)-1, m.roleCursor+1)
		return m, nil, true
	case "right", "l":
		m.cycleRoleModel(1)
		return m, nil, true
	case "left", "h":
		m.cycleRoleModel(-1)
		return m, nil, true
	case "e":
		m.cycleRoleEffort()
		return m, nil, true
	case "enter":
		action := "latest"
		if m.mgr.Source != "" {
			action = "local"
		}
		m.action = action
		m.title = "Revisando plan con modelos seleccionados"
		m.screen = busy
		m.offset = 0
		m.fault = ""
		m.log = ""
		return m, prepareCmd(m.mgr, action, false), true
	}
	return m, nil, true
}

func (m model) crtModels(width int) (string, string) {
	m.ensureModelState()
	var b strings.Builder
	spec, _ := manager.FindClient(m.mgr.Target)

	if width < 60 {
		b.WriteString(crtGreen.Render("Modelos · "+spec.Short) + " (" + m.mgr.Preset + ")\n")
	} else {
		b.WriteString(crtGreen.Render("Modelos y Presets") + "  " + renderClientTabs(spec.ID) + "\n")
		b.WriteString(crtAmber.Render("Preset: "+m.mgr.Preset) + " · P rotar preset · Tab cliente\n\n")
		b.WriteString(crtBeige.Render("   ROL               │ MODELO                  │ ESFUERZO") + "\n")
		b.WriteString(crtCaseSide.Render("   ──────────────────┼─────────────────────────┼─────────") + "\n")
	}

	roles := manager.Roles()
	for i, r := range roles {
		rm := m.mgr.RoleModels[r.ID]
		if width < 60 {
			roleLabel := strings.TrimPrefix(r.ID, "chuli_")
			if len(roleLabel) > 9 {
				roleLabel = roleLabel[:9]
			}
			line := fmt.Sprintf("%-9s %s (%s)", roleLabel, rm.Model, rm.Effort)
			if i == m.roleCursor {
				b.WriteString(crtNeon.Render("› "+line) + "\n")
			} else {
				b.WriteString("  " + line + "\n")
			}
			continue
		}
		sep := crtCaseSide.Render(" │ ")
		tag := effortTag(rm.Effort)
		if i == m.roleCursor {
			row := crtNeon.Render(fmt.Sprintf("▎› %-17s", r.ID)) + sep +
				crtAmber.Render("◀ ") + crtNeon.Render(fmt.Sprintf("%-19s", rm.Model)) + crtAmber.Render(" ▶") +
				sep + tag
			b.WriteString(row + "\n")
		} else {
			row := fmt.Sprintf("   %-17s", r.ID) + sep + fmt.Sprintf("  %-19s  ", rm.Model) + sep + tag
			b.WriteString(row + "\n")
		}
	}
	keys := "↑↓ rol · ←→ modelo · E esfuerzo · P preset · Tab cliente\nEnter previsualizar e instalar · Esc volver · q salir"
	if width < 65 {
		keys = "↑↓ rol · ←→ modelo · E esfuerzo · P preset\nTab cliente · Enter aplicar · Esc volver"
	}
	return strings.TrimRight(b.String(), "\n"), keys
}
