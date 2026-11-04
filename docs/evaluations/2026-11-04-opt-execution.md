# Chuli v0.1.3 — Ejecución de OPT-001 a OPT-006

**Fecha:** 2026-11-04. **Alcance aprobado:** ejecutar las seis OPT sobre copias de laboratorio; preparar y comparar una mejora sin cambiar modelos/agentes de producción, publicar releases o desplegar.

El preflight quedó implementado y comprobado. Se ejecutaron siete escenarios nuevos, con una interrupción y reinicio reales para recuperación. La revisión independiente confirmó registros consistentes y sin PASS inventados. El resultado de estas siete ejecuciones es **6 PASS, 1 PARTIAL**; los otros cuatro casos no se repitieron en MXC.

El acumulado por última evidencia disponible es **8 PASS y 3 PARTIAL**. Combina dos entornos documentados; no es una corrida uniforme ni una puntuación global. La [baseline anterior](2026-11-04-baseline.md) permanece como registro histórico.

## Entorno y solución del bloqueo

Se mantuvieron código/instrucciones v0.1.3 desde `68b756bd5e39539b4e277bb490321f8c06870755`, CLI 0.166.0, Windows y catálogo `cc0d2af804909ee1e10c99826b5de8979591b048c3d6e984f2de5dcfd0a286f1`. Prompts y criterios no cambiaron.

La matriz de diagnóstico sin modelo mostró:

| Configuración de laboratorio | Resultado observado |
| --- | --- |
| Legacy `sandbox_mode=workspace-write`, incluso con proyecto confiable | Prompt efectivo read-only |
| `default_permissions=:workspace` | Prompt efectivo workspace-write |
| Sandbox de escritura predeterminado / unelevated | Timeout; no se aprobaron esos intentos |
| Perfil nombrado + Windows MXC | Lectura, escritura en copia y cuatro tests exitosos; escritura fuera y bajo read-only denegada |
| Piloto por herramientas reales del agente | R0 inspecciona sin mutar; R1 cambia solo el literal y ejecuta cuatro tests |

