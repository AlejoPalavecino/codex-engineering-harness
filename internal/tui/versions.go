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
	updateScreen   = "update-view"
	versionsScreen = "versions-view"
)

type updatesLoadedMsg struct {
	report   manager.UpdateReport
	releases []manager.Release
	cached   []string
	statuses []manager.ClientStatus
	err      error
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

func (m model) updateVersions(message tea.Msg) (model, tea.Cmd, bool) {
	if msg, ok := message.(updatesLoadedMsg); ok {
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
	}

	msg, ok := message.(tea.KeyPressMsg)
	if !ok {
		return m, nil, false
	}
	key := strings.ToLower(msg.String())

	switch m.screen {
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

func (m model) crtVersions(width int, page string) (string, string) {
	var b strings.Builder
	switch m.screen {
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
