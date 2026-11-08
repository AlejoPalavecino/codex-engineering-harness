# Trayectoria de Chuli

## Decisiones vigentes en 1.0

| Decisión | Motivo y consecuencia |
| --- | --- |
| Instalar en el perfil local habitual de Codex | La CLI y el modo Codex local de la app pueden consumir el mismo perfil. El usuario continúa iniciando su cliente de la forma habitual. |
| Usar la TUI para administrar versiones | Importa, verifica y revierte recursos. El perfil incluye las instrucciones y los roles que el propio Codex ejecuta en sus sesiones. |
| Adoptar ODD proporcional y TDD aplicable | Conserva el tablero TTP para trabajos recuperables. Los casos pequeños se resuelven con el contexto mínimo y una comprobación enfocada. |
| Fijar el catálogo de Gentle AI en v4.4.0 | Las skills, recursos, licencias y adaptaciones se pueden auditar por versión. Una actualización de upstream entra mediante revisión y pruebas. |
| Mantener modelos por rol y respetar la configuración personal | El orquestador puede delegar con objetivos concretos. Se distingue el modelo configurado del efectivamente observado en una ejecución. |

## Hitos

| Etapa | Resultado |
| --- | --- |
| 0.1–0.1.2 | Protocolos EWP, MRP, ASEP, TKA, CES, tablero TTP, instalación con respaldo y reversión. |
| 0.1.3 | Primer Manager en Go y distribución de binarios. |
| 0.2.0 | Evaluador de once escenarios, preflight, recursos TTP y evidencia histórica de Codex CLI en Windows. |
| 0.2.1 | El [PR #14](https://github.com/AlejoPalavecino/codex-engineering-harness/pull/14) integró la recuperación del inventario, verificación obligatoria y control del ejecutor de rollback. Esos cambios forman parte de 1.0. |
| Propuesta 0.3 | El [PR #15](https://github.com/AlejoPalavecino/codex-engineering-harness/pull/15) documentó un runtime separado y el lanzamiento de Codex por etapa. La decisión de producto 1.0 sustituyó esa dirección. La investigación de permisos conserva valor para escenarios estrictos. |
| 1.0.0 | Perfil ODD instalable, seis especialistas, catálogo completo, paquete versionado, recuperación local y distribución ligada al CI. [Seguimiento de entrega](https://github.com/AlejoPalavecino/codex-engineering-harness/issues/17). |

## Próxima evidencia

Medir cambios de comportamiento y consumo comparando por tarea la calidad de aceptación, las regresiones, los reintentos, las delegaciones y los tokens efectivamente observados. Registrar el entorno y el modelo antes de atribuir una mejora al perfil. Las campañas históricas conservan sus propias versiones y límites.
