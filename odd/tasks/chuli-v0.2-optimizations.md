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
- **State:** DONE
- **Depends on:** OPT-001
- **Owner:** test_engineer
- **Acceptance:** Una ejecución por caso con prompts/criterios originales y misma huella; R0 sin mutaciones; R1 solo cambia literal y ejecuta pruebas; observar cliente/modelo sin inferencias.
- **Evidence:** Piloto R0/R1 real con MXC: lectura, snapshots, diff mínimo y 4 tests del agente comprobados; fuente privada cases/*-mxc
- **Blocker:** none


### OPT-003 — Completar escenarios de implementación y calidad
- **State:** DONE
- **Depends on:** OPT-002
- **Owner:** orchestrator
- **Acceptance:** Implementación, TDD y relectura de diseño/código con evidencia; RED temporal anterior a GREEN; permisos ajenos preservados; no afirmar protección de tests sin comprobación.
- **Evidence:** Dos escenarios reales, contratos y RED/GREEN verificados; evidencia privada cases/r2-implementation-mxc y tdd-qa-mxc. Limitación del tablero TTP registrada
- **Blocker:** none


### OPT-004 — Comprobar recuperación y autoridad con memoria aislada
- **State:** DONE
- **Depends on:** OPT-002
- **Owner:** test_engineer
- **Acceptance:** Interrupción y nueva sesión reales; TTP/spec/diff cotejados; A usa ADR vigente y no historia/B; instancia ficticia autorizada o integración Engram NOT_TESTED.
- **Evidence:** CLI interrumpido tras mutación real de app.py; nueva thread distinta sin transcript; 5 tests y taskboard canónico PASS. Engram3.4.0 en data-dir ficticio: A14/B90 separados, otra base vacía sin datos; agente eligió ADR A30 y solo mem_current_project; conteos intactos. Recall de contenido por agente y Desktop no verificados.
- **Blocker:** none


### OPT-005 — Medir delegación y modelos observables
- **State:** DONE
- **Depends on:** OPT-003
- **Owner:** quality_reviewer
- **Acceptance:** Delegaciones justificadas y contratos evaluados con fuentes; modelo configurado/observado separados; ausencia de datos explícita; sin repeticiones extensivas ni costos inventados.
- **Evidence:** MRP real: funciones correctas, decisión inline justificada, sin modelo inventado. Siete escenarios nuevos sin delegaciones observadas; turn_context Sol/medium en sesiones raíz (dos para recuperación). Especialistas/backend NO VERIFICABLES; reviewer independiente verificó registros. Auditoría del helper sí delegó implementación y revisión con hallazgos reproducibles, separada del benchmark.
- **Blocker:** none


### OPT-006 — Proponer y comparar una mejora de v0.2
- **State:** DONE
- **Depends on:** OPT-003, OPT-004, OPT-005
- **Owner:** orchestrator
- **Acceptance:** Una variable, beneficio y regresiones medibles; aprobación de diseño e implementación por separado; sin release ni deploy.
- **Evidence:** Candidata preflight aprobada: replay de 3 registros reales exit0 → BLOCKED; 2 probes reales → READY, sin llamadas nuevas al modelo. 13 tests del gate, 29 Python, Go/vet/mod/format, smoke/rollback PASS. CI código helper f2b2c2d: 41906029549 cinco jobs SUCCESS. Informe, JSON y A/B validados; revisión independiente de registros y cierre documental VERIFIED; credenciales temporales retiradas y hashes personales intactos.
- **Blocker:** none
