package tui

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"

	tea "charm.land/bubbletea/v2"
	"github.com/AlejoPalavecino/codex-engineering-harness/v2/internal/manager"
)

const (
	installPlatformsScreen = "install-platforms"
	installProfileScreen   = "install-profile"
	installConfirmScreen   = "install-confirm"
	installGentleScreen    = "install-gentle"
	updateScreen           = "update-view"
	versionsScreen         = "versions-view"
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

type wizardBatchResultMsg struct {
	title     string
	summary   string
	techLog   string
	nextSteps string
	err       error
	statuses  []manager.ClientStatus
	retry     func(*progressTracker) tea.Cmd
}

type updatesLoadedMsg struct {
	report   manager.UpdateReport
	releases []manager.Release
	cached   []string
	statuses []manager.ClientStatus
	err      error
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

func checkUpdatesCmd(mgr manager.Manager, managerVersion string) tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 8*time.Second)
		defer cancel()

		var (
			wg       sync.WaitGroup
			statuses []manager.ClientStatus
			cached   []string
			report   manager.UpdateReport
			releases []manager.Release
			err      error
		)
		wg.Add(3)
		go func() {
			defer wg.Done()
			statuses = mgr.DetectClients()
			cached, _ = mgr.CachedTags()
		}()
		go func() {
			defer wg.Done()
			report, err = mgr.CheckUpdates(ctx, managerVersion)
		}()
		go func() {
			defer wg.Done()
			releases, _ = mgr.ListReleases(ctx, 10)
		}()
		wg.Wait()

		return updatesLoadedMsg{
			report:   report,
			releases: releases,
			cached:   cached,
			statuses: statuses,
			err:      err,
		}
	}
}

type wizardClientConfig struct {
	Preset string
	Roles  map[string]manager.RoleModel
}

