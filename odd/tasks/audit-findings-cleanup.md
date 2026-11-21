# Feature: Resolución de hallazgos técnicos de auditoría (v2.1.2)

- **Spec:** #especificacion
- **Design:** #diseno-tecnico
- **Approval:** approved
- **Evidence mode:** candidate-v1
- **Trace mode:** requirements-v1

## Especificacion

- **REQ-001:** `config/models.json` es la única fuente de verdad del catálogo de roles, clientes, modelos, esfuerzos y presets para Go (`internal/manager`) y Python (`scripts/client_targets.py`), sin duplicación manual y sin incluir archivos `.go` dentro del paquete de release `chuli-profile-v*.zip`.
- **REQ-002:** Los scripts de instalación, verificación y rollback con `pwsh` + Python se ejecutan en CI también sobre Linux y macOS (`ubuntu-latest`, `macos-latest`), preservando la compatibilidad POSIX (incluidos los alias del sistema `/tmp` y `/var` en macOS).
- **REQ-003:** `docs/ROADMAP.md` refleja con precisión la trayectoria real del proyecto hasta `v2.1.2` (multi-cliente Codex/Claude/AGY, consola CRT, integraciones MCP, gestión de skills y cierre del roadmap de ciclo de vida `#63`).
- **REQ-004:** `internal/tui` separa su estado por dominio y divide `wizard.go` por flujo sin cambiar atajos ni comportamiento observable; `internal/manager` permite un ciclo rápido con `go test -short ./...` omitiendo solo las pruebas pesadas de PowerShell mientras `go test ./...` mantiene la corrida completa.
- **REQ-005:** `scripts/verify_release.py` conserva intacto el pin upstream de `skills/diagram-design` y valida presupuestos de tamaño y cantidad de entradas del paquete `chuli-profile-v*.zip` alineados con los límites de `internal/manager/archive.go` (`maxArchive`, `maxExpanded`, `maxEntry`, `maxEntries`).

## Diseno tecnico

- **Catálogo único (`REQ-001`):** Exponer `config/models.json` embebido en Go mediante `config/models_embed.go` (`package config`, `//go:embed models.json`), excluir `*.go` del empaquetado de `config/` en `scripts/package_release.py`, y parsear el JSON en `internal/manager/catalog.go` preservando el orden de clientes y presets. Agregar test unitario en Go que verifica la carga y la paridad exacta contra `../../config/models.json`.
- **Cobertura cross-platform (`REQ-002`):** Añadir la excepción segura de `/tmp` y `/var` en macOS en `Assert-ChuliSafeRegularTarget` (`scripts/profile.ps1`), agregar un job `posix-installer` en `.github/workflows/validate.yml` (`ubuntu-latest` y `macos-latest`) que ejecute `tests/windows-smoke.ps1` y `tests/windows-multiclient.ps1`, y documentar el alcance en `docs/TESTING.md`.
- **Roadmap actualizado (`REQ-003`):** Actualizar `docs/ROADMAP.md` manteniendo las referencias verificables a releases, PRs e issues ya integradas.
- **Modularización TUI y tests rápidos (`REQ-004`):** Agrupar los campos de `model` en `internal/tui/tui.go` en sub-structs embebidos por dominio (`integrationState`, `updateState`, `wizardState`, `modelPickerState`, `skillsState`, `progressState`) y extraer las pantallas de actualizaciones y gestión de versiones de `internal/tui/wizard.go` a `internal/tui/versions.go`. En `internal/manager`, marcar los tests que lanzan PowerShell real con `if testing.Short()` y documentar `go test -short ./...` en `docs/TESTING.md`.
- **Guardrails de paquete (`REQ-005`):** Validar en `scripts/verify_release.py` el presupuesto del ZIP del perfil contra las constantes de `internal/manager/archive.go` (16 MiB comprimido, 32 MiB expandido, 4 MiB por archivo, 2048 entradas) y cubrir casos de rechazo en `tests/test_verify_release.py`.

## Tasks

### TASK-001 — Unificar catálogo de modelos desde config/models.json en Go y Python
- **State:** DONE
- **Depends on:** none
- **Owner:** orchestrator
- **Requirements:** REQ-001
- **Verification:** go test ./internal/manager -run TestCatalogMatchesConfigModelsJSON; python -B -m unittest tests/test_verify_release.py
- **Acceptance:** `internal/manager` carga roles, clientes, modelos, esfuerzos y presets directamente desde `config/models.json` embebido preservando su orden; un test comprueba paridad exacta con el archivo en disco y `package_release.py` no incluye archivos `.go` en el ZIP del perfil.
- **Scope:** ["config/models_embed.go", "config/models.json", "internal/manager/catalog.go", "internal/manager/catalog_test.go", "scripts/package_release.py", "tests/test_verify_release.py", ".github/workflows/validate.yml"]
- **Evidence:** {"command":"go test -short ./...; python -B -m unittest tests/test_verify_release.py","result":"TestCatalogMatchesConfigModelsJSON, TestParseCatalogRejectsInvalidPayload y test_profile_files_excludes_go_sources verdes; gofmt y go vet sin errores.","candidate":"sha256:628f1301365963fb326244012c169f39f4a8a132fa9d4e474203e51e22c8c103"}
- **Blocker:** none

