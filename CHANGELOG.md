# Changelog

## [Unreleased] — v0.2.1 (maintenance candidate)

### Fixed
- Conserva un inventario completo de hashes administrados aunque una reinstalación no modifique archivos.
- Recupera los baselines oficiales de instalaciones v0.1.3 y v0.2.0 cuyos manifiestos quedaron vacíos por la regresión de reinstalación.
- verify.ps1 comprueba la presencia de los 13 recursos obligatorios y devuelve código distinto de cero si falta alguno.
- El rollback valida versión del ejecutor, esquema del manifiesto, rutas, duplicados y JSON anterior antes de modificar archivos.
- Chuli Manager rechaza checkouts anteriores a v0.2.1 para ejecutar rollback y no delega la recuperación a scripts conocidos como inseguros.

### Added
- managed_files, schema_version y rollback_schema_version en el estado de instalación.
- Baselines SHA-256 históricos y pruebas de migración desde manifiestos v0.1.3/v0.2.0.
- Regresiones Windows para reinstalación, recursos obligatorios ausentes, backups faltantes y manifiestos anteriores corruptos.
- Criterios de compatibilidad de rollback en las pruebas Go del Manager.

### Validation scope
- CI valida estructura, pruebas Python/Go, instalador y rollback en Windows PowerShell 5.1 y PowerShell 7.
- Los baselines históricos son una migración acotada a v0.1.3 y v0.2.0; futuras versiones deben agregar sus propios hashes antes de retirar archivos del estado.
- La candidata no publica binarios ni declara soporte de instalación/rollback fuera de Windows.

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
