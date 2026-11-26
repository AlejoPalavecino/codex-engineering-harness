package tui

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	tea "charm.land/bubbletea/v2"
	"github.com/AlejoPalavecino/codex-engineering-harness/v2/internal/manager"
)

const (
	installPlatformsScreen = "install-platforms"
	installProfileScreen   = "install-profile"
	installConfirmScreen   = "install-confirm"
	installGentleScreen    = "install-gentle"
)

type progressTracker struct {
	ch    chan progressEvent
	retry func(*progressTracker) tea.Cmd
}

type progressEvent struct {
	step  int
	total int
	label string
}

type wizardStageMsg struct {
	tracker *progressTracker
	event   progressEvent
	ok      bool
}

type wizardPreflightMsg struct {
	conflicts map[string]bool
	runtime   manager.RuntimePreflight
}

type wizardBatchResultMsg struct {
	title     string
	summary   string
	techLog   string
	nextSteps string
	err       error
	statuses  []manager.ClientStatus
	retry     func(*progressTracker) tea.Cmd
}

type animTickMsg time.Time

var spinnerFrames = []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"}

func animTickCmd() tea.Cmd {
	return tea.Tick(120*time.Millisecond, func(t time.Time) tea.Msg {
		return animTickMsg(t)
	})
}

func waitProgressCmd(tr *progressTracker) tea.Cmd {
	if tr == nil {
		return nil
	}
	return func() tea.Msg {
		ev, ok := <-tr.ch
		return wizardStageMsg{tracker: tr, event: ev, ok: ok}
	}
}

type wizardClientConfig struct {
	Preset string
	Roles  map[string]manager.RoleModel
}

type wizardProfileOption struct {
	kind  string // "keep-custom" | "preset" | "saved" | "custom-edit"
	id    string
	label string
	desc  string
	roles map[string]manager.RoleModel
}

func cloneRoleMap(src map[string]manager.RoleModel) map[string]manager.RoleModel {
	out := make(map[string]manager.RoleModel, len(src))
	for k, v := range src {
		out[k] = v
	}
	return out
}

func (m *model) initWizard() {
	m.statuses = m.mgr.DetectClients()
	m.wizardSelected = map[string]bool{}
	m.wizardConfigs = map[string]wizardClientConfig{}
	m.wizardSavedName = map[string]string{}
	m.wizardAdoptGentle = map[string]bool{}
	m.wizardPreflight = map[string]bool{}
	m.wizardRuntime = manager.RuntimePreflight{}
	m.wizardPreflightDone = false
	m.wizardCameFromFastPath = false
	anyDetected := false
	for _, st := range m.statuses {
		spec := st.Spec
		preset := spec.DefaultPreset
		_, roles, _ := manager.DefaultRolesForPreset(spec.ID, preset)
		if st.Preset != "" {
			if _, pr, err := manager.DefaultRolesForPreset(spec.ID, st.Preset); err == nil {
				preset = st.Preset
				roles = pr
			} else if st.Preset == "custom" {
				preset = "custom"
			}
		} else if st.InstalledVersion != "" && len(st.RoleModels) > 0 {
			preset = "custom"
		}
		if len(st.RoleModels) > 0 {
			for k, v := range st.RoleModels {
				roles[k] = v
			}
		}
		m.wizardConfigs[spec.ID] = wizardClientConfig{Preset: preset, Roles: roles}
		if st.Detected {
			m.wizardSelected[spec.ID] = true
			anyDetected = true
		}
	}
	if !anyDetected {
		target := m.mgr.Target
		if target == "" {
			target = "codex"
		}
		m.wizardSelected[target] = true
	}
	m.wizardUseLocal = m.mgr.Source != ""
	m.wizardCursor = 0
	m.wizardProviderIdx = 0
	m.wizardConfirmCursor = 0
	m.wizardActiveID = m.firstSelectedClient()
	m.syncWizardProfileCursor()
	m.adoptGentle = false
}

func (m model) selectedClientIDs() []string {
	var ids []string
	for _, c := range manager.Clients() {
		if m.wizardSelected[c.ID] {
			ids = append(ids, c.ID)
		}
	}
	return ids
}

func (m model) firstSelectedClient() string {
	ids := m.selectedClientIDs()
	if len(ids) > 0 {
		return ids[0]
	}
	return "codex"
}

