package tui

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/AlejoPalavecino/codex-engineering-harness/v2/internal/manager"
	"github.com/charmbracelet/x/ansi"
)

func TestHomeNumericShortcutsAndTabClientSwitch(t *testing.T) {
	codex := canonicalTUITestHome(t)
	claude := canonicalTUITestHome(t)
	agy := canonicalTUITestHome(t)
	t.Setenv("CODEX_HOME", codex)
	t.Setenv("CLAUDE_CONFIG_DIR", claude)
	t.Setenv("AGY_HOME", agy)

	base := model{
		screen:         home,
		width:          100,
		height:         30,
		managerVersion: "2.1.5",
		mgr:            manager.Manager{Home: codex, Target: "codex"},
	}

	view := ansi.Strip(base.renderCRT().Content)
	if !strings.Contains(view, "[● Codex]") || strings.Contains(view, "1:Codex") {
		t.Fatalf("expected client tabs without numeric prefix:\n%s", view)
	}
	if !strings.Contains(view, "Primera vez:") {
		t.Fatalf("expected onboarding hint when no client is installed:\n%s", view)
	}

	// Tab rotates client without activating menu
	m, _ := sendKey(base, "tab")
	if m.mgr.Target != "claude" || m.screen != home {
		t.Fatalf("expected Tab to switch to claude on home, got target=%s screen=%s", m.mgr.Target, m.screen)
	}
	// Shift+Tab rotates back
	next, _ := m.Update(tea.KeyPressMsg{Code: tea.KeyTab, Mod: tea.ModShift})
	m = next.(model)
	if m.mgr.Target != "codex" {
		t.Fatalf("expected Shift+Tab to return to codex, got %s", m.mgr.Target)
	}

	// Keys 1..6 open corresponding menu items
	cases := []struct {
		key        string
		wantScreen string
	}{
		{"1", installPlatformsScreen},
		{"4", modelsScreen},
		{"5", integrationList},
		{"6", skillsScreen},
	}
	for _, tc := range cases {
		opened, _ := sendKey(base, tc.key)
		if opened.screen != tc.wantScreen {
			t.Fatalf("key %s opened %s, want %s", tc.key, opened.screen, tc.wantScreen)
		}
	}
}

func TestFastPathThreeEntersAndEscReturnsToStep1(t *testing.T) {
	codex := canonicalTUITestHome(t)
	missing := canonicalTUITestHome(t)
	t.Setenv("PATH", "")
	t.Setenv("CODEX_HOME", codex)
	t.Setenv("CLAUDE_CONFIG_DIR", filepath.Join(missing, "no-claude"))
	t.Setenv("AGY_HOME", filepath.Join(missing, "no-agy"))

	m := model{
		screen:         home,
		width:          100,
		height:         30,
		managerVersion: "2.1.5",
		mgr:            manager.Manager{Home: codex, Target: "codex"},
	}

	// Enter 1: home -> Paso 1
	m, _ = sendKey(m, "enter")
	if m.screen != installPlatformsScreen {
		t.Fatalf("expected installPlatformsScreen, got %s", m.screen)
	}
	// Enter 2: Paso 1 -> Paso 3 directly (fast path)
	m, preflightCmd := sendKey(m, "enter")
	if m.screen != installConfirmScreen || !m.wizardCameFromFastPath || preflightCmd == nil {
		t.Fatalf("expected fast path to Paso 3 with preflight cmd, got screen=%s fast=%v cmd=%v", m.screen, m.wizardCameFromFastPath, preflightCmd)
	}
	view := ansi.Strip(m.renderCRT().Content)
	if !strings.Contains(view, "Equilibrado (Astra 6.2 + Sol 6.2)") || !strings.Contains(view, "Paso 3/4: Confirmar") {
		t.Fatalf("expected friendly preset name and breadcrumb in Paso 3:\n%s", view)
	}
	// Esc from fast-path Paso 3 returns to Paso 1 (not skipped Paso 2)
	m, _ = sendKey(m, "esc")
	if m.screen != installPlatformsScreen {
		t.Fatalf("expected Esc from fast-path Paso 3 to return to Paso 1, got %s", m.screen)
	}
	// Go forward to Paso 3 again and Enter 3 -> Paso 4 (busy)
	m, _ = sendKey(m, "enter")
	m, installCmd := sendKey(m, "enter")
	if m.screen != busy || installCmd == nil {
		t.Fatalf("expected Enter 3 to start install, got screen=%s", m.screen)
	}
}

