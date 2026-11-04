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
- **State:** DONE
- **Depends on:** TASK-003
- **Owner:** test_engineer
- **Acceptance:** Unit tests y fixtures pasan en CI para Windows, Linux y macOS, sin ejecutar agentes ni modificar la configuración.
- **Evidence:** GitHub Actions 41714827622 PASS (structure, windows-installer, chuli-go ubuntu/windows/macos); fingerprint 7ee5b69e... idéntico entre SO
- **Blocker:** none

### TASK-005 — Preparar revisión humana del PR
- **State:** DONE
- **Depends on:** TASK-004
- **Owner:** orchestrator
- **Acceptance:** PR documentado, green CI, limitaciones explícitas, comparación conductual pendiente para la fase de evaluación.
- **Evidence:** Revisión independiente sobre 8b52754: cuatro defectos reproducidos y corregidos con RED/GREEN; go test ./..., go vet, Python y regresiones instalador/rollback PASS. CI 41898483959 y 41898494409: todos los jobs PASS para 8b527547e5a538fa673db09d1bbc304aaf57ac55. Usuario autorizó squash merge condicional el 2026-11-04; baseline conductual sigue separado.
- **Blocker:** none

## Validación adicional del 2026-11-04

Correcciones: comparación por criterio y NOT_TESTED, creación exclusiva de snapshots y rechazo de alias al workspace, cierre SQLite en éxito/error. Los fixtures Python se ejecutan también en los tres SO. Nueva huella: `cc0d2af804909ee1e10c99826b5de8979591b048c3d6e984f2de5dcfd0a286f1`; prompts y criterios originales preservados. Límites de rutas concurrentes documentados en docs/EVALUATION.md.
