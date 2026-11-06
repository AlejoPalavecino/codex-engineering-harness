# Recursos TTP: distribución y validación

Diseño e implementación autorizados el 2026-11-05. Candidato de código: `cd89872175facc383fa9b061b3233617a0c83a91`, sobre `edf2d4f`. No constituye una nueva release ni una certificación de v0.1.3. [Especificación](../specs/ttp-resources.md) · [tareas](../../odd/tasks/chuli-ttp-resources.md).

## Cambio entregado

El instalador distribuye las fuentes canónicas en `CODEX_HOME/harness/templates/feature-tasks.md` y `harness/docs/protocols/TTP.md`, junto al validador existente en `harness/bin/taskboard.py`. Usa el mismo manifiesto, hashes, preservación, backups y rollback. Las instrucciones y la skill de especificación indican dónde leer los recursos y validar el tablero antes del cierre. El enlace relativo guía/plantilla funciona tanto en el repositorio como instalado. R0/R1 siguen exentos de crear tableros.

No cambian estados, formato, validador ni modelos. Se corrigió además una incompatibilidad preexistente de `verify.ps1` con el paso de comillas a Python en Windows PowerShell 5.1, reproducida en el código anterior.

## Evidencia determinista

Contrato test-first preparado por test engineer independiente: Python RED por referencias ausentes y seis escenarios PowerShell RED por recursos ausentes. El autor corrigió posteriormente la captura de stderr del test en PS5.1; conservó las aserciones de códigos, hashes y snapshots.

| Verificación | Resultado |
| --- | --- |
| Python completo (`unittest discover`, 32 tests) y estructura | PASS |
| Seis escenarios de recursos en PowerShell 5.1 | 6/6 PASS |
| Los mismos escenarios en PowerShell 7 | 6/6 PASS |
| Smoke y regresión Windows existentes | PASS |
| Go tests, vet, dependencias y formato | PASS |
| Upgrade desde instalación real de `edf2d4f`, verify PS5.1 y rollback | PASS |

Los seis escenarios cubren dry-run sin escritura; instalación con hashes/manifiesto; plantilla aceptada por el validador instalado; reinstalación y doble rollback; preservación de personalizaciones; reemplazo con backup y restauración; y rechazo de rollback ante deriva sin cambios parciales. Incluyen archivos ajenos. El upgrade real restaura hashes y manifiesto previos y retira los recursos nuevos.

[CI del candidato de código](https://github.com/AlejoPalavecino/codex-engineering-harness/actions/runs/42141105008): cinco jobs SUCCESS, Go en Windows/Linux/macOS y contratos del instalador en ambos motores PowerShell. La revisión independiente también ejecutó ambos motores y verificó el candidato.

## Generación observada en sesiones reales

Tres sesiones nuevas de Codex CLI 0.166.0, Windows, perfil MXC y modelo observado en metadatos `gpt-6.1-sol`, con el paquete instalado realmente en homes descartables. Se mantuvieron fixture, prompt, criterios y catálogo de evaluación. Los snapshots previos coinciden con sus controles históricos. Sin editar los resultados para hacerlos aprobar.

| Caso | Evidencia observada | Resultado |
| --- | --- | --- |
| R2 implementación | Leyó plantilla/guía instaladas; generó tres tareas TASK-001/002/003 con metadatos, dependencias, aceptación y evidencia. Dos checks del agente y un check observador con el validador instalado: exit 0. Seis tests PASS y comprobación adicional de idempotencia/autorización. | PASS |
| R0 lectura | Snapshots idénticos; auditoría fundamentada y límites explícitos; sin tablero. | PASS |
| R1 cambio mínimo | Solo el literal autorizado, tests originales intactos; comprobación del agente con Python/diff; cuatro tests posteriores del observador PASS; sin tablero ni especificación. | PASS |

El tablero R2 quedó en DONE con evidencia RED/GREEN registrada. El agente corrigió un valor inicial inválido de Approval antes del cierre; el resultado final coincide por hash con el snapshot. El observador no modificó ese archivo. La revisión independiente concluyó VERIFIED sin bloqueantes para este alcance. El control histórico `r2-implementation-mxc` produjo T1/T2/T3 y fue rechazado (exit 2); permanece conservado. El validador tiene el mismo contenido, normalizando saltos de línea. La nueva observación muestra que el paquete permitió generar un tablero canónico; una comparación histórica y una muestra no demuestran causalidad universal ni ahorro de tokens.

Los otros ocho escenarios quedan NOT_TESTED para esta candidata. No se atribuyen a ella resultados anteriores ni se repite el benchmark completo. La prueba del instalador es Windows; el CI multiplataforma de Go no demuestra instalación en otros sistemas.

## Privacidad, reversión y entrega

Logs, sesiones, snapshots, hashes, reportes de protocolo y comparación permanecen fuera de Git. La copia temporal del login fue autorizada; todas las copias `auth.json` dentro del laboratorio se retiraron después de ejecutar los casos. Los hashes de los doce archivos personales controlados permanecen iguales. No se consultaron ni escribieron memorias reales.

Reversión de una instalación: rollback existente con su manifiesto; comprobado también ante personalización y deriva. Reversión del código: retirar las entradas de assets y referencias añadidas en la unidad `6eea748`, conservando separada la corrección de compatibilidad `cd89872`. Sin merge, release ni instalación en el home personal.
