# Diseñar tareas y comprobar interfaces con egolite

- Aprobación: el usuario autorizó la primera entrega recomendada en esta conversación el 2026-11-09.
- Alcance: evolucionar harness-ux-ui y harness-tdd-qa, referencias instalables y un escenario sintético de formulario, errores y teclado. Sin nuevos agentes, stack impuesto, instalación de terceros o cambios del perfil personal. El usuario autorizó crear el issue y abrir la PR de esta unidad el 2026-11-09.
- Release: incluida en el alcance de v1.0.1 por decisión del usuario el 2026-11-09, junto con las mejoras de preguntas/autorización y voz del orquestador. Esta decisión no autoriza merge ni publicación de la release.
- Seguimiento: [issue #49](https://github.com/AlejoPalavecino/codex-engineering-harness/issues/49); cierre al integrar la PR elegido por el usuario.

## Diseño del procedimiento

Definir persona, tarea, información, acción principal, inicio, resultado y recuperación antes de elegir componentes. Respetar diseño existente y obtener aprobación de flujos/dirección visual sustanciales antes de implementar. Conservar decisiones en los documentos existentes; crear solo el mínimo recuperable en R2.

Usar las skills externas como apoyo opcional revisado; elegir una fuente de dirección visual por tarea. Egolite es la preferencia para la revisión web si está disponible y autorizado. Su guía instalada determina las APIs: no se incluye ni se instala el navegador con Chuli.

La revisión distingue apariencia, interacción y resultado observable. Comprobar tareas y estados relevantes con datos de prueba, capturas acotadas, teclado y tamaños justificados. Corregir dentro del alcance y repetir comprobaciones afectadas; ante falta de progreso, revisar diagnóstico sin iteración indefinida. Sin navegador, declarar lo no comprobado y usar una alternativa autorizada si existe.

La inspección por un agente no demuestra comprensión humana ni conformidad completa de accesibilidad. Para validar comprensión, proponer observar a una persona representativa completar la tarea sin indicarle dónde pulsar; registrar dudas y límites sin inventar investigación.

## Aceptación

| ID | Resultado verificable |
| --- | --- |
| UX-01 | La skill define tarea, información, acción, resultado y recuperación antes del aspecto visual. |
| UX-02 | Conserva aprobación de flujos/UI, stack y diseño existentes, y proporcionalidad R1/R2. |
| UX-03 | La revisión prefiere egolite y consulta su guía real; no instala ni presume APIs o permisos. |
| UX-04 | El guion separa apariencia, interacción y resultado; cubre errores, teclado y tamaños pertinentes. |
| UX-05 | Correcciones requieren nueva evidencia afectada; bloqueos se declaran sin aprobar por inferencia. |
| UX-06 | Guías y enlaces funcionan en el paquete instalado; no alteran el catálogo histórico. |
| UX-07 | Existe un escenario reproducible y se distingue su ejecución manual de la evaluación conductual del orquestador y la comprensión humana. |

## Riesgos y reversión

Evitar estética arbitraria, puntuaciones sin fundamento, datos reales y capturas usadas como prueba funcional. No fijar reglas de una biblioteca como requisitos universales de accesibilidad. Sin migración: distribución de referencias por el inventario existente, con preservación de personalizaciones y rollback habituales. Verificar los recursos en perfil temporal; no tocar el perfil real.

Revertir solo las modificaciones de ambas skills, docs/protocols/SKILLS.md y docs/EVALUATION.md y retirar referencias, fixture, test, escenario, especificación y tablero de esta unidad. Conservar las otras PR. Evidencia y estado: odd/tasks/task-based-ux-review.md.
