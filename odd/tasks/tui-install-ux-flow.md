# Feature: Rediseño integral de UX de instalación y navegación en la TUI

- **Spec:** #especificacion
- **Design:** #diseno-tecnico
- **Approval:** approved
- **Evidence mode:** candidate-v1
- **Trace mode:** requirements-v1

## Especificacion

- **REQ-001:** La portada de Chuli Manager guía el primer uso cuando no hay instalaciones, asigna `1`–`6` a las acciones del menú, usa `Tab`/`Shift+Tab`/`←`/`→` para cambiar de cliente sin prefijos numéricos en pestañas, reemplaza el encabezado en subpantallas por una línea de breadcrumb sin agregar filas verticales, restringe `q` a pantallas finales/portada y colorea la barra inferior de atajos manteniendo su texto plano intacto.
- **REQ-002:** El Paso 1 del asistente (`installPlatformsScreen`) usa `Espacio` (y `1`–`3`) para marcar/desmarcar plataformas, permite avanzar directo desde cualquier fila con `Enter` hacia el Paso 3 conservando personalizaciones instaladas o usando `Equilibrado` (vía rápida), y ofrece acceso explícito al Paso 2 mediante fila dedicada o teclas `m`/`c`.
- **REQ-003:** El Paso 2 (`installProfileScreen` y `modelsScreen`) permite cambiar entre plataformas elegidas con `Tab`/`Shift+Tab`, muestra presets oficiales junto a perfiles guardados del usuario, evita sobreescribir configuraciones `custom` al mover el cursor con `↑`/`↓` (fijando cambios con `Espacio` o `Enter`), y agrega atajo `s`/`Ctrl+S` (más `Tab` en modo wizard) en `modelsScreen`.
- **REQ-004:** El Paso 3 (`installConfirmScreen`) muestra nombres legibles de presets/perfiles y acción por plataforma (instalación nueva o actualización), ejecuta un preflight asíncrono no destructivo de PowerShell/Python y detección de conflictos de Gentle AI, y permite autorizar el respaldo de Gentle AI antes de iniciar el Paso 4; volver con `Esc` respeta si el usuario llegó por vía rápida o desde el Paso 2.
- **REQ-005:** La pantalla de resultado (`result`) presenta acciones navegables con `↑`/`↓` y `Enter` (volver al inicio, abrir MCPs, abrir Skills, ver detalle técnico o volver a confirmación ante fallo en asistente), toda la TUI ajusta en `40×24`, `80×24` y `120×32`, y las pruebas y documentación (`README.md`, `docs/TUI.md`) reflejan el flujo implementado.

## Diseno tecnico

- **Preflight no destructivo (`REQ-004`):** `internal/manager/preflight.go` expone `HasGentleRoutingConflict`, `CheckRuntimePreflight` (con timeout acotado para no bloquear con `PATH` vacío) y `PresetDisplayName`, sin escribir en disco.
- **Navegación global, portada, breadcrumbs y `result` (`REQ-001`, `REQ-005`):** `internal/tui/presentation.go` y `internal/tui/tui.go` unifican `renderClientTabs`, `crtBreadcrumb` en 1 línea, sincronización de cabecera con `wizardActiveID` (preservando rutas sintéticas cuando coincide con `mgr.Target`), `renderKeyBar` con contraste ANSI, alcance seguro de `q` y acciones navegables en `result`.
- **Asistente 1–4 y editor de roles (`REQ-002`, `REQ-003`, `REQ-004`):** `internal/tui/wizard.go` y `internal/tui/models.go` separan vía rápida vs personalización, protegen el estado `custom` de la navegación por cursor, integran perfiles guardados y preflight en Paso 3, y preservan la autorización individual de Gentle AI en reintentos batch.
- **Pruebas y documentación (`REQ-005`):** Se actualizan `internal/tui/presentation_test.go` y `internal/tui/ux_v21_test.go`, se agrega `internal/tui/ux_flow_test.go` y se actualizan `docs/TUI.md` y `README.md`.

## Tasks

### TASK-001 — Helpers no destructivos de preflight y nombres legibles en internal/manager
- **State:** DONE
- **Depends on:** none
- **Owner:** orchestrator
- **Requirements:** REQ-004
- **Verification:** go test -short ./internal/manager
- **Acceptance:** `HasGentleRoutingConflict` detecta marcadores de enrutamiento de Gentle AI fuera de bloques cercados Markdown sin modificar archivos; `CheckRuntimePreflight` reporta disponibilidad de PowerShell y Python 3.11+ en lectura; `PresetDisplayName` devuelve el título legible de cada preset o `"Personalizado por rol"` para `custom`.
- **Scope:** ["internal/manager/preflight.go", "internal/manager/preflight_test.go"]
- **Evidence:** {"command": "go test -count=1 -short ./internal/manager", "result": "ok internal/manager (4.666s); preflight y PresetDisplayName en verde", "candidate": "sha256:52aa3157b5c178c380c0c753019b72971c06cb0f5cc6ad39082357c55c742d65"}
- **Blocker:** none

