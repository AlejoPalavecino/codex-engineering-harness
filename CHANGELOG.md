# Changelog

## [Unreleased] — v0.2 (Evaluation Foundation / PR #5)

### Added
- Evaluation Protocol v1.0 con 11 escenarios sintéticos R0/R1/R2/C, UX/UI, TDD, TTP, Engram y MRP; baseline anclado a v0.1.3.
- Biblioteca Go de catálogo/rúbricas, huella reproducible, snapshots SHA-256 y reportes comparativos derivados de observaciones por criterio.
- Comandos `chuli eval check|list|fingerprint|snapshot|diff|assert-unchanged|report`.
- Tests de unidades Go, fixtures Python y documentación de ejecución manual segura.

### Limitations
- El evaluador no ejecuta Codex, no valida modelos efectivos por sí solo ni certifica permisos/sandbox.
- Las métricas de tokens/modelos solo se aceptan con fuente declarada; no hay baseline conductual todavía.
- Esta PR no cambia ni reemplaza MRP, agentes, skills funcionales o TTP.

## [0.1.3] — 2026-11-03

### Added
- Identidad Chuli; comando `chuli` y Manager TUI en Go (Bubble Tea v2).
- Preview/confirmación para instalar desde release estable o checkout local, diagnóstico y rollback en Codex Windows.
- Tests de unidad y compilaciones cruzadas en CI.

### Compatibility
- Repositorio GitHub y módulo Go conservan temporalmente la ruta previa.
- No se alteran agentes ni skills existentes ni se autoactualiza el binario de Chuli.

## [0.1.2] — 2026-11-02

### Fixed
- P1-1: parseo TOML real y validación previa; encabezados con comentarios no generan tablas duplicadas.
- P1-2: rollback con prevalidación de todos los backups y manifiesto anterior.
- P1-3: plan de actualizaciones explícito y preservación de archivos y claves personalizadas salvo autorización.
- P2-1: verificación distingue sintaxis y deriva de claves administradas de cambios válidos en claves ajenas.

### Added
- Regression fixtures Windows y tests de TOML sin dependencias externas de Python.
- Workflow de publicación manual por SHA aprobado y CI exitoso.

## [0.1.1] — 2026-11-02

### Added
- TTP v1.0 con tareas, dependencias y evidencias.
- Seguimiento nativo `update_plan` configurable, validador de tareas y scripts de reversión.
