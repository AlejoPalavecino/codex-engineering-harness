# Chuli v0.2 — Próxima etapa de evaluación y optimización

**Estado:** diseño aprobado y ejecución de OPT-001 a OPT-006 autorizada por el usuario el 2026-11-04. Candidatas sustanciales de producción aún no definidas requieren revisión específica.
**Fuente:** [baseline del 2026-11-04](../evaluations/2026-11-04-baseline.md), protocolo [Evaluation](../EVALUATION.md) y pedido del usuario del 2026-11-04.

La primera mejora propuesta es comprobar las capacidades efectivas del cliente antes de consumir una ejecución conductual. Todavía no hay evidencia suficiente para cambiar modelos, roles o contratos de producción.

## Decisiones y fronteras

- Foundation está integrada en `main` por PR #5. Su rollback es revertir `8b15e082fd9716172238da6a756da704e5e9f6e4`.
- v0.1.3 / `68b756bd5e39539b4e277bb490321f8c06870755` sigue como referencia y última release estable.
- Conservar EWP, CES, TTP, TKA, MRP, ASEP, instalador, rollback y Chuli Manager.
- Excluir cambios globales personales, memorias reales, datos productivos, publicación de v0.2 y despliegues.
- No modificar prompts ni rúbricas para conseguir mejores resultados. Toda corrección futura de fixtures exige nueva huella y declaración de comparabilidad.
- El usuario aprobó el diseño y autorizó ejecutar las seis OPT. Las dependencias y la evidencia controlan las transiciones; una candidata sustancial de producción requiere revisión específica. Registro: [tareas](../../odd/tasks/chuli-v0.2-optimizations.md).

## OPT-001 — Preflight de capacidades y procedencia

**Problema y evidencia:** los pilotos emitieron `turn.completed` y exit 0 aun cuando todas las lecturas fueron rechazadas. R1 solicitó `workspace-write`, pero `turn_context` registró `read-only`. El primer home temporal también recibió una advertencia sobre helpers bajo el temporal del sistema. Estos resultados no justifican cambiar instrucciones o modelos de Chuli.

**Propuesta recomendada:** agregar un procedimiento local de diagnóstico a la guía de evaluación, acompañado por un comprobador pequeño de los metadatos de la sesión piloto si se aprueba. Separar ejecución del proceso, disponibilidad de herramientas y resultado de la rúbrica. No crear un orquestador nuevo ni cambiar políticas del host.

| Alternativa | Beneficio | Costo / límite |
| --- | --- | --- |
| Checklist manual documentado | Mínima superficie y fácil reversión | Depende del observador |
| Checklist + comprobador local de metadatos | Detecta discrepancias repetibles; recomendada | Hay que mantener el contrato de eventos por versión |
| Runner integral nuevo | Más automatización | Complejidad injustificada por esta muestra; diferir |

**Componentes posibles:** `docs/EVALUATION.md`, `docs/TESTING.md` y, solo si se aprueba, un helper bajo `scripts/` con pruebas sintéticas. No cambiar agentes, modelos ni configuración personal.

**Criterios de aceptación:**

1. Antes del benchmark queda registrado cliente, versión, commit evaluado, huella y overrides; nunca contenido de credenciales o configuración personal.
2. Se contrasta permiso solicitado con permiso observado. Si difieren, se declara el bloqueo y se evita iniciar escenarios que dependen de esa capacidad.
3. Un probe descartable demuestra lectura y, cuando corresponda, escritura/pruebas. Exit 0 sin herramientas exitosas no satisface el gate.
4. Un cliente sin metadatos suficientes queda NO VERIFICABLE; no se relajan sus restricciones ni se simulan aprobaciones.
5. Evidencia privada, referencias sanitizadas y cleanup de la credencial temporal quedan comprobados.
6. Pruebas del comprobador cubren eventos ausentes, errores, salida de proceso exitosa con herramientas bloqueadas y discrepancias de permisos.

**Riesgos:** atribuir diferencias del host al harness; depender de eventos incompletos; confundir políticas declaradas con aislamiento efectivo. Mitigar con versión fijada y verificación de comportamiento. Reversión: quitar únicamente guía/helper nuevos.

**Comparación antes/después:** reproducir los eventos sintéticos de los pilotos bloqueados. Antes se consumió una ejecución completa; después el diagnóstico debe detectar el bloqueo antes de ampliar la evaluación. Si cambia el host, registrar esa variable; no atribuir la diferencia a una mejora de razonamiento.

## OPT-002 / OPT-003 — Completar la baseline

**Problema:** lectura/escritura bloqueadas impidieron medir implementación, TDD y routing; las respuestas de diseño no prueban fidelidad al código.

**Propuesta:** resolver la capacidad del cliente únicamente en la instalación descartable o por intervención humana fuera de este alcance; repetir piloto R0/R1 una vez y después los escenarios pendientes. Conservar la ejecución del 2026-11-04 como evidencia histórica, sin reemplazarla por una corrida favorable.

