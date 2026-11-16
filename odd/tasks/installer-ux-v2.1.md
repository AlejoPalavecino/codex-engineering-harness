# Feature: Chuli UX de Instalación, Multi-Plataforma, Versiones, Perfiles, MCPs y Skills

- **Spec:** docs/specs/installer-ux-v2.1.md
- **Design:** docs/BRAND.md
- **Approval:** approved

El usuario aprobó el diseño de la nueva experiencia TUI en terminal manteniendo el banner CRT retro (`renderChuliLogo`), flujo de instalación multi-plataforma sin ruido visual, notificaciones y resumen de actualizaciones, gestor de versiones, perfiles de AI por proveedor, MCPs multi-cliente, gestor de skills y los arreglos base de AGY y PowerShell.

## Tasks

### TASK-001 — Arreglos base de subagentes/settings AGY y SHA-256 portable en PowerShell
- **State:** DONE
- **Depends on:** none
- **Owner:** chuli_worker
- **Acceptance:** `scripts/client_targets.py`, `scripts/profile_resources.py`, `scripts/profile.ps1`, `scripts/install.ps1`, `scripts/verify.ps1`, `scripts/rollback.ps1` y `scripts/install-manager.ps1` producen agentes AGY válidos (`agy plugin validate` con 0 errores y 6 agentes), sincronizan `antigravity-cli/settings.json` cuando corresponde y calculan SHA-256 sin `Get-FileHash`; pruebas Python y PowerShell pasan.
- **Evidence:** test_client_targets.py (6 OK), test_profile_resources.py (9 OK), agy plugin validate (6 agents, 0 errors), windows-multiclient.ps1 + windows-smoke.ps1 + windows-manager-install.ps1 + windows-regression.ps1 (PS 5.1 y pwsh) OK
- **Blocker:** none

### TASK-002 — Motor Manager: autodetección de plataformas, releases con novedades, versiones y aplicación local
- **State:** DONE
- **Depends on:** TASK-001
- **Owner:** chuli_worker
- **Acceptance:** `internal/manager` detecta plataformas instaladas (`codex`, `claude`, `agy`), obtiene lista de releases + release notes (`body`), detecta actualizaciones de perfil y de binario Manager, instala/cambia versiones específicas y aplica cambios de perfil localmente sin re-descargar `latest`, cubierto por `go test ./internal/manager`.
- **Evidence:** go test ./internal/manager: TestDetectClientsAndForTarget, TestListReleasesAndSummarizeNotes, TestCheckUpdatesDetectsOutdatedProfilesAndManager (y suite previa completa salvo ajuste de ruta .gemini en TestDetectClientsAndForTarget; corrida completa final en TASK-005)
- **Blocker:** none

### TASK-003 — Perfiles nombrados por proveedor, MCPs multi-cliente/custom y gestión de Skills
- **State:** DONE
- **Depends on:** TASK-002
- **Owner:** chuli_worker
- **Acceptance:** Existen almacenamiento/carga de perfiles propios por proveedor, soporte MCP multi-cliente (`codex`, `claude`, `agy`) + MCPs personalizados en `internal/integrations`, y activar/desactivar/agregar/quitar skills compatible con `state.json` y `verify.ps1`, con pruebas verdes en Go y Python.
- **Evidence:** Python unittest (19 OK, 2 skipped), go test ./internal/integrations (all PASS), TestSavedProfilesCRUDAndValidation + TestSkillsListDisableInInstallVerifyRollbackAndCustom PASS
- **Blocker:** none

### TASK-004 — Rediseño TUI CRT: banner preservado, menú de 6 botones, wizard animado y pantallas nuevas
- **State:** DONE
- **Depends on:** TASK-003
- **Owner:** chuli_worker
- **Acceptance:** `internal/tui` conserva `renderChuliLogo`, agrega banner de nueva release, menú de 6 acciones, wizard de instalación multi-plataforma con animación por etapas que tapa el texto técnico y pantalla de éxito, más pantallas de Actualizar, Versiones, Perfiles de AI, MCPs y Skills responsivas en `40x24`, `80x24` y `120x32` con `go test ./internal/tui ./cmd/chuli` verde.
- **Evidence:** go test -v ./internal/tui (17/17 PASS en 40x24, 80x24 y 120x32)
- **Blocker:** none

### TASK-005 — Documentación, verificación integral y cierre del repositorio
- **State:** DONE
- **Depends on:** TASK-004
- **Owner:** quality_reviewer
- **Acceptance:** `README.md`, `docs/TUI.md`, `docs/INSTALLATION.md` y `docs/AGENTS.md` reflejan el nuevo flujo, `python tests/validate_repository.py`, `go test ./...`, `python -m unittest discover -s tests`, scripts PowerShell de regresión y `agy plugin validate` pasan sin fallos, y `taskboard.py check` valida el tablero.
- **Evidence:** validate_repository.py OK; python unittest 62 tests OK (2 skipped); go test ./... all 5 packages OK; windows-multiclient.ps1, windows-smoke.ps1, windows-regression.ps1, windows-manager-install.ps1 y windows-ttp-resources.ps1 OK
- **Blocker:** none

### TASK-006 — Bugfix de búsqueda de actualizaciones y navegación secuencial con flechas + Enter
- **State:** DONE
- **Depends on:** TASK-005
- **Owner:** chuli_worker
- **Acceptance:** Al buscar actualizaciones (`R` o botón en `updateScreen`), `updatesLoadedMsg` regresa de `busy` a `updateScreen` (y `Esc` cancela consultas solo-lectura); el wizard guía paso a paso con flechas + `Enter` seleccionando en vivo, configura los modelos proveedor por proveedor de forma secuencial e incluye selector de rol con `↑↓` + `Enter`, validado con pruebas Go de flujo completo.
- **Evidence:** go test -v ./internal/tui (19/19 PASS), go build ./cmd/chuli + go test ./cmd/chuli OK, validate_repository.py OK
- **Blocker:** none
