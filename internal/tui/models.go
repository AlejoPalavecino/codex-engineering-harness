package tui

import (
	"context"
	"fmt"
	"slices"
	"strings"

	tea "charm.land/bubbletea/v2"
	"github.com/AlejoPalavecino/codex-engineering-harness/v2/internal/manager"
)

type modelsLoadedMsg struct {
	target, source string
	request        int
	ids            []string
	labels         map[string]string
}

func loadModelsCmd(mgr manager.Manager, request int) tea.Cmd {
	return func() tea.Msg {
		spec, _ := manager.FindClient(mgr.Target)
		ids, source, labels := mgr.AvailableModelCatalog(context.Background(), spec)
		return modelsLoadedMsg{target: mgr.Target, source: source, request: request, ids: ids, labels: labels}
	}
}

func (m model) choicesForRole(spec manager.ClientSpec, role string) []string {
	ids := m.modelChoices
	if len(ids) == 0 {
		ids = spec.Models
	}
	// AGY plugin agents support tiers; exact model IDs configure its main agent.
	if spec.ID == "agy" && role != "orchestrator" {
		ids = []string{"inherit", "pro", "flash", "flash_lite"}
	}
	if current := m.mgr.RoleModels[role].Model; current != "" && !slices.Contains(ids, current) {
		return append(slices.Clone(ids), current)
	}
	return ids
}

const (
	modelsScreen     = "models"
	modelsSaveScreen = "models-save"
	rolePickerScreen = "models-role-picker"
)

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

func (m *model) reloadSavedProfiles() {
	target := m.mgr.Target
	if target == "" {
		target = "codex"
	}
	list, err := m.mgr.ListSavedProfiles(target)
	if err == nil {
		m.savedProfiles = list
	} else {
		m.savedProfiles = nil
	}
	m.savedProfileIdx = -1
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
	m.reloadSavedProfiles()
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
	m.savedProfileIdx = -1
}

func (m *model) cycleSavedProfile() {
	if len(m.savedProfiles) == 0 {
		m.reloadSavedProfiles()
	}
	if len(m.savedProfiles) == 0 {
		m.fault = "Todavía no guardaste perfiles propios para este proveedor (usá Guardar perfil)."
		return
	}
	m.fault = ""
	m.savedProfileIdx = (m.savedProfileIdx + 1) % len(m.savedProfiles)
	chosen := m.savedProfiles[m.savedProfileIdx]
	cloned := make(map[string]manager.RoleModel, len(chosen.RoleModels))
	for k, v := range chosen.RoleModels {
		cloned[k] = v
	}
	m.mgr.RoleModels = cloned
	m.mgr.Preset = "custom"
}

func (m *model) cycleRoleModel(delta int) {
	m.ensureModelState()
	spec, ok := manager.FindClient(m.mgr.Target)
	if !ok || len(spec.Models) == 0 {
		return
	}
	roles := manager.Roles()
	if m.roleCursor < 0 || m.roleCursor >= len(roles) {
		return
	}
	roleID := roles[m.roleCursor].ID
	current := m.mgr.RoleModels[roleID]
	spec.Models = m.choicesForRole(spec, roleID)
	idx := 0
	for i, mod := range spec.Models {
		if mod == current.Model {
			idx = (i + delta + len(spec.Models)) % len(spec.Models)
			break
		}
	}
	if current.Model != spec.Models[idx] || current.Label == "" {
		current.Label = m.modelLabels[spec.Models[idx]]
	}
	current.Model = spec.Models[idx]
	if current.Effort == "" && len(spec.Efforts) > 0 {
		current.Effort = "medium"
	}
	m.mgr.RoleModels[roleID] = current
	m.mgr.Preset = "custom"
	m.savedProfileIdx = -1
}

func (m *model) cycleRoleEffort() {
	m.ensureModelState()
	spec, ok := manager.FindClient(m.mgr.Target)
	if !ok || len(spec.Efforts) == 0 {
		return
	}
	roles := manager.Roles()
	if m.roleCursor < 0 || m.roleCursor >= len(roles) {
		return
	}
	roleID := roles[m.roleCursor].ID
	current := m.mgr.RoleModels[roleID]
	if m.mgr.Target == "agy" && roleID == "orchestrator" && current.Label != "" {
		return // The exact variant already encodes effort in its ID/label.
	}
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
	m.savedProfileIdx = -1
}

func (m *model) openRolePicker() {
	m.ensureModelState()
	roles := manager.Roles()
	if m.roleCursor < 0 || m.roleCursor >= len(roles) {
		return
	}
	spec, _ := manager.FindClient(m.mgr.Target)
	current := m.mgr.RoleModels[roles[m.roleCursor].ID]
	m.pickerChoices = m.choicesForRole(spec, roles[m.roleCursor].ID)
	m.rolePickerModelIdx = 0
	for i, mod := range m.pickerChoices {
		if mod == current.Model {
			m.rolePickerModelIdx = i
			break
		}
	}
	m.screen = rolePickerScreen
}