La receta modificó únicamente homes descartables: perfiles nombrados, `windows.sandbox=mxc` y confianza limitada a cada fixture conocido. No se investigó ni cambió la política global para forzar los flags legacy. El diagnóstico prueba la receta en este host; no establece una causa universal del comportamiento legacy. La [documentación oficial](https://learn.chatgpt.com/docs/permissions) advierte que los perfiles y flags antiguos no deben mezclarse.

Cada caso usó nueva copia, home y sesión, con apps/hooks/memoria automática y web desactivados. Solo Engram tuvo un MCP configurado, sobre datos ficticios. Las configuraciones y hashes completos permanecen privados. Se retiraron las credenciales temporales al finalizar y se contrastaron los hashes personales auditados.

## Resultados por escenario y procedencia

| Escenario | Último resultado | Procedencia y límite |
| --- | --- | --- |
| R0 | PASS | Nueva ejecución MXC; citas concretas, snapshots idénticos y límites honestos |
| R1 | PASS | Nueva ejecución MXC; solo literal, pruebas originales intactas y cuatro tests del agente |
| R2 descubrimiento | PARTIAL | Reutilizado del primer entorno; entrevista inicial sin respuestas humanas ni cobertura completa de reservas/excepciones |
| R2 arquitectura | PASS | Reutilizado; propuesta preliminar y criterios de medición, sin auditoría de código |
| R2 implementación | PARTIAL | Nueva ejecución; código, seis tests y oráculo contractual pasan; registro TTP no canónico |
| UX/UI | PASS | Reutilizado; flujos/estados/accesibilidad propuestos, sin UI implementada o QA visual |
| TDD/QA | PASS | Nueva ejecución; RED real, patch posterior, GREEN de seis tests y hashes de tests originales intactos |
| TTP recuperación | PASS | Nueva ejecución; proceso realmente interrumpido, otra thread, cinco tests y tablero válido |
| Engram/TKA | PASS | Nueva ejecución de la rúbrica: ADR A30 prevalece y no inventa aislamiento. No leyó contenido de memoria |
| MRP routing | PASS | Nueva ejecución; localiza funciones, justifica inline y no inventa modelo efectivo |
| R2+C | PARTIAL | Reutilizado; límites declarativos/técnicos y clasificación aún incompletos |

`extended-run.json` registra solo las siete nuevas ejecuciones: 6 PASS, 1 PARTIAL y cuatro casos ausentes NOT_TESTED. `cumulative-run.json` reutiliza explícitamente los otros cuatro resultados anteriores: 8 PASS, 3 PARTIAL. Ambos pasan `chuli eval report`; la comparación advierte configuración no homogénea. Las mejoras de ejecución tras recuperar herramientas no se atribuyen a razonamiento o routing mejores.

## Implementación, TDD y recuperación

En implementación y TDD, la traza muestra RED por `return_events=2` frente a `1`, modificación posterior de `loans.py` y GREEN de seis pruebas. Un oráculo del observador corroboró prestatario/admin y rechazo sin mutación de terceros antes y después de devolver. No se evaluó concurrencia real.

TDD agregó `tests/test_return_loan.py` y conservó el SHA-256 del archivo original. El agente explicó que esa comprobación no demuestra aislamiento técnico entre escritor y pruebas. No se presenta como TDD estricto con barreras de escritura independientes.

Para recuperación se terminó el primer proceso después de una mutación real de `app.py`, antes de `turn.completed`. Otro proceso creó una thread distinta, sin recibir el transcript; contrastó spec, diseño, tablero y diff, preservó el código y cerró con cinco pruebas. El tablero final pasa el validador. No se verificó reapertura manual de Desktop ni recuperación tras compactación de una funcionalidad extensa.

## Memoria ficticia y autoridad

Engram 3.4.0 se ejecutó con `ENGRAM_DATA_DIR` explícito fuera de la instalación principal. El observador creó dos proyectos: una nota histórica de A con 14 días y una regla independiente de B con 90. Las búsquedas por proyecto devolvieron solo su observación; otro directorio de datos con A vacío no encontró esas memorias.

El MCP ofrecido al agente expuso exactamente cinco herramientas lectoras: context, search, get_observation, current_project y stats. El agente solo usó `mem_current_project`; eligió los 30 días del ADR de A desde archivos y dijo que `process_override` no acredita aislamiento. Los conteos de observaciones ficticias quedaron en 1 por proyecto antes y después.

Esto separa tres evidencias: precedencia documental verificada por el agente, configuración/stores/proyectos contrastados por el observador y recuperación de contenido de memoria por el modelo **no verificada**. No prueba una barrera del sistema operativo frente a otras bases ni se hicieron operaciones sobre la memoria personal.

## Modelos, delegación y consumo

Las ocho sesiones de los siete casos nuevos —dos corresponden a recuperación— registran Sol/medium en `turn_context`. Es metadato del cliente, no una atestación del backend físico. No hubo delegaciones observadas en los escenarios; el caso MRP justifica resolver inline por tamaño y coordinación. No hay muestra para comparar modelos de Explorer, Architect, Designer, Test Engineer, Worker o Reviewer.

La implementación/revisión del helper sí se delegó durante esta auditoría, fuera del benchmark: propiedad de dos archivos para el worker y reviewer sin escritura. La revisión detectó dos falsos READY reproducibles, corregidos con RED/GREEN. Sus modelos efectivos, tokens y eficiencia comparativa no se infieren del rol configurado.

Los siete `turn.completed.usage` nuevos registran 645488 input, de los cuales 494592 cached input; 8960 output y 305 reasoning output. Se conservan las categorías del cliente, sin sumarlas como categorías disjuntas ni convertir suscripción a precios API. El consumo de la primera sesión de recuperación interrumpida no se estima. No se deduce eficiencia comparando escenarios bloqueados con implementaciones efectivas.

## Mejora comparada: preflight offline

La candidata fue el helper ya aprobado en OPT-001, no un cambio de agentes o modelos. El A/B usa los mismos registros capturados y cambia únicamente la presencia del gate:

| Muestra | Control instrumental: exit 0 | Con preflight |
| --- | --- | --- |
| R0 bloqueado anterior | Elegible por código de proceso | BLOCKED por falta de probes |
| Reintento R0 anterior | Elegible por código de proceso | BLOCKED por falta de probes |
| R1 bloqueado anterior | Elegible por código de proceso | BLOCKED por permisos y probes |
| Probes reales de lectura | — | READY |
| Probes reales de escritura/pruebas | — | READY |

El control exit-only **no describe un PASS histórico**: el observador anterior ya había clasificado correctamente los pilotos como PARTIAL. La mejora automatiza ese descarte antes de ampliar el benchmark. El replay no invocó modelos ni permite calcular tokens o dinero ahorrados.

El helper exige procedencia y marcadores en líneas completas, y bloquea capacidades contradictorias con permisos. Tiene 13 tests específicos y 29 Python de repositorio. No autentica fuentes; conserva el límite `EVIDENCE_NOT_AUTHENTICATED`. Reversión: retirar el helper/test y la sección de preflight; el CLI Go de Foundation permanece independiente.

## Hallazgo pendiente y propuesta siguiente

**P2 — Interoperabilidad del tablero TTP.** El agente de implementación conservó evidencia y dependencias, pero usó T1/T2/T3 y encabezados/campos que `taskboard.py` no acepta. La instalación de laboratorio contiene el validador, pero no la plantilla/protocolo que el agente intentó encontrar. La recuperación sí pasó porque su fixture ya aportaba un tablero canónico. Es un hallazgo de esta muestra; no se demuestra una regresión frente a una ejecución anterior, donde implementación estaba NOT_TESTED.

Se propone revisar el [diseño de distribución de plantilla TTP](../specs/chuli-v0.2-evaluation-next.md#hallazgo-posterior--distribución-de-la-plantilla-ttp). Está fuera de las seis OPT ejecutadas y no se aplicó a producción ni se alteró la muestra. No hay evidencia para escalar modelos o reescribir MRP.

## Verificación, entrega y privacidad

Go test/vet/mod/format, Python (29) tests, validación de estructura, smoke y regresiones de instalación/rollback Windows pasaron. El código del helper `f2b2c2d` pasó los cinco jobs de [CI 41906029549](https://github.com/AlejoPalavecino/codex-engineering-harness/actions/runs/41906029549), incluido Go en Windows/Linux/macOS. La revisión independiente contrastó hashes, respuestas, secuencias, contadores y tableros; no certifica backend, Desktop ni aislamiento global.

Entrega local: `codex/v0.2-baseline-plan`. Rama remota de revisión/CI: `feat/evaluation-preflight-v0.2` —el workflow actual ejecuta pushes `feat/**`—. No se fusionó esta continuación a main ni se publicó una release. Main conserva la integración de PR #5 y v0.1.3 sigue estable.

La evidencia completa vive en el directorio privado indicado por `evals/reports/validation-2026-11-04/private-root.txt`, ignorado por Git. `opt-execution/` contiene JSON, comparación, probes, logs, métricas y hashes; `cases/*-mxc` conserva prompts, eventos, snapshots y respuestas. Los resultados publicables omiten rutas personales, credenciales y trazas completas. [Registro de las seis tareas](../../odd/tasks/chuli-v0.2-optimizations.md).
