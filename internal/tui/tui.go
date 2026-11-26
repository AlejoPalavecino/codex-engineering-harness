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

type integrationState struct {
	integrationID     string
	integrationCursor int
	integrationPlan   integrations.Plan
	configuredMCPs    []integrations.ConfiguredServer
	selectedCustomMCP integrations.ConfiguredServer
	customMCPField    int
	customMCPName     string
	customMCPIsCmd    bool
	customMCPTarget   string
	customMCPEnv      string
}

type updateState struct {
	statuses       []manager.ClientStatus
	updatesChecked bool
	updateCheckErr string
	updateReport   manager.UpdateReport
	releases       []manager.Release
	cachedTags     []string
	versionCursor  int
	updateCursor   int
}

type wizardState struct {
	wizardSelected         map[string]bool
	wizardConfigs          map[string]wizardClientConfig
	wizardCursor           int
	wizardProfileCursor    int
	wizardActiveID         string
	wizardProviderIdx      int
	wizardConfirmCursor    int
	wizardUseLocal         bool
	wizardReturn           bool
	wizardCameFromFastPath bool
	wizardSavedName        map[string]string
	wizardAdoptGentle      map[string]bool
	wizardPreflight        map[string]bool
	wizardRuntime          manager.RuntimePreflight
	wizardPreflightDone    bool
	adoptGentle            bool
}

type modelPickerState struct {
	roleCursor         int
	rolePickerModelIdx int
	modelChoices       []string
	modelLabels        map[string]string
	pickerChoices      []string
	modelSource        string
	modelRequest       int
	savedProfiles      []manager.SavedProfile
	savedProfileIdx    int
	profileNameInput   string
}

type progressState struct {
	progressStep     int
	progressTotal    int
	progressLabel    string
	spinnerIdx       int
	activeTracker    *progressTracker
	busyCancelable   bool
	busyReturnScreen string
}

type skillsState struct {
	skillsList          []manager.SkillInfo
	skillsDraftDisabled map[string]bool
	skillsCursor        int
	skillsDirty         bool
	skillPathInput      string
	skillNotice         string
}

type model struct {
	width, height, offset int
	mgr                   manager.Manager
	managerVersion        string
	installed             string
	cursor                int
	screen                string
	title                 string
	action                string
	prepared              manager.Prepared
	log                   string
	fault                 string
	resultSummary         string
	resultNext            string
	resultCursor          int
	resultFromWizard      bool
	showTechLog           bool
	integrationState
	updateState
	wizardState
	modelPickerState
	progressState
	skillsState
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
	if catalog, ok := message.(modelsLoadedMsg); ok {
		if catalog.request == m.modelRequest && catalog.target == m.mgr.Target {
			m.modelChoices, m.modelSource = catalog.ids, catalog.source
			m.modelLabels = catalog.labels
			if m.screen == rolePickerScreen {
				m.openRolePicker()
			}
		}
		return m, nil
	}
	next, cmd := m.update(message)
	wasEditing := m.screen == modelsScreen || m.screen == rolePickerScreen || m.screen == modelsSaveScreen
	if next.screen == modelsScreen && (!wasEditing || m.mgr.Target != next.mgr.Target) {
		next.modelChoices, next.modelSource = nil, ""
		next.modelLabels = nil
		next.modelRequest++
		return next, tea.Batch(cmd, loadModelsCmd(next.mgr, next.modelRequest))
	}
	return next, cmd
}

