# Changelog

## [2.1.2] — 2026-11-19

- En Codex, la instalación y la actualización muestran los modelos y esfuerzos de los siete roles en AGENTS.md también cuando coinciden con el preset equilibrado.

## [2.1.1] — 2026-11-19

- TTP y la skill de especificación identifican la raíz del perfil de Codex, Claude Code y AGY para encontrar plantillas y ejecutar el validador.
- Antigravity carga completa la entrada de diagram-design y conserva la guía original junto a sus referencias. Codex y Claude Code reciben la distribución original.

## [2.1.0] — 2026-11-19

- Chuli revisa criterios ambiguos antes de planificar e identifica las decisiones necesarias para probarlos.
- El tablero permite consultar trazabilidad, impacto por archivo y decisiones de arquitectura. Los tableros anteriores siguen siendo compatibles.
- Los controles de calidad detectan evidencia vencida por cambios en archivos o dependencias. Los tres clientes comparten el flujo Git y GitHub.
- Se incorporan controles de revisión de seguridad, guías de operación y evaluación, y publicación desde el artefacto exacto validado por CI.

## [2.0.3] — 2026-11-17

- El preset equilibrado de Codex usa GPT-6.2 Astra para orquestación y arquitectura, y GPT-6.2 Sol para los demás roles.
- Claude Code usa los alias Opus, Sonnet y Haiku. Al configurar los modelos explícitamente, Chuli retira de `settings.json` la variable global que podía anular la elección por rol.
- AGY combina Orchestrator (Argon) con los niveles nativos `pro`, `flash` y `flash_lite`. El README detalla las asignaciones de los tres clientes.

## [2.0.2] — 2026-11-17

- Las tareas sustanciales se organizan por comportamiento y el cierre de un tablero nuevo vincula la prueba al contenido de los archivos declarados.
- La delegación identifica la pregunta, el trabajo independiente y la evidencia requerida. El comparador offline contrasta tres recorridos e informa fricción y calidad por separado.
- El selector descubre modelos en las cachés de Codex y Claude Code y en AGY CLI. Se incorporan GPT-6.2 Sol y Astra al catálogo de respaldo.
- El asistente informa el resultado de cada plataforma y permite retomar una instalación tras un conflicto con Gentle. Se corrigen los rótulos de `status` y una prueba de codificación en Windows.
- La distribución incluye la skill `frontend-design`. El README ilustra la arquitectura multicliente y el recorrido ODD.

## [2.0.1] — 2026-11-16

- Consola CRT renovada en `chuli` preservando el banner retro de la PC y bloque `CHULI`, con menú de 6 accesos directos (**Instalar**, **Actualizar**, **Gestionar versiones**, **Perfiles de AI**, **MCPs** y **Skills**) y alerta no bloqueante cuando hay una nueva release.
- Asistente guiado de instalación en 4 pasos (`1/4` detección y selección de plataformas, `2/4` preset o configuración de modelos proveedor por proveedor y rol por rol con flechas y `Enter`, `3/4` confirmación y `4/4` animación limpia que oculta los logs crudos del instalador).
- Pantalla **Actualizar** con resumen de novedades de la release y búsqueda cancelable con `Esc`/`q`, junto con **Gestionar versiones** para cambiar o restaurar (`rollback`) versiones en caché o publicadas.
- Gestión de **Perfiles de AI** guardados por proveedor en `~/.chuli/`, soporte multi-cliente para servidores **MCP** (incluidos servidores personalizados en Codex, Claude Code y AGY) y gestión de **Skills** con activación/desactivación y alta de skills propias.
- Corrección de tiers de subagentes AGY (`pro`, `flash`, `flash_lite`, `inherit`), sincronización de `~/.gemini/antigravity-cli/settings.json` con backup/rollback y cálculo SHA-256 portátil entre PowerShell 5.1 y PowerShell 7+.

## [2.0.0] — 2026-11-16