### TASK-002 — Agregar cobertura funcional cross-platform (pwsh) en CI y soporte POSIX
- **State:** DONE
- **Depends on:** none
- **Owner:** orchestrator
- **Requirements:** REQ-002
- **Verification:** pwsh -NoProfile -File tests/windows-smoke.ps1; pwsh -NoProfile -File tests/windows-multiclient.ps1; python tests/validate_repository.py
- **Acceptance:** `.github/workflows/validate.yml` ejecuta pruebas funcionales de instalación, verificación y rollback con `pwsh` en `ubuntu-latest` y `macos-latest`; `Assert-ChuliSafeRegularTarget` contempla los alias del sistema `/tmp` y `/var` en macOS y las suites `windows-smoke.ps1` y `windows-multiclient.ps1` pasan localmente.
- **Scope:** ["scripts/profile.ps1", ".github/workflows/validate.yml", "docs/TESTING.md"]
- **Evidence:** {"command":"pwsh -NoProfile -File tests/windows-smoke.ps1; pwsh -NoProfile -File tests/windows-multiclient.ps1; powershell.exe StrictMode check scripts/profile.ps1; python tests/validate_repository.py","result":"SMOKE TESTS OK, MULTICLIENT INSTALL, VERIFY AND ROLLBACK TESTS OK, PS 5.1 Assert-ChuliSafeRegularTarget OK y validate_repository OK. Ejecución real en runners Linux/macOS queda pendiente de CI remoto.","candidate":"sha256:c40f39fa561fff08a5ec1aa3a1145d8ba532ccb1fac84dba42bbca189f9adadc"}
- **Blocker:** none

### TASK-003 — Actualizar docs/ROADMAP.md hasta v2.1.2
- **State:** DONE
- **Depends on:** none
- **Owner:** orchestrator
- **Requirements:** REQ-003
- **Verification:** python tests/validate_repository.py
- **Acceptance:** `docs/ROADMAP.md` documenta las decisiones vigentes multi-cliente y los hitos verificados desde v0.1 hasta v2.1.2, enlazando únicamente archivos e issues existentes.
- **Scope:** ["docs/ROADMAP.md"]
- **Evidence:** {"command":"python tests/validate_repository.py","result":"OK: 6 agents, 25 skills, EWP/CES and install scripts; enlaces locales a docs/protocols/ASEP.md, docs/protocols/OPERATIONS.md, docs/evaluations/workflow-friction.md y odd/tasks/roadmap-ciclo-de-vida.md verificados.","candidate":"sha256:929b4a23af936a7391d738bc7358457a8f4ef6122d369155fe886b0ff5e52683"}
- **Blocker:** none

### TASK-004 — Modularizar internal/tui y agregar modo rápido testing.Short() en internal/manager
- **State:** DONE
- **Depends on:** TASK-001
- **Owner:** orchestrator
- **Requirements:** REQ-004
- **Verification:** go test -short ./...; go test ./internal/tui; go vet ./...
- **Acceptance:** `internal/tui/tui.go` organiza el estado en sub-structs por dominio y `wizard.go` separa las vistas/actualizaciones de versiones en `versions.go` manteniendo todas las pruebas TUI verdes; `go test -short ./...` corre sin invocar PowerShell mientras `go test ./...` sigue ejecutando la suite completa.
- **Scope:** ["internal/tui/tui.go", "internal/tui/wizard.go", "internal/tui/versions.go", "internal/tui/presentation.go", "internal/tui/integration_commands_test.go", "internal/tui/integrations_test.go", "internal/tui/model_discovery_test.go", "internal/tui/presentation_test.go", "internal/tui/ux_v21_test.go", "internal/tui/wizard_batch_test.go", "internal/manager/agy_models_test.go", "internal/manager/archive_test.go", "internal/manager/manager_v21_test.go", "internal/manager/profiles_skills_test.go", "docs/TESTING.md"]
- **Evidence:** {"command":"gofmt -l; go vet ./...; go test -short ./...","result":"5/5 paquetes Go con tests verdes en modo -short (internal/manager pasó de ~205s a 9.5s en fría); internal/tui verde con wizard.go reducido y versions.go separado.","candidate":"sha256:fb32e3b92b9e5625475af280ef5442dfecb198c81f7a828717a04c24acd55287"}
- **Blocker:** none

### TASK-005 — Agregar guardrails de presupuesto del paquete en verify_release.py
- **State:** DONE
- **Depends on:** TASK-001
- **Owner:** orchestrator
- **Requirements:** REQ-005
- **Verification:** python -B -m unittest tests/test_verify_release.py tests/test_diagram_design_resources.py
- **Acceptance:** `scripts/verify_release.py` rechaza paquetes de perfil que excedan los límites de `internal/manager/archive.go` (16 MiB comprimido, 32 MiB expandido, 4 MiB por entrada, 2048 entradas), conserva intacto el pin upstream de `diagram-design` y cuenta con pruebas unitarias de rechazo y aceptación.
- **Scope:** ["scripts/verify_release.py", "tests/test_verify_release.py", "internal/manager/archive.go"]
- **Evidence:** {"command":"python -B -m unittest tests/test_verify_release.py tests/test_diagram_design_resources.py","result":"18 tests OK (4 casos de límite de presupuesto, paridad exacta con internal/manager/archive.go, verificación del ZIP real empaquetado y pin íntegro de los 293 archivos de diagram-design).","candidate":"sha256:73f63629c19c77309d773e4786f6bd6915cc57653b4f27e2c193ec3959e227b5"}
- **Blocker:** none
