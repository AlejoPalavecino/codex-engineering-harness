# Operación y evolución después de la entrega

Usá esta guía cuando necesites detectar un problema en servicio, sostener una entrega o comprobar si un cambio resuelve la necesidad que lo originó. Retomá la versión, requisitos y pruebas de la entrega y registrá los hallazgos donde el equipo ya mantiene su historial.

1. Identificá el flujo, el entorno, la versión desplegada y el resultado esperado. Para un incidente activo, priorizá el impacto y la mitigación según el plan autorizado del servicio.
2. Revisá la evidencia disponible. Marcá cada afirmación como **observada** con su fuente, **hipótesis** con una forma de comprobarla, o **no disponible** con el dato que falta. Un cero requiere medición válida.
3. Registrá la decisión, quién se ocupa y cómo se verificará. Usá [EWP](EWP.md) para elegir R0, R1 o R2 y sumar criticidad cuando corresponda. Vinculá el cambio con el registro del proyecto y usá [TTP](TTP.md) para el trabajo R2 autorizado que requiera seguimiento.

## Detectar problemas

Elegí señales por flujo crítico. Para cada alerta acordá la fuente, ventana, umbral, responsable y respuesta inicial. Justificá el umbral con un objetivo de servicio o evidencia disponible. En un entorno de prueba, provocá un fallo conocido y verificá que llegue el aviso esperado y que la recuperación lo resuelva. Registrá por separado una caída del sistema de medición y una caída del servicio.

| Señal | Qué registrar y revisar |
| --- | --- |
| Logs | Eventos estructurados con hora y zona, entorno, versión, operación, resultado y código de error. Usá un identificador opaco de correlación si existe. |
| Métricas | Volumen, errores, latencia y saturación según el riesgo. Usá dimensiones acotadas como operación y versión. Revisá el denominador y si faltaron datos durante la ventana. |
| Trazas | Propagá la correlación por los componentes relevantes. Revisá duración y estado del tramo que falla. Registrá el muestreo si afecta la interpretación. |
| Detección | Usá señales del impacto percibido, controles sintéticos o reportes de soporte. Documentá la cobertura y las dependencias que podrían quedar fuera. |

## Resolver un incidente

Usá el registro de incidentes del equipo. Para un incidente pequeño alcanza con estos campos en el registro existente.

| Campo | Qué dejar asentado |
| --- | --- |
| Identidad e impacto | ID, responsable, entorno, hora o ventana, versión afectada, flujo, alcance observado y alcance pendiente de confirmar. Referenciá requisito, release o cambio relacionado si está identificado. |
| Cronología | Detección, mitigación, recuperación y verificación con hora y referencias. Identificá la versión o configuración de cada intervención. |
| Causa | Qué se confirmó y con qué evidencia. Para cada hipótesis abierta, dejá la comprobación pendiente. |
| Acciones | Mitigación, corrección y prevención con responsable, estado, fecha objetivo o revisión, y evidencia de aceptación. Vinculá el PR, issue o prueba cuando exista. |

Comprobá la recuperación con una señal ligada al impacto y, cuando sea viable, una prueba de regresión sobre el recorrido afectado. Registrá la versión verificada, ventana y límites. Las acciones preventivas que queden abiertas conservan responsable y referencia tras el cierre del incidente.

## Mantener y retirar funcionalidades

Priorizá con evidencia de impacto, criticidad, frecuencia, esfuerzo y reversibilidad. El issue, decisión o tablero existente puede guardar problema, versión y consumidores afectados, requisito, responsable, comprobación y ruta de reversión.

| Trabajo | Evidencia antes de cerrar |
| --- | --- |
| Dependencias | Revisá compatibilidad, soporte y avisos a partir de la versión instalada y la fuente del proveedor. Probá contratos y recuperación con la candidata. |
| Deuda técnica | Vinculá la deuda con una limitación o riesgo observado. Fijá una revisión y un resultado verificable. |
| Refactorización | Fijá el contrato a preservar y ejecutá las mismas pruebas relevantes antes y después. Registrá qué cambio de mantenibilidad se midió. |
| Retirada | Inventariá consumidores, contratos y datos. Acordá aviso, transición, tratamiento de datos y fecha. Verificá la alternativa y el plan de reversión antes de retirar. |

## Comprobar calidad y valor del producto

Para cada medida registrá definición, fuente, versión, población o conjunto medido, ventana y exclusiones. Si comparás períodos, revisá diferencias de tráfico, exposición, configuración o método que afecten la conclusión. Elegí las que respondan a una decisión del equipo.

