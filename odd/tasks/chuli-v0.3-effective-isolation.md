# Feature: Chuli v0.3 — Aislamiento efectivo de ejecución

- **Spec:** docs/specs/chuli-v0.3-effective-isolation.md
- **Design:** docs/adr/ADR-0001-chuli-runtime-isolation.md
- **Approval:** pending

> Investigación M0 documentada. Arquitectura y modo de lanzamiento pendientes de aprobación. No implementar antes de esa aprobación.

## M0 — Investigación y decisiones

### TASK-001 — Verificar Permission Profiles y precedencia
- **State:** PLANNED
- **Depends on:** none
- **Owner:** solution_architect
- **Acceptance:** documentar sintaxis, versionado beta, precedencia, incompatibilidad legacy y reglas de paths.
- **Evidence:** matriz oficial y probes reproducibles en Codex CLI.
- **Blocker:** aprobación de arquitectura pendiente.

### TASK-002 — Verificar backends y capacidades por plataforma
- **State:** PLANNED
- **Depends on:** TASK-001
- **Owner:** test_engineer
- **Acceptance:** detectar versión, backend, read/write/deny, procesos hijos y red en Windows, Linux y macOS.
- **Evidence:** informes de capability y probes por backend.
- **Blocker:** requiere máquinas compatibles.

### TASK-003 — Aprobar runtime y modo de lanzamiento
- **State:** PLANNED
- **Depends on:** TASK-001, TASK-002
- **Owner:** orchestrator
- **Acceptance:** aprobar CODEX_HOME aislado, login separado, modo TUI/CLI, defaults de red/MCP y alcance CLI.
- **Evidence:** ADR aprobado y decisiones registradas.
- **Blocker:** decisión del usuario.

## M1 — Policy Core

### TASK-004 — Definir esquema neutral de políticas
- **State:** PLANNED
- **Depends on:** TASK-003
- **Owner:** solution_architect
- **Acceptance:** schema versionado para roles, roots, paths read/write/deny, red, MCP y verificación.
- **Evidence:** fixtures válidos/inválidos y documentación del esquema.
- **Blocker:** M0 pendiente.

### TASK-005 — Normalizar y validar paths
- **State:** PLANNED
- **Depends on:** TASK-004
- **Owner:** worker
- **Acceptance:** bloquear traversal, rutas ambiguas, duplicados y enlaces no verificables.
- **Evidence:** unit tests y fixtures adversariales.
- **Blocker:** M0 pendiente.

## M2 — Runtime y adaptador Codex

### TASK-006 — Inicializar CODEX_HOME aislado
- **State:** PLANNED
- **Depends on:** TASK-003, TASK-004
- **Owner:** worker
- **Acceptance:** crear runtime separado, instalar recursos Chuli y no copiar credenciales ni MCP.
- **Evidence:** pruebas de instalación, estado y cleanup.
- **Blocker:** M0 pendiente.

### TASK-007 — Generar perfiles por rol
- **State:** PLANNED
- **Depends on:** TASK-004, TASK-005, TASK-006
- **Owner:** worker
- **Acceptance:** perfiles planner, test-author, implementation, reviewer y verifier sin sandbox legacy.
- **Evidence:** TOML validado y pruebas codex sandbox por perfil.
- **Blocker:** M0 pendiente.

### TASK-008 — Implementar capability detector y fail-closed
- **State:** PLANNED
- **Depends on:** TASK-001, TASK-002, TASK-007
- **Owner:** test_engineer
- **Acceptance:** detectar versión, plataforma, backend, config legacy, perfil efectivo, red y capacidades.
- **Evidence:** READY/BLOCKED/NOT_VERIFIABLE con casos sintéticos y entorno real.
- **Blocker:** M0 pendiente.

## M3 — Worktrees y procesos

### TASK-009 — Administrar worktrees por etapa
- **State:** PLANNED
- **Depends on:** TASK-003, TASK-005
- **Owner:** worker
- **Acceptance:** crear, validar, recuperar y limpiar worktrees sin descartar cambios locales.
- **Evidence:** tests de lifecycle, concurrencia, interrupción y cleanup.
- **Blocker:** M0 pendiente.

### TASK-010 — Lanzar procesos Codex con perfiles distintos
- **State:** PLANNED
- **Depends on:** TASK-007, TASK-008, TASK-009
- **Owner:** worker
- **Acceptance:** cada etapa inicia proceso principal con CODEX_HOME, perfil y worktree correctos.
- **Evidence:** proceso y perfil observados.
- **Blocker:** modo de lanzamiento pendiente.

### TASK-011 — Congelar tests aprobados
- **State:** PLANNED
- **Depends on:** TASK-009, TASK-010
- **Owner:** worker
- **Acceptance:** registrar commit y SHA-256; implementation puede leer tests pero no modificarlos.
- **Evidence:** shell y apply_patch denegados, hashes intactos.
- **Blocker:** M0 pendiente.

## M4 — Pruebas adversariales

### TASK-012 — Probar filesystem y procesos hijos
- **State:** PLANNED
- **Depends on:** TASK-008, TASK-010
- **Owner:** test_engineer
- **Acceptance:** read/write/deny, delete, rename, traversal, enlaces y subprocessos cumplen la política.
- **Evidence:** logs y hashes por backend.
- **Blocker:** requiere backend real.

### TASK-013 — Probar sesión real, apply_patch y subagentes
- **State:** PLANNED
- **Depends on:** TASK-011, TASK-012
- **Owner:** quality_reviewer
- **Acceptance:** Codex y subagentes no editan tests congelados ni obtienen escalada automática.
- **Evidence:** transcript sanitizado, denegación y hashes.
- **Blocker:** requiere sesión Codex compatible.

### TASK-014 — Probar red y frontera MCP
- **State:** PLANNED
- **Depends on:** TASK-008, TASK-010
- **Owner:** test_engineer
- **Acceptance:** red deshabilitada por defecto; proxy/allowlist y MCP cumplen política.
- **Evidence:** probes permitidos/denegados.
- **Blocker:** M0 pendiente.

## M5 — CLI, TUI e informes

### TASK-015 — Integrar comandos isolate
- **State:** PLANNED
- **Depends on:** TASK-008, TASK-009, TASK-010, TASK-012
- **Owner:** worker
- **Acceptance:** check, plan, probe, freeze-tests, launch, verify y report con preview/confirmación.
- **Evidence:** tests Go, recorrido TUI y tests de integración.
- **Blocker:** M0 pendiente.

### TASK-016 — Generar reporte reproducible
- **State:** PLANNED
- **Depends on:** TASK-011, TASK-012, TASK-013, TASK-014
- **Owner:** worker
- **Acceptance:** JSON registra versión, backend, perfil, paths, hashes, probes, red, MCP y estado.
- **Evidence:** schema, fixtures y snapshots deterministas.
- **Blocker:** M0 pendiente.

## M6 — CI y release

### TASK-017 — Validar plataformas y preparar v0.3.0
- **State:** PLANNED
- **Depends on:** TASK-015, TASK-016
- **Owner:** quality_reviewer
- **Acceptance:** CI verde, pruebas conductuales de release, matriz real de soporte y documentación.
- **Evidence:** informes por backend, revisión independiente y checklist.
- **Blocker:** pruebas reales pendientes.

## Gate de salida

- Tests congelados no modificables desde implementation.
- Reviewer read-only.
- Verifier sobre copia temporal; candidato original intacto.
- Sin fallback silencioso ni READY sin probes.
- MCP y red tratados como superficies separadas.
- Solo plataformas que superen la suite se declaran soportadas.
- Todas las tareas necesarias en DONE con evidencia.
