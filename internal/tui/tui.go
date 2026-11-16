// Package tui provides the interactive Chuli Manager terminal application.
package tui

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"github.com/AlejoPalavecino/codex-engineering-harness/v2/internal/integrations"
	"github.com/AlejoPalavecino/codex-engineering-harness/v2/internal/manager"
)

const (
	home     = "home"
	busy     = "busy"
	preview  = "preview"
	conflict = "conflict"
	result   = "result"
	about    = "about"
)

var (
	heading = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#79CFFF"))
	muted   = lipgloss.NewStyle().Foreground(lipgloss.Color("#8794AA"))
	active  = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#FFFFFF")).Background(lipgloss.Color("#245178"))
	good    = lipgloss.NewStyle().Foreground(lipgloss.Color("#60D6A1"))
	bad     = lipgloss.NewStyle().Foreground(lipgloss.Color("#FF919B"))
)

type menuItem struct{ label, action string }

func choices(mgr manager.Manager) []menuItem {
	return []menuItem{
		{"Instalar", "install"},
		{"Actualizar", "update"},
		{"Gestionar versiones", "versions"},
		{"Perfiles de AI", "models"},
		{"MCPs", "integrations"},
		{"Skills", "skills"},
	}
}

type model struct {
	width               int
	integrationID       string
	integrationCursor   int
	integrationPlan     integrations.Plan
	configuredMCPs      []integrations.ConfiguredServer
	selectedCustomMCP   integrations.ConfiguredServer
	customMCPField      int
	customMCPName       string
	customMCPIsCmd      bool
	customMCPTarget     string
	customMCPEnv        string
	roleCursor          int
	height, offset      int
	mgr                 manager.Manager
	managerVersion      string
	installed           string
	cursor              int
	screen              string
	title               string
	action              string
	prepared            manager.Prepared
	log                 string
	fault               string
	resultSummary       string
	resultNext          string
	showTechLog         bool
	statuses            []manager.ClientStatus
	updatesChecked      bool
	updateCheckErr      string
	updateReport        manager.UpdateReport
	releases            []manager.Release
	cachedTags          []string
	wizardSelected      map[string]bool
	wizardConfigs       map[string]wizardClientConfig
	wizardCursor        int
	wizardProfileCursor int
	wizardActiveID      string
	wizardProviderIdx   int
	wizardConfirmCursor int
	wizardUseLocal      bool
	wizardReturn        bool
	adoptGentle         bool
	versionCursor       int
	updateCursor        int
	rolePickerModelIdx  int
	progressStep        int
	progressTotal       int
	progressLabel       string
	spinnerIdx          int
	activeTracker       *progressTracker
	busyCancelable      bool
	busyReturnScreen    string
	savedProfiles       []manager.SavedProfile
	savedProfileIdx     int
	profileNameInput    string
	skillsList          []manager.SkillInfo
	skillsDraftDisabled map[string]bool
	skillsCursor        int
	skillsDirty         bool
	skillPathInput      string
	skillNotice         string
}

type preparedMsg struct {
	plan manager.Prepared
	err  error
}

type finishedMsg struct {
	log string
	err error
}

func (m model) Init() tea.Cmd {
	return checkUpdatesCmd(m.mgr, m.managerVersion)
}

