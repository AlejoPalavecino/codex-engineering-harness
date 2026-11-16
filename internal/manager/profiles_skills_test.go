package manager

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestSavedProfilesCRUDAndValidation(t *testing.T) {
	m := Manager{cacheDir: filepath.Join(t.TempDir(), "chuli", "profiles")}
	_, roles, err := DefaultRolesForPreset("agy", "agy-balanced")
	if err != nil {
		t.Fatal(err)
	}
	roles["chuli_worker"] = RoleModel{Model: "gemini-4-argon", Effort: "high"}
	if err := m.SaveProfile("agy", "Trabajo Diario", roles); err != nil {
		t.Fatal(err)
	}
	list, err := m.ListSavedProfiles("agy")
	if err != nil || len(list) != 1 || list[0].Name != "Trabajo Diario" {
		t.Fatalf("unexpected profiles: %+v, err=%v", list, err)
	}
	if list[0].RoleModels["chuli_worker"].Model != "gemini-4-argon" {
		t.Fatalf("unexpected worker model: %+v", list[0].RoleModels["chuli_worker"])
	}
	if err := m.SaveProfile("agy", "../bad", roles); err == nil {
		t.Fatal("accepted unsafe profile name")
	}
	if err := m.DeleteSavedProfile("agy", "Trabajo Diario"); err != nil {
		t.Fatal(err)
	}
	list, err = m.ListSavedProfiles("agy")
	if err != nil || len(list) != 0 {
		t.Fatalf("expected empty list after delete: %+v, %v", list, err)
	}
}

func TestSkillsListDisableInInstallVerifyRollbackAndCustom(t *testing.T) {
	if _, err := shell(); err != nil {
		t.Skip("requires PowerShell and Python 3.11+")
	}
	repo, err := filepath.Abs("../..")
	if err != nil {
		t.Fatal(err)
	}
	homeDir, err := filepath.EvalSymlinks(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	m := Manager{Home: homeDir, Source: repo, Target: "agy", cacheDir: filepath.Join(t.TempDir(), "chuli", "profiles")}

	// Initial install with all skills enabled.
	p1, err := m.PrepareWithOptions(t.Context(), PrepareOptions{Action: "local"})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := m.Apply(t.Context(), p1); err != nil {
		p1.Cleanup()
		t.Fatal(err)
	}
	p1.Cleanup()

	// Add custom skill and verify it appears alongside catalog skills.
	customSrc := filepath.Join(t.TempDir(), "mi-skill-propia")
	if err := os.MkdirAll(customSrc, 0700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(customSrc, "SKILL.md"), []byte("---\nname: mi-skill-propia\ndescription: Skill de prueba\n---\n\n# Prueba\n"), 0600); err != nil {
		t.Fatal(err)
	}
	added, err := m.AddCustomSkill(customSrc)
	if err != nil || added.Source != "custom" {
		t.Fatalf("AddCustomSkill = %+v, err=%v", added, err)
	}

	// Disable gentle-ai-bench and apply locally; verify.ps1 must pass.
	m.DisabledSkills = []string{"gentle-ai-bench"}
	p2, err := m.PrepareWithOptions(t.Context(), PrepareOptions{Action: "apply-local"})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := m.Apply(t.Context(), p2); err != nil {
		p2.Cleanup()
		t.Fatal(err)
	}
	p2.Cleanup()

	if _, err := os.Stat(filepath.Join(homeDir, "plugins", "chuli", "skills", "gentle-ai-bench", "SKILL.md")); !os.IsNotExist(err) {
		t.Fatal("disabled skill was not removed from active plugin dir")
	}
	if log, err := m.Doctor(t.Context()); err != nil {
		t.Fatalf("doctor failed with disabled skill: %v\n%s", err, log)
	}

	skills, err := m.ListSkills()
	if err != nil {
		t.Fatal(err)
	}
	foundDisabled, foundCustom := false, false
	for _, s := range skills {
		if s.Name == "gentle-ai-bench" {
			foundDisabled = true
			if s.Enabled || s.Source != "chuli" || !strings.Contains(strings.ToLower(s.Description), "bench") {
				t.Fatalf("unexpected disabled skill info: %+v", s)
			}
		}
		if s.Name == "mi-skill-propia" {
			foundCustom = true
			if s.Source != "custom" || !s.Enabled {
				t.Fatalf("unexpected custom skill info: %+v", s)
			}
		}
	}
	if !foundDisabled || !foundCustom {
		t.Fatalf("missing skills in list: disabled=%v custom=%v", foundDisabled, foundCustom)
	}

	if err := m.RemoveCustomSkill("gentle-ai-bench"); err == nil {
		t.Fatal("must refuse deleting catalog skill directly")
	}
	if err := m.RemoveCustomSkill("mi-skill-propia"); err != nil {
		t.Fatal(err)
	}
}
