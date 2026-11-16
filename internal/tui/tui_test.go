package tui

import (
	tea "charm.land/bubbletea/v2"
	"github.com/AlejoPalavecino/codex-engineering-harness/v2/internal/manager"
	"strings"
	"testing"
)

func TestDefaultMenuAndExplicitLocalSourceInWizard(t *testing.T) {
	items := choices(manager.Manager{})
	want := []string{"install", "update", "versions", "models", "integrations", "skills"}
	if len(items) != len(want) {
		t.Fatalf("expected %d main menu buttons, got %d", len(want), len(items))
	}
	for i, w := range want {
		if items[i].action != w {
			t.Fatalf("menu[%d] = %q, want %q", i, items[i].action, w)
		}
	}
	for _, s := range []string{"", "/trusted"} {
		m := model{mgr: manager.Manager{Home: canonicalTUITestHome(t), Source: s}}
		m.initWizard()
		if m.wizardUseLocal != (s != "") {
			t.Fatalf("wizardUseLocal=%v for Source=%q", m.wizardUseLocal, s)
		}
	}
}
func TestPreviewConfirmationAndCancellation(t *testing.T) {
	base := model{screen: preview, height: 15, log: "one\ntwo\nthree\nfour"}
	if !strings.Contains(base.page(), "one") || strings.Contains(base.page(), "four") {
		t.Fatal(base.page())
	}
	next, _ := base.Update(tea.KeyPressMsg{Code: 'j', Text: "j"})
	if !strings.Contains(next.(model).page(), "four") {
		t.Fatal("inaccessible preview line")
	}
	next, cmd := base.Update(tea.KeyPressMsg{Code: 'n', Text: "n"})
	if next.(model).screen != home || cmd != nil {
		t.Fatal("cancel invoked an action")
	}
	next, cmd = base.Update(tea.KeyPressMsg{Code: 's', Text: "s"})
	if next.(model).screen != busy || cmd == nil {
		t.Fatal("no explicit confirmation")
	}
}

func TestGentleConflictOffersAnotherPreview(t *testing.T) {
	m, cmd := (model{action: "latest", screen: busy}).Update(preparedMsg{err: manager.ErrGentleConflict})
	if m.(model).screen != conflict || cmd != nil {
		t.Fatal("conflict should await user's choice")
	}
	m, cmd = m.(model).Update(tea.KeyPressMsg{Code: 's', Text: "s"})
	if m.(model).screen != conflict || cmd != nil {
		t.Fatal("conflict screen must not apply")
	}
	m, cmd = m.(model).Update(tea.KeyPressMsg{Code: 'a', Text: "a"})
	if m.(model).screen != busy || cmd == nil {
		t.Fatal("explicit adoption must request a new preflight")
	}
	m, cmd = m.(model).Update(preparedMsg{plan: manager.Prepared{Preview: "nuevo plan"}})
	if m.(model).screen != preview || m.(model).log != "nuevo plan" || cmd != nil {
		t.Fatal("adoption must await final confirmation")
	}
}
