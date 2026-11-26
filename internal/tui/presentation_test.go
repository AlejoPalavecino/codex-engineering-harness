package tui

import (
	"fmt"
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/AlejoPalavecino/codex-engineering-harness/v2/internal/manager"
	"github.com/charmbracelet/x/ansi"
)

func TestCRTResponsiveHome(t *testing.T) {
	for _, width := range []int{32, 60, 99, 100, 140} {
		t.Run(fmt.Sprint(width), func(t *testing.T) {
			m := model{screen: home, width: width, managerVersion: "1.2.3", installed: "v2.0.0", mgr: manager.Manager{Home: "/temporary/codex"}}
			v := m.renderCRT()
			text := ansi.Strip(v.Content)
			for _, want := range []string{"CHULI", "Manager: 1.2.3", "Perfil: v2.0.0", "Enter elegir", "Instalar", "Codex"} {
				if !strings.Contains(text, want) {
					t.Fatalf("missing %q in\n%s", want, text)
				}
			}
			if width >= 60 {
				for _, art := range []string{"▄█████▄", "▄█▀▀▀▀▀▀▀▀▀▀▀▀█▄", "▄▄▄▄▄▄▄████▄▄▄▄▄▄▄"} {
					if !strings.Contains(text, art) {
						t.Fatalf("missing pixel logo token %q at width %d:\n%s", art, width, text)
					}
				}
			}
			for _, line := range strings.Split(text, "\n") {
				if lipgloss.Width(line) > width {
					t.Fatalf("line exceeds %d columns: %q", width, line)
				}
			}
			if !v.AltScreen || v.WindowTitle != "Chuli Manager" {
				t.Fatal("terminal view metadata lost")
			}
			selected := strings.Index(text, "›")
			next, cmd := m.Update(tea.KeyPressMsg{Code: 'j', Text: "j"})
			if cmd != nil || strings.Index(ansi.Strip(next.(model).renderCRT().Content), "›") <= selected {
				t.Fatal("selection marker did not move")
			}
			if width >= 100 {
				found := false
				for _, line := range strings.Split(text, "\n") {
					if strings.Contains(line, "›") && strings.Contains(line, "Elegí AGY, Claude o Codex") {
						found = true
					}
				}
				if !found {
					t.Fatal("wide explanation is not beside menu")
				}
			}
		})
	}
}

func TestCRTStatesAndScrollableLogs(t *testing.T) {
	states := map[string]string{busy: "EN CURSO", preview: "REVISIÓN PREVIA", conflict: "CONFLICTO", result: "COMPLETADO", about: "Acerca de Chuli", integrationList: "Integraciones", integrationDetail: "V comprobar", integrationPreview: "S / Y confirmar", integrationResult: "COMPLETADO"}
	for screen, want := range states {
		m := model{screen: screen, width: 60, height: 24, integrationState: integrationState{integrationID: "exa"}, mgr: manager.Manager{Home: "/temporary/codex"}, log: "cambios revisables"}
		text := ansi.Strip(m.renderCRT().Content)
		if !strings.Contains(text, want) {
			t.Errorf("%s missing %q: %s", screen, want, text)
		}
		if screen == preview || screen == integrationPreview {
			for _, label := range []string{"Destino:", "Cambios:", "N cancelar"} {
				if !strings.Contains(text, label) {
					t.Errorf("%s missing %s", screen, label)
				}
			}
		}
	}
	for _, screen := range []string{preview, conflict, result, integrationDetail, integrationPreview, integrationResult} {
		m := model{screen: screen, width: 60, height: 24, integrationState: integrationState{integrationID: "exa"}, log: strings.Repeat("registro\n", 40) + "ULTIMA LINEA"}
		for range 50 {
			next, _ := m.Update(tea.KeyPressMsg{Code: 'j', Text: "j"})
			m = next.(model)
		}
		if !strings.Contains(m.renderCRT().Content, "ULTIMA LINEA") {
			t.Errorf("%s lost final log line", screen)
		}
	}
}

func TestCRTAllScreensFitTerminalWidth(t *testing.T) {
	screens := []string{
		home, busy, preview, conflict, result, about,
		integrationList, integrationDetail, integrationPreview, integrationResult, integrationCustomForm,
		installPlatformsScreen, installProfileScreen, installConfirmScreen, installGentleScreen,
		updateScreen, versionsScreen, modelsScreen, modelsSaveScreen, rolePickerScreen, skillsScreen, skillsAddScreen,
	}
	for _, size := range [][2]int{{40, 24}, {80, 24}, {120, 32}} {
		for _, screen := range screens {
			m := model{
				screen:           screen,
				width:            size[0],
				height:           size[1],
				integrationState: integrationState{integrationID: "github"},
				wizardState: wizardState{
					wizardActiveID: "codex",
					wizardSelected: map[string]bool{"codex": true},
					wizardConfigs:  map[string]wizardClientConfig{"codex": {Preset: "codex-balanced"}},
				},
				updateState: updateState{
					updatesChecked: true,
					updateReport:   manager.UpdateReport{Latest: manager.Release{Tag: "v2.1.0"}, Summary: "• Mejoras generales"},
				},
				skillsState: skillsState{skillsDraftDisabled: map[string]bool{}},
				title:       "Comprobando integración",
				log:         strings.Repeat("cambio revisable ", 15),
				mgr:         manager.Manager{Home: "/temporary/" + strings.Repeat("destination/", 12)},
			}
			text := ansi.Strip(m.renderCRT().Content)
			for _, line := range strings.Split(text, "\n") {
				if lipgloss.Width(line) > size[0] {
					t.Errorf("%s at %dx%d: line exceeds width: %q", screen, size[0], size[1], line)
				}
			}
			if !strings.Contains(text, "Carpeta Codex:") || !strings.Contains(text, "destination/") {
				t.Errorf("%s lost destination:\n%s", screen, text)
			}
		}
	}
}