func (m model) wizardProfileOptions(spec manager.ClientSpec) []wizardProfileOption {
	cfg := m.wizardConfigs[spec.ID]
	var saved []manager.SavedProfile
	if m.mgr.Home != "" {
		if list, err := m.mgr.ListSavedProfiles(spec.ID); err == nil {
			saved = list
		}
	}
	var opts []wizardProfileOption
	savedMatched := false
	if cfg.Preset == "custom" {
		if savedName := m.wizardSavedName[spec.ID]; savedName != "" {
			for _, sp := range saved {
				if sp.Name == savedName {
					savedMatched = true
					break
				}
			}
		}
		if !savedMatched {
			opts = append(opts, wizardProfileOption{
				kind:  "keep-custom",
				id:    "custom",
				label: "Equipo personalizado actual",
				desc:  "Conserva tus asignaciones actuales rol por rol.",
				roles: cfg.Roles,
			})
		}
	}
	for _, p := range spec.Presets {
		opts = append(opts, wizardProfileOption{
			kind:  "preset",
			id:    p.ID,
			label: presetFriendlyName(p),
			desc:  p.Description,
			roles: p.Roles,
		})
	}
	for _, sp := range saved {
		opts = append(opts, wizardProfileOption{
			kind:  "saved",
			id:    sp.Name,
			label: "★ " + sp.Name,
			desc:  "Perfil guardado para " + spec.Short,
			roles: sp.RoleModels,
		})
	}
	opts = append(opts, wizardProfileOption{
		kind:  "custom-edit",
		id:    "custom-edit",
		label: "⚙ Personalizar cada rol a gusto →",
		desc:  "Abre la tabla de los 7 roles para elegir modelo y esfuerzo.",
	})
	return opts
}

func (m *model) syncWizardProfileCursor() {
	spec, ok := manager.FindClient(m.wizardActiveID)
	if !ok {
		m.wizardProfileCursor = 0
		return
	}
	cfg := m.wizardConfigs[spec.ID]
	opts := m.wizardProfileOptions(spec)
	if cfg.Preset == "custom" {
		if savedName := m.wizardSavedName[spec.ID]; savedName != "" {
			for i, opt := range opts {
				if opt.kind == "saved" && opt.id == savedName {
					m.wizardProfileCursor = i
					return
				}
			}
		}
		for i, opt := range opts {
			if opt.kind == "keep-custom" {
				m.wizardProfileCursor = i
				return
			}
		}
	}
	for i, opt := range opts {
		if opt.kind == "preset" && opt.id == cfg.Preset {
			m.wizardProfileCursor = i
			return
		}
	}
	m.wizardProfileCursor = 0
}

func (m *model) applyWizardProfileOption(clientID string, opt wizardProfileOption) {
	if m.wizardConfigs == nil {
		m.wizardConfigs = map[string]wizardClientConfig{}
	}
	if m.wizardSavedName == nil {
		m.wizardSavedName = map[string]string{}
	}
	switch opt.kind {
	case "preset":
		_, roles, err := manager.DefaultRolesForPreset(clientID, opt.id)
		if err != nil {
			roles = cloneRoleMap(opt.roles)
		}
		m.wizardConfigs[clientID] = wizardClientConfig{Preset: opt.id, Roles: roles}
		delete(m.wizardSavedName, clientID)
	case "saved":
		m.wizardConfigs[clientID] = wizardClientConfig{Preset: "custom", Roles: cloneRoleMap(opt.roles)}
		m.wizardSavedName[clientID] = opt.id
	case "keep-custom":
		// Keep current custom RoleModels untouched.
	}
}

func runWizardPreflightCmd(mgr manager.Manager, ids []string) tea.Cmd {
	selected := append([]string(nil), ids...)
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()
		conflicts := make(map[string]bool, len(selected))
		statuses := mgr.DetectClients()
		homes := make(map[string]string, len(statuses))
		for _, st := range statuses {
			if st.Home != "" {
				homes[st.Spec.ID] = st.Home
			}
		}
		for _, id := range selected {
			home := homes[id]
			if home == "" {
				if h, err := manager.ResolveClientHome(id); err == nil {
					home = h
				}
			}
			if home != "" && manager.HasGentleRoutingConflict(id, home) {
				conflicts[id] = true
			}
		}
		runtime := manager.CheckRuntimePreflight(ctx)
		return wizardPreflightMsg{conflicts: conflicts, runtime: runtime}
	}
}

func (m model) enterInstallProfile() (model, tea.Cmd) {
	ids := m.selectedClientIDs()
	m.fault = ""
	m.wizardCameFromFastPath = false
	m.wizardProviderIdx = 0
	if len(ids) > 0 {
		m.wizardActiveID = ids[0]
	} else {
		m.wizardActiveID = "codex"
	}
	m.syncWizardProfileCursor()
	m.screen = installProfileScreen
	return m, nil
}

func (m model) enterInstallConfirm(fastPath bool) (model, tea.Cmd) {
	m.fault = ""
	m.wizardCameFromFastPath = fastPath
	m.wizardConfirmCursor = 0
	m.wizardPreflightDone = false
	m.wizardPreflight = map[string]bool{}
	m.wizardActiveID = m.firstSelectedClient()
	m.screen = installConfirmScreen
	m.offset = 0
	return m, runWizardPreflightCmd(m.mgr, m.selectedClientIDs())
}

