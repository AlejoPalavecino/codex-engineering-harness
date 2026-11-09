# Diseñar desde una tarea

Aplicar junto a [harness-ux-ui](../SKILL.md). Resolver lo conocido leyendo el proyecto; preguntar solo decisiones pendientes que cambien el diseño. Una corrección pequeña no necesita todos estos artefactos.

1. Precisar persona y contexto, objetivo, punto de entrada y permisos. Describir qué necesita decidir o completar y cómo reconocer el resultado.
2. Ordenar información y acciones por pasos. Definir acción principal, salida/cancelación, confirmación cuando corresponda y recuperación sin pérdida inesperada de datos.
3. Cubrir estados pertinentes: inicio/vacío, carga, error de campo o red, éxito y acceso denegado. No inventar permisos ni reglas. Los textos explican qué pasó y cómo seguir.
4. Para flujos nuevos, presentar esquema y dirección visual antes de implementar: jerarquía, tipografía, colores por función, espaciado, componentes y adaptación a dispositivos. Justificar decisiones en lenguaje cotidiano y obtener la aprobación aplicable.
5. Reutilizar stack, componentes, identidad y convenciones. Si falta dirección, usar una biblioteca revisada disponible como apoyo opcional; no mezclar prescripciones estéticas incompatibles ni tomar sugerencias como aprobación. No instalar herramientas o dependencias por esta guía.
6. Registrar lo recuperable en la especificación o diseño ya existentes. No generar un MASTER.md o prototipos alternativos por obligación; ofrecer variantes cuando resuelvan una decisión real.

## Entrega para revisión

Por tarea: persona/permiso, inicio y datos sintéticos, acciones, resultado esperado, estados a provocar, dispositivos/tamaños justificados y criterio observable. Incluir navegación, teclado, foco, legibilidad y feedback según riesgo. Continuar con la [guía de navegador](../../harness-tdd-qa/references/browser-review.md).

Ejemplo: «Desde préstamos activos, devolver L-01; enviar sin código muestra un error que explica cómo corregirlo; tras corregir, la devolución aparece una sola vez y el resultado queda visible». La prueba debe comprobar el efecto, no solo la presencia del botón.

## Comprensión humana

Proponer a una persona representativa completar la tarea sin indicar dónde pulsar; observar dudas, errores y recuperación. Registrar contexto, consentimiento aplicable y límites sin recolectar datos personales innecesarios. Si no se hizo, indicar que la comprensión humana sigue pendiente. No presentar un recorrido automatizado como investigación con usuarios.
