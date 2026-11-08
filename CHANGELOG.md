# Changelog

## [1.0.0] — Perfil personal para Codex

- Manager orientado a instalar y mantener el perfil del usuario. El paquete del perfil y los binarios publicados provienen del CI del commit etiquetado.
- Orquestación ODD/TDD, seis especialistas y 23 skills con inventario, atribución y referencias instalables.
- Actualización del perfil, conservación del contexto personal y recuperación desde el paquete oficial.
- La [trayectoria](docs/ROADMAP.md) registra la sustitución del diseño de runtime separado.

## [0.2.1] — Mantenimiento integrado en main, incluido en 1.0.0

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

## [0.2.0] — 2026-11-06

### Added
- Evaluation Protocol v1.0 con 11 escenarios sintéticos R0/R1/R2/C, UX/UI, TDD, TTP, Engram y MRP; baseline anclado a v0.1.3.
- Biblioteca Go de catálogo/rúbricas, huella reproducible, snapshots SHA-256 y reportes comparativos derivados de observaciones por criterio.
- Comandos `chuli eval check|list|fingerprint|snapshot|diff|assert-unchanged|report`.
- Preflight offline de procedencia, permisos declarados, probes y eventos; distingue READY, BLOCKED y NOT_VERIFIABLE.
- Plantilla y guía TTP instaladas junto a taskboard.py, con manifiesto, hashes, preservación de personalizaciones, backup y rollback.
- Formulario de issues, catálogo de etiquetas y plantilla de PR para revisión trazable.

### Limitations
- El evaluador no ejecuta Codex ni valida por sí solo modelos, permisos efectivos o sandbox.
- Las muestras conductuales corresponden a Codex CLI en Windows; no certifican Codex Desktop, éxito universal, ahorro de tokens ni aislamiento global.
- El preflight no autentica fuentes y las comprobaciones TTP verifican estructura, no la autenticidad de pruebas declaradas.
- Los binarios Linux/macOS se compilan; la instalación y el rollback funcionales se validaron solo en Windows.

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