func (m model) advanceWizardProvider() (model, tea.Cmd) {
	ids := m.selectedClientIDs()
	if m.wizardProviderIdx+1 < len(ids) {
		m.wizardProviderIdx++
		m.wizardActiveID = ids[m.wizardProviderIdx]
		m.syncWizardProfileCursor()
		m.screen = installProfileScreen
		return m, nil
	}
	return m.enterInstallConfirm(false)
}

func (m *model) stepBackWizardProvider() {
	ids := m.selectedClientIDs()
	if m.wizardProviderIdx > 0 && len(ids) > 0 {
		m.wizardProviderIdx--
		m.wizardActiveID = ids[m.wizardProviderIdx]
		m.syncWizardProfileCursor()
		m.screen = installProfileScreen
		return
	}
	m.screen = installPlatformsScreen
}

func (m model) hasAnyWizardGentleConflict() bool {
	for _, id := range m.selectedClientIDs() {
		if m.wizardPreflight[id] {
			return true
		}
	}
	return false
}

func (m model) wizardGentleConflictNames() []string {
	var names []string
	for _, id := range m.selectedClientIDs() {
		if m.wizardPreflight[id] {
			if spec, ok := manager.FindClient(id); ok {
				names = append(names, spec.Short)
			} else {
				names = append(names, id)
			}
		}
	}
	return names
}

func (m model) allWizardGentleAdopted() bool {
	found := false
	for _, id := range m.selectedClientIDs() {
		if m.wizardPreflight[id] {
			found = true
			if !m.wizardAdoptGentle[id] {
				return false
			}
		}
	}
	return found
}

func presetFriendlyName(p manager.PresetSpec) string {
	id := strings.ToLower(p.ID)
	switch {
	case strings.Contains(id, "balanced"):
		return "Equilibrado (recomendado)"
	case strings.Contains(id, "max"):
		return "Máxima capacidad"
	case strings.Contains(id, "fast"):
		return "Rápido y económico"
	default:
		return p.Name
	}
}

func runInstallBatchCmd(mgr manager.Manager, ids []string, configs map[string]wizardClientConfig, action string, useLocal, adoptGentle, allowDowngrade bool, tr *progressTracker) tea.Cmd {
	return runInstallBatchWithAdoptionCmd(mgr, ids, configs, action, useLocal, adoptGentle, nil, allowDowngrade, tr)
}

func runInstallBatchWithAdoptionCmd(mgr manager.Manager, ids []string, configs map[string]wizardClientConfig, action string, useLocal, adoptGentle bool, adoptByClient map[string]bool, allowDowngrade bool, tr *progressTracker) tea.Cmd {
	return runInstallBatchRemainingCmd(mgr, ids, configs, action, useLocal, adoptGentle, adoptByClient, allowDowngrade, tr, nil, "")
}

