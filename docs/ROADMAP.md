# Roadmap de Chuli

## v0.1 — Fundamentos reutilizables
- Protocolos, agentes, cinco skills originales, instalador Windows, verificación y documentación.
- Preservación de MCP y memoria existentes.

## v0.1.1 — Seguimiento de trabajo
- TTP, tablero persistente, dependencias y evidencia.
- Integración con update_plan cuando está disponible.

## v0.1.2 — Instalación y rollback seguros
- Validación TOML, preservación de personalizaciones, backup y rollback prevalidado.

## v0.1.3 — Identidad y TUI
- Marca Chuli, Manager Go/Bubble Tea, releases multiplataforma y binarios con SHA-256.
- Instalación de Codex validada en Windows.

## v0.2.0 — Evaluation Foundation
- 11 escenarios, rúbricas, snapshots SHA-256, registros de observación, reportes comparativos y preflight offline.
- No certifica comportamiento de modelos ni aislamiento efectivo.

## v0.2.1 — Mantenimiento y confiabilidad
- Inventario persistente de hashes, migración de manifiestos históricos, verificación bloqueante y rollback compatible.
- Integrada en main el 2026-11-07 mediante PR #14 y commit 7c15a83. Tag/release oficial pendiente.

## v0.3.0 — Aislamiento efectivo (M0 en investigación)
- Permission Profiles nativos de Codex y CODEX_HOME separado.
- Procesos principales separados por nivel de confianza; subagentes heredan política.
- Git Worktrees, tests congelados, reviewer read-only y verifier en copia temporal.
- Probes reales de filesystem, apply_patch, procesos hijos, red y MCP.
- CI y matriz de plataformas basada en backends efectivamente probados.
- No crear sandbox propio ni prometer aislamiento a partir de prompts.

## v0.4 — Adaptadores y operación
- Evaluar otros agentes/proveedores después de estabilizar el contrato neutral de políticas.
- Mejoras de mantenimiento, recuperación, UX del Manager y compatibilidad según demanda.

## v1.0 — Release estable
- Instalación, rollback, compatibilidad, seguridad y documentación comprobadas en entornos reales.
- Garantías respaldadas por pruebas reproducibles.
