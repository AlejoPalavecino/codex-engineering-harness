package tui

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	tea "charm.land/bubbletea/v2"
	"github.com/AlejoPalavecino/codex-engineering-harness/v2/internal/manager"
	"github.com/charmbracelet/x/ansi"
)

func sendKey(m model, s string) (model, tea.Cmd) {
	var msg tea.KeyPressMsg
	switch s {
	case "enter":
		msg = tea.KeyPressMsg{Code: tea.KeyEnter}
	case "esc":
		msg = tea.KeyPressMsg{Code: tea.KeyEscape}
	case "tab":
		msg = tea.KeyPressMsg{Code: tea.KeyTab}
	case "down":
		msg = tea.KeyPressMsg{Code: tea.KeyDown}
	case "up":
		msg = tea.KeyPressMsg{Code: tea.KeyUp}
	case "space":
		msg = tea.KeyPressMsg{Code: tea.KeySpace, Text: " "}
	case "backspace":
		msg = tea.KeyPressMsg{Code: tea.KeyBackspace}
	default:
		msg = tea.KeyPressMsg{Code: rune(s[0]), Text: s}
	}
	next, cmd := m.Update(msg)
	return next.(model), cmd
}

func TestHomeShowsReleaseBannerWhilePreservingLogo(t *testing.T) {
	m := model{
		screen:         home,
		width:          100,
		height:         32,
		managerVersion: "2.0.0",
		installed:      "2.0.0",
		mgr:            manager.Manager{Home: "/tmp/codex", Target: "codex"},
	}
	next, _ := m.Update(updatesLoadedMsg{
		report: manager.UpdateReport{
			Latest:           manager.Release{Tag: "v2.1.0", Name: "v2.1.0"},
			Summary:          "• Nuevo asistente en terminal\n• MCPs para AGY y Claude",
			ManagerNewer:     true,
			AnyProfileUpdate: true,
		},
		releases: []manager.Release{{Tag: "v2.1.0"}, {Tag: "v2.0.0"}},
	})
	m = next.(model)
	text := ansi.Strip(m.renderCRT().Content)
	for _, want := range []string{
		"▄█▀▀▀▀▀▀▀▀▀▀▀▀█▄",
		"▄█████▄",
		"Nueva versión v2.1.0 disponible",
		"[NUEVA v2.1.0]",
		"1. Instalar",
		"2. Actualizar",
		"3. Gestionar versiones",
		"4. Perfiles de AI",
		"5. MCPs",
		"6. Skills",
	} {
		if !strings.Contains(text, want) {
			t.Fatalf("missing %q in home view:\n%s", want, text)
		}
	}
}

func TestInstallWizardAnimatesCleanlyHidingTechnicalNoise(t *testing.T) {
	codexHome := canonicalTUITestHome(t)
	claudeHome := canonicalTUITestHome(t)
	agyHome := canonicalTUITestHome(t)
	t.Setenv("CODEX_HOME", codexHome)
	t.Setenv("CLAUDE_CONFIG_DIR", claudeHome)
	t.Setenv("AGY_HOME", agyHome)

	m := model{
		screen:         home,
		width:          100,
		height:         30,
		managerVersion: "2.0.0",
		mgr:            manager.Manager{Home: codexHome, Target: "codex"},
	}
	// Step 1: enter Install wizard
	m, _ = sendKey(m, "enter")
	if m.screen != installPlatformsScreen {
		t.Fatalf("expected install-platforms, got %s", m.screen)
	}
	// Toggle AGY (index 2) and open Step 2 (presets/models) with "m"
	m, _ = sendKey(m, "3")
	m, _ = sendKey(m, "m")
	if m.screen != installProfileScreen {
		t.Fatalf("expected install-profile, got %s", m.screen)
	}
	// Open customizer from wizard and accept directly with "s"
	m, _ = sendKey(m, "c")
	if m.screen != modelsScreen || !m.wizardReturn {
		t.Fatalf("expected modelsScreen with wizardReturn=true, got %s", m.screen)
	}
	// Accept customizer -> advances through remaining selected providers to confirm
	for m.screen == modelsScreen || m.screen == installProfileScreen {
		if m.screen == modelsScreen {
			m, _ = sendKey(m, "s")
		} else {
			m, _ = sendKey(m, "enter")
		}
	}
	if m.screen != installConfirmScreen {
		t.Fatalf("expected installConfirmScreen after providers, got %s", m.screen)
	}

	// Simulate running stage: spinner + stage label must hide technical script text
	m = m.startWizardInstall(false)
	next, _ := m.Update(wizardStageMsg{ok: true, event: progressEvent{step: 2, total: 4, label: "[AGY] Creando respaldo de seguridad..."}})
	m = next.(model)
	next, _ = m.Update(animTickMsg(time.Now()))
	m = next.(model)
	busyView := ansi.Strip(m.renderCRT().Content)
	if !strings.Contains(busyView, "Creando respaldo de seguridad") || strings.Contains(busyView, "CREATE") {
		t.Fatalf("unexpected busy view:\n%s", busyView)
	}

	// Simulate completion: clean success screen hides raw log unless 'v' is toggled
	next, _ = m.Update(wizardBatchResultMsg{
		title:     "¡Listo! Chuli quedó instalado",
		summary:   "Plataformas listas: AGY\nVersión aplicada: v2.1.0",
		nextSteps: "Abrí agy en tu proyecto.",
		techLog:   "CREATE skills/harness-discovery/SKILL.md\nRAW_POWERSHELL_OUTPUT",
	})
	m = next.(model)
	cleanResult := ansi.Strip(m.renderCRT().Content)
	if !strings.Contains(cleanResult, "¡Listo! Chuli quedó instalado") || strings.Contains(cleanResult, "RAW_POWERSHELL_OUTPUT") {
		t.Fatalf("clean result should hide raw tech log:\n%s", cleanResult)
	}
	m, _ = sendKey(m, "v")
	techResult := ansi.Strip(m.renderCRT().Content)
	if !strings.Contains(techResult, "RAW_POWERSHELL_OUTPUT") {
		t.Fatalf("expected raw tech log after pressing v:\n%s", techResult)
	}
}