func runInstallBatchRemainingCmd(mgr manager.Manager, ids []string, configs map[string]wizardClientConfig, action string, useLocal, adoptGentle bool, adoptByClient map[string]bool, allowDowngrade bool, tr *progressTracker, completed []string, previousLog string) tea.Cmd {
	return func() tea.Msg {
		defer close(tr.ch)
		ctx, cancel := context.WithTimeout(context.Background(), 6*time.Minute)
		defer cancel()

		var logs strings.Builder
		logs.WriteString(previousLog)
		installedNames := append([]string(nil), completed...)
		var installedVersion string
		totalSteps := len(ids) * 4
		send := func(step int, label string) {
			select {
			case tr.ch <- progressEvent{step: step, total: totalSteps, label: label}:
			default:
			}
		}
		failed := func(i int, err error) wizardBatchResultMsg {
			var summary strings.Builder
			for _, name := range installedNames {
				fmt.Fprintf(&summary, "%s: completado\n", name)
			}
			for j, id := range ids[i:] {
				spec, ok := manager.FindClient(id)
				name := spec.Short
				if !ok {
					name = id
				}
				state := "pendiente (sin ejecutar)"
				if j == 0 {
					state = "fallido (revisá el motivo y el detalle técnico)"
					if errors.Is(err, manager.ErrGentleConflict) {
						state = "pendiente (conflicto Gentle; sin aplicar)"
					}
				}
				fmt.Fprintf(&summary, "%s: %s\n", name, state)
			}
			msg := wizardBatchResultMsg{err: err, summary: strings.TrimSpace(summary.String()), techLog: logs.String(), statuses: mgr.DetectClients(), nextSteps: "Revisá el motivo antes de volver a instalar las plataformas pendientes."}
			if errors.Is(err, manager.ErrGentleConflict) {
				msg.retry = func(next *progressTracker) tea.Cmd {
					return runInstallBatchRemainingCmd(mgr, ids[i:], configs, action, useLocal, true, adoptByClient, allowDowngrade, next, installedNames, logs.String())
				}
			}
			return msg
		}

		batchAction := action
		if ((batchAction == "" && !useLocal) || batchAction == "latest") && len(ids) > 1 {
			if rel, err := mgr.Latest(ctx); err == nil && rel.Tag != "" {
				batchAction = "tag:" + rel.Tag
			}
		}

		for i, id := range ids {
			spec, _ := manager.FindClient(id)
			base := i * 4
			send(base+1, fmt.Sprintf("[%s] Descargando y comprobando paquete oficial...", spec.Short))
			cm, err := mgr.ForTarget(id)
			if err != nil {
				return failed(i, err)
			}
			if !useLocal {
				cm.Source = ""
			}
			if cfg, ok := configs[id]; ok {
				cm.Preset = cfg.Preset
				if len(cfg.Roles) > 0 {
					cm.RoleModels = cfg.Roles
				}
			}
			for _, st := range mgr.DetectClients() {
				if st.Spec.ID == id && len(st.DisabledSkills) > 0 {
					cm.DisabledSkills = st.DisabledSkills
				}
			}
			act := batchAction
			if act == "" {
				act = "latest"
				if useLocal && cm.Source != "" {
					act = "local"
				}
			}
			adoptClient := (adoptGentle && i == 0) || adoptByClient[id]
			p, err := cm.PrepareWithOptions(ctx, manager.PrepareOptions{
				Action:             act,
				AdoptGentleRouting: adoptClient,
				AllowDowngrade:     allowDowngrade,
				ConfigureModels:    true,
			})
			if err != nil {
				return failed(i, fmt.Errorf("%s: %w", spec.Short, err))
			}

			send(base+2, fmt.Sprintf("[%s] Creando respaldo de seguridad...", spec.Short))
			out, err := cm.ApplyWithProgress(ctx, p, func(stage, _ string) {
				switch stage {
				case "apply":
					send(base+3, fmt.Sprintf("[%s] Instalando especialistas, reglas y skills...", spec.Short))
				case "verify":
					send(base+4, fmt.Sprintf("[%s] Comprobando que todo quedó funcionando...", spec.Short))
				}
			})
			installedVersion = p.Version
			p.Cleanup()
			logs.WriteString("=== " + spec.Name + " ===\n" + out + "\n")
			if err != nil {
				return failed(i, fmt.Errorf("%s: %w", spec.Short, err))
			}
			installedNames = append(installedNames, spec.Short)
		}

		nextHint := "Abrí tu herramienta (" + strings.Join(installedNames, " / ") + ") en la carpeta de tu proyecto\ny probá decirle: «Explicame cómo funciona este módulo»."
		return wizardBatchResultMsg{
			title:     "¡Listo! Chuli quedó instalado",
			summary:   fmt.Sprintf("Plataformas listas: %s\nVersión aplicada: v%s\nRespaldo automático guardado sin pisar tu configuración personal.", strings.Join(installedNames, ", "), installedVersion),
			techLog:   logs.String(),
			nextSteps: nextHint,
			statuses:  mgr.DetectClients(),
		}
	}
}