func TestQDoesNotQuitInsideSubflowsButQuitsOnHomeAndSummaryResult(t *testing.T) {
	codex := canonicalTUITestHome(t)
	subflows := []model{
		{screen: installPlatformsScreen, mgr: manager.Manager{Home: codex, Target: "codex"}},
		{screen: installProfileScreen, mgr: manager.Manager{Home: codex, Target: "codex"}, wizardState: wizardState{wizardActiveID: "codex", wizardSelected: map[string]bool{"codex": true}}},
		{screen: installConfirmScreen, mgr: manager.Manager{Home: codex, Target: "codex"}, wizardState: wizardState{wizardActiveID: "codex", wizardSelected: map[string]bool{"codex": true}}},
		{screen: modelsScreen, mgr: manager.Manager{Home: codex, Target: "codex"}},
		{screen: skillsScreen, mgr: manager.Manager{Home: codex, Target: "codex"}},
		{screen: versionsScreen, mgr: manager.Manager{Home: codex, Target: "codex"}},
		{screen: integrationList, mgr: manager.Manager{Home: codex, Target: "codex"}},
		{screen: preview, mgr: manager.Manager{Home: codex, Target: "codex"}, log: "plan"},
		{screen: result, mgr: manager.Manager{Home: codex, Target: "codex"}, resultSummary: "ok", showTechLog: true, log: "raw"},
	}
	for _, m := range subflows {
		next, cmd := sendKey(m, "q")
		if cmd != nil || next.screen != m.screen {
			t.Fatalf("q must not quit or exit subflow %s (cmd=%v, next=%s)", m.screen, cmd, next.screen)
		}
	}

	for _, screen := range []string{home, about, result, integrationResult} {
		m := model{screen: screen, mgr: manager.Manager{Home: codex, Target: "codex"}}
		if screen == result {
			m.resultSummary = "Listo"
		}
		_, cmd := sendKey(m, "q")
		if cmd == nil {
			t.Fatalf("q must quit on %s", screen)
		}
		if _, ok := cmd().(tea.QuitMsg); !ok {
			t.Fatalf("q on %s did not return tea.QuitMsg", screen)
		}
	}
}

func TestStep2PreservesCustomOnCursorMoveAndShowsSavedProfiles(t *testing.T) {
	cfgDir := canonicalTUITestHome(t)
	codex := canonicalTUITestHome(t)
	t.Setenv("APPDATA", cfgDir)
	t.Setenv("XDG_CONFIG_HOME", cfgDir)
	t.Setenv("CODEX_HOME", codex)
	t.Setenv("CLAUDE_CONFIG_DIR", canonicalTUITestHome(t))

	mgr, err := manager.NewForClient("", "codex")
	if err != nil {
		t.Fatal(err)
	}
	_, customRoles, _ := manager.DefaultRolesForPreset("codex", "codex-balanced")
	customRoles["orchestrator"] = manager.RoleModel{Model: "gpt-6.1-sol", Effort: "low"}
	if err := mgr.SaveProfile("codex", "Diario", customRoles); err != nil {
		t.Fatal(err)
	}

	m := model{
		screen: installProfileScreen,
		width:  100,
		height: 30,
		mgr:    mgr,
		wizardState: wizardState{
			wizardSelected:  map[string]bool{"codex": true, "claude": true},
			wizardActiveID:  "codex",
			wizardSavedName: map[string]string{},
			wizardConfigs: map[string]wizardClientConfig{
				"codex":  {Preset: "custom", Roles: customRoles},
				"claude": {Preset: "claude-balanced"},
			},
		},
	}
	m.syncWizardProfileCursor()

	// Moving cursor down over presets must NOT overwrite custom config
	m, _ = sendKey(m, "down")
	m, _ = sendKey(m, "down")
	if m.wizardConfigs["codex"].Preset != "custom" || m.wizardConfigs["codex"].Roles["orchestrator"].Model != "gpt-6.1-sol" {
		t.Fatalf("cursor navigation overwrote custom config: %+v", m.wizardConfigs["codex"])
	}

	view := ansi.Strip(m.renderCRT().Content)
	if !strings.Contains(view, "★ Diario") || !strings.Contains(view, "Equipo personalizado actual") {
		t.Fatalf("expected saved profile and keep-custom option in Step 2:\n%s", view)
	}

	// Tab switches to claude without losing codex custom config
	m, _ = sendKey(m, "tab")
	if m.wizardActiveID != "claude" {
		t.Fatalf("expected Tab to switch to claude, got %s", m.wizardActiveID)
	}
	if m.wizardConfigs["codex"].Preset != "custom" {
		t.Fatalf("switching tabs lost codex config")
	}
}