func TestUpdateAndVersionsScreensRenderSummaryAndTags(t *testing.T) {
	m := model{
		screen:         home,
		width:          90,
		height:         28,
		managerVersion: "2.0.0",
		installed:      "2.0.0",
		mgr:            manager.Manager{Home: "/tmp/codex", Target: "codex"},
		updateState: updateState{
			updatesChecked: true,
			releases:       []manager.Release{{Tag: "v2.1.0"}, {Tag: "v2.0.0"}},
			cachedTags:     []string{"v2.0.0"},
			updateReport: manager.UpdateReport{
				Latest:           manager.Release{Tag: "v2.1.0"},
				Summary:          "• Instalación animada sin ruido\n• Gestión de Skills",
				ManagerNewer:     true,
				AnyProfileUpdate: true,
				Clients: []manager.ClientUpdateStatus{
					{ID: "codex", Name: "Codex", Installed: "2.0.0", HasInstall: true, HasUpdate: true},
				},
			},
		},
	}
	// Navigate to Actualizar (cursor 1)
	m, _ = sendKey(m, "j")
	m, _ = sendKey(m, "enter")
	if m.screen != updateScreen {
		t.Fatalf("expected updateScreen, got %s", m.screen)
	}
	updView := ansi.Strip(m.renderCRT().Content)
	for _, want := range []string{"v2.1.0", "versión nueva del binario Manager", "Instalación animada sin ruido"} {
		if !strings.Contains(updView, want) {
			t.Fatalf("missing %q in update view:\n%s", want, updView)
		}
	}
	m, _ = sendKey(m, "esc")
	// Navigate to Gestionar versiones (cursor 2)
	m, _ = sendKey(m, "j")
	m, _ = sendKey(m, "enter")
	if m.screen != versionsScreen {
		t.Fatalf("expected versionsScreen, got %s", m.screen)
	}
	verView := ansi.Strip(m.renderCRT().Content)
	if !strings.Contains(verView, "v2.0.0 (Instalada, En caché)") || !strings.Contains(verView, "R restaurar anterior") {
		t.Fatalf("unexpected versions view:\n%s", verView)
	}
}

