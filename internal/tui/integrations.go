package tui

import (
	"strings"

	tea "charm.land/bubbletea/v2"
	"github.com/AlejoPalavecino/codex-engineering-harness/v2/internal/integrations"
)

type mcpListItem struct {
	ID      string
	Name    string
	Purpose string
	Catalog bool
	Active  bool
	Server  integrations.ConfiguredServer
}

func (m model) mcpListItems() []mcpListItem {
	byID := map[string]integrations.ConfiguredServer{}
	for _, s := range m.configuredMCPs {
		byID[s.ID] = s
	}
	catalog := integrations.Catalog()
	items := make([]mcpListItem, 0, len(catalog)+len(m.configuredMCPs))
	seen := map[string]bool{}
	for _, c := range catalog {
		srv, active := byID[c.ID]
		items = append(items, mcpListItem{
			ID:      c.ID,
			Name:    c.Name,
			Purpose: c.Purpose,
			Catalog: true,
			Active:  active,
			Server:  srv,
		})
		seen[c.ID] = true
	}
	for _, s := range m.configuredMCPs {
		if seen[s.ID] {
			continue
		}
		items = append(items, mcpListItem{
			ID:      s.ID,
			Name:    s.ID,
			Purpose: s.Transport + " · " + s.Target,
			Catalog: false,
			Active:  true,
			Server:  s,
		})
	}
	return items
}