func (m model) updateWizard(message tea.Msg) (model, tea.Cmd, bool) {
	switch msg := message.(type) {
	case animTickMsg:
		if m.screen == busy {
			m.spinnerIdx = (m.spinnerIdx + 1) % len(spinnerFrames)
			return m, animTickCmd(), true
		}
		return m, nil, false
	case wizardStageMsg:
		if msg.ok && m.screen == busy {
			m.progressStep = msg.event.step
			m.progressTotal = msg.event.total
			m.progressLabel = msg.event.label
			return m, waitProgressCmd(msg.tracker), true
		}
		return m, nil, true
	case wizardPreflightMsg:
		m.wizardPreflightDone = true
		m.wizardRuntime = msg.runtime
		if msg.conflicts != nil {
			m.wizardPreflight = msg.conflicts
		} else {
			m.wizardPreflight = map[string]bool{}
		}
		return m, nil, true
	case wizardBatchResultMsg:
		m.prepared.Cleanup()
		m.prepared = manager.Prepared{}
		m.busyCancelable = false
		m.busyReturnScreen = ""
		if len(msg.statuses) > 0 {
			m.statuses = msg.statuses
		}
		if installed, err := m.mgr.Installed(); err == nil {
			m.installed = installed
		}
		m.resultSummary = msg.summary
		m.resultNext = msg.nextSteps
		m.resultCursor = 0
		m.log = msg.techLog
		if msg.err != nil {
			if errors.Is(msg.err, manager.ErrGentleConflict) {
				m.screen = installGentleScreen
				m.fault = msg.err.Error()
				m.activeTracker = &progressTracker{retry: msg.retry}
				m.offset = 0
				return m, nil, true
			}
			m.screen = result
			m.title = "No se pudo completar"
			m.fault = msg.err.Error()
			m.showTechLog = false
			m.offset = 0
			return m, nil, true
		}
		m.screen = result
		m.title = msg.title
		m.fault = ""
		m.showTechLog = false
		m.offset = 0
		return m, checkUpdatesCmd(m.mgr, m.managerVersion), true
	}

	msg, ok := message.(tea.KeyPressMsg)
	if !ok {
		return m, nil, false
	}
	key := strings.ToLower(msg.String())

	switch m.screen {
	case installPlatformsScreen:
		clients := manager.Clients()
		maxCursor := len(clients) + 1 // platforms (0..2), fast-path (3), customize (4)
		switch key {
		case "ctrl+c":
			return m, tea.Quit, true
		case "esc":
			m.fault = ""
			m.screen = home
			return m, nil, true
		case "up", "k":
			m.wizardCursor = max(0, m.wizardCursor-1)
		case "down", "j":
			m.wizardCursor = min(maxCursor, m.wizardCursor+1)
		case " ", "space":
			if m.wizardCursor < len(clients) {
				id := clients[m.wizardCursor].ID
				m.wizardSelected[id] = !m.wizardSelected[id]
				m.fault = ""
				return m, nil, true
			}
			ids := m.selectedClientIDs()
			if len(ids) == 0 {
				m.fault = "Marcá al menos una plataforma con Espacio antes de continuar."
				return m, nil, true
			}
			if m.wizardCursor == len(clients)+1 {
				next, cmd := m.enterInstallProfile()
				return next, cmd, true
			}
			next, cmd := m.enterInstallConfirm(true)
			return next, cmd, true
		case "1", "2", "3":
			idx := int(key[0] - '1')
			if idx >= 0 && idx < len(clients) {
				m.wizardCursor = idx
				id := clients[idx].ID
				m.wizardSelected[id] = !m.wizardSelected[id]
				m.fault = ""
			}
		case "s":
			if m.mgr.Source != "" {
				m.wizardUseLocal = !m.wizardUseLocal
			}
		case "m", "c":
			ids := m.selectedClientIDs()
			if len(ids) == 0 {
				m.fault = "Marcá al menos una plataforma con Espacio antes de continuar."
				return m, nil, true
			}
			next, cmd := m.enterInstallProfile()
			return next, cmd, true
		case "enter":
			ids := m.selectedClientIDs()
			if len(ids) == 0 {
				m.fault = "Marcá al menos una plataforma con Espacio antes de continuar."
				return m, nil, true
			}
			if m.wizardCursor == len(clients)+1 {
				next, cmd := m.enterInstallProfile()
				return next, cmd, true
			}
			next, cmd := m.enterInstallConfirm(true)
			return next, cmd, true
		}
		return m, nil, true

	case installProfileScreen:
		spec, _ := manager.FindClient(m.wizardActiveID)
		opts := m.wizardProfileOptions(spec)
		maxIdx := max(0, len(opts)-1)
		switch key {
		case "ctrl+c":
			return m, tea.Quit, true
		case "esc":
			m.stepBackWizardProvider()
			return m, nil, true
		case "tab":
			ids := m.selectedClientIDs()
			if len(ids) > 1 {
				m.wizardProviderIdx = (m.wizardProviderIdx + 1) % len(ids)
				m.wizardActiveID = ids[m.wizardProviderIdx]
				m.syncWizardProfileCursor()
			}
			return m, nil, true
		case "shift+tab":
			ids := m.selectedClientIDs()
			if len(ids) > 1 {
				m.wizardProviderIdx = (m.wizardProviderIdx - 1 + len(ids)) % len(ids)
				m.wizardActiveID = ids[m.wizardProviderIdx]
				m.syncWizardProfileCursor()
			}
			return m, nil, true
		case "up", "k":
			m.wizardProfileCursor = max(0, m.wizardProfileCursor-1)
		case "down", "j":
			m.wizardProfileCursor = min(maxIdx, m.wizardProfileCursor+1)
		case " ", "space":
			if len(opts) > 0 {
				chosen := opts[min(m.wizardProfileCursor, len(opts)-1)]
				if chosen.kind == "custom-edit" {
					m.openWizardCustomModels()
					return m, nil, true
				}
				m.applyWizardProfileOption(spec.ID, chosen)
			}
		case "c":
			m.openWizardCustomModels()
			return m, nil, true
		case "enter":
			if len(opts) == 0 {
				next, cmd := m.advanceWizardProvider()
				return next, cmd, true
			}
			chosen := opts[min(m.wizardProfileCursor, len(opts)-1)]
			if chosen.kind == "custom-edit" {
				m.openWizardCustomModels()
				return m, nil, true
			}
			m.applyWizardProfileOption(spec.ID, chosen)
			next, cmd := m.advanceWizardProvider()
			return next, cmd, true
		}
		return m, nil, true

	case installConfirmScreen:
		hasGentle := m.hasAnyWizardGentleConflict()
		maxCursor := 1
		if hasGentle {
			maxCursor = 2
		}
		goBack := func() (model, tea.Cmd, bool) {
			if m.wizardCameFromFastPath {
				m.screen = installPlatformsScreen
				return m, nil, true
			}
			ids := m.selectedClientIDs()
			if len(ids) > 0 {
				m.wizardProviderIdx = len(ids) - 1
				m.wizardActiveID = ids[m.wizardProviderIdx]
				m.syncWizardProfileCursor()
			}
			m.screen = installProfileScreen
			return m, nil, true
		}
		toggleGentle := func() {
			if !hasGentle {
				return
			}
			if m.wizardAdoptGentle == nil {
				m.wizardAdoptGentle = map[string]bool{}
			}
			targetVal := !m.allWizardGentleAdopted()
			for _, id := range m.selectedClientIDs() {
				if m.wizardPreflight[id] {
					m.wizardAdoptGentle[id] = targetVal
				}
			}
		}
		switch key {
		case "ctrl+c":
			return m, tea.Quit, true
		case "esc", "n":
			return goBack()
		case "up", "k":
			m.wizardConfirmCursor = max(0, m.wizardConfirmCursor-1)
		case "down", "j":
			m.wizardConfirmCursor = min(maxCursor, m.wizardConfirmCursor+1)
		case "a":
			toggleGentle()
		case " ", "space":
			if hasGentle && m.wizardConfirmCursor == 1 {
				toggleGentle()
			}
		case "s", "y":
			next := m.startWizardInstall(false)
			return next, tea.Batch(next.startWizardInstallCmd(false), animTickCmd()), true
		case "enter":
			if m.wizardConfirmCursor == maxCursor {
				return goBack()
			}
			if hasGentle && m.wizardConfirmCursor == 1 {
				toggleGentle()
				return m, nil, true
			}
			next := m.startWizardInstall(false)
			return next, tea.Batch(next.startWizardInstallCmd(false), animTickCmd()), true
		}
		return m, nil, true

	case installGentleScreen:
		switch key {
		case "ctrl+c":
			return m, tea.Quit, true
		case "esc", "n":
			m.screen = home
			m.fault = ""
			return m, nil, true
		case "a", "enter", "s":
			m.adoptGentle = true
			if m.activeTracker != nil && m.activeTracker.retry != nil {
				retry := m.activeTracker.retry
				next := m.startWizardInstall(true)
				tr := &progressTracker{ch: make(chan progressEvent, 16)}
				next.activeTracker = tr
				return next, tea.Batch(retry(tr), waitProgressCmd(tr), animTickCmd()), true
			}
			next := m.startWizardInstall(true)
			return next, tea.Batch(next.startWizardInstallCmd(true), animTickCmd()), true
		}
		return m, nil, true
	}
	return m, nil, false
}