- Soporte multi-cliente completo para **OpenAI Codex (`~/.codex`)**, **Claude Code (`~/.claude`)** y **AGY CLI (`~/.gemini/config`)** desde una única fuente canónica (`config/models.json` y `scripts/client_targets.py`).
- Catálogo actualizado de modelos y 9 presets rápidos: familia **GPT-6.1 (`Sol`, `Terra`, `Luna`, `Astra`)**, **Claude 5.5 (`Opus 5.5`, `Sonnet 5.5`, `Haiku 5.5`) + `Fable 5.2`**, y **Gemini 4 (`Carbon`, `Argon`, `Hydrogen`, `Orchestrator`)**.
- Configurador interactivo de modelos y nivel de esfuerzo rol por rol en la consola CRT (`Modelos y presets`) y subcomando CLI `chuli [--client codex|claude|agy] models` / `install --preset ... --role ROL=MODELO[:ESFUERZO]`.
- Rediseño visual de la consola CRT con monitor enmarcado, selector rápido de cliente (`Tab` / `1`/`2`/`3`) y tarjeta de equipo activo en terminales anchas.
- Preservación de personalizaciones, validación estructural (`verify.ps1`) y `rollback.ps1` byte a byte para los tres clientes (`config.toml`, `settings.json` con soporte UTF-8 BOM, `AGENTS.md`, `CLAUDE.md` y plugin nativo `plugins/chuli/` en AGY).

## [1.1.2] — 2026-11-16

- Instalación multiplataforma con un comando mediante `go install github.com/AlejoPalavecino/codex-engineering-harness/cmd/chuli@v1.1.2` (o `@latest`) y detección automática de la versión del módulo desde `runtime/debug.ReadBuildInfo()`.
- Fijación de `LF` para `tests/fixtures/**` en `.gitattributes` y ejecución de `tests/validate_repository.py` en el job de Windows en CI para asegurar validación nativa con `core.autocrlf=true`.
- Reintentos acotados en `scripts/install-manager.ps1` (`Restore-ManagerFile`) ante bloqueos transitorios `IO.IOException` durante el rollback, con detalle de error en advertencia y prueba determinista en `tests/windows-manager-install.ps1`.
- Limpieza de `.github/workflows/release.yml` retirando el disparador residual `workflow_run` y el valor por defecto `'1.0.0'`, conservando únicamente `workflow_dispatch`.

## [1.1.1] — 2026-11-09

- Instalador opcional de Chuli Manager para Windows por usuario: descarga oficial, SHA-256, registro de `chuli` en PATH y vista previa/confirmación (#58, issue #57).
- Preservación del PATH y archivos existentes, rechazo de destinos ambiguos y recuperación ante fallos; pruebas en PowerShell 7 y Windows PowerShell 5.1.
- `install-chuli-windows.ps1` incluido como archivo descargable de la release con su checksum, generado desde la fuente revisada por CI. Los ZIP/TAR portables permanecen disponibles.
- Sin instalación permanente automática en la PC del usuario ni cambios al perfil Codex al ejecutar la vista previa.

## [1.1.0] — 2026-11-09

- Asistente opcional de integraciones en TUI y CLI: Context7, Engram, Exa, GitHub y Egolite; vista previa, confirmación, comprobación y reversión de conexiones propias (#52–#56, issue #51).
- Consola CRT: marca en texto inspirada en el logo, verde fósforo/beige, explicación contextual, anchos adaptables y contenido paginado con destino completo.
- Diagram Design 2.6 fijada a 6a7f19f: 293 recursos originales y licencia MIT, integridad verificable y activación condicional para documentación. Catálogo de 24 skills.
- README ampliado, logo retro y diagramas de arquitectura y recorrido con fuentes HTML/SVG.
- Pruebas del asistente con perfiles temporales, handshake MCP sin tools/call y compatibilidad de fixtures con macOS. Sin nuevos modelos ni permisos automáticos.
- El perfil se actualiza desde el Manager existente; el diseño CRT y el asistente requieren descargar el ejecutable v1.1.0. No hay autoactualización del ejecutable.

## [1.0.1] — 2026-11-09

- Preguntas enfocadas en decisiones pendientes, investigación del contexto disponible y continuidad de la autorización hasta verificar el resultado (#45, issue #47).
- Voz argentina más cercana y explicaciones cotidianas antes de introducir conceptos técnicos cuando aportan aprendizaje (#46, issue #48).
- Diseño UX/UI por tareas y guion de revisión con egolite: apariencia, interacción y resultado comprobados por separado; referencias portables y fixture sintético (#50, issue #49).
- Escenarios complementarios separados del catálogo histórico. Su entrega no acredita por sí sola comportamiento observado del modelo ni comprensión humana.
- Distribución por el CI del SHA final, con perfil y Managers nativos versionados y sumas SHA-256. Sin nuevos modelos, agentes, dependencias ni cambios de permisos.

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
