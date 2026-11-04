# Feature: Validación integral y baseline de Chuli

- **Spec:** docs/EVALUATION.md y pedido autorizado del usuario del 2026-11-04
- **Design:** docs/EVALUATION.md
- **Approval:** approved

> Autorizados: pruebas sintéticas, correcciones de Foundation, integración condicional de PR #5 y planificación. Excluidos: release, deploy, cambios de instalación o memorias reales y optimizaciones sustanciales sin aprobación.

## Tasks

### TASK-001 — Identificar entorno y aislar evidencia
- **State:** DONE
- **Depends on:** none
- **Owner:** orchestrator
- **Acceptance:** Versiones y diferencias con baseline identificadas; evidencia privada excluida de Git; instalación personal preservada.
- **Evidence:** CLI 0.166.0; Desktop 26.1029.7803.0/runtime alpha.6; manifiesto instalado 0.1.1 personalizado; baseline inmutable 68b756b. Worktree separado y home sintético instalado desde SHA. Probe autenticado exit 0; reutilización temporal de login autorizada por el usuario.
- **Blocker:** none

### TASK-002 — Validar Foundation y resolver defectos
- **State:** DONE
- **Depends on:** TASK-001
- **Owner:** orchestrator
- **Acceptance:** Go/Python/instalador/rollback, catálogo, huellas, snapshots y revisión independiente; sin bloqueantes.
- **Evidence:** Go test/vet/mod/format PASS; Python repositorio 16, loans 4, importer 2 PASS; smoke/rollback Windows PASS. Cuatro defectos corregidos con RED/GREEN; revisión de cierre VERIFIED. CI 41898759757/41898767437: diez checks SUCCESS sobre 6c8e481.
- **Blocker:** none

### TASK-003 — Integrar Foundation y verificar main
- **State:** DONE
- **Depends on:** TASK-002
- **Owner:** orchestrator
- **Acceptance:** PR #5 squash solo con CI verde del SHA actual; pruebas postmerge; v0.1.3 preservada.
- **Evidence:** Squash 8b15e082fd9716172238da6a756da704e5e9f6e4; Go y eval check postmerge PASS; CI main 41899170236 cinco jobs SUCCESS. Última release pública v0.1.3.
- **Blocker:** none

### TASK-004 — Ejecutar evaluación conductual factible
- **State:** DONE
- **Depends on:** TASK-003
- **Owner:** orchestrator
- **Acceptance:** Piloto R0/R1 primero; escenarios factibles con evidencia; resto NOT_TESTED/PARTIAL explicado; prompts/criterios preservados; sin aprobaciones simuladas.
- **Evidence:** Seis escenarios y un reintento R0: 2 PASS, 4 PARTIAL, 5 NOT_TESTED. baseline-run.json validado por eval report. Reviewer independiente verificó consistencia de estados, snapshots y métricas. Límites del cliente, memoria y recuperación registrados en docs/evaluations/2026-11-04-baseline.md.
- **Blocker:** none; escenarios pendientes por bloqueo del cliente se trasladan a OPT-001/OPT-002, no se declaran ejecutados.

### TASK-005 — Entregar baseline y propuestas v0.2
- **State:** DONE
- **Depends on:** TASK-004
- **Owner:** orchestrator
- **Acceptance:** Informe A-H, JSON validado, especificación y tareas pendientes sin aprobaciones inventadas; resumen público sanitizado y evidencia privada preservada.
- **Evidence:** docs/evaluations/2026-11-04-baseline.md; docs/specs/chuli-v0.2-evaluation-next.md; odd/tasks/chuli-v0.2-optimizations.md; rama codex/v0.2-baseline-plan. Verificación final: tres tableros válidos, validate_repository PASS, diff --check PASS; rutas privadas ignoradas por Git; hashes personales intactos y cero auth.json temporales. Revisión independiente confirmó consistencia del registro. Entrega local sanitizada de 297 líneas, sin cambios funcionales adicionales.
- **Blocker:** none