func (m model) confirmModelsAction() (model, tea.Cmd) {
	if m.wizardReturn {
		if m.wizardConfigs == nil {
			m.wizardConfigs = map[string]wizardClientConfig{}
		}
		m.wizardConfigs[m.mgr.Target] = wizardClientConfig{Preset: m.mgr.Preset, Roles: m.mgr.RoleModels}
		m.wizardReturn = false
		m.advanceWizardProvider()
		return m, nil
	}
	if m.installed == "" && m.mgr.Source == "" {
		m.initWizard()
		m.wizardSelected = map[string]bool{m.mgr.Target: true}
		m.wizardConfigs[m.mgr.Target] = wizardClientConfig{Preset: m.mgr.Preset, Roles: m.mgr.RoleModels}
		m.wizardProviderIdx = 0
		m.wizardActiveID = m.mgr.Target
		m.wizardConfirmCursor = 0
		m.screen = installConfirmScreen
		return m, nil
	}
	applyMgr := m.mgr
	for _, st := range m.statuses {
		if st.Spec.ID == applyMgr.Target && len(st.DisabledSkills) > 0 {
			applyMgr.DisabledSkills = st.DisabledSkills
		}
	}
	m.busyCancelable = false
	m.busyReturnScreen = ""
	m.screen = busy
	m.title = "Aplicando perfil de AI"
	m.progressStep, m.progressTotal = 1, 3
	m.progressLabel = "Guardando perfil de modelos en " + applyMgr.Target + "..."
	m.fault = ""
	m.log = ""
	tr := &progressTracker{ch: make(chan progressEvent, 16)}
	m.activeTracker = tr
	action := "apply-local"
	if m.installed == "" && m.mgr.Source != "" {
		action = "local"
	}
	return m, tea.Batch(
		runSingleActionCmd(applyMgr, action, manager.PrepareOptions{Action: action, ConfigureModels: true}, "Perfil de AI aplicado", tr),
		waitProgressCmd(tr),
		animTickCmd(),
	)
}

