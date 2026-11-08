// Package tui provides the interactive Chuli Manager terminal application.
package tui

import (
	"context"
	"fmt"
	"strings"
	"time"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"github.com/AlejoPalavecino/codex-engineering-harness/internal/manager"
)

const (
	home    = "home"
	busy    = "busy"
	preview = "preview"
	result  = "result"
	about   = "about"
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
	return append(items, menuItem{"Ver modelos, agentes y skills registrados", "profile"}, menuItem{"Verificar el perfil instalado", "doctor"}, menuItem{"Restaurar la instalación anterior", "rollback"}, menuItem{"Acerca de Chuli", "about"})
}

type model struct {
	height, offset int
	mgr            manager.Manager
	managerVersion string
	installed      string
	cursor         int
	screen         string
	title          string
	prepared       manager.Prepared
	log            string
	fault          string
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
	switch msg := message.(type) {
	case tea.WindowSizeMsg:
		m.height = msg.Height
	case tea.KeyPressMsg:
		key := msg.String()
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
		if m.screen == preview || m.screen == result {
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
				m.fault = ""
				m.log = ""
				if choice.action == "about" {
					m.screen = about
					return m, nil
				}
				m.screen = busy
				if choice.action == "profile" {
					return m, func() tea.Msg { summary, err := m.mgr.Profile(); return finishedMsg{summary, err} }
				}
				if choice.action == "doctor" {
					return m, diagnoseCmd(m.mgr)
				}
				return m, prepareCmd(m.mgr, choice.action)
			}
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

func prepareCmd(m manager.Manager, action string) tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
		defer cancel()
		plan, err := m.Prepare(ctx, action)
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
	return max(1, m.height-12)
}

func (m model) lines() []string {
	log := m.log
	if m.fault != "" {
		log = m.fault + "\n" + log
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

func (m model) View() tea.View {
	var b strings.Builder
	b.WriteString(heading.Render("chuli  /  Engineering Harness Manager") + "    " + muted.Render("v"+m.managerVersion) + "\n")
	b.WriteString(muted.Render(strings.Repeat("─", 57)) + "\n")
	installed := m.installed
	if installed == "" {
		installed = "No instalado"
	}
	b.WriteString("Perfil Chuli para Codex   |   Versión: " + installed + "\n")
	b.WriteString(muted.Render("CODEX_HOME: ") + m.mgr.Home + "\n\n")

	switch m.screen {
	case home:
		b.WriteString("Seleccioná una acción:\n\n")
		for i, item := range choices(m.mgr) {
			label := fmt.Sprintf("  %d. %s", i+1, item.label)
			if i == m.cursor {
				b.WriteString(active.Render("› "+label) + "\n")
			} else {
				b.WriteString("  " + label + "\n")
			}
		}
		b.WriteString("\n" + muted.Render("↑↓ / j k navegar  ·  Enter elegir  ·  q salir") + "\n")
	case busy:
		b.WriteString(heading.Render(m.title) + "\n\n")
		b.WriteString("Ejecutando comprobaciones. No cierres la terminal durante una escritura.\n")
	case preview:
		b.WriteString(heading.Render("Revisión previa - "+m.prepared.Version) + "\n")
		b.WriteString("Revisá los cambios antes de confirmar.\n\n")
		b.WriteString(m.page() + "\n\n")
		b.WriteString(good.Render("S / Y confirmar") + "    " + muted.Render("N cancelar · Esc volver") + "\n")
	case result:
		if m.fault != "" {
			b.WriteString(bad.Render("No se completó la operación") + "\n\n")
		} else {
			b.WriteString(good.Render(m.title) + "\n\n")
		}
		if m.log != "" || m.fault != "" {
			b.WriteString(m.page() + "\n\n")
		}
		b.WriteString(muted.Render("Enter / Esc volver · q salir") + "\n")
	case about:
		b.WriteString(heading.Render("Chuli — AI Engineering Harness") + "\n")
		b.WriteString("Instala y actualiza el perfil de ingeniería en CODEX_HOME.\n")
		b.WriteString("Después de instalar, abrí Codex como siempre.\n\n")
		b.WriteString("El paquete oficial se verifica con SHA-256 y se conserva para recuperación.\n")
		b.WriteString("Requisitos: Python 3.11+ y PowerShell (pwsh en Linux y macOS).\n\n")
		b.WriteString(muted.Render("Esc volver · q salir") + "\n")
	}
	view := tea.NewView(b.String())
	view.AltScreen = true
	view.WindowTitle = "Chuli Manager"
	return view
}

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