func TestCRTRejectsTerminalControls(t *testing.T) {
	attack := "visible\x1b[2J\x1b]52;c;c2VjcmV0\a\rFORGED\u009b2J"
	m := model{screen: result, width: 60, log: attack, fault: attack, title: attack, installed: attack, managerVersion: attack, mgr: manager.Manager{Home: attack}}
	text := ansi.Strip(m.renderCRT().Content)
	if strings.ContainsAny(text, "\x1b\a\r\u009b") || strings.Contains(text, "52;c;") {
		t.Fatalf("terminal control escaped sanitizer: %q", text)
	}
	if !strings.Contains(text, "ERROR") || !strings.Contains(text, "visible") {
		t.Fatal("error details lost")
	}
}

func TestCRTHeightKeepsControlsVisible(t *testing.T) {
	screens := []string{
		home, modelsScreen, modelsSaveScreen, rolePickerScreen, busy, preview, conflict, result, about,
		integrationList, integrationPreview, integrationDetail, integrationResult, integrationCustomForm,
		installPlatformsScreen, installProfileScreen, installConfirmScreen, installGentleScreen,
		updateScreen, versionsScreen, skillsScreen, skillsAddScreen,
	}
	for _, size := range [][2]int{{40, 24}, {80, 24}, {120, 32}} {
		for _, screen := range screens {
			m := model{
				screen:           screen,
				width:            size[0],
				height:           size[1],
				integrationState: integrationState{integrationID: "exa"},
				wizardState: wizardState{
					wizardActiveID: "codex",
					wizardSelected: map[string]bool{"codex": true},
					wizardConfigs:  map[string]wizardClientConfig{"codex": {Preset: "codex-balanced"}},
				},
				updateState: updateState{
					updatesChecked: true,
					updateReport:   manager.UpdateReport{Latest: manager.Release{Tag: "v2.1.0"}, Summary: strings.Repeat("• Novedad importante\n", 30)},
				},
				skillsState: skillsState{skillsDraftDisabled: map[string]bool{}},
				mgr:         manager.Manager{Home: "/temporary/" + strings.Repeat("destination/", 12)},
				log:         strings.Repeat("cambio revisable largo ", 100),
				fault:       "un error recuperable",
			}
			content := m.renderCRT().Content
			if rows := lipgloss.Height(content); rows > size[1] {
				t.Errorf("%s %dx%d used %d rows", screen, size[0], size[1], rows)
			}
		}
	}
}

func TestModelsConfiguratorAndClientSwitch(t *testing.T) {
	m := model{screen: home, width: 100, height: 28, mgr: manager.Manager{Home: "/temporary/codex", Target: "codex"}}
	next, _ := m.Update(tea.KeyPressMsg{Code: 'm', Text: "m"})
	m = next.(model)
	if m.screen != modelsScreen {
		t.Fatalf("expected models screen, got %s", m.screen)
	}
	next, _ = m.Update(tea.KeyPressMsg{Code: '2', Text: "2"})
	m = next.(model)
	if m.mgr.Target != "claude" || m.mgr.Preset != "claude-balanced" {
		t.Fatalf("expected claude-balanced, got %s / %s", m.mgr.Target, m.mgr.Preset)
	}
	if m.mgr.RoleModels["solution_architect"].Model != "opus" {
		t.Fatalf("unexpected architect model: %+v", m.mgr.RoleModels["solution_architect"])
	}
	next, _ = m.Update(tea.KeyPressMsg{Code: 'p', Text: "p"})
	m = next.(model)
	if m.mgr.Preset != "claude-fable-max" || m.mgr.RoleModels["orchestrator"].Model != "claude-fable-5-2" {
		t.Fatalf("preset cycle failed: %s / %+v", m.mgr.Preset, m.mgr.RoleModels["orchestrator"])
	}
	next, _ = m.Update(tea.KeyPressMsg{Code: 'l', Text: "l"})
	m = next.(model)
	if m.mgr.Preset != "custom" || m.mgr.RoleModels["orchestrator"].Model != "claude-opus-5-5" {
		t.Fatalf("role model cycle failed: %s / %+v", m.mgr.Preset, m.mgr.RoleModels["orchestrator"])
	}
	next, _ = m.Update(tea.KeyPressMsg{Code: 'e', Text: "e"})
	m = next.(model)
	if m.mgr.RoleModels["orchestrator"].Effort != "low" {
		t.Fatalf("role effort cycle failed: %+v", m.mgr.RoleModels["orchestrator"])
	}
	next, _ = m.Update(tea.KeyPressMsg{Code: '3', Text: "3"})
	m = next.(model)
	if m.mgr.Target != "agy" || m.mgr.Preset != "agy-balanced" || m.mgr.RoleModels["orchestrator"].Model != "orchestrator-argon" {
		t.Fatalf("expected agy-balanced, got %s / %s", m.mgr.Target, m.mgr.Preset)
	}
	text := ansi.Strip(m.renderCRT().Content)
	for _, want := range []string{"Modelos y Presets", "[● AGY]", "agy-balanced", "orchestrator-argon"} {
		if !strings.Contains(text, want) {
			t.Fatalf("missing %q in models screen:\n%s", want, text)
		}
	}
}