func (m model) Update(message tea.Msg) (tea.Model, tea.Cmd) {
	if next, cmd, handled := m.updateWizard(message); handled {
		return next, cmd
	}
	if next, cmd, handled := m.updateSkills(message); handled {
		return next, cmd
	}
	if next, cmd, handled := m.updateIntegrations(message); handled {
		return next, cmd
	}
	if next, cmd, handled := m.updateModels(message); handled {
		return next, cmd
	}
	switch msg := message.(type) {
	case tea.WindowSizeMsg:
		m.height = msg.Height
		m.width = msg.Width
	case tea.KeyPressMsg:
		key := strings.ToLower(msg.String())
		if m.screen == busy {
			if m.busyCancelable {
				switch key {
				case "ctrl+c", "q":
					return m, tea.Quit
				case "esc":
					ret := m.busyReturnScreen
					if ret == "" {
						ret = home
					}
					m.busyCancelable = false
					m.busyReturnScreen = ""
					m.screen = ret
					return m, nil
				}
			}
			return m, nil
		} // Do not detach a mutating process.
		switch key {
		case "ctrl+c", "q":
			m.prepared.Cleanup()
			return m, tea.Quit
		case "esc":
			m.prepared.Cleanup()
			m.prepared = manager.Prepared{}
			m.screen = home
			m.fault = ""
			m.log = ""
			m.resultSummary = ""
			m.resultNext = ""
			m.showTechLog = false
			return m, nil
		}
		if m.screen == preview || m.screen == result || m.screen == conflict {
			switch key {
			case "down", "j", "pgdown", "up", "k", "pgup":
				step := 1
				if key == "pgdown" || key == "pgup" {
					step = m.pageSize()
				}
				if key == "up" || key == "k" || key == "pgup" {
					step = -step
				}
				m.offset = max(0, min(m.offset+step, len(m.lines())-m.pageSize()))
				return m, nil
			}
		}
		if m.screen == result && key == "v" {
			m.showTechLog = !m.showTechLog
			m.offset = 0
			return m, nil
		}
		if m.screen == home {
			options := choices(m.mgr)
			switch key {
			case "tab":
				m.nextClient()
			case "1":
				m.switchClient("codex")
			case "2":
				m.switchClient("claude")
			case "3":
				m.switchClient("agy")
			case "m":
				m.ensureModelState()
				m.reloadSavedProfiles()
				m.wizardReturn = false
				m.screen = modelsScreen
				return m, nil
			case "?", "a":
				m.screen = about
				return m, nil
			case "d":
				m.offset = 0
				m.title = "Verificar el perfil instalado"
				m.action = "doctor"
				m.fault = ""
				m.log = ""
				m.resultSummary = ""
				m.resultNext = ""
				m.showTechLog = true
				m.progressStep, m.progressTotal = 1, 1
				m.progressLabel = "Comprobando archivos, reglas y configuración..."
				m.screen = busy
				return m, diagnoseCmd(m.mgr)
			case "up", "k":
				if m.cursor > 0 {
					m.cursor--
				}
			case "down", "j":
				if m.cursor < len(options)-1 {
					m.cursor++
				}
			case "enter":
				choice := options[m.cursor]
				m.offset = 0
				m.title = choice.label
				m.action = choice.action
				m.fault = ""
				m.log = ""
				m.resultSummary = ""
				m.resultNext = ""
				m.showTechLog = false
				switch choice.action {
				case "install":
					m.initWizard()
					m.screen = installPlatformsScreen
					return m, nil
				case "update":
					m.updateCursor = 0
					if !m.updatesChecked {
						m.screen = busy
						m.busyCancelable = true
						m.busyReturnScreen = updateScreen
						m.title = "Buscando actualizaciones"
						m.progressStep, m.progressTotal = 1, 1
						m.progressLabel = "Consultando últimas releases oficiales en GitHub..."
						return m, tea.Batch(checkUpdatesCmd(m.mgr, m.managerVersion), animTickCmd())
					}
					m.screen = updateScreen
					m.log = m.updateReport.Summary
					return m, nil
				case "versions":
					m.versionCursor = 0
					if !m.updatesChecked {
						m.screen = busy
						m.busyCancelable = true
						m.busyReturnScreen = versionsScreen
						m.title = "Cargando versiones"
						m.progressStep, m.progressTotal = 1, 1
						m.progressLabel = "Consultando releases disponibles en GitHub..."
						return m, tea.Batch(checkUpdatesCmd(m.mgr, m.managerVersion), animTickCmd())
					}
					m.screen = versionsScreen
					return m, nil
				case "models":
					m.ensureModelState()
					m.reloadSavedProfiles()
					m.wizardReturn = false
					m.screen = modelsScreen
					return m, nil
				case "integrations":
					m.integrationCursor = 0
					m.screen = integrationList
					return m, nil
				case "skills":
					m.skillsCursor = 0
					m.skillNotice = ""
					m.screen = skillsScreen
					return m, loadSkillsCmd(m.mgr)
				}
			}
		} else if m.screen == conflict && key == "a" {
			m.screen = busy
			m.title = "Usar Chuli como orquestador y respaldar Gentle"
			m.progressLabel = "Respaldando instrucciones anteriores y preparando plan..."
			m.log, m.fault = "", ""
			m.offset = 0
			return m, prepareCmd(m.mgr, m.action, true)
		} else if m.screen == preview {
			switch key {
			case "y", "s":
				m.screen = busy
				m.title = "Aplicando: " + m.title
				m.progressLabel = "Aplicando cambios de forma segura..."
				return m, applyCmd(m.mgr, m.prepared)
			case "n":
				m.prepared.Cleanup()
				m.prepared = manager.Prepared{}
				m.screen = home
			}
		} else if m.screen == result && key == "enter" {
			m.screen = home
			m.resultSummary = ""
			m.resultNext = ""
			m.showTechLog = false
			m.fault = ""
			m.log = ""
		}
	case preparedMsg:
		if msg.err != nil {
			m.screen = result
			if errors.Is(msg.err, manager.ErrGentleConflict) {
				m.screen = conflict
			}
			m.fault = msg.err.Error()
			return m, nil
		}
		m.offset = 0
		m.prepared = msg.plan
		m.log = msg.plan.Preview
		m.screen = preview
		return m, nil
	case finishedMsg:
		m.prepared.Cleanup()
		m.prepared = manager.Prepared{}
		m.screen = result
		m.offset = 0
		m.log = msg.log
		if msg.err != nil {
			m.fault = msg.err.Error()
		} else {
			m.fault = ""
		}
		if installed, err := m.mgr.Installed(); err == nil {
			m.installed = installed
		}
		return m, nil
	}
	return m, nil
}

