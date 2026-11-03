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

var menu = []struct {
	label  string
	action string
}{
	{"Instalar / actualizar ultima release estable", "latest"},
	{"Instalar desde un checkout local", "local"},
	{"Diagnosticar instalacion", "doctor"},
	{"Rollback de la ultima instalacion", "rollback"},
	{"Acerca de Chuli", "about"},
}

type model struct {
	mgr            manager.Manager
	managerVersion string
	installed      string
	cursor         int
	screen         string
	operation      string
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
		if m.screen == home {
			switch key {
			case "up", "k":
				if m.cursor > 0 {
					m.cursor--
				}
			case "down", "j":
				if m.cursor < len(menu)-1 {
					m.cursor++
				}
			case "enter":
				choice := menu[m.cursor]
				m.operation = choice.action
				m.title = choice.label
				m.fault = ""
				m.log = ""
				if choice.action == "about" {
					m.screen = about
					return m, nil
				}
				m.screen = busy
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
		m.prepared = msg.plan
		m.log = msg.plan.Preview
		m.screen = preview
		return m, nil
	case finishedMsg:
		m.prepared.Cleanup()
		m.prepared = manager.Prepared{}
		m.screen = result
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
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
		defer cancel()
		log, err := m.Apply(ctx, plan)
		return finishedMsg{log, err}
	}
}
func diagnoseCmd(m manager.Manager) tea.Cmd {
	return func() tea.Msg {
		root, err := m.Local()
		if err != nil {
			return finishedMsg{"", err}
		}
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
		defer cancel()
		log, err := m.Script(ctx, root, "verify.ps1", false)
		return finishedMsg{log, err}
	}
}

func cutLog(input string, lines int) string {
	chunks := strings.Split(strings.ReplaceAll(input, "\r\n", "\n"), "\n")
	if len(chunks) > lines {
		chunks = append([]string{"... (mostrando las últimas líneas)"}, chunks[len(chunks)-lines:]...)
	}
	for i, line := range chunks {
		if len([]rune(line)) > 110 {
			chunks[i] = string([]rune(line)[:110]) + "..."
		}
	}
	return strings.Join(chunks, "\n")
}

func (m model) View() tea.View {
	var b strings.Builder
	b.WriteString(heading.Render("chuli  /  Engineering Harness Manager") + "    " + muted.Render("v"+m.managerVersion) + "\n")
	b.WriteString(muted.Render(strings.Repeat("─", 57)) + "\n")
	installed := m.installed
	if installed == "" {
		installed = "No instalado"
	}
	b.WriteString("Destino: Codex   |   Harness: " + installed + "\n")
	b.WriteString(muted.Render("CODEX_HOME: ") + m.mgr.Home + "\n\n")

	switch m.screen {
	case home:
		b.WriteString("Seleccioná una acción:\n\n")
		for i, item := range menu {
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
		b.WriteString("Todavía NO se modificó tu instalación.\n\n")
		b.WriteString(cutLog(m.log, 19) + "\n\n")
		b.WriteString(good.Render("S / Y confirmar") + "    " + muted.Render("N cancelar · Esc volver") + "\n")
	case result:
		if m.fault != "" {
			b.WriteString(bad.Render("No se completó la operación") + "\n" + m.fault + "\n\n")
		} else {
			b.WriteString(good.Render("Operación finalizada") + "\n\n")
		}
		if m.log != "" {
			b.WriteString(cutLog(m.log, 21) + "\n\n")
		}
		b.WriteString(muted.Render("Enter / Esc volver · q salir") + "\n")
	case about:
		b.WriteString(heading.Render("Chuli — AI Engineering Harness") + "\n")
		b.WriteString("La misma ingeniería, independiente del proveedor.\n\n")
		b.WriteString("EWP · MRP · ASEP · TKA · CES · TTP\n")
		b.WriteString("Este MVP administra Codex en Windows reutilizando los scripts publicados.\n")
		b.WriteString("Antigravity, otros sistemas y autoactualización del ejecutable: futuros adaptadores.\n\n")
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
