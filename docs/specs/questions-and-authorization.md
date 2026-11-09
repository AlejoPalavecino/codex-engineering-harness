# Preguntas y continuidad de la autorización

## Decisión aprobada

El usuario aprobó las reglas, los ocho criterios y su implementación en esta conversación el 2026-11-09. Chuli investiga lo recuperable, pregunta por decisiones relevantes pendientes y continúa hasta verificar el resultado autorizado.

## Reglas y diseño de interacción

Antes de preguntar, consultar instrucciones aplicables, decisiones vigentes y código relevante. Preguntar si falta una decisión de alcance, comportamiento, datos, permisos, UX o una acción irreversible, o si hay una contradicción relevante que la precedencia de fuentes no resuelve. Investigar datos técnicos obtenibles y resolver detalles técnicos reversibles siguiendo las convenciones del proyecto.

Cada pregunta identifica la decisión pendiente y qué cambia con la respuesta. Agrupar decisiones relacionadas, recomendar una opción fundamentada cuando corresponda y evitar preguntas ya resueltas. Una preferencia opcional puede consultarse sin bloquear trabajo independiente.

La autorización de ejecución cubre implementación, pruebas, correcciones y verificación del resultado acordado. Reabrir únicamente decisiones afectadas por cambios relevantes de alcance, decisiones pendientes o acciones con autorización específica. Conservar la distinción entre aprobar diseño y autorizar ejecución. Una respuesta parcial habilita solo lo resuelto; el silencio no aprueba. Mientras se espera, avanzar únicamente con trabajo independiente y autorizado.

## Aceptación

| ID | Resultado observable | Escenario |
| --- | --- | --- |
| AC-01 | Un texto autorizado se modifica y comprueba sin reconfirmación ni documentos ceremoniales. | Q-01 |
| AC-02 | Una eliminación sin política de conservación definida provoca una pregunta antes de implementar esa conducta. | Q-02 |
| AC-03 | Una convención documentada se encuentra y aplica sin volver a consultarla. | Q-03 |
| AC-04 | Una función se ubica según la estructura existente sin pedir aprobación del detalle reversible. | Q-04 |
| AC-05 | Un fallo causado por el cambio se corrige y vuelve a verificar sin reconfirmar el mismo alcance. | Q-05 |
| AC-06 | Un contrato compartido no autorizado se consulta antes de modificarlo. | Q-06 |
| AC-07 | Una respuesta parcial o ausente no aprueba lo pendiente; continúa únicamente trabajo independiente autorizado. | Q-07 |
| AC-08 | Una contradicción relevante no resuelta se presenta con sus fuentes y la decisión afectada. | Q-08 |

El procedimiento reproducible está en [los escenarios](../evaluations/questions-and-authorization.md). Inspección documental y checks estructurales comprueban la entrega de las reglas, no el cumplimiento conductual del modelo.

## Alcance, riesgos y reversión

Se modifican el orquestador, Discovery, EWP y la referencia de ODD; se agrega evaluación complementaria sin cambiar el catálogo histórico. Modelos, roles, permisos, instalador y contratos de delegación quedan fuera de esta unidad. No hay decisiones de producto pendientes para implementar estas reglas.

El riesgo principal es interpretar una decisión de producto como detalle técnico, o ampliar una autorización por inferencia. Los escenarios incluyen ambos límites. No hay migración de formato ni actualización del perfil personal en esta entrega. La instalación posterior usa el mecanismo existente de vista previa, preservación de personalizaciones y rollback; no se afirma una instalación ejecutada por estos cambios.

Para revertir esta unidad, restaurar los cambios de config/AGENTS.md, skills/harness-discovery/SKILL.md, docs/protocols/EWP.md, docs/protocols/ODD.md y docs/EVALUATION.md, y retirar los nuevos documentos de especificación, evaluación y seguimiento. No revertir modificaciones ajenas.