func (m *model) initWizard() {
	m.statuses = m.mgr.DetectClients()
	m.wizardSelected = map[string]bool{}
	m.wizardConfigs = map[string]wizardClientConfig{}
	anyDetected := false
	for _, st := range m.statuses {
		spec := st.Spec
		preset := spec.DefaultPreset
		_, roles, _ := manager.DefaultRolesForPreset(spec.ID, preset)
		if st.Preset != "" {
			if _, pr, err := manager.DefaultRolesForPreset(spec.ID, st.Preset); err == nil {
				preset = st.Preset
				roles = pr
			}
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

func (m *model) syncWizardProfileCursor() {
	spec, ok := manager.FindClient(m.wizardActiveID)
	if !ok {
		m.wizardProfileCursor = 0
		return
	}
	cfg := m.wizardConfigs[spec.ID]
	if cfg.Preset == "custom" {
		m.wizardProfileCursor = len(spec.Presets)
		return
	}
	for i, p := range spec.Presets {
		if p.ID == cfg.Preset {
			m.wizardProfileCursor = i
			return
		}
	}
	m.wizardProfileCursor = 0
}

func (m *model) applyProfileCursorLive() {
	spec, ok := manager.FindClient(m.wizardActiveID)
	if !ok {
		return
	}
	if m.wizardProfileCursor >= 0 && m.wizardProfileCursor < len(spec.Presets) {
		chosen := spec.Presets[m.wizardProfileCursor]
		_, roles, _ := manager.DefaultRolesForPreset(spec.ID, chosen.ID)
		m.wizardConfigs[spec.ID] = wizardClientConfig{Preset: chosen.ID, Roles: roles}
	}
}

func (m *model) advanceWizardProvider() {
	ids := m.selectedClientIDs()
	if m.wizardProviderIdx+1 < len(ids) {
		m.wizardProviderIdx++
		m.wizardActiveID = ids[m.wizardProviderIdx]
		m.syncWizardProfileCursor()
		m.screen = installProfileScreen
		return
	}
	m.wizardConfirmCursor = 0
	m.screen = installConfirmScreen
	m.offset = 0
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
	return runInstallBatchRemainingCmd(mgr, ids, configs, action, useLocal, adoptGentle, allowDowngrade, tr, nil, "")
}

func runInstallBatchRemainingCmd(mgr manager.Manager, ids []string, configs map[string]wizardClientConfig, action string, useLocal, adoptGentle, allowDowngrade bool, tr *progressTracker, completed []string, previousLog string) tea.Cmd {
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
					return runInstallBatchRemainingCmd(mgr, ids[i:], configs, action, useLocal, true, allowDowngrade, next, installedNames, logs.String())
				}
			}
			return msg
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
			act := action
			if act == "" {
				act = "latest"
				if useLocal && cm.Source != "" {
					act = "local"
				}
			}
			p, err := cm.PrepareWithOptions(ctx, manager.PrepareOptions{
				Action:             act,
				AdoptGentleRouting: adoptGentle && i == 0,
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

func runSingleActionCmd(mgr manager.Manager, action string, opts manager.PrepareOptions, successTitle string, tr *progressTracker) tea.Cmd {
	return func() tea.Msg {
		defer close(tr.ch)
		ctx, cancel := context.WithTimeout(context.Background(), 4*time.Minute)
		defer cancel()

		spec, _ := manager.FindClient(mgr.Target)
		send := func(step int, label string) {
			select {
			case tr.ch <- progressEvent{step: step, total: 3, label: label}:
			default:
			}
		}
		send(1, fmt.Sprintf("[%s] Preparando cambios y comprobando integridad...", spec.Short))
		p, err := mgr.PrepareWithOptions(ctx, opts)
		if err != nil {
			if errors.Is(err, manager.ErrGentleConflict) {
				return wizardBatchResultMsg{err: err, summary: spec.Short + ": pendiente de confirmación", retry: func(next *progressTracker) tea.Cmd {
					opts.AdoptGentleRouting = true
					return runSingleActionCmd(mgr, action, opts, successTitle, next)
				}}
			}
			return wizardBatchResultMsg{err: err}
		}
		defer p.Cleanup()
		send(2, fmt.Sprintf("[%s] Aplicando operación de forma segura...", spec.Short))
		out, err := mgr.ApplyWithProgress(ctx, p, func(stage, _ string) {
			if stage == "verify" {
				send(3, fmt.Sprintf("[%s] Verificando resultado final...", spec.Short))
			}
		})
		if err != nil {
			return wizardBatchResultMsg{err: err, techLog: out}
		}
		ver := p.Version
		summary := fmt.Sprintf("Plataforma: %s\nDestino: %s", spec.Name, mgr.Home)
		if action == "rollback" {
			summary += "\nSe restauró la instalación anterior desde el respaldo."
		} else if ver != "" {
			summary += "\nVersión activa: v" + ver
		}
		return wizardBatchResultMsg{
			title:     successTitle,
			summary:   summary,
			techLog:   out,
			nextSteps: "Reiniciá " + spec.Short + " si lo tenías abierto para tomar los cambios.",
			statuses:  mgr.DetectClients(),
		}
	}
}

func (m model) startUpdateSearch() (model, tea.Cmd) {
	m.screen = busy
	m.busyCancelable = true
	m.busyReturnScreen = updateScreen
	m.title = "Buscando actualizaciones"
	m.progressLabel = "Consultando últimas releases oficiales en GitHub..."
	m.progressStep, m.progressTotal = 1, 1
	return m, tea.Batch(checkUpdatesCmd(m.mgr, m.managerVersion), animTickCmd())
}

func (m model) startApplyUpdates() (model, tea.Cmd) {
	var outdated []string
	for _, cu := range m.updateReport.Clients {
		if cu.HasUpdate {
			outdated = append(outdated, cu.ID)
		}
	}
	if len(outdated) == 0 {
		return m, nil
	}
	m.wizardSelected = map[string]bool{}
	m.wizardConfigs = map[string]wizardClientConfig{}
	for _, st := range m.statuses {
		for _, id := range outdated {
			if st.Spec.ID == id {
				m.wizardSelected[id] = true
				m.wizardConfigs[id] = wizardClientConfig{Preset: st.Preset, Roles: st.RoleModels}
			}
		}
	}
	m.wizardUseLocal = false
	m.busyCancelable = false
	m.busyReturnScreen = ""
	m.screen = busy
	m.title = "Actualizando Chuli"
	m.progressStep, m.progressTotal = 1, len(outdated)*4
	m.progressLabel = "Iniciando actualización..."
	tr := &progressTracker{ch: make(chan progressEvent, 16)}
	m.activeTracker = tr
	return m, tea.Batch(
		runInstallBatchCmd(m.mgr, outdated, m.wizardConfigs, "latest", false, false, false, tr),
		waitProgressCmd(tr),
		animTickCmd(),
	)
}

func (m model) updateActions() []struct{ id, label string } {
	var actions []struct{ id, label string }
	if m.updateReport.AnyProfileUpdate && m.updateCheckErr == "" {
		actions = append(actions, struct{ id, label string }{"apply", "▶ Actualizar plataformas ahora"})
	}
	actions = append(actions,
		struct{ id, label string }{"refresh", "↻ Buscar actualizaciones de nuevo"},
		struct{ id, label string }{"back", "← Volver al menú principal"},
	)
	return actions
}

func (m model) updateWizard(message tea.Msg) (model, tea.Cmd, bool) {
	switch msg := message.(type) {
	case updatesLoadedMsg:
		m.updatesChecked = true
		if len(msg.statuses) > 0 {
			m.statuses = msg.statuses
		}
		if len(msg.cached) > 0 {
			m.cachedTags = msg.cached
		}
		if msg.err == nil {
			m.updateCheckErr = ""
			m.updateReport = msg.report
			m.releases = msg.releases
		} else {
			m.updateCheckErr = msg.err.Error()
		}
		if m.screen == busy && m.busyReturnScreen != "" {
			m.screen = m.busyReturnScreen
			m.busyCancelable = false
			m.busyReturnScreen = ""
			m.updateCursor = 0
			m.offset = 0
		}
		return m, nil, true
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
		maxCursor := len(clients) // last row = "Continuar"
		switch key {
		case "q", "ctrl+c":
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
			}
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
		case "enter":
			if m.wizardCursor < len(clients) {
				id := clients[m.wizardCursor].ID
				m.wizardSelected[id] = !m.wizardSelected[id]
				m.fault = ""
				return m, nil, true
			}
			ids := m.selectedClientIDs()
			if len(ids) == 0 {
				m.fault = "Marcá al menos una plataforma antes de continuar."
				return m, nil, true
			}
			m.fault = ""
			m.wizardProviderIdx = 0
			m.wizardActiveID = ids[0]
			m.syncWizardProfileCursor()
			m.screen = installProfileScreen
		}
		return m, nil, true

	case installProfileScreen:
		spec, _ := manager.FindClient(m.wizardActiveID)
		maxIdx := len(spec.Presets) // last item = Personalizar a gusto
		switch key {
		case "q", "ctrl+c":
			return m, tea.Quit, true
		case "esc":
			m.stepBackWizardProvider()
			return m, nil, true
		case "up", "k":
			m.wizardProfileCursor = max(0, m.wizardProfileCursor-1)
			m.applyProfileCursorLive()
		case "down", "j":
			m.wizardProfileCursor = min(maxIdx, m.wizardProfileCursor+1)
			m.applyProfileCursorLive()
		case " ", "space":
			m.applyProfileCursorLive()
		case "c":
			m.openWizardCustomModels()
			return m, nil, true
		case "enter":
			if m.wizardProfileCursor == len(spec.Presets) {
				m.openWizardCustomModels()
				return m, nil, true
			}
			m.applyProfileCursorLive()
			m.advanceWizardProvider()
		}
		return m, nil, true

	case installConfirmScreen:
		switch key {
		case "q", "ctrl+c":
			return m, tea.Quit, true
		case "esc", "n":
			ids := m.selectedClientIDs()
			if len(ids) > 0 {
				m.wizardProviderIdx = len(ids) - 1
				m.wizardActiveID = ids[m.wizardProviderIdx]
				m.syncWizardProfileCursor()
			}
			m.screen = installProfileScreen
			return m, nil, true
		case "up", "k":
			m.wizardConfirmCursor = 0
		case "down", "j":
			m.wizardConfirmCursor = 1
		case "s", "y":
			return m.startWizardInstall(false), tea.Batch(m.startWizardInstallCmd(false), animTickCmd()), true
		case "enter":
			if m.wizardConfirmCursor == 1 {
				ids := m.selectedClientIDs()
				if len(ids) > 0 {
					m.wizardProviderIdx = len(ids) - 1
					m.wizardActiveID = ids[m.wizardProviderIdx]
					m.syncWizardProfileCursor()
				}
				m.screen = installProfileScreen
				return m, nil, true
			}
			return m.startWizardInstall(false), tea.Batch(m.startWizardInstallCmd(false), animTickCmd()), true
		}
		return m, nil, true

	case installGentleScreen:
		switch key {
		case "q", "ctrl+c":
			return m, tea.Quit, true
		case "esc", "n":
			m.screen = home
			m.fault = ""
			return m, nil, true
		case "a", "enter", "s":
			m.adoptGentle = true
			if m.activeTracker != nil && m.activeTracker.retry != nil {
				retry := m.activeTracker.retry
				m = m.startWizardInstall(true)
				tr := &progressTracker{ch: make(chan progressEvent, 16)}
				m.activeTracker = tr
				return m, tea.Batch(retry(tr), waitProgressCmd(tr), animTickCmd()), true
			}
			return m.startWizardInstall(true), tea.Batch(m.startWizardInstallCmd(true), animTickCmd()), true
		}
		return m, nil, true

	case updateScreen:
		actions := m.updateActions()
		switch key {
		case "q", "ctrl+c":
			return m, tea.Quit, true
		case "esc":
			m.screen = home
			m.offset = 0
			return m, nil, true
		case "up", "k":
			m.updateCursor = max(0, m.updateCursor-1)
		case "down", "j":
			m.updateCursor = min(len(actions)-1, m.updateCursor+1)
		case "pgup":
			m.offset = max(0, m.offset-max(1, m.pageSize()))
		case "pgdown":
			m.offset = max(0, min(m.offset+max(1, m.pageSize()), max(0, len(m.lines())-m.pageSize())))
		case "r":
			next, cmd := m.startUpdateSearch()
			return next, cmd, true
		case "s", "y":
			next, cmd := m.startApplyUpdates()
			return next, cmd, true
		case "enter":
			chosen := actions[min(m.updateCursor, len(actions)-1)]
			switch chosen.id {
			case "apply":
				next, cmd := m.startApplyUpdates()
				return next, cmd, true
			case "refresh":
				next, cmd := m.startUpdateSearch()
				return next, cmd, true
			case "back":
				m.screen = home
				m.offset = 0
			}
		}
		return m, nil, true

	case versionsScreen:
		items := m.availableVersionTags()
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
		case "up", "k":
			m.versionCursor = max(0, m.versionCursor-1)
		case "down", "j":
			m.versionCursor = min(max(0, len(items)-1), m.versionCursor+1)
		case "r":
			if m.installed == "" {
				m.fault = "No hay una instalación previa para restaurar en " + m.mgr.Target
				return m, nil, true
			}
			m.busyCancelable = false
			m.busyReturnScreen = ""
			m.screen = busy
			m.title = "Restaurando versión anterior"
			m.progressStep, m.progressTotal = 1, 3
			m.progressLabel = "Verificando respaldo anterior..."
			tr := &progressTracker{ch: make(chan progressEvent, 16)}
			m.activeTracker = tr
			return m, tea.Batch(
				runSingleActionCmd(m.mgr, "rollback", manager.PrepareOptions{Action: "rollback"}, "Versión anterior restaurada", tr),
				waitProgressCmd(tr),
				animTickCmd(),
			), true
		case "enter", "i":
			if len(items) == 0 {
				return m, nil, true
			}
			tag := items[m.versionCursor]
			m.busyCancelable = false
			m.busyReturnScreen = ""
			m.screen = busy
			m.title = "Aplicando versión " + tag
			m.progressStep, m.progressTotal = 1, 3
			m.progressLabel = "Descargando / verificando " + tag + "..."
			tr := &progressTracker{ch: make(chan progressEvent, 16)}
			m.activeTracker = tr
			opts := manager.PrepareOptions{
				Action:          "tag:" + tag,
				AllowDowngrade:  true,
				ConfigureModels: true,
			}
			return m, tea.Batch(
				runSingleActionCmd(m.mgr, "install", opts, "Versión "+tag+" aplicada", tr),
				waitProgressCmd(tr),
				animTickCmd(),
			), true
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
	m.showTechLog = false
	return m
}

func (m *model) startWizardInstallCmd(adopt bool) tea.Cmd {
	ids := m.selectedClientIDs()
	tr := &progressTracker{ch: make(chan progressEvent, 16)}
	m.activeTracker = tr
	return tea.Batch(
		runInstallBatchCmd(m.mgr, ids, m.wizardConfigs, "", m.wizardUseLocal, adopt, false, tr),
		waitProgressCmd(tr),
	)
}

func (m model) availableVersionTags() []string {
	seen := map[string]bool{}
	var tags []string
	for _, r := range m.releases {
		if !seen[r.Tag] {
			seen[r.Tag] = true
			tags = append(tags, r.Tag)
		}
	}
	for _, t := range m.cachedTags {
		if !seen[t] {
			seen[t] = true
			tags = append(tags, t)
		}
	}
	if len(tags) == 0 && m.updateReport.Latest.Tag != "" {
		tags = append(tags, m.updateReport.Latest.Tag)
	}
	return tags
}

func (m model) crtWizard(width int, page string) (string, string) {
	var b strings.Builder
	switch m.screen {
	case installPlatformsScreen:
		b.WriteString(crtGreen.Render("Paso 1 de 4 · Elegí en qué plataforma(s) instalar") + "\n\n")
		b.WriteString("Movete con ↑↓ y usá Enter o Espacio para marcar/desmarcar:\n\n")
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
		contLabel := fmt.Sprintf("▶ Continuar con las seleccionadas (%d) →", selectedCount)
		b.WriteString("\n")
		if m.wizardCursor == len(clients) {
			b.WriteString(crtNeon.Render("› " + contLabel))
		} else {
			b.WriteString("  " + crtAmber.Render(contLabel))
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
		return strings.TrimRight(b.String(), "\n"), "↑↓ mover · Enter / Espacio marcar o continuar · Esc volver"

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
				if i < m.wizardProviderIdx {
					steps = append(steps, crtGreen.Render("[✓ "+cs.Short+"]"))
				} else if i == m.wizardProviderIdx {
					steps = append(steps, crtNeon.Render("[● "+cs.Short+"]"))
				} else {
					steps = append(steps, crtBeige.Render("[○ "+cs.Short+"]"))
				}
			}
			b.WriteString("Orden: " + strings.Join(steps, " → ") + "\n")
		}
		b.WriteString("\nElegí con ↑↓ (se selecciona al pasar) y Enter para aceptar:\n\n")
		for i, p := range spec.Presets {
			mark := "( )"
			if cfg.Preset == p.ID {
				mark = "(●)"
			}
			line := fmt.Sprintf("%s %s", mark, presetFriendlyName(p))
			if i == m.wizardProfileCursor {
				b.WriteString(crtNeon.Render("› "+line) + "\n")
				if width >= 56 {
					b.WriteString("    " + crtBeige.Render(p.Description) + "\n")
				}
			} else {
				b.WriteString("  " + line + "\n")
			}
		}
		customMark := "( )"
		if cfg.Preset == "custom" {
			customMark = "(●)"
		}
		customLine := customMark + " ⚙ Personalizar cada rol a gusto →"
		if m.wizardProfileCursor == len(spec.Presets) {
			b.WriteString(crtNeon.Render("› "+customLine) + "\n")
		} else {
			b.WriteString("  " + customLine + "\n")
		}
		if len(cfg.Roles) > 0 && width >= 58 {
			b.WriteString("\n" + crtBeige.Render(fmt.Sprintf("Resumen: Orquestador %s · Worker %s · Arquitecto %s",
				cfg.Roles["orchestrator"].Model, cfg.Roles["chuli_worker"].Model, cfg.Roles["solution_architect"].Model)))
		}
		return strings.TrimRight(b.String(), "\n"), "↑↓ elegir · Enter aceptar o personalizar · Esc atrás"

	case installConfirmScreen:
		b.WriteString(crtGreen.Render("Paso 3 de 4 · Confirmación final") + "\n\n")
		for _, id := range m.selectedClientIDs() {
			spec, _ := manager.FindClient(id)
			home, _ := manager.ResolveClientHome(id)
			cfg := m.wizardConfigs[id]
			b.WriteString(crtAmber.Render("• "+spec.Name) + "\n")
			b.WriteString("  Perfil: " + cfg.Preset + "\n")
			if width >= 54 {
				b.WriteString("  Carpeta: " + home + "\n")
			}
		}
		b.WriteString("\n" + crtBeige.Render("Tus archivos y configuraciones actuales se respaldan automáticamente antes de escribir.") + "\n\n")
		opts := []string{"▶ Confirmar e instalar ahora", "← Volver atrás para cambiar"}
		for i, label := range opts {
			if i == m.wizardConfirmCursor {
				b.WriteString(crtNeon.Render("› "+label) + "\n")
			} else {
				b.WriteString("  " + label + "\n")
			}
		}
		return strings.TrimRight(b.String(), "\n"), "↑↓ elegir · Enter aceptar · S confirmar · Esc atrás"

	case installGentleScreen:
		b.WriteString(crtAmber.Render("Atención · Encontramos instrucciones anteriores de Gentle AI") + "\n\n")
		b.WriteString("La plataforma en conflicto se detuvo antes de aplicar cambios.\n")
		if m.resultSummary != "" {
			b.WriteString("\n" + crtText(m.resultSummary) + "\n\n")
		}
		b.WriteString("Si aceptás, guardamos un respaldo completo de lo anterior y activamos Chuli.\n")
		return strings.TrimRight(b.String(), "\n"), "A / Enter respaldar y activar Chuli · Esc cancelar"

	case updateScreen:
		b.WriteString(crtGreen.Render("Actualizar Chuli") + "\n\n")
		if m.updateCheckErr != "" {
			b.WriteString(crtAmber.Render("No se pudo consultar GitHub: ") + m.updateCheckErr + "\n\n")
		} else if !m.updatesChecked {
			b.WriteString("Todavía no se consultó GitHub.\n\n")
		} else {
			b.WriteString("Última versión oficial: " + crtNeon.Render(m.updateReport.Latest.Tag) + "\n")
			if m.updateReport.ManagerNewer {
				b.WriteString(crtAmber.Render("★ Hay una versión nueva del binario Manager ("+m.updateReport.Latest.Tag+"). Actualizalo con install-chuli-windows.ps1 o go install.") + "\n")
			}
			b.WriteString("\nEstado en tu equipo:\n")
			for _, cu := range m.updateReport.Clients {
				if !cu.HasInstall {
					b.WriteString(fmt.Sprintf("  • %-8s Sin instalar\n", cu.Name))
				} else if cu.HasUpdate {
					b.WriteString(fmt.Sprintf("  • %-8s v%s → %s (actualizable)\n", cu.Name, cu.Installed, crtNeon.Render(m.updateReport.Latest.Tag)))
				} else {
					b.WriteString(fmt.Sprintf("  • %-8s v%s (al día)\n", cu.Name, cu.Installed))
				}
			}
			b.WriteString("\n" + crtAmber.Render("Qué trae "+m.updateReport.Latest.Tag+":") + "\n" + page + "\n")
		}
		b.WriteString("\n")
		for i, act := range m.updateActions() {
			if i == m.updateCursor {
				b.WriteString(crtNeon.Render("› "+act.label) + "\n")
			} else {
				b.WriteString("  " + act.label + "\n")
			}
		}
		return strings.TrimRight(b.String(), "\n"), "↑↓ elegir · Enter aceptar · R buscar de nuevo · PgUp/PgDn novedades · Esc volver"

	case versionsScreen:
		spec, _ := manager.FindClient(m.mgr.Target)
		b.WriteString(crtGreen.Render("Gestionar versiones") + "  " + renderClientTabs(spec.ID) + "\n\n")
		cur := "Sin instalar"
		if m.installed != "" {
			cur = "v" + m.installed
		}
		b.WriteString("Versión actual en " + spec.Short + ": " + crtAmber.Render(cur) + "\n\n")
		tags := m.availableVersionTags()
		if len(tags) == 0 {
			b.WriteString("No hay releases cargadas todavía (verificá conexión a GitHub).\n")
		} else {
			cachedSet := map[string]bool{}
			for _, t := range m.cachedTags {
				cachedSet[t] = true
			}
			for i, tag := range tags {
				var badges []string
				if "v"+m.installed == tag {
					badges = append(badges, "Instalada")
				}
				if cachedSet[tag] {
					badges = append(badges, "En caché")
				}
				line := tag
				if len(badges) > 0 {
					line += " (" + strings.Join(badges, ", ") + ")"
				}
				if i == m.versionCursor {
					b.WriteString(crtNeon.Render("› "+line) + "\n")
				} else {
					b.WriteString("  " + line + "\n")
				}
			}
		}
		if m.fault != "" {
			b.WriteString("\n" + crtAmber.Render(m.fault))
		}
		return strings.TrimRight(b.String(), "\n"), "↑↓ versión · Enter instalar/cambiar · R restaurar anterior · Tab cliente · Esc volver"
	}
	return "", "Esc volver"
}
