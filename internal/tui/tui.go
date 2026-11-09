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

	"github.com/AlejoPalavecino/codex-engineering-harness/internal/integrations"
	"github.com/AlejoPalavecino/codex-engineering-harness/internal/manager"
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
	items := []menuItem{{"Instalar o actualizar el perfil oficial", "latest"}}
	if mgr.Source != "" {
		items = append(items, menuItem{"Instalar desde --source", "local"})
	}
	return append(items, menuItem{"Integraciones: configurar herramientas", "integrations"}, menuItem{"Ver modelos, agentes y skills registrados", "profile"}, menuItem{"Verificar el perfil instalado", "doctor"}, menuItem{"Restaurar la instalación anterior", "rollback"}, menuItem{"Acerca de Chuli", "about"})
}

type model struct {
	width             int
	integrationID     string
	integrationCursor int
	integrationPlan   integrations.Plan
	height, offset    int
	mgr               manager.Manager
	managerVersion    string
	installed         string
	cursor            int
	screen            string
	title             string
	action            string
	prepared          manager.Prepared
	log               string
	fault             string
}

type preparedMsg struct {
	plan manager.Prepared
	err  error
}

type finishedMsg struct {
	log string
	err error
}

func (m model) Init() tea.Cmd { return nil }

func (m model) Update(message tea.Msg) (tea.Model, tea.Cmd) {
	if next, cmd, handled := m.updateIntegrations(message); handled {
		return next, cmd
	}
	switch msg := message.(type) {
	case tea.WindowSizeMsg:
		m.height = msg.Height
		m.width = msg.Width
	case tea.KeyPressMsg:
		key := strings.ToLower(msg.String())
		if m.screen == busy {
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
		if m.screen == home {
			options := choices(m.mgr)
			switch key {
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
				if choice.action == "about" {
					m.screen = about
					return m, nil
				}
				if choice.action == "integrations" {
					m.screen = integrationList
					return m, nil
				}
				m.screen = busy
				if choice.action == "profile" {
					return m, func() tea.Msg { summary, err := m.mgr.Profile(); return finishedMsg{summary, err} }
				}
				if choice.action == "doctor" {
					return m, diagnoseCmd(m.mgr)
				}
				return m, prepareCmd(m.mgr, choice.action, false)
			}
		} else if m.screen == conflict && key == "a" {
			m.screen = busy
			m.title = "Usar Chuli como orquestador y respaldar Gentle"
			m.log, m.fault = "", ""
			m.offset = 0
			return m, prepareCmd(m.mgr, m.action, true)
		} else if m.screen == preview {
			switch key {
			case "y", "s":
				m.screen = busy
				m.title = "Aplicando: " + m.title
				return m, applyCmd(m.mgr, m.prepared)
			case "n":
				m.prepared.Cleanup()
				m.prepared = manager.Prepared{}
				m.screen = home
			}
		} else if m.screen == result && key == "enter" {
			m.screen = home
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
	if m.fault != "" {
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
	initial := model{mgr: mgr, managerVersion: version, installed: installed, screen: home}
	final, err := tea.NewProgram(initial).Run()
	if finalModel, ok := final.(model); ok {
		finalModel.prepared.Cleanup()
	}
	return err
}
