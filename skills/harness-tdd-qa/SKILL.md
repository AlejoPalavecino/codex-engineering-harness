---
name: harness-tdd-qa
description: "Aplicar test-first cuando corresponde y seleccionar verificaciones según riesgo: unitarias, contratos, E2E, seguridad, accesibilidad, lint y CI; evidencia por resultado."
---

# TDD & Quality

1. Leer criterios de aceptación y contratos aprobados.
2. Elegir pruebas que detecten reglas de negocio y regresiones reales; no inflar la suite por estética.
3. En test-first: crear escenario de comportamiento, ejecutar y documentar fallo RED por la razón esperada, implementar GREEN, refactorizar con la suite verde.
4. Si ASEP exige aislamiento, separar ejecutor de pruebas e implementador mediante entornos/permisos efectivos. Un prompt no alcanza.
5. Si una prueba aprobada está mal, detenerse para revisión; el implementador no puede modificarla para autoaprobarse.
6. Verificar types, lint, build, integración, E2E, CI, seguridad, accesibilidad y rendimiento según corresponda.
7. Registrar comandos y resultados; declarar límites y omisiones. No afirmar «verificado» si existen bloqueantes.

## Conexión con seguimiento TTP

Devolvé resultados de pruebas, comandos y evidencia por ID de tarea. Cuando la implementación finalice, el orquestador pasa la tarea a VERIFYING; solo después de satisfacer sus criterios pasa a DONE y se registra evidencia. Si falla, mantener IN_PROGRESS o BLOCKED con motivo. El verificador no debe editar directamente el registro de tareas, y el cierre no depende de lo que afirma el implementador.
