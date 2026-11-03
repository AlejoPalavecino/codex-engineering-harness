# Changelog

## [Unreleased] — v0.1.2 candidate

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
