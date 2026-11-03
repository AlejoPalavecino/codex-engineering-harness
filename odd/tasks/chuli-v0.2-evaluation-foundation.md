# Feature: Chuli v0.2 — Evaluation Foundation (PR #5)

- **Spec:** docs/EVALUATION.md
- **Design:** docs/EVALUATION.md
- **Approval:** approved

> Alcance autorizado: fundación de evaluación offline, sin cambiar roles, modelos ni instalar componentes. La publicación de v0.2 y las optimizaciones funcionales requieren aprobación posterior.

## Tasks

### TASK-001 — Definir catálogo y criterios de evaluación
- **State:** DONE
- **Depends on:** none
- **Owner:** orchestrator
- **Acceptance:** Casos R0/R1/R2/C con prompts, rubricas y baseline v0.1.3 identificable.
- **Evidence:** evals/manifest.json + evals/cases + TestActualRepositoryCatalog; CI 41714293479 Go / structure PASS
- **Blocker:** none

### TASK-002 — Preparar fixtures aislados y repetibles
- **State:** DONE
- **Depends on:** TASK-001
- **Owner:** worker
- **Acceptance:** Fixtures sin datos reales, pruebas básicas ejecutables, sin necesidad de acceso externo.
- **Evidence:** evals/fixtures, unittest loans/importer en CI 41714293479 structure PASS
- **Blocker:** none

### TASK-003 — Implementar herramientas de evaluación Go
- **State:** DONE
- **Depends on:** TASK-002
- **Owner:** worker
- **Acceptance:** CLI offline valida catálogo y registros, captura snapshots, compara evidencia y rechaza valores inventados sin fuente.
- **Evidence:** internal/eval y tests Go en Ubuntu/Windows, CI 41714293479 PASS
- **Blocker:** none

### TASK-004 — Verificar regresiones de la fundación
- **State:** VERIFYING
- **Depends on:** TASK-003
- **Owner:** test_engineer
- **Acceptance:** Unit tests y fixtures pasan en CI para Windows, Linux y macOS, sin ejecutar agentes ni modificar la configuración.
- **Evidence:** none
- **Blocker:** none

### TASK-005 — Preparar revisión humana del PR
- **State:** PLANNED
- **Depends on:** TASK-004
- **Owner:** orchestrator
- **Acceptance:** PR documentado, green CI, limitaciones explícitas, comparación conductual pendiente para la fase de evaluación.
- **Evidence:** none
- **Blocker:** none