func (m *model) openWizardCustomModels() {
	m.switchClient(m.wizardActiveID)
	if cfg, ok := m.wizardConfigs[m.wizardActiveID]; ok {
		m.mgr.Preset = cfg.Preset
		m.mgr.RoleModels = cfg.Roles
	}
	m.wizardReturn = true
	m.roleCursor = 0
	m.screen = modelsScreen
}

func (m model) startWizardInstall(adopt bool) model {
	ids := m.selectedClientIDs()
	m.busyCancelable = false
	m.busyReturnScreen = ""
	m.screen = busy
	m.title = "Paso 4 de 4 · Instalando Chuli"
	m.progressStep = 1
	m.progressTotal = max(1, len(ids)*4)
	m.progressLabel = "Iniciando instalación..."
	m.fault = ""
	m.resultSummary = ""
	m.resultNext = ""
	m.resultCursor = 0
	m.resultFromWizard = true
	m.showTechLog = false
	return m
}

func (m *model) startWizardInstallCmd(adopt bool) tea.Cmd {
	ids := m.selectedClientIDs()
	tr := &progressTracker{ch: make(chan progressEvent, 16)}
	m.activeTracker = tr
	adoptMap := make(map[string]bool, len(m.wizardAdoptGentle))
	for k, v := range m.wizardAdoptGentle {
		adoptMap[k] = v
	}
	return tea.Batch(
		runInstallBatchWithAdoptionCmd(m.mgr, ids, m.wizardConfigs, "", m.wizardUseLocal, adopt, adoptMap, false, tr),
		waitProgressCmd(tr),
	)
}