func prepareCmd(m manager.Manager, action string, adoptGentleRouting bool) tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
		defer cancel()
		plan, err := m.Prepare(ctx, action, adoptGentleRouting)
		return preparedMsg{plan, err}
	}
}
func applyCmd(m manager.Manager, plan manager.Prepared) tea.Cmd {
	return func() tea.Msg {
		log, err := m.Apply(context.Background(), plan)
		return finishedMsg{log, err}
	}
}
func diagnoseCmd(m manager.Manager) tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
		defer cancel()
		log, err := m.Doctor(ctx)
		return finishedMsg{log, err}
	}
}

func (m model) pageSize() int {
	if m.height == 0 {
		return 15
	}
	if m.width == 0 {
		return max(1, m.height-12)
	}
	if m.crtPaged() {
		return m.crtPageSize()
	}
	return max(1, m.height-12)
}

func (m model) lines() []string {
	log := crtText(m.log)
	if m.screen == updateScreen {
		log = crtText(m.updateReport.Summary)
		if strings.TrimSpace(log) == "" {
			log = "No hay notas adicionales para esta release."
		}
	}
	if m.fault != "" && m.screen != updateScreen {
		log = crtText(m.fault) + "\n" + log
	}
	if m.width > 0 && (m.screen == preview || m.screen == integrationPreview) {
		prefix := "Cambios: revisá el plan antes de confirmar.\n\n"
		if !strings.Contains(log, "Destino: "+crtText(m.mgr.Home)) {
			prefix = "Destino: " + crtText(m.mgr.Home) + "\n" + prefix
		}
		log = prefix + log
	}
	if m.crtPaged() && m.width > 0 {
		log = lipgloss.NewStyle().Width(max(20, m.width-2)).Render(log)
	}
	return strings.Split(strings.TrimSpace(strings.ReplaceAll(log, "\r\n", "\n")), "\n")
}

func (m model) page() string {
	lines := m.lines()
	start := min(m.offset, max(0, len(lines)-m.pageSize()))
	end := min(len(lines), start+m.pageSize())
	page := strings.Join(lines[start:end], "\n")
	if end-start < len(lines) {
		page += "\n" + muted.Render(fmt.Sprintf("Líneas %d–%d de %d · ↑↓ / PgUp / PgDn", start+1, end, len(lines)))
	}
	return page
}

func (m model) View() tea.View { return m.renderCRT() }

func Run(mgr manager.Manager, version string) error {
	installed, err := mgr.Installed()
	if err != nil {
		return err
	}
	statuses := mgr.DetectClients()
	initial := model{
		mgr:                 mgr,
		managerVersion:      version,
		installed:           installed,
		statuses:            statuses,
		screen:              home,
		savedProfileIdx:     -1,
		skillsDraftDisabled: map[string]bool{},
	}
	initial.reloadSavedProfiles()
	if servers, err := initial.integrationService().ListConfigured(context.Background()); err == nil {
		initial.configuredMCPs = servers
	}
	final, err := tea.NewProgram(initial).Run()
	if finalModel, ok := final.(model); ok {
		finalModel.prepared.Cleanup()
	}
	return err
}
