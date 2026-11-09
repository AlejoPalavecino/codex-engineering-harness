package tui

import (
	tea "charm.land/bubbletea/v2"
	"context"
	"github.com/AlejoPalavecino/codex-engineering-harness/internal/integrations"
	"time"
)

const (
	integrationList    = "integration-list"
	integrationDetail  = "integration-detail"
	integrationPreview = "integration-preview"
	integrationResult  = "integration-result"
)

type integrationMsg struct {
	screen, log string
	plan        integrations.Plan
	err         error
}

func (m model) integrationCmd(action string) tea.Cmd {
	s := integrations.Service{Home: m.mgr.Home}
	id, plan := m.integrationID, m.integrationPlan
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 25*time.Second)
		defer cancel()
		message := integrationMsg{screen: integrationResult}
		switch action {
		case "detail":
			message.screen = integrationDetail
			message.log, message.err = s.Describe(ctx, id)
		case "configure", "undo":
			message.plan, message.err = s.Prepare(ctx, id, action)
			message.log = message.plan.Preview
			if message.err == nil {
				message.screen = integrationPreview
			}
		case "apply":
			message.log, message.err = s.Apply(ctx, plan)
		case "verify":
			message.log, message.err = s.Verify(ctx, id)
		case "open":
			message.err = integrations.OpenGuide(ctx, id)
			message.log = "Se solicitó abrir el sitio oficial en tu navegador. Volvé al Manager después de completar los pasos."
		}
		return message
	}
}