**Aceptación:** misma huella, prompts y criterios; sesiones independientes; snapshots/diffs y pruebas observadas; tablas por criterio; resultados no ejecutables en NOT_TESTED. Aprobaciones conversacionales reales antes de cualquier fase que las requiera. Separar datos CLI de Desktop y versión oficial de instalación personalizada.

**Alternativa:** ejecución manual guiada en Desktop con intervención del usuario. No es equivalente al CLI: requiere registro propio. **Riesgo:** confundir el cambio de cliente/configuración con una optimización de Chuli.

## OPT-004 — TTP, TKA y memoria ficticia

**Problema:** los tests deterministas de TTP no prueban recuperación de una sesión ni aislamiento de Engram.

**Propuesta:** una interrupción real antes de cerrar TASK-002, nueva sesión sin transcript previo y contraste con spec/diseño/tablero/diff; prueba documental A/B y luego instancia Engram aislada si el usuario aprueba su preparación.

**Aceptación:** registrar las dos sesiones e interrupción; conservar IDs y dependencias; no repetir tareas verificadas; DONE exige pruebas. ADR vigente de A prevalece sobre historia y B. No consultar ni escribir memoria real. Si no hay instancia aislada, integración Engram queda NOT_TESTED. Reversión: descartar solo entornos sintéticos identificados.

## OPT-005 — Delegación y routing basados en evidencia

**Problema:** no se observaron delegaciones de los siete roles en el benchmark. Los TOML y los metadatos de herramientas no prueban por sí solos qué modelo ejecutó un subagente.

**Propuesta:** observar delegaciones justificadas durante escenarios útiles; registrar propósito, contexto, alcance, herramientas, evidencia, duplicación y resultado. No invocar todos los roles solo para completar una tabla.

**Aceptación:** modelo configurado separado de observado; falta de evento = NO VERIFICABLE; tokens/reintentos solo desde fuente identificada; sin costos de API derivados de suscripción. Cualquier cambio de modelo/esfuerzo/contrato requiere diseño aprobado y comparación de una variable, con el mismo fixture y sin regresiones críticas.

**Alternativa recomendada mientras faltan datos:** conservar routing vigente. **Riesgo:** optimizar una muestra bloqueada o confundir una preferencia con una mejora comprobada.

## OPT-006 — Optimización candidata y gate de entrega

Elegir una sola mejora después de OPT-002 a OPT-005. Documentar problema reproducible, beneficio esperado, componente, alternativa de no cambiar, riesgo y prueba A/B. No se define ahora una reescritura de MRP ni un cambio de agentes porque no existe evidencia suficiente.

**Aceptación:** propuesta revisada por el usuario, implementación autorizada por separado, criterios críticos sin regresiones y todas las omisiones visibles; Go/Python/instalador/rollback/CI siguen verdes. Publicación y despliegue siguen excluidos.

### Resultado de las seis OPT

La candidata elegida para OPT-006 fue el preflight de OPT-001 ya aprobado: comparación offline sobre los mismos registros, sin cambios de modelos ni nuevas llamadas de benchmark. [Resultados y límites](../evaluations/2026-11-04-opt-execution.md). Las siete ejecuciones nuevas completan la cobertura conductual disponible; las cuatro observaciones anteriores se conservan separadas en el acumulado.

## Hallazgo posterior — Distribución de la plantilla TTP

**Estado:** propuesta nueva pendiente de revisión; no incluida en la autorización de las seis OPT. No se implementó ni se corrigió el tablero del benchmark.

**Problema:** implementación produjo un tablero legible con evidencia, pero no aceptado por `taskboard.py`. El home instalado entrega el validador y carece de la plantilla/protocolo que el agente buscó. Recuperación funcionó con un tablero canónico preexistente. Evidencia: caso `r2-implementation-mxc`, log del validador exit 2 y caso `ttp-recovery-mxc` exit 0.

**Alternativas:** documentar cómo obtener la plantilla, incorporar plantilla y guía como assets propios administrados por el instalador, o convertir tableros manualmente. Se recomienda evaluar assets propios con referencia de uso y validación antes de DONE; convertir la muestra para aprobarla queda excluido.

**Superficie posible:** manifiesto/instalador/rollback, plantilla existente, documentación y guía de uso; cualquier referencia nueva en instrucciones requiere aprobación específica. No cambiar modelos, estados ni esquema de datos de TTP.

**Riesgos:** sobrescribir personalizaciones, crear otro formato o ampliar indebidamente el paquete. Mantener backup/preservación/rollback existentes y usar una sola plantilla canónica.

**Aceptación propuesta:** instalación descartable entrega plantilla y herramienta; modelo encuentra el recurso; tablero generado pasa check con IDs/dependencias/evidencia; no se cierra antes de probar; R0/R1 siguen proporcionales. Comparar una variable —disponibilidad/guía del recurso— con el mismo prompt/fixture/rúbrica, conservar ambos resultados y no atribuir el efecto al modelo. Reversión: quitar solo esos assets/referencias y comprobar rollback en copia sintética.
