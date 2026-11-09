# EWP v1.0 — Engineering Workflow Protocol

La cantidad de código **no** determina cuánto proceso corresponde. Primero evaluar intención, incertidumbre, riesgo e interrupciones posibles.

| Ruta | Entrada | Comportamiento |
| --- | --- | --- |
| R0 (consulta) | Investigación, análisis o diseño sin autorización de modificar | Solo lectura y diagnóstico |
| R1 (directa) | Resultado claro, riesgo acotado y cambio autorizado | Explorar → implementar → comprobar |
| R2 (guiada) | Idea nueva, decisiones importantes pendientes, trabajo sustancial | Relevamiento → diseño → especificación → desarrollo → verificación → entrega |
| C (criticidad) | Riesgo significativo de seguridad, datos, contratos o producción | Controles reforzados **sobre R1 o R2** |

## Relevamiento R2

Comprobar problema, éxito esperado, actores, frontera dentro/fuera, procesos, datos, validación, permisos, reglas, integraciones, restricciones no funcionales, UX y operación. Preguntar por bloques que resuelvan dependencias; no repetir datos conocidos ni preguntar por detalles irrelevantes.

## Decidir cuándo preguntar

Antes de consultar al usuario, revisar instrucciones aplicables, decisiones vigentes y código relevante. La lista de relevamiento identifica vacíos; no obliga a entrevistar sobre todos los temas.

| Situación | Acción |
| --- | --- |
| Dato técnico obtenible inspeccionando o probando | Investigar. |
| Detalle técnico reversible dentro del alcance autorizado | Resolver siguiendo convenciones del proyecto. |
| Decisión pendiente de alcance, comportamiento, datos, permisos, UX o acción irreversible | Preguntar antes del trabajo dependiente. |
| Contradicción relevante no resuelta por precedencia de fuentes | Mostrar las fuentes y pedir resolver la decisión afectada. |
| Preferencia opcional que mejoraría el resultado | Puede consultarse sin bloquear trabajo independiente. |

Cada pregunta explica qué decisión falta y qué cambia con la respuesta. Agrupar decisiones relacionadas, recomendar una opción fundamentada cuando corresponda y evitar preguntas ya respondidas. Un cambio observable de producto no se convierte en detalle técnico por ser fácil de implementar.

## Puertas humanas

Requieren aprobación: alcance, stack principal, arquitectura importante, modelo de datos relevante, flujos y dirección UI, ampliación de alcance y despliegue. No confundir aprobación de diseño con permiso para implementarlo. Agrupar aprobaciones relacionadas cuando sea claro.

La autorización de ejecución vigente cubre implementación, pruebas, correcciones y verificación necesarias para alcanzar el resultado acordado. No pedir reconfirmación por avanzar de etapa o corregir un fallo propio dentro de ese alcance. Reabrir únicamente decisiones afectadas por cambios relevantes de alcance, decisiones pendientes o acciones que requieren autorización específica. No extender una autorización a un contrato compartido no autorizado.

Una respuesta parcial habilita solo lo resuelto. El silencio no es aprobación. Mientras se espera, continuar únicamente con trabajo independiente y autorizado; mantener pendiente lo que depende de la respuesta.

## Etapas

1. **Discovery & Scope:** entender problema y aprobar alcance.
2. **Solution & Experience Design:** comparar soluciones, stack, datos y UX/UI; aprobar lo relevante.
3. **Specification & Planning:** definir contratos, tareas y aceptación verificable.
4. **Engineering:** implementar unidades de comportamiento con pruebas pertinentes.
5. **Verification & Validation:** evidencia técnica y validación de utilidad/UX.
6. **Delivery:** preparar entrega; despliegue solo con autorización.

No imponer documentos en R1. La etapa R2 se documenta cuando no pueda recuperarse desde el diff y las decisiones. La prueba crítica fallida impide declarar `VERIFIED`. La autorización pendiente impide avanzar sobre el trabajo dependiente; un modelo no puede aprobar en nombre del usuario.

La selección de pruebas, unidades y cortes de revisión se detalla en [ODD](ODD.md).

## Recuperación y cambios

Si cambia alcance o contrato, reabrir únicamente lo afectado. Estado posible: DRAFT, NEEDS_APPROVAL, READY_TO_BUILD, IMPLEMENTING, VERIFYING, READY_FOR_DELIVERY, BLOCKED. Los estados son conceptuales: esta versión no incluye motor formal que los imponga.
