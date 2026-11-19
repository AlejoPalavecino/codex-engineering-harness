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

## Revisión de código y deuda

- Identificá el candidato y los criterios. Revisá legibilidad, responsabilidades, convenciones del repositorio, contratos y manejo de errores. Vinculá cada hallazgo con su impacto y archivo/línea.
- Corré los chequeos estáticos que el proyecto ya use (lint, tipos, análisis de seguridad según riesgo). Registrá el alcance, la versión y el resultado. Si un chequeo aplicable falta o queda incompleto, indicá qué riesgo queda sin verificar.
- Derivá la regresión desde el diff, los consumidores y los casos vinculados. Priorizá el flujo más usado y los cambios que pueden afectar permisos, persistencia, contratos compartidos o producción. En éstos verificá también los rechazos y el estado preexistente.
- Para deuda que pueda diferirse, registrá impacto, probabilidad, esfuerzo y un enlace al seguimiento acordado. Priorizá por impacto y probabilidad, y usá el esfuerzo para ordenar trabajos de impacto similar. Un criterio fallido o riesgo bloqueante requiere resolución antes de declarar la entrega verificada.

## Revisión de interfaces

Para cambios de interfaz, leer el guion y los criterios aprobados y aplicar [revisión en navegador](references/browser-review.md). Preferir ego-browser/egolite si está disponible y autorizado; consultar su skill y guía instalada antes de operar. Separar apariencia, interacción y resultado observable, con datos de prueba y tamaños pertinentes. Una captura o un clic enviado no prueban que la tarea funcione. Repetir comprobaciones afectadas tras corregir; declarar lo no probado cuando falte herramienta o evidencia. No sustituir investigación con personas por opiniones del agente.

## Conexión con seguimiento TTP

Devolvé resultados de pruebas, comandos y evidencia por ID de tarea. Cuando la implementación finalice, el orquestador pasa la tarea a VERIFYING; solo después de satisfacer sus criterios pasa a DONE y se registra evidencia. Si falla, mantener IN_PROGRESS o BLOCKED con motivo. El verificador no debe editar directamente el registro de tareas, y el cierre no depende de lo que afirma el implementador.

Para un control reproducible por caso, consultá `quality-v1` en `harness/docs/protocols/TTP.md` bajo el home del perfil. Capturá `fingerprint` antes de ejecutar, consultá `impact` para seleccionar regresiones y cotejá `quality` con los resultados observados. La salida técnica de esos comandos se reporta por separado de las aprobaciones y la autorización de entrega.
