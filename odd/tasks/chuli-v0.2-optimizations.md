# Feature: Chuli v0.2 — Evaluación completa y optimizaciones demostrables

- **Spec:** docs/specs/chuli-v0.2-evaluation-next.md
- **Design:** docs/specs/chuli-v0.2-evaluation-next.md
- **Approval:** pending

> Plan preparado por pedido del usuario del 2026-11-04. Las propuestas no fueron aprobadas ni autorizadas para implementar. Foundation ya integrada se registra por separado. No se modifica producción ni se publica v0.2.

## Tasks

### OPT-001 — Acordar y validar preflight del cliente
- **State:** BLOCKED
- **Depends on:** none
- **Owner:** orchestrator
- **Acceptance:** Diseño aprobado; permisos solicitados/observados y lectura/escritura de copia sintética comprobados; exit 0 no se interpreta como benchmark exitoso; sin secretos en evidencia.
- **Evidence:** docs/evaluations/2026-11-04-baseline.md; pilotos R0/R1 con comandos rechazados y discrepancia workspace-write/read-only.
- **Blocker:** Falta aprobación del diseño propuesto y habilitar una capacidad de ejecución de prueba sin modificar la instalación personal.

### OPT-002 — Verificar piloto R0/R1 con capacidades disponibles
- **State:** PLANNED
- **Depends on:** OPT-001
- **Owner:** test_engineer
- **Acceptance:** Una ejecución por caso con prompts/criterios originales y misma huella; R0 sin mutaciones; R1 solo cambia literal y ejecuta pruebas; observar cliente/modelo sin inferencias.
- **Evidence:** none
- **Blocker:** Diseño pendiente y dependencia OPT-001.

### OPT-003 — Completar escenarios de implementación y calidad
- **State:** PLANNED
- **Depends on:** OPT-002
- **Owner:** orchestrator
- **Acceptance:** Implementación, TDD y relectura de diseño/código con evidencia; RED temporal anterior a GREEN; permisos ajenos preservados; no afirmar protección de tests sin comprobación.
- **Evidence:** none
- **Blocker:** Diseño pendiente y capacidad del cliente aún no comprobada.

### OPT-004 — Comprobar recuperación y autoridad con memoria aislada
- **State:** PLANNED
- **Depends on:** OPT-002
- **Owner:** test_engineer
- **Acceptance:** Interrupción y nueva sesión reales; TTP/spec/diff cotejados; A usa ADR vigente y no historia/B; instancia ficticia autorizada o integración Engram NOT_TESTED.
- **Evidence:** none
- **Blocker:** Diseño pendiente; no hay instancia de memoria aislada autorizada/preparada.

### OPT-005 — Medir delegación y modelos observables
- **State:** PLANNED
- **Depends on:** OPT-003
- **Owner:** quality_reviewer
- **Acceptance:** Delegaciones justificadas y contratos evaluados con fuentes; modelo configurado/observado separados; ausencia de datos explícita; sin repeticiones extensivas ni costos inventados.
- **Evidence:** none
- **Blocker:** No hay muestra suficiente de delegación del benchmark; diseño pendiente.

### OPT-006 — Proponer y comparar una mejora de v0.2
- **State:** BLOCKED
- **Depends on:** OPT-003, OPT-004, OPT-005
- **Owner:** orchestrator
- **Acceptance:** Una variable, beneficio y regresiones medibles; aprobación de diseño e implementación por separado; sin release ni deploy.
- **Evidence:** none
- **Blocker:** Falta evidencia conductual completa y decisión humana sobre la mejora candidata; no cambiar modelos ni agentes por intuición.