| Qué interesa | Ejemplo de medida y contexto |
| --- | --- |
| Defectos | Defectos confirmados y deduplicados por versión y severidad. Explicitá qué canales o población se monitorearon. |
| Regresiones | Casos antes verificados que fallan en la candidata bajo la misma prueba. Indicá cuántos casos previos se volvieron a ejecutar. |
| Cobertura pertinente | Criterios o riesgos críticos comprobados del total identificado. Enumerá los pendientes y marcá la cobertura de líneas por separado si se mide. |
| Mantenibilidad | Fricción observada en un cambio repetible, como módulos que hay que tocar para extender el mismo contrato. Registrá el recorrido y las condiciones. |
| Estabilidad | Errores del flujo sobre intentos válidos o disponibilidad durante una ventana con medición válida. Informá volumen y brechas de observación. |

La evaluación de producto (PDLC) parte de la necesidad, la población y el cambio esperado. Antes de medir, acordá un resultado observable, fuente, línea base, objetivo y ventana. Usá una señal de resultado y, si el cambio puede perjudicar otro recorrido, una señal de resguardo. Combiná datos de uso autorizados con pruebas de tarea o retroalimentación según la pregunta. Registrá versión, muestra, sesgos y decisión humana de mantener, ajustar o retirar. Si falta evidencia, dejá abierta la hipótesis. Pasar las pruebas técnicas acredita los criterios ejecutados.

## Evaluar una versión de Chuli

Para el arnés, registrá baseline y candidata por versión o commit. Mantené equivalentes el escenario, prompt, fixture, rúbrica, cliente y versión, entorno, herramientas, permisos observados y modelo efectivo. Identificá qué cambió y ejecutá cada variante en una sesión nueva. Revisá primero aceptación y regresiones. Después contrastá retrabajo, preguntas, tiempo o costos según la evidencia disponible. El piloto y el comparador instalados describen cómo registrar condiciones incompatibles y métricas desconocidas. [Comparación de fricción del flujo](../evaluations/workflow-friction.md)

Conservá resultados del producto y del arnés en registros distintos. La primera evaluación fundamenta decisiones sobre el producto. La segunda fundamenta decisiones sobre instrucciones, skills o flujos de Chuli. Usá tiempos, tokens y costos sólo cuando exista una fuente observada con unidad y método. La campaña sintética comprueba el procedimiento y sus herramientas.

## Privacidad y límites

Usá entornos y cuentas de prueba por defecto. Acordá autorización específica antes de acceder a producción o a datos sensibles. Revisá acceso, retención y borrado con las reglas del proyecto. Limitá la recolección a lo necesario y excluí payloads, credenciales, identificadores directos y transcripciones privadas de logs, trazas, issues y reportes compartidos. Conservá evidencia cruda autorizada en el destino privado acordado y compartí referencias sanitizadas. Los identificadores opacos también requieren cuidado cuando permiten vincular registros. Chuli distribuye la guía y el comparador offline. La recolección, las alertas, la autenticidad de la evidencia y la intervención operativa dependen de las herramientas y permisos del proyecto.

## Ejercicio sintético · devolución repetida

Desde el repositorio fuente, copiá `evals/fixtures/loans` a un directorio descartable. Registrá el commit de origen con `git rev-parse HEAD` y ejecutá estos comandos en la copia.

```powershell
python -B -m unittest discover -s tests
@'
from loans import checkout, return_loan

loan = checkout("X1", "u1", asset_status="available", user_active=True)
for attempt in range(2):
    result = return_loan(loan, "u1", "member")
    print(result["return_events"])

assert loan["return_events"] == 1, loan["return_events"]
'@ | python -B -
```

- **Versión e impacto observado.** En el fixture de origen, la suite existente pasa. En un préstamo de prueba, dos devoluciones autorizadas producen los recuentos `1` y `2`. La aserción falla contra el criterio 2 de `docs/SPEC_APPROVED.md` en la copia.
- **Causa y alcance.** La segunda llamada incrementa `return_events` aunque el estado ya es `returned`. La frecuencia real, el número de personas afectadas y la utilidad del flujo quedan como no disponibles. Un reintento tras una respuesta demorada es una hipótesis que requiere evidencia propia.
- **Acción y cierre.** El responsable técnico vincula el cambio y la prueba de regresión con el requisito y la versión candidata. La corrección se acepta cuando la primera y la segunda devolución autorizada imprimen `1`, la persona ajena sigue recibiendo `LoanError` sin mutación y la suite pasa. En un incidente real queda por comprobar la recuperación durante la ventana acordada con evidencia autorizada.
- **Validación posterior.** Para evaluar el producto, acordá con las personas responsables del proceso cómo comprobar que completan la devolución con reintentos y que pueden operar con un único evento registrado. Para evaluar Chuli, usá el piloto de fricción y verificá que una métrica ausente quede `UNKNOWN`, y que un fixture distinto produzca `NOT_COMPARABLE` sin deltas para ese recorrido. Son comprobaciones independientes.
