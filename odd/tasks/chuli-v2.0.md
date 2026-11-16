# Chuli v2.0 · Multi-Cliente, Modelos por Agente y TUI Visual

- **Spec:** docs/specs/multi-client-v2.md
- **Design:** docs/AGENTS.md
- **Approval:** approved

El usuario aprobó la arquitectura multi-cliente (Codex, Claude Code y AGY CLI), los presets actualizados de modelos (GPT-6.1, Claude 5.5/Fable 5.2, Gemini 4 Carbon/Argon/Hydrogen) y la mejora visual de la TUI CRT para Chuli v2.0.

### TASK-001 — Catálogo de modelos y renderizador multi-cliente (Codex, Claude, AGY)
- **State:** DONE
- **Depends on:** none
- **Owner:** orchestrator
- **Acceptance:** `config/models.json`, `scripts/client_targets.py`, `scripts/toml_config.py` y `scripts/profile_resources.py` generan instrucciones, agentes y configuraciones válidas para `codex`, `claude` y `agy` con pruebas unitarias en Python.
- **Evidence:** `tests/test_client_targets.py` y suite Python `unittest` (60 tests OK, 2 skipped por privilegio de symlink en Windows)
- **Blocker:** none

### TASK-002 — Instalador, verificador y rollback multi-cliente con modelos personalizados
- **State:** DONE
- **Depends on:** TASK-001
- **Owner:** orchestrator
- **Acceptance:** `install.ps1`, `verify.ps1`, `rollback.ps1` e `internal/manager` soportan `-Client codex|claude|agy`, `-Preset` y `-RoleModelsJson` preservando personalizaciones y rollback exacto.
- **Evidence:** `tests/windows-multiclient.ps1`, `tests/windows-smoke.ps1`, `tests/windows-regression.ps1` y `go test ./internal/manager` OK
- **Blocker:** none

### TASK-003 — Rediseño visual TUI CRT, selector de cliente y configurador de modelos
- **State:** DONE
- **Depends on:** TASK-002
- **Owner:** orchestrator
- **Acceptance:** TUI CRT con barra visual de clientes, tarjetas de estado, pantalla interactiva de presets y edición rol por rol, más subcomando `chuli models` y tests Go verdes en 40x24, 80x24 y 120x32.
- **Evidence:** `internal/tui/models.go`, `internal/tui/presentation.go`, `internal/tui/presentation_test.go` y `go test ./internal/tui` (13 tests OK)
- **Blocker:** none

### TASK-004 — Documentación v2.0, validación integral y preparación de release
- **State:** DONE
- **Depends on:** TASK-003
- **Owner:** orchestrator
- **Acceptance:** `VERSION` en `2.0.0`, `README.md`, `docs/INSTALLATION.md`, `docs/TUI.md`, `docs/AGENTS.md`, `CHANGELOG.md` y `docs/releases/v2.0.0.md` actualizados y suites Python, Go y PowerShell pasando.
- **Evidence:** `python tests/validate_repository.py`, `python scripts/taskboard.py check odd/tasks/chuli-v2.0.md`, suites Go/Python/PowerShell OK
- **Blocker:** none