func (m model) updateModels(message tea.Msg) (model, tea.Cmd, bool) {
	if m.screen != modelsScreen && m.screen != modelsSaveScreen && m.screen != rolePickerScreen {
		return m, nil, false
	}
	msg, ok := message.(tea.KeyPressMsg)
	if !ok {
		return m, nil, false
	}
	m.ensureModelState()
	key := strings.ToLower(msg.String())

	if m.screen == rolePickerScreen {
		spec, _ := manager.FindClient(m.mgr.Target)
		spec.Models = m.pickerChoices
		roles := manager.Roles()
		roleID := roles[min(m.roleCursor, len(roles)-1)].ID
		applyModelIdx := func(idx int) {
			if idx < 0 || idx >= len(spec.Models) {
				return
			}
			cur := m.mgr.RoleModels[roleID]
			if cur.Model != spec.Models[idx] || cur.Label == "" {
				cur.Label = m.modelLabels[spec.Models[idx]]
			}
			cur.Model = spec.Models[idx]
			if cur.Effort == "" && len(spec.Efforts) > 0 {
				cur.Effort = "medium"
			}
			m.mgr.RoleModels[roleID] = cur
			m.mgr.Preset = "custom"
			m.savedProfileIdx = -1
		}
		switch key {
		case "q", "ctrl+c":
			return m, tea.Quit, true
		case "esc":
			m.screen = modelsScreen
			return m, nil, true
		case "up", "k":
			m.rolePickerModelIdx = max(0, m.rolePickerModelIdx-1)
			applyModelIdx(m.rolePickerModelIdx)
			return m, nil, true
		case "down", "j":
			if len(spec.Models) > 0 {
				m.rolePickerModelIdx = min(len(spec.Models)-1, m.rolePickerModelIdx+1)
				applyModelIdx(m.rolePickerModelIdx)
			}
			return m, nil, true
		case "left", "right", "h", "l", "e":
			m.cycleRoleEffort()
			return m, nil, true
		case "enter":
			applyModelIdx(m.rolePickerModelIdx)
			m.screen = modelsScreen
			return m, nil, true
		}
		return m, nil, true
	}

	if m.screen == modelsSaveScreen {
		switch key {
		case "ctrl+c":
			return m, tea.Quit, true
		case "esc":
			m.screen = modelsScreen
			m.profileNameInput = ""
			m.fault = ""
			return m, nil, true
		case "backspace":
			if len(m.profileNameInput) > 0 {
				runes := []rune(m.profileNameInput)
				m.profileNameInput = string(runes[:len(runes)-1])
			}
			return m, nil, true
		case "enter":
			name := strings.TrimSpace(m.profileNameInput)
			if err := m.mgr.SaveProfile(m.mgr.Target, name, m.mgr.RoleModels); err != nil {
				m.fault = err.Error()
				return m, nil, true
			}
			m.profileNameInput = ""
			m.fault = ""
			m.reloadSavedProfiles()
			for i, sp := range m.savedProfiles {
				if strings.EqualFold(sp.Name, name) {
					m.savedProfileIdx = i
					break
				}
			}
			m.screen = modelsScreen
			return m, nil, true
		default:
			if len(msg.Text) > 0 {
				m.profileNameInput += msg.Text
			}
			return m, nil, true
		}
	}

	roles := manager.Roles()
	maxRoleCursor := len(roles) + 1 // roles (0..6), Accept/Apply (7), Save custom (8)
	switch key {
	case "q", "ctrl+c":
		return m, tea.Quit, true
	case "esc":
		m.fault = ""
		if m.wizardReturn {
			m.wizardReturn = false
			m.screen = installProfileScreen
			return m, nil, true
		}
		m.screen = home
		return m, nil, true
	case "tab":
		if !m.wizardReturn {
			m.nextClient()
		}
		return m, nil, true
	case "1":
		if !m.wizardReturn {
			m.switchClient("codex")
		}
		return m, nil, true
	case "2":
		if !m.wizardReturn {
			m.switchClient("claude")
		}
		return m, nil, true
	case "3":
		if !m.wizardReturn {
			m.switchClient("agy")
		}
		return m, nil, true
	case "p":
		m.fault = ""
		m.cyclePreset()
		return m, nil, true
	case "n":
		m.cycleSavedProfile()
		return m, nil, true
	case "g":
		m.fault = ""
		m.profileNameInput = ""
		m.screen = modelsSaveScreen
		return m, nil, true
	case "x":
		if m.savedProfileIdx >= 0 && m.savedProfileIdx < len(m.savedProfiles) {
			name := m.savedProfiles[m.savedProfileIdx].Name
			if err := m.mgr.DeleteSavedProfile(m.mgr.Target, name); err != nil {
				m.fault = err.Error()
			} else {
				m.reloadSavedProfiles()
				m.fault = ""
			}
		}
		return m, nil, true
	case "up", "k":
		m.roleCursor = max(0, m.roleCursor-1)
		return m, nil, true
	case "down", "j":
		m.roleCursor = min(maxRoleCursor, m.roleCursor+1)
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
		if m.roleCursor < len(roles) {
			m.openRolePicker()
			return m, nil, true
		}
		if m.roleCursor == len(roles)+1 {
			m.fault = ""
			m.profileNameInput = ""
			m.screen = modelsSaveScreen
			return m, nil, true
		}
		next, cmd := m.confirmModelsAction()
		return next, cmd, true
	}
	return m, nil, true
}