func TestStep3GentlePreflightCheckboxAndActionableResult(t *testing.T) {
	codex := canonicalTUITestHome(t)
	t.Setenv("CODEX_HOME", codex)
	if err := os.WriteFile(filepath.Join(codex, "AGENTS.md"), []byte("<!-- gentle-ai:orchestrator -->\nold\n<!-- /gentle-ai:orchestrator -->\n"), 0600); err != nil {
		t.Fatal(err)
	}
	mgr := manager.Manager{Home: codex, Target: "codex"}
	m := model{
		screen: installPlatformsScreen,
		width:  100,
		height: 30,
		mgr:    mgr,
	}
	m.initWizard()
	m.wizardSelected = map[string]bool{"codex": true}
	m, preflightCmd := m.enterInstallConfirm(true)
	next, _ := m.Update(preflightCmd())
	m = next.(model)

	view := ansi.Strip(m.renderCRT().Content)
	if !strings.Contains(view, "Respaldar Gentle AI") {
		t.Fatalf("expected Gentle preflight option in Paso 3:\n%s", view)
	}
	m, _ = sendKey(m, "a")
	if !m.wizardAdoptGentle["codex"] {
		t.Fatal("expected Gentle adoption flag toggled with 'a'")
	}

	// Simulate completion and navigate result actions to Skills
	next, _ = m.Update(wizardBatchResultMsg{
		title:     "¡Listo! Chuli quedó instalado",
		summary:   "Plataformas listas: Codex",
		nextSteps: "Abrí codex.",
		techLog:   "ok",
	})
	m = next.(model)
	m.resultFromWizard = true
	resView := ansi.Strip(m.renderCRT().Content)
	if !strings.Contains(resView, "Configurar MCPs opcionales") || !strings.Contains(resView, "Revisar Skills instaladas") {
		t.Fatalf("expected actionable next steps on result screen:\n%s", resView)
	}
	m, _ = sendKey(m, "down") // MCPs
	m, _ = sendKey(m, "down") // Skills
	m, cmd := sendKey(m, "enter")
	if m.screen != skillsScreen || cmd == nil {
		t.Fatalf("expected result action to open skillsScreen, got screen=%s cmd=%v", m.screen, cmd)
	}
}

func TestRenderKeyBarPreservesPlainText(t *testing.T) {
	inputs := []string{
		"↑↓ / 1-6 navegar · Enter elegir · Tab cliente · D verificar · ? ayuda · q salir",
		"S / Y confirmar · N cancelar · Esc volver · q salir\n↑↓ / j k / PgUp / PgDn desplazar",
		"O abrir sitio oficial\nEnter revisar ficha\nEsc volver",
	}
	for _, in := range inputs {
		if got := ansi.Strip(renderKeyBar(in)); got != in {
			t.Fatalf("renderKeyBar changed stripped text:\ngot:  %q\nwant: %q", got, in)
		}
	}
}