func (m model) crtWizard(width int, page string) (string, string) {
	compactHeight := m.height > 0 && m.height <= 24
	var b strings.Builder
	switch m.screen {
	case installPlatformsScreen:
		b.WriteString(crtGreen.Render("Paso 1 de 4 · Elegí en qué plataforma(s) instalar") + "\n")
		if !compactHeight {
			b.WriteString("\n")
		}
		if width >= 54 {
			b.WriteString("Espacio marca/desmarca · Enter continúa directo:\n\n")
		} else {
			b.WriteString("Espacio marca · Enter sigue:\n\n")
		}
		statusByID := map[string]manager.ClientStatus{}
		for _, st := range m.statuses {
			statusByID[st.Spec.ID] = st
		}
		clients := manager.Clients()
		for i, c := range clients {
			check := "[ ]"
			if m.wizardSelected[c.ID] {
				check = "[✓]"
			}
			st := statusByID[c.ID]
			badge := "No detectado"
			if st.InstalledVersion != "" {
				badge = "Instalado v" + st.InstalledVersion
			} else if st.Detected {
				badge = "Detectado en tu PC"
			}
			line := fmt.Sprintf("%s %s", check, c.Name)
			if width >= 62 {
				line += " · " + badge
			}
			if i == m.wizardCursor {
				b.WriteString(crtNeon.Render("› "+line) + "\n")
			} else {
				b.WriteString("  " + line + "\n")
			}
		}
		selectedCount := len(m.selectedClientIDs())
		fastLabel := fmt.Sprintf("▶ Continuar con recomendados (%d) →", selectedCount)
		customLabel := "⚙ Elegir presets o personalizar modelos →"
		if width < 48 {
			fastLabel = fmt.Sprintf("▶ Continuar (%d) →", selectedCount)
			customLabel = "⚙ Personalizar modelos →"
		}
		b.WriteString("\n")
		if m.wizardCursor == len(clients) {
			b.WriteString(crtNeon.Render("› "+fastLabel) + "\n")
		} else {
			b.WriteString("  " + crtAmber.Render(fastLabel) + "\n")
		}
		if m.wizardCursor == len(clients)+1 {
			b.WriteString(crtNeon.Render("› " + customLabel))
		} else {
			b.WriteString("  " + customLabel)
		}
		if m.mgr.Source != "" {
			mode := "Release oficial (GitHub)"
			if m.wizardUseLocal {
				mode = "Carpeta local (--source)"
			}
			b.WriteString("\n\n" + crtAmber.Render("Origen: "+mode) + " (S para cambiar)")
		}
		if m.fault != "" {
			b.WriteString("\n\n" + crtAmber.Render(m.fault))
		}
		keys := "↑↓ mover · Espacio marcar · Enter continuar · M personalizar · Esc volver"
		if width < 72 {
			keys = "↑↓ mover · Espacio marcar\nEnter continuar · M modelos · Esc volver"
		}
		return strings.TrimRight(b.String(), "\n"), keys

	case installProfileScreen:
		ids := m.selectedClientIDs()
		spec, _ := manager.FindClient(m.wizardActiveID)
		cfg := m.wizardConfigs[spec.ID]
		totalProviders := max(1, len(ids))
		currentNum := min(totalProviders, m.wizardProviderIdx+1)

		b.WriteString(crtGreen.Render(fmt.Sprintf("Paso 2 de 4 · Modelos para %s (%d de %d)", spec.Name, currentNum, totalProviders)) + "\n")
		if totalProviders > 1 {
			var steps []string
			for i, id := range ids {
				cs, _ := manager.FindClient(id)
				if i == m.wizardProviderIdx {
					steps = append(steps, crtNeon.Render("[● "+cs.Short+"]"))
				} else if i < m.wizardProviderIdx {
					steps = append(steps, crtGreen.Render("[✓ "+cs.Short+"]"))
				} else {
					steps = append(steps, crtBeige.Render("[○ "+cs.Short+"]"))
				}
			}
			b.WriteString("Proveedores: " + strings.Join(steps, " ") + "\n")
		}
		b.WriteString("\nElegí con ↑↓ y Enter para aplicar y continuar:\n\n")

		opts := m.wizardProfileOptions(spec)
		cursor := min(m.wizardProfileCursor, max(0, len(opts)-1))
		maxRows := len(opts)
		if compactHeight && maxRows > 4 {
			maxRows = 4
		} else if m.height > 0 && m.height <= 28 && maxRows > 6 {
			maxRows = 6
		}
		start := max(0, min(cursor-maxRows/2, len(opts)-maxRows))
		end := min(len(opts), start+maxRows)

		for i := start; i < end; i++ {
			opt := opts[i]
			selected := false
			switch opt.kind {
			case "preset":
				selected = cfg.Preset == opt.id
			case "saved":
				selected = cfg.Preset == "custom" && m.wizardSavedName[spec.ID] == opt.id
			case "keep-custom":
				selected = cfg.Preset == "custom"
			}
			mark := "( )"
			if selected {
				mark = "(●)"
			}
			line := fmt.Sprintf("%s %s", mark, opt.label)
			if opt.kind == "custom-edit" {
				line = opt.label
			}
			if i == cursor {
				b.WriteString(crtNeon.Render("› "+line) + "\n")
				if width >= 56 && !compactHeight && opt.desc != "" {
					b.WriteString("    " + crtBeige.Render(opt.desc) + "\n")
				}
			} else {
				b.WriteString("  " + line + "\n")
			}
		}

		if len(opts) > 0 && width >= 58 {
			focused := opts[cursor]
			previewRoles := cfg.Roles
			prefix := "Resumen"
			if len(focused.roles) > 0 {
				previewRoles = focused.roles
				if (focused.kind == "preset" && focused.id != cfg.Preset) || (focused.kind == "saved" && m.wizardSavedName[spec.ID] != focused.id) {
					prefix = "Vista previa"
				}
			}
			if len(previewRoles) > 0 {
				b.WriteString("\n" + crtBeige.Render(fmt.Sprintf("%s: Orquestador %s · Worker %s · Arquitecto %s",
					prefix, previewRoles["orchestrator"].Model, previewRoles["chuli_worker"].Model, previewRoles["solution_architect"].Model)))
			}
		}
		keys := "↑↓ navegar · Espacio marcar · Enter continuar · Tab proveedor · Esc atrás"
		if width < 72 {
			keys = "↑↓ nav · Enter continuar\nTab proveedor · C editar · Esc atrás"
		}
		return strings.TrimRight(b.String(), "\n"), keys

	case installConfirmScreen:
		b.WriteString(crtGreen.Render("Paso 3 de 4 · Chequeo y confirmación") + "\n")
		if !compactHeight {
			b.WriteString("\n")
		}
		statusByID := map[string]manager.ClientStatus{}
		for _, st := range m.statuses {
			statusByID[st.Spec.ID] = st
		}
		selectedIDs := m.selectedClientIDs()
		for _, id := range selectedIDs {
			spec, _ := manager.FindClient(id)
			home := ""
			if st, ok := statusByID[id]; ok && st.Home != "" {
				home = st.Home
			} else {
				home, _ = manager.ResolveClientHome(id)
			}
			cfg := m.wizardConfigs[id]
			actionLabel := "Instalación nueva"
			if st, ok := statusByID[id]; ok && st.InstalledVersion != "" {
				actionLabel = "Actualizar desde v" + st.InstalledVersion
			}
			presetLabel := manager.PresetDisplayName(id, cfg.Preset)
			if name := m.wizardSavedName[id]; cfg.Preset == "custom" && name != "" {
				presetLabel = "Perfil propio: " + name
			}
			if compactHeight {
				b.WriteString(crtAmber.Render("• "+spec.Short) + " · " + actionLabel + " · " + presetLabel + "\n")
			} else {
				b.WriteString(crtAmber.Render("• "+spec.Name) + " (" + actionLabel + ")\n")
				b.WriteString("  Perfil: " + presetLabel + "\n")
				if width >= 54 {
					b.WriteString("  Carpeta: " + home + "\n")
				}
			}
		}
		if m.wizardPreflightDone {
			if m.wizardRuntime.PowerShellOK && m.wizardRuntime.PythonOK {
				b.WriteString("\n" + crtGreen.Render("✓ Requisitos listos (PowerShell y Python 3.11+)") + "\n")
			} else if m.wizardRuntime.Detail != "" {
				b.WriteString("\n" + crtAmber.Render("⚠ "+m.wizardRuntime.Detail) + "\n")
			}
		} else {
			b.WriteString("\n" + crtBeige.Render("Tus configuraciones actuales se respaldan automáticamente.") + "\n")
		}

		hasGentle := m.hasAnyWizardGentleConflict()
		opts := []string{"▶ Confirmar e instalar ahora"}
		if hasGentle {
			mark := "[ ]"
			if m.allWizardGentleAdopted() {
				mark = "[✓]"
			}
			gentleLabel := fmt.Sprintf("%s Respaldar Gentle AI (%s) y activar Chuli", mark, strings.Join(m.wizardGentleConflictNames(), ", "))
			if width < 56 {
				gentleLabel = fmt.Sprintf("%s Respaldar Gentle AI", mark)
			}
			opts = append(opts, gentleLabel)
		}
		opts = append(opts, "← Volver atrás para cambiar")

		b.WriteString("\n")
		selected := min(m.wizardConfirmCursor, len(opts)-1)
		for i, label := range opts {
			if i == selected {
				b.WriteString(crtNeon.Render("› "+label) + "\n")
			} else {
				b.WriteString("  " + label + "\n")
			}
		}
		keys := "↑↓ elegir · Enter aceptar · S confirmar · Esc atrás"
		if hasGentle {
			keys = "↑↓ elegir · Espacio marcar Gentle · Enter aceptar · S confirmar · Esc atrás"
			if width < 72 {
				keys = "↑↓ elegir · A marcar Gentle\nEnter aceptar · Esc atrás"
			}
		}
		return strings.TrimRight(b.String(), "\n"), keys

	case installGentleScreen:
		b.WriteString(crtAmber.Render("Atención · Encontramos instrucciones anteriores de Gentle AI") + "\n\n")
		b.WriteString("La plataforma en conflicto se detuvo antes de aplicar cambios.\n")
		if m.resultSummary != "" {
			b.WriteString("\n" + crtText(m.resultSummary) + "\n\n")
		}
		b.WriteString("Si aceptás, guardamos un respaldo completo de lo anterior y activamos Chuli.\n")
		return strings.TrimRight(b.String(), "\n"), "A / Enter respaldar y activar Chuli · Esc cancelar"
	}
	return "", "Esc volver"
}