func (m model) crtModels(width int) (string, string) {
	m.ensureModelState()
	spec, _ := manager.FindClient(m.mgr.Target)
	roles := manager.Roles()

	if m.screen == rolePickerScreen {
		roleSpec := roles[min(m.roleCursor, len(roles)-1)]
		cur := m.mgr.RoleModels[roleSpec.ID]
		var b strings.Builder
		if m.wizardReturn {
			ids := m.selectedClientIDs()
			b.WriteString(crtGreen.Render(fmt.Sprintf("Paso 2 de 4 · %s (%d de %d)", spec.Short, min(max(1, len(ids)), m.wizardProviderIdx+1), max(1, len(ids)))) + "\n")
		} else {
			b.WriteString(crtGreen.Render("Perfiles de AI · "+spec.Short) + "\n")
		}
		b.WriteString(crtAmber.Render("Elegir modelo para: "+roleSpec.Name) + "\n")
		if spec.ID == "agy" && roleSpec.ID != "orchestrator" {
			b.WriteString("Niveles de agente de AGY\n")
		} else if m.modelSource != "" {
			b.WriteString("Fuente: " + m.modelSource + "\n")
		}
		if spec.ID == "agy" && roleSpec.ID == "orchestrator" && cur.Label != "" {
			b.WriteString("Variante: " + cur.Label + " · esfuerzo definido por AGY\n\n")
		} else {
			b.WriteString("Esfuerzo actual: " + effortTag(cur.Effort) + " (←→ o E para cambiar)\n\n")
		}
		rows := max(3, m.height-16)
		start := max(0, min(m.rolePickerModelIdx-rows/2, len(m.pickerChoices)-rows))
		for i := start; i < min(start+rows, len(m.pickerChoices)); i++ {
			mod := m.pickerChoices[i]
			mark := "( )"
			if mod == cur.Model {
				mark = "(●)"
			}
			line := mark + " " + mod
			if i == m.rolePickerModelIdx {
				b.WriteString(crtNeon.Render("› "+line) + "\n")
			} else {
				b.WriteString("  " + line + "\n")
			}
		}
		help := "↑↓ elegir modelo · ←→ esfuerzo · Enter aceptar · Esc volver"
		if spec.ID == "agy" && roleSpec.ID == "orchestrator" && cur.Label != "" {
			help = "↑↓ elegir variante · Enter aceptar · Esc volver"
		}
		return strings.TrimRight(b.String(), "\n"), help
	}

	if m.screen == modelsSaveScreen {
		var b strings.Builder
		b.WriteString(crtGreen.Render("Guardar perfil propio") + " · " + spec.Short + "\n\n")
		b.WriteString("Escribí un nombre para esta configuración (ej. Trabajo diario):\n\n")
		b.WriteString(crtNeon.Render("> "+m.profileNameInput+"_") + "\n")
		if m.fault != "" {
			b.WriteString("\n" + crtAmber.Render(m.fault))
		}
		return b.String(), "Enter guardar · Esc cancelar"
	}

	var b strings.Builder
	savedNote := ""
	if m.savedProfileIdx >= 0 && m.savedProfileIdx < len(m.savedProfiles) {
		savedNote = " · Perfil: " + m.savedProfiles[m.savedProfileIdx].Name
	} else if len(m.savedProfiles) > 0 {
		savedNote = fmt.Sprintf(" · %d guardados (N)", len(m.savedProfiles))
	}

	if m.wizardReturn {
		ids := m.selectedClientIDs()
		b.WriteString(crtGreen.Render(fmt.Sprintf("Paso 2 de 4 · Personalizar %s (%d de %d)", spec.Name, min(max(1, len(ids)), m.wizardProviderIdx+1), max(1, len(ids)))) + "\n")
		b.WriteString(crtAmber.Render("Preset base: "+m.mgr.Preset) + " · Enter en un rol para elegir modelo\n\n")
	} else if width < 60 {
		b.WriteString(crtGreen.Render("Modelos · "+spec.Short) + " (" + m.mgr.Preset + ")\n")
	} else {
		b.WriteString(crtGreen.Render("Modelos y Presets") + "  " + renderClientTabs(spec.ID) + "\n")
		b.WriteString(crtAmber.Render("Preset: "+m.mgr.Preset+savedNote) + " · P preset · N cargar\n\n")
	}

	if width >= 60 {
		b.WriteString(crtBeige.Render("   ROL               │ MODELO                  │ ESFUERZO") + "\n")
		b.WriteString(crtCaseSide.Render("   ──────────────────┼─────────────────────────┼─────────") + "\n")
	}

	for i, r := range roles {
		rm := m.mgr.RoleModels[r.ID]
		effort := rm.Effort
		if spec.ID == "agy" && r.ID == "orchestrator" && rm.Label != "" {
			effort = "variante AGY"
		}
		if width < 60 {
			roleLabel := strings.TrimPrefix(r.ID, "chuli_")
			if len(roleLabel) > 9 {
				roleLabel = roleLabel[:9]
			}
			line := fmt.Sprintf("%-9s %s (%s)", roleLabel, rm.Model, effort)
			if i == m.roleCursor {
				b.WriteString(crtNeon.Render("› "+line) + "\n")
			} else {
				b.WriteString("  " + line + "\n")
			}
			continue
		}
		sep := crtCaseSide.Render(" │ ")
		tag := effortTag(effort)
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

	acceptLabel := "▶ Aplicar este perfil ahora"
	if m.wizardReturn {
		acceptLabel = "▶ Aceptar equipo de " + spec.Short + " y continuar →"
	}
	saveLabel := "💾 Guardar como perfil propio"
	b.WriteString("\n")
	if m.roleCursor == len(roles) {
		b.WriteString(crtNeon.Render("› "+acceptLabel) + "\n")
	} else {
		b.WriteString("  " + crtAmber.Render(acceptLabel) + "\n")
	}
	if m.roleCursor == len(roles)+1 {
		b.WriteString(crtNeon.Render("› " + saveLabel))
	} else {
		b.WriteString("  " + saveLabel)
	}

	if m.fault != "" {
		b.WriteString("\n\n" + crtAmber.Render(m.fault))
	}
	keys := "↑↓ mover · Enter elegir modelo / aceptar · ←→ modelo · E esfuerzo · P preset · Esc volver"
	if m.wizardReturn {
		keys = "↑↓ mover · Enter editar rol o continuar · ←→ modelo · E esfuerzo · Esc atrás"
	} else if width < 70 {
		keys = "↑↓ mover · Enter editar / aplicar\n←→ modelo · E esfuerzo · Esc volver"
	}
	return strings.TrimRight(b.String(), "\n"), keys
}
