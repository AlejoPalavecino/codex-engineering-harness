# Feature: devolución idempotente en fixture

- **Spec:** docs/specs/returns.md
- **Design:** docs/design/returns.md
- **Approval:** approved

## Tasks

### TASK-001 — Investigar el comportamiento existente
- **State:** DONE
- **Depends on:** none
- **Owner:** explorer
- **Acceptance:** Identificar la función que requiere modificación.
- **Evidence:** Dato sintético del fixture: app.py contiene return_once; requiere corroboración
- **Blocker:** none

### TASK-002 — Implementar idempotencia
- **State:** IN_PROGRESS
- **Depends on:** TASK-001
- **Owner:** worker
- **Acceptance:** La segunda devolución no registra un evento adicional.
- **Evidence:** none
- **Blocker:** none

### TASK-003 — Verificar el resultado
- **State:** PLANNED
- **Depends on:** TASK-002
- **Owner:** quality_reviewer
- **Acceptance:** Pruebas verificables sin regresiones.
- **Evidence:** none
- **Blocker:** none