func TestSavedProfileAndCustomMCPAndSkillsScreens(t *testing.T) {
	homeDir := canonicalTUITestHome(t)
	mgr, _ := manager.NewForClient("", "agy")
	mgr.Home = homeDir
	// Isolated profile file in temp config dir using a Manager constructed cleanly
	tmpCfg := canonicalTUITestHome(t)
	t.Setenv("APPDATA", tmpCfg)
	t.Setenv("XDG_CONFIG_HOME", tmpCfg)
	t.Setenv("AGY_HOME", homeDir)
	mgr, _ = manager.NewForClient("", "agy")

	// 1. Save, cycle and delete a custom AI profile
	m := model{screen: modelsScreen, width: 90, height: 28, mgr: mgr, modelPickerState: modelPickerState{savedProfileIdx: -1}}
	m.ensureModelState()
	m, _ = sendKey(m, "g")
	if m.screen != modelsSaveScreen {
		t.Fatalf("expected modelsSaveScreen, got %s", m.screen)
	}
	for _, ch := range []string{"m", "i", "-", "p", "e", "r", "f"} {
		m, _ = sendKey(m, ch)
	}
	m, _ = sendKey(m, "enter")
	if m.screen != modelsScreen || len(m.savedProfiles) != 1 || m.savedProfiles[0].Name != "mi-perf" {
		t.Fatalf("saved profile failed: screen=%s profiles=%+v fault=%s", m.screen, m.savedProfiles, m.fault)
	}
	m, _ = sendKey(m, "x")
	if len(m.savedProfiles) != 0 {
		t.Fatalf("expected saved profile deleted, got %+v", m.savedProfiles)
	}

	// 2. Custom MCP form prepares without writing
	m = model{screen: integrationList, width: 90, height: 28, mgr: mgr}
	m, _ = sendKey(m, "n")
	if m.screen != integrationCustomForm {
		t.Fatalf("expected integrationCustomForm, got %s", m.screen)
	}
	for _, ch := range []string{"a", "p", "i"} {
		m, _ = sendKey(m, ch)
	}
	m, _ = sendKey(m, "enter") // move to type
	m, _ = sendKey(m, "enter") // move to URL
	for _, ch := range strings.Split("https://mcp.ejemplo.test/mcp", "") {
		m, _ = sendKey(m, ch)
	}
	m, _ = sendKey(m, "enter") // move to env
	m, cmd := sendKey(m, "enter")
	if cmd == nil || m.screen != busy {
		t.Fatalf("expected custom MCP validation cmd, screen=%s", m.screen)
	}

	// 3. Skills screen toggles draft enable/disable cleanly
	skillsDir := filepath.Join(homeDir, "plugins", "chuli", "skills", "harness-discovery")
	if err := os.MkdirAll(skillsDir, 0700); err != nil {
		t.Fatal(err)
	}
	m = model{
		screen:      skillsScreen,
		width:       90,
		height:      28,
		mgr:         mgr,
		skillsState: skillsState{skillsDraftDisabled: map[string]bool{}},
	}
	next, _ := m.Update(skillsLoadedMsg{
		items: []manager.SkillInfo{{Name: "harness-discovery", Description: "Discovery R2", Source: "chuli", Enabled: true}},
	})
	m = next.(model)
	m, _ = sendKey(m, "space")
	if !m.skillsDirty || !m.skillsDraftDisabled["harness-discovery"] {
		t.Fatalf("expected harness-discovery toggled disabled in draft")
	}
	skillsView := ansi.Strip(m.renderCRT().Content)
	if !strings.Contains(skillsView, "Cambios sin aplicar") || !strings.Contains(skillsView, "Discovery R2") {
		t.Fatalf("unexpected skills view:\n%s", skillsView)
	}
}

func TestUpdateSearchReturnsFromBusyAndAllowsCancel(t *testing.T) {
	m := model{
		screen:         updateScreen,
		width:          90,
		height:         28,
		managerVersion: "2.0.0",
		mgr:            manager.Manager{Home: "/tmp/codex", Target: "codex"},
		updateState:    updateState{updatesChecked: true},
	}
	// Pressing R enters cancelable busy screen and returns to updateScreen on updatesLoadedMsg
	m, cmd := sendKey(m, "r")
	if m.screen != busy || cmd == nil || !m.busyCancelable || m.busyReturnScreen != updateScreen {
		t.Fatalf("expected cancelable busy returning to updateScreen, got screen=%s cancelable=%v return=%s", m.screen, m.busyCancelable, m.busyReturnScreen)
	}
	// Pressing Esc during read-only search cancels back to updateScreen instead of hanging
	canceled, _ := sendKey(m, "esc")
	if canceled.screen != updateScreen {
		t.Fatalf("expected esc during search to return to updateScreen, got %s", canceled.screen)
	}
	// Receiving updatesLoadedMsg (even with an error) must leave busy and show updateScreen
	next, _ := m.Update(updatesLoadedMsg{err: os.ErrNotExist})
	m = next.(model)
	if m.screen != updateScreen || m.updateCheckErr == "" {
		t.Fatalf("expected return to updateScreen with error, got screen=%s err=%q", m.screen, m.updateCheckErr)
	}
}