func (m model) updateIntegrations(message tea.Msg) (model, tea.Cmd, bool) {
	if msg, ok := message.(mcpListLoadedMsg); ok {
		if msg.err != nil {
			m.fault = msg.err.Error()
		} else {
			m.configuredMCPs = msg.servers
		}
		return m, nil, true
	}
	if msg, ok := message.(integrationMsg); ok {
		m.screen, m.log, m.offset = msg.screen, msg.log, 0
		m.integrationPlan = msg.plan
		if msg.plan.ID != "" {
			m.integrationID = msg.plan.ID
		}
		m.fault = ""
		if msg.err != nil {
			m.fault = msg.err.Error()
		}
		if msg.screen == integrationResult && msg.err == nil {
			return m, loadConfiguredMCPsCmd(m), true
		}
		return m, nil, true
	}
	if m.screen != integrationList && m.screen != integrationDetail && m.screen != integrationPreview && m.screen != integrationResult && m.screen != integrationCustomForm {
		return m, nil, false
	}
	msg, ok := message.(tea.KeyPressMsg)
	if !ok {
		return m, nil, false
	}
	key := strings.ToLower(msg.String())

	if m.screen == integrationCustomForm {
		switch key {
		case "ctrl+c":
			return m, tea.Quit, true
		case "esc":
			m.screen = integrationList
			m.fault = ""
			return m, nil, true
		case "up", "shift+tab":
			m.customMCPField = max(0, m.customMCPField-1)
			return m, nil, true
		case "down", "tab":
			m.customMCPField = min(3, m.customMCPField+1)
			return m, nil, true
		case "left", "right", " ", "space":
			if m.customMCPField == 1 {
				m.customMCPIsCmd = !m.customMCPIsCmd
				return m, nil, true
			}
		case "backspace":
			switch m.customMCPField {
			case 0:
				m.customMCPName = dropLastRune(m.customMCPName)
			case 2:
				m.customMCPTarget = dropLastRune(m.customMCPTarget)
			case 3:
				m.customMCPEnv = dropLastRune(m.customMCPEnv)
			}
			return m, nil, true
		case "enter":
			if m.customMCPField < 3 {
				m.customMCPField++
				return m, nil, true
			}
			m.screen = busy
			m.title = "Validando MCP personalizado"
			m.progressStep, m.progressTotal = 1, 2
			m.progressLabel = "Revisando configuración de " + m.mgr.Target + "..."
			return m, tea.Batch(m.prepareCustomMCPCmd(), animTickCmd()), true
		}
		if len(msg.Text) > 0 && m.customMCPField != 1 {
			switch m.customMCPField {
			case 0:
				m.customMCPName += msg.Text
			case 2:
				m.customMCPTarget += msg.Text
			case 3:
				m.customMCPEnv += msg.Text
			}
		}
		return m, nil, true
	}

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
			if _, isCat := integrations.Find(m.integrationID); !isCat && m.integrationPlan.Action == "configure" {
				m.screen = integrationList
			} else {
				m.screen = integrationDetail
				m.log = "Elegí Configurar, Comprobar o Guía para continuar."
			}
		}
		m.fault, m.offset, m.integrationPlan = "", 0, integrations.Plan{}
		return m, nil, true
	}
	if m.screen == integrationList {
		items := m.mcpListItems()
		switch key {
		case "tab":
			m.nextClient()
			m.integrationCursor = 0
			m.fault = ""
			return m, loadConfiguredMCPsCmd(m), true
		case "1":
			m.switchClient("codex")
			m.integrationCursor = 0
			m.fault = ""
			return m, loadConfiguredMCPsCmd(m), true
		case "2":
			m.switchClient("claude")
			m.integrationCursor = 0
			m.fault = ""
			return m, loadConfiguredMCPsCmd(m), true
		case "3":
			m.switchClient("agy")
			m.integrationCursor = 0
			m.fault = ""
			return m, loadConfiguredMCPsCmd(m), true
		case "n":
			m.screen = integrationCustomForm
			m.customMCPField = 0
			m.customMCPName = ""
			m.customMCPIsCmd = false
			m.customMCPTarget = ""
			m.customMCPEnv = ""
			m.fault = ""
			return m, nil, true
		case "down", "j":
			if len(items) > 0 {
				m.integrationCursor = min(len(items)-1, m.integrationCursor+1)
			}
		case "up", "k":
			m.integrationCursor = max(0, m.integrationCursor-1)
		case "enter":
			if len(items) == 0 {
				return m, nil, true
			}
			chosen := items[m.integrationCursor]
			m.integrationID = chosen.ID
			m.selectedCustomMCP = chosen.Server
			m.title = "Revisando " + chosen.Name
			m.progressStep, m.progressTotal = 1, 1
			m.progressLabel = "Leyendo estado en " + m.mgr.Target + "..."
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
			m.title = "Aplicando configuración MCP"
			m.progressStep, m.progressTotal = 1, 1
			m.progressLabel = "Actualizando servidor MCP..."
			return m, m.integrationCmd("apply"), true
		}
		return m, nil, true
	}
	if m.screen == integrationResult && key == "enter" {
		if _, isCat := integrations.Find(m.integrationID); !isCat {
			m.screen = integrationList
			return m, loadConfiguredMCPsCmd(m), true
		}
		m.screen = busy
		m.title = "Revisando MCP"
		m.progressStep, m.progressTotal = 1, 1
		m.progressLabel = "Comprobando estado..."
		return m, m.integrationCmd("detail"), true
	}
	_, isCatalog := integrations.Find(m.integrationID)
	if key == "g" && isCatalog {
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
		if isCatalog {
			action = "configure"
		}
	case "d":
		action = "undo"
	case "v":
		action = "verify"
	case "o":
		if isCatalog {
			action = "open"
		}
	}
	if action != "" {
		m.screen = busy
		m.title = map[string]string{"configure": "Preparando conexión", "undo": "Preparando reversión", "verify": "Comprobando MCP", "open": "Abriendo sitio oficial"}[action]
		m.progressStep, m.progressTotal = 1, 1
		m.progressLabel = m.title + "..."
		return m, m.integrationCmd(action), true
	}
	return m, nil, true
}

func dropLastRune(s string) string {
	r := []rune(s)
	if len(r) == 0 {
		return ""
	}
	return string(r[:len(r)-1])
}
