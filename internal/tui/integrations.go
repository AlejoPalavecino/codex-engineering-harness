package tui

import (
	"fmt"
	"strings"

	tea "charm.land/bubbletea/v2"
	"github.com/AlejoPalavecino/codex-engineering-harness/internal/integrations"
)

func (m model) updateIntegrations(message tea.Msg) (model, tea.Cmd, bool) {
	if msg, ok := message.(integrationMsg); ok {
		m.screen, m.log, m.offset = msg.screen, msg.log, 0
		m.integrationPlan = msg.plan
		m.fault = ""
		if msg.err != nil {
			m.fault = msg.err.Error()
		}
		return m, nil, true
	}
	if m.screen != integrationList && m.screen != integrationDetail && m.screen != integrationPreview && m.screen != integrationResult {
		return m, nil, false
	}
	msg, ok := message.(tea.KeyPressMsg)
	if !ok {
		return m, nil, false
	}
	key := strings.ToLower(msg.String())
	if key == "q" || key == "ctrl+c" {
		return m, tea.Quit, true
	}
	if key == "esc" || (key == "n" && m.screen == integrationPreview) {
		switch m.screen {
		case integrationList:
			m.screen = home
			m.integrationID = ""
		case integrationDetail:
			m.screen = integrationList
		default:
			m.screen = integrationDetail
			m.log = "Elegí Configurar, Comprobar o Guía para continuar."
		}
		m.fault, m.offset, m.integrationPlan = "", 0, integrations.Plan{}
		return m, nil, true
	}
	if m.screen == integrationList {
		items := integrations.Catalog()
		switch key {
		case "down", "j":
			m.integrationCursor = min(len(items)-1, m.integrationCursor+1)
		case "up", "k":
			m.integrationCursor = max(0, m.integrationCursor-1)
		case "enter":
			m.integrationID = items[m.integrationCursor].ID
			m.title = "Revisando " + items[m.integrationCursor].Name
			m.screen = busy
			return m, m.integrationCmd("detail"), true
		}
		return m, nil, true
	}
	if key == "down" || key == "j" || key == "up" || key == "k" || key == "pgdown" || key == "pgup" {
		step := 1
		if key == "pgdown" || key == "pgup" {
			step = m.pageSize()
		}
		if key == "up" || key == "k" || key == "pgup" {
			step = -step
		}
		m.offset = max(0, min(m.offset+step, len(m.lines())-m.pageSize()))
		return m, nil, true
	}
	if m.screen == integrationPreview {
		if key == "s" || key == "y" {
			m.screen = busy
			m.title = "Aplicando integración"
			return m, m.integrationCmd("apply"), true
		}
		return m, nil, true
	}
	if m.screen == integrationResult && key == "enter" {
		m.screen = busy
		m.title = "Revisando integración"
		return m, m.integrationCmd("detail"), true
	}
	if key == "g" {
		i, _ := integrations.Find(m.integrationID)
		m.screen = integrationResult
		m.log = i.Instructions + "\n\nGuía / descarga oficial:\n" + i.GuideURL
		m.fault = ""
		m.offset = 0
		return m, nil, true
	}
	action := ""
	switch key {
	case "c":
		action = "configure"
	case "d":
		action = "undo"
	case "v":
		action = "verify"
	case "o":
		action = "open"
	}
	if action != "" {
		m.screen = busy
		m.title = map[string]string{"configure": "Preparando conexión", "undo": "Preparando reversión", "verify": "Comprobando integración", "open": "Abriendo sitio oficial"}[action]
		return m, m.integrationCmd(action), true
	}
	return m, nil, true
}

func (m model) integrationView() string {
	var b strings.Builder
	b.WriteString(heading.Render("Integraciones") + "\n\n")
	if m.screen == integrationList {
		for index, item := range integrations.Catalog() {
			label := fmt.Sprintf("%s · %s", item.Name, item.Purpose)
			if m.width > 0 && m.width < 90 {
				label = item.Name
			}
			if index == m.integrationCursor {
				b.WriteString(active.Render("› "+label) + "\n")
			} else {
				b.WriteString("  " + label + "\n")
			}
		}
		b.WriteString("\nElegí una opción para revisar estado, requisitos y guía.\n")
		b.WriteString(muted.Render("↑↓ / j k navegar · Enter elegir · Esc volver · q salir"))
		return b.String()
	}
	i, _ := integrations.Find(m.integrationID)
	b.WriteString(i.Name + "\n")
	if m.screen == integrationPreview {
		b.WriteString("Revisá los cambios antes de confirmar.\n")
	}
	if m.fault != "" {
		b.WriteString(bad.Render("No se completó la operación") + "\n")
	}
	b.WriteString(m.page() + "\n\n")
	if m.screen == integrationPreview {
		b.WriteString(good.Render("S / Y confirmar") + " · N cancelar · Esc volver")
	} else {
		separator := " · "
		if m.width > 0 && m.width < 70 {
			separator = "\n"
		}
		if m.integrationID != "egolite" {
			b.WriteString("C configurar · D deshacer" + separator)
		}
		b.WriteString("V comprobar · G guía" + separator + "O abrir sitio oficial\nEnter revisar ficha" + separator + "Esc volver · q salir")
	}
	return b.String()
}
