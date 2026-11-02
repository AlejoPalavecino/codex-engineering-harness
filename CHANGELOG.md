# Changelog

## [0.1.1] — 2026-11-02

### Added
- TTP v1.0: tareas persistentes por funcionalidad con dependencias, aceptación, evidencias y estados.
- `update_plan` habilitado desde el instalador, sin depender del formato de visualización de Codex Desktop.
- Plantilla de tareas y validador/actualizador opcional en Python estándar.
- Pruebas de transiciones, bloqueo, recuperación del instalador y config nativa.

### Limitations
- No incluye panel de tareas propio ni sincronización automática con GitHub Projects.
- El aislamiento de agentes y la representación real de la UI deben comprobarse en cada instalación.

## [0.1.0-rc.1] — 2026-11-02

### Added
- Protocolos EWP, MRP, ASEP, TKA y CES con responsabilidades explícitas.
- Orquestación global, cuatro subagentes y cinco skills propias.
- Instalación Windows PowerShell con dry-run, backups, verificación y reversión.
- Inventario de MCP recomendado sin credenciales ni instalación automática de terceros.
- Documentación para contribuciones y validaciones de estructura.

### Not yet guaranteed
- Aislamiento estricto real de permisos entre subagentes.
- Selección automática del modelo efectivo en todas las versiones de Codex.
- Instalador multiplataforma o actualizador de componentes externos.