func TestWizardSequentialMultiProviderWithArrowsAndEnter(t *testing.T) {
	codexHome := canonicalTUITestHome(t)
	missingRoot := canonicalTUITestHome(t)
	claudeHome := filepath.Join(missingRoot, "no-claude")
	agyHome := filepath.Join(missingRoot, "no-agy")
	t.Setenv("PATH", "")
	t.Setenv("CODEX_HOME", codexHome)
	t.Setenv("CLAUDE_CONFIG_DIR", claudeHome)
	t.Setenv("AGY_HOME", agyHome)

	m := model{
		screen:         home,
		width:          100,
		height:         30,
		managerVersion: "2.0.0",
		mgr:            manager.Manager{Home: codexHome, Target: "codex"},
	}
	// Enter Paso 1: platforms
	m, _ = sendKey(m, "enter")
	if m.screen != installPlatformsScreen {
		t.Fatalf("expected installPlatformsScreen, got %s", m.screen)
	}
	// Move down to AGY (index 2) and toggle with Space
	m, _ = sendKey(m, "down")
	m, _ = sendKey(m, "down")
	m, _ = sendKey(m, "space")
	if !m.wizardSelected["agy"] || !m.wizardSelected["codex"] {
		t.Fatalf("expected codex and agy selected, got %+v", m.wizardSelected)
	}
	// Move down to "Personalizar modelos" (index 4) and press Enter -> Paso 2, provider 1 of 2 (Codex)
	m, _ = sendKey(m, "down")
	m, _ = sendKey(m, "down")
	m, _ = sendKey(m, "enter")
	if m.screen != installProfileScreen || m.wizardActiveID != "codex" || m.wizardProviderIdx != 0 {
		t.Fatalf("expected Paso 2 for codex (1/2), got screen=%s active=%s idx=%d", m.screen, m.wizardActiveID, m.wizardProviderIdx)
	}
	step2View := ansi.Strip(m.renderCRT().Content)
	if !strings.Contains(step2View, "1 de 2") || !strings.Contains(step2View, "Paso 2 de 4") {
		t.Fatalf("missing step/provider indicator in Paso 2:\n%s", step2View)
	}

	// Move down with arrows to "Personalizar" (last option = index 3) and enter with Enter
	m, _ = sendKey(m, "down")
	m, _ = sendKey(m, "down")
	m, _ = sendKey(m, "down")
	m, _ = sendKey(m, "enter")
	if m.screen != modelsScreen || !m.wizardReturn {
		t.Fatalf("expected modelsScreen from Personalizar, got %s", m.screen)
	}
	// Open role picker for orchestrator with Enter, pick 2nd model with Down+Enter
	m, _ = sendKey(m, "enter")
	if m.screen != rolePickerScreen {
		t.Fatalf("expected rolePickerScreen, got %s", m.screen)
	}
	m, _ = sendKey(m, "down")
	m, _ = sendKey(m, "enter")
	if m.screen != modelsScreen || m.mgr.RoleModels["orchestrator"].Model == "gpt-6.2-astra" {
		t.Fatalf("expected updated orchestrator model after picker, got screen=%s model=%s", m.screen, m.mgr.RoleModels["orchestrator"].Model)
	}
	// Navigate down to "Aceptar y continuar" (index 7) and press Enter -> moves to provider 2 of 2 (AGY)!
	for range 7 {
		m, _ = sendKey(m, "down")
	}
	m, _ = sendKey(m, "enter")
	if m.screen != installProfileScreen || m.wizardActiveID != "agy" || m.wizardProviderIdx != 1 {
		t.Fatalf("expected Paso 2 to advance sequentially to AGY (2/2), got screen=%s active=%s idx=%d", m.screen, m.wizardActiveID, m.wizardProviderIdx)
	}
	agyView := ansi.Strip(m.renderCRT().Content)
	if !strings.Contains(agyView, "2 de 2") || !strings.Contains(agyView, "AGY") {
		t.Fatalf("missing 2 de 2 AGY indicator:\n%s", agyView)
	}
	// Moving down previews without overwriting; Enter applies AGY preset and moves to Paso 3 (confirm)
	m, _ = sendKey(m, "down")
	if m.wizardConfigs["agy"].Preset != "agy-balanced" {
		t.Fatalf("moving cursor must not overwrite preset before selection, got %s", m.wizardConfigs["agy"].Preset)
	}
	m, _ = sendKey(m, "enter")
	if m.wizardConfigs["agy"].Preset != "agy-carbon-max" {
		t.Fatalf("expected agy-carbon-max applied on enter, got %s", m.wizardConfigs["agy"].Preset)
	}
	if m.screen != installConfirmScreen {
		t.Fatalf("expected Paso 3 installConfirmScreen after finishing 2/2 providers, got %s", m.screen)
	}
}
