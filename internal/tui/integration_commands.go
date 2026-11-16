package tui

import (
	"context"
	"strings"
	"time"

	tea "charm.land/bubbletea/v2"
	"github.com/AlejoPalavecino/codex-engineering-harness/v2/internal/integrations"
)

const (
	integrationList       = "integration-list"
	integrationDetail     = "integration-detail"
	integrationPreview    = "integration-preview"
	integrationResult     = "integration-result"
	integrationCustomForm = "integration-custom-form"
)

type integrationMsg struct {
	screen, log string
	plan        integrations.Plan
	err         error
}

type mcpListLoadedMsg struct {
	servers []integrations.ConfiguredServer
	err     error
}

func (m model) integrationService() integrations.Service {
	target := m.mgr.Target
	if target == "" {
		target = "codex"
	}
	return integrations.Service{Home: m.mgr.Home, Client: target}
}

func loadConfiguredMCPsCmd(m model) tea.Cmd {
	s := m.integrationService()
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
		defer cancel()
		servers, err := s.ListConfigured(ctx)
		return mcpListLoadedMsg{servers: servers, err: err}
	}
}

func (m model) integrationCmd(action string) tea.Cmd {
	s := m.integrationService()
	id, plan := m.integrationID, m.integrationPlan
	customSrv := m.selectedCustomMCP
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 25*time.Second)
		defer cancel()
		message := integrationMsg{screen: integrationResult}
		_, isCatalog := integrations.Find(id)
		switch action {
		case "detail":
			message.screen = integrationDetail
			if isCatalog {
				message.log, message.err = s.Describe(ctx, id)
			} else {
				origin := "externo (configuración existente preservada)"
				if customSrv.Managed {
					origin = "agregado por Chuli"
				}
				message.log = customSrv.ID + " — MCP personalizado (" + s.Client + ")\nTransporte: " + customSrv.Transport + "\nDestino: " + customSrv.Target + "\nOrigen: " + origin
			}
		case "configure", "undo":
			if isCatalog {
				message.plan, message.err = s.Prepare(ctx, id, action)
			} else if action == "undo" {
				message.plan, message.err = s.PrepareRemoveCustom(ctx, customSrv)
			}
			message.log = message.plan.Preview
			if message.err == nil {
				message.screen = integrationPreview
			}
		case "apply":
			message.log, message.err = s.Apply(ctx, plan)
		case "verify":
			if isCatalog {
				message.log, message.err = s.Verify(ctx, id)
			} else {
				message.log, message.err = s.VerifyConfigured(ctx, customSrv)
			}
		case "open":
			message.err = integrations.OpenGuide(ctx, id)
			message.log = "Se solicitó abrir el sitio oficial en tu navegador. Volvé al Manager después de completar los pasos."
		}
		return message
	}
}

func (m model) prepareCustomMCPCmd() tea.Cmd {
	s := m.integrationService()
	input := integrations.CustomMCP{
		ID:       strings.TrimSpace(m.customMCPName),
		TokenEnv: strings.TrimSpace(m.customMCPEnv),
	}
	if m.customMCPIsCmd {
		parts := strings.Fields(m.customMCPTarget)
		if len(parts) > 0 {
			input.Command = parts[0]
			input.Args = parts[1:]
		}
	} else {
		input.URL = strings.TrimSpace(m.customMCPTarget)
	}
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
		defer cancel()
		plan, err := s.PrepareCustom(ctx, input)
		msg := integrationMsg{screen: integrationResult, plan: plan, log: plan.Preview, err: err}
		if err == nil {
			msg.screen = integrationPreview
		}
		return msg
	}
}
