# Trayectoria de Chuli

## Decisiones vigentes (v2.1.2)

| Decisión | Motivo y consecuencia |
| --- | --- |
| Instalar en el perfil nativo de cada cliente (`~/.codex`, `~/.claude`, `~/.gemini/config`) | Un mismo núcleo de protocolos, roles y skills se proyecta al formato de OpenAI Codex, Claude Code y AGY CLI desde `config/models.json` y `scripts/client_targets.py`. Cada persona sigue trabajando en su cliente habitual. |
| Usar Chuli Manager (CLI + TUI CRT en Go) para gestionar el ciclo de vida | Descarga, verifica por SHA-256, instala, actualiza, diagnostica y revierte perfiles, además de administrar modelos por rol, perfiles de AI guardados, integraciones MCP multi-cliente y skills activas. |
| Adoptar ODD proporcional, TDD aplicable y TTP trazable | Los cambios chicos se resuelven con contexto mínimo y prueba focalizada. El trabajo sustancial conserva tablero en `odd/tasks/<slug>.md` con trazabilidad (`REQ-###`), impacto por archivo, quality gates y huellas SHA-256. |
| Fijar y auditar el catálogo de 25 skills | Cinco skills propias, 18 de Gentle AI v4.4.0, Diagram Design 2.6 (293 archivos pineados en `UPSTREAM.json`) y Frontend Design (Anthropic). Cualquier actualización upstream entra con revisión y pruebas. |
| Mantener modelos configurables por rol y preservar personalizaciones | El orquestador y los seis especialistas se configuran por preset, perfil guardado o descubrimiento local del cliente. Se distingue el modelo configurado del efectivamente observado en una sesión. |
| Publicar desde el artefacto exacto validado por CI | El workflow de validación verifica seguridad, estructura, instalador multiplataforma y checksums; el despacho manual publica únicamente ese candidato aprobado. |

## Hitos

| Etapa | Resultado |
| --- | --- |
| 0.1–0.1.2 | Protocolos EWP, MRP, ASEP, TKA, CES, tablero TTP, instalación con respaldo y reversión prevalidada. |
| 0.1.3 | Identidad Chuli, primer Manager TUI en Go (Bubble Tea v2) y compilación multiplataforma. |
| 0.2.0–0.2.1 | Evaluador offline de 11 escenarios, preflight, recursos TTP, inventario `managed_files` y recuperación segura de rollback ([PR #14](https://github.com/AlejoPalavecino/codex-engineering-harness/pull/14)). |
| Propuesta 0.3 (sustituida) | El [PR #15](https://github.com/AlejoPalavecino/codex-engineering-harness/pull/15) exploró un runtime separado por etapa; la decisión 1.0 priorizó instalar sobre el perfil local. Sus hallazgos sobre permisos siguen vigentes en [ASEP](protocols/ASEP.md). |
| 1.0.0–1.0.1 | Perfil ODD instalable, seis especialistas, catálogo completo, voz colaborativa argentina, diseño UX/UI por tareas y distribución ligada al CI ([issue #17](https://github.com/AlejoPalavecino/codex-engineering-harness/issues/17)). |
| 1.1.0–1.1.2 | Consola CRT, asistente opcional de integraciones MCP, Diagram Design 2.6, instalador de `chuli` para Windows e instalación portable con `go install`. |
| 2.0.0–2.0.3 | Soporte multi-cliente completo (Codex, Claude Code y AGY CLI), asistente guiado en 4 pasos, perfiles de AI guardados, gestión de skills, MCPs multi-cliente, descubrimiento dinámico de modelos, comparador de fricción y Frontend Design. |
| 2.1.0–2.1.2 | Cierre del roadmap del ciclo de vida ([issue #63](https://github.com/AlejoPalavecino/codex-engineering-harness/issues/63), [tablero](../odd/tasks/roadmap-ciclo-de-vida.md)): flujo Git/GitHub compartido en los tres clientes, revisión de criterios ambiguos, trazabilidad por requisito, impacto por archivo, quality gates, revisión de seguridad, publicación del artefacto exacto de CI, guía de operación y correcciones de recursos TTP/AGY/Codex. |

## Próximos frentes y evidencia

- **Paridad y calidad de ingeniería continua:** mantener `config/models.json` como fuente única embebida en Go y Python, validar presupuestos de tamaño del paquete de perfil en CI y ampliar la cobertura funcional del instalador `pwsh` en Linux y macOS.
- **Evidencia conductual comparable:** medir cambios de instrucciones, skills o modelos comparando por tarea aceptación, regresiones, preguntas repetidas, retrabajo y tiempo hasta la primera prueba útil ([fricción del flujo](evaluations/workflow-friction.md)), registrando siempre cliente, entorno y modelo efectivamente observado.
- **Operación y evolución:** separar la evaluación del producto entregado de las comparaciones entre versiones de Chuli según [Operación y evolución](protocols/OPERATIONS.md).
