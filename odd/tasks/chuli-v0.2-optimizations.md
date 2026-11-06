# Feature: Chuli v0.2 — Evaluación completa y optimizaciones demostrables

- **Spec:** docs/specs/chuli-v0.2-evaluation-next.md
- **Design:** docs/specs/chuli-v0.2-evaluation-next.md
- **Approval:** approved

> Plan preparado por pedido del usuario del 2026-11-04. El usuario aprobó el diseño y autorizó ejecutar las seis OPT el 2026-11-04. Cambios sustanciales de producción o una candidata aún no definida requieren aprobación específica. Foundation ya integrada se registra por separado. No se modifica producción ni se publica v0.2.

## Tasks

### OPT-001 — Acordar y validar preflight del cliente
- **State:** DONE
- **Depends on:** none
- **Owner:** orchestrator
- **Acceptance:** Diseño aprobado; permisos solicitados/observados y lectura/escritura de copia sintética comprobados; exit 0 no se interpreta como benchmark exitoso; sin secretos en evidencia.
- **Evidence:** Diagnóstico sin modelo: legacy mantiene read-only; default_permissions reconoce perfiles. MXC real: lectura, escritura dentro, 4 tests y rechazo de escrituras fuera/readonly comprobados. eval_preflight: RED real, 10 tests GREEN y 26 tests repositorio PASS. Evidencia local opt-execution; reviewer independiente VERIFIED tras corregir dos falsos READY (permisos contradictorios y marcador en prosa); 13 tests específicos y 29 de repositorio PASS.
- **Blocker:** none

### OPT-002 — Verificar piloto R0/R1 con capacidades disponibles
- **State:** IN_PROGRESS
- **Depends on:** OPT-001
- **Owner:** test_engineer
- **Acceptance:** Una ejecución por caso con prompts/criterios originales y misma huella; R0 sin mutaciones; R1 solo cambia literal y ejecuta pruebas; observar cliente/modelo sin inferencias.
- **Evidence:** none
- **Blocker:** Dependencia OPT-001 verificada; piloto autorizado.

### OPT-003 — Completar escenarios de implementación y calidad
- **State:** PLANNED
- **Depends on:** OPT-002
- **Owner:** orchestrator
- **Acceptance:** Implementación, TDD y relectura de diseño/código con evidencia; RED temporal anterior a GREEN; permisos ajenos preservados; no afirmar protección de tests sin comprobación.
- **Evidence:** none
- **Blocker:** Espera OPT-002; diseño y ejecución aprobados.

### OPT-004 — Comprobar recuperación y autoridad con memoria aislada
- **State:** PLANNED
- **Depends on:** OPT-002
- **Owner:** test_engineer
- **Acceptance:** Interrupción y nueva sesión reales; TTP/spec/diff cotejados; A usa ADR vigente y no historia/B; instancia ficticia autorizada o integración Engram NOT_TESTED.
- **Evidence:** none
- **Blocker:** Espera OPT-002; instancia ficticia autorizada y creada con ENGRAM_DATA_DIR separado.

### OPT-005 — Medir delegación y modelos observables
- **State:** PLANNED
- **Depends on:** OPT-003
- **Owner:** quality_reviewer
- **Acceptance:** Delegaciones justificadas y contratos evaluados con fuentes; modelo configurado/observado separados; ausencia de datos explícita; sin repeticiones extensivas ni costos inventados.
- **Evidence:** none
- **Blocker:** Espera OPT-003 y muestra observada; diseño aprobado.

### OPT-006 — Proponer y comparar una mejora de v0.2
- **State:** BLOCKED
- **Depends on:** OPT-003, OPT-004, OPT-005
- **Owner:** orchestrator
- **Acceptance:** Una variable, beneficio y regresiones medibles; aprobación de diseño e implementación por separado; sin release ni deploy.
- **Evidence:** none
- **Blocker:** Falta evidencia conductual completa y decisión humana sobre la mejora candidata; no cambiar modelos ni agentes por intuición.