func (m model) update(message tea.Msg) (model, tea.Cmd) {
	if next, cmd, handled := m.updateVersions(message); handled {
		return next, cmd
	}
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
		case "ctrl+c":
			m.prepared.Cleanup()
			return m, tea.Quit
		case "q":
			if m.screen == home || m.screen == about || (m.screen == result && !m.showTechLog) {
				m.prepared.Cleanup()
				return m, tea.Quit
			}
			return m, nil
		case "esc":
			if m.screen == result && m.showTechLog && m.resultSummary != "" {
				m.showTechLog = false
				m.offset = 0
				return m, nil
			}
			m.prepared.Cleanup()
			m.prepared = manager.Prepared{}
			m.screen = home
			m.fault = ""
			m.log = ""
			m.resultSummary = ""
			m.resultNext = ""
			m.resultCursor = 0
			m.resultFromWizard = false
			m.showTechLog = false
			return m, nil
		}
		if m.screen == preview || m.screen == conflict || (m.screen == result && (m.resultSummary == "" || m.showTechLog)) {
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
			case "tab", "right", "l":
				m.nextClient()
			case "shift+tab", "left", "h":
				m.prevClient()
			case "1", "2", "3", "4", "5", "6":
				idx := int(key[0] - '1')
				if idx >= 0 && idx < len(options) {
					return m.activateHomeChoice(idx)
				}
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
				m.resultCursor = 0
				m.resultFromWizard = false
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
				return m.activateHomeChoice(m.cursor)
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
		} else if m.screen == result {
			if m.resultSummary != "" && !m.showTechLog {
				actions := m.resultActions()
				switch key {
				case "up", "k":
					m.resultCursor = max(0, m.resultCursor-1)
					return m, nil
				case "down", "j":
					m.resultCursor = min(max(0, len(actions)-1), m.resultCursor+1)
					return m, nil
				case "enter":
					if len(actions) > 0 {
						chosen := actions[min(m.resultCursor, len(actions)-1)].id
						return m.executeResultAction(chosen)
					}
				}
			} else if key == "enter" {
				m.screen = home
				m.resultSummary = ""
				m.resultNext = ""
				m.resultCursor = 0
				m.resultFromWizard = false
				m.showTechLog = false
				m.fault = ""
				m.log = ""
			}
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
		m.resultCursor = 0
		m.resultFromWizard = false
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

func (m model) hasAnyInstall() bool {
	if m.installed != "" {
		return true
	}
	for _, st := range m.statuses {
		if st.InstalledVersion != "" {
			return true
		}
	}
	return false
}

func (m model) activateHomeChoice(idx int) (model, tea.Cmd) {
	options := choices(m.mgr)
	if idx < 0 || idx >= len(options) {
		return m, nil
	}
	m.cursor = idx
	choice := options[idx]
	m.offset = 0
	m.title = choice.label
	m.action = choice.action
	m.fault = ""
	m.log = ""
	m.resultSummary = ""
	m.resultNext = ""
	m.resultCursor = 0
	m.resultFromWizard = false
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
	return m, nil
}

func (m model) resultActions() []struct{ id, label string } {
	var actions []struct{ id, label string }
	if m.fault == "" {
		return []struct{ id, label string }{
			{"home", "▶ Volver al menú principal"},
			{"mcps", "🔌 Configurar MCPs opcionales"},
			{"skills", "🧩 Revisar Skills instaladas"},
			{"quit", "⏻ Salir para empezar a usar Chuli"},
		}
	}
	if m.log != "" {
		actions = append(actions, struct{ id, label string }{"log", "🔍 Ver detalle técnico del error (V)"})
	}
	if m.resultFromWizard {
		actions = append(actions, struct{ id, label string }{"retry-wizard", "← Volver a confirmación del asistente"})
	}
	actions = append(actions, struct{ id, label string }{"home", "⌂ Volver al menú principal"})
	return actions
}

func (m model) executeResultAction(id string) (model, tea.Cmd) {
	switch id {
	case "mcps":
		if m.resultFromWizard {
			if sel := m.selectedClientIDs(); len(sel) > 0 {
				m.switchClient(sel[0])
			}
		}
		m.resultSummary = ""
		m.resultNext = ""
		m.resultCursor = 0
		m.resultFromWizard = false
		m.showTechLog = false
		m.fault = ""
		m.log = ""
		m.integrationCursor = 0
		m.screen = integrationList
		return m, loadConfiguredMCPsCmd(m)
	case "skills":
		if m.resultFromWizard {
			if sel := m.selectedClientIDs(); len(sel) > 0 {
				m.switchClient(sel[0])
			}
		}
		m.resultSummary = ""
		m.resultNext = ""
		m.resultCursor = 0
		m.resultFromWizard = false
		m.showTechLog = false
		m.fault = ""
		m.log = ""
		m.skillsCursor = 0
		m.skillNotice = ""
		m.screen = skillsScreen
		return m, loadSkillsCmd(m.mgr)
	case "retry-wizard":
		m.resultSummary = ""
		m.resultNext = ""
		m.resultCursor = 0
		m.showTechLog = false
		m.fault = ""
		m.log = ""
		return m.enterInstallConfirm(m.wizardCameFromFastPath)
	case "log":
		m.showTechLog = true
		m.offset = 0
		return m, nil
	case "quit":
		m.prepared.Cleanup()
		return m, tea.Quit
	default:
		m.screen = home
		m.resultSummary = ""
		m.resultNext = ""
		m.resultCursor = 0
		m.resultFromWizard = false
		m.showTechLog = false
		m.fault = ""
		m.log = ""
		return m, nil
	}
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
		mgr:              mgr,
		managerVersion:   version,
		installed:        installed,
		screen:           home,
		updateState:      updateState{statuses: statuses},
		modelPickerState: modelPickerState{savedProfileIdx: -1},
		skillsState:      skillsState{skillsDraftDisabled: map[string]bool{}},
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