### TASK-002 — Navegación global, portada adaptativa, breadcrumbs, barra de atajos y cierre accionable
- **State:** DONE
- **Depends on:** none
- **Owner:** orchestrator
- **Requirements:** REQ-001, REQ-005
- **Verification:** go test -short ./internal/tui
- **Acceptance:** En `home` las teclas `1`–`6` activan opciones del menú y `Tab`/`Shift+Tab`/`←`/`→` cambian cliente; `renderClientTabs` omite prefijos numéricos; `crtBreadcrumb` reemplaza la línea superior en subpantallas sin exceder `40×24`; `q` cierra solo en `home`, `about`, `result` resumido e `integrationResult`; `renderKeyBar` preserva el texto plano tras `ansi.Strip`; `result` permite navegar acciones con `↑`/`↓` y `Enter`.
- **Scope:** ["internal/tui/tui.go", "internal/tui/presentation.go", "internal/tui/versions.go", "internal/tui/skills.go", "internal/tui/integrations.go"]
- **Evidence:** {"command": "go test -count=1 -short ./internal/tui", "result": "ok internal/tui (6.450s); portada, breadcrumbs, q seguro, renderKeyBar y result en verde", "candidate": "sha256:fd7549358de544c70641d2dc898cb82fac6a4dad9a0c022d436e38319209264a"}
- **Blocker:** none

### TASK-003 — Rediseño del asistente Pasos 1–3, preflight y atajos del editor de modelos
- **State:** DONE
- **Depends on:** TASK-001, TASK-002
- **Owner:** orchestrator
- **Requirements:** REQ-002, REQ-003, REQ-004
- **Verification:** go test -short ./internal/tui ./internal/manager
- **Acceptance:** Paso 1 marca con `Espacio` y avanza a Paso 3 con `Enter` (o a Paso 2 por opción/tecla `m`/`c`); Paso 2 permite cambiar de plataforma con `Tab`/`Shift+Tab`, muestra perfiles guardados y conserva equipos `custom` sin sobreescribir al mover `↑`/`↓`; `modelsScreen` confirma con `s`/`Ctrl+S`; Paso 3 ejecuta preflight asíncrono, muestra nombres legibles y permite marcar adopción de Gentle AI antes del Paso 4.
- **Scope:** ["internal/tui/wizard.go", "internal/tui/models.go", "internal/tui/tui.go"]
- **Evidence:** {"command": "go test -count=1 -short ./internal/tui ./internal/manager", "result": "ok internal/tui (6.450s) y internal/manager (4.666s); wizard pasos 1-3 y modelsScreen en verde", "candidate": "sha256:83be2f1da9ca3a74ff5bdb834be79c4f411bb05705db8602766b4bf6f9dbdab6"}
- **Blocker:** none

### TASK-004 — Pruebas de flujo UX, actualización de documentación y verificación integral
- **State:** DONE
- **Depends on:** TASK-002, TASK-003
- **Owner:** orchestrator
- **Requirements:** REQ-001, REQ-002, REQ-003, REQ-004, REQ-005
- **Verification:** gofmt -l cmd/chuli config internal/manager internal/tui internal/eval internal/integrations; go vet ./...; go test -short ./internal/manager ./internal/tui ./internal/eval ./internal/integrations; python tests/validate_repository.py
- **Acceptance:** Las suites `internal/manager` e `internal/tui` (incluyendo `internal/tui/ux_flow_test.go`, `TestCRTAllScreensFitTerminalWidth` y `TestCRTHeightKeepsControlsVisible`) pasan sin regresiones, `docs/TUI.md` y `README.md` describen el flujo vigente y `tests/validate_repository.py` termina sin errores.
- **Scope:** ["internal/tui/presentation_test.go", "internal/tui/ux_v21_test.go", "internal/tui/ux_flow_test.go", "docs/TUI.md", "README.md"]
- **Evidence:** {"command": "gofmt -l; go vet ./...; go test -count=1 -short ./internal/manager ./internal/tui ./internal/eval ./internal/integrations; python tests/validate_repository.py", "result": "Todos los paquetes Go (-short) en verde, gofmt/go vet limpios y validate_repository OK", "candidate": "sha256:93246796dbed0b45ebf73799ed49b7ce2e918d844c91733a9ad8872d9cfd90e3"}
- **Blocker:** none
