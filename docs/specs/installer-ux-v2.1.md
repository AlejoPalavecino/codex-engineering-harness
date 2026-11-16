# Especificación: Chuli UX de Instalación, Multi-Plataforma, Versiones, Perfiles, MCPs y Skills

## Objetivo
Simplificar el flujo de instalación y administración en la terminal CRT de Chuli Manager para que cualquier persona pueda instalar, actualizar y gestionar **OpenAI Codex (`codex`)**, **Claude Code (`claude`)** y **AGY CLI (`agy`)** de forma guiada y sin ruido visual técnico, manteniendo la identidad visual retro (PC + wordmark `CHULI` en bloques y paleta CRT) y corrigiendo las incompatibilidades detectadas en AGY y PowerShell.

## Criterios de aceptación
1. **Arreglos base AGY y PowerShell (SHA-256 portable):**
   - `scripts/client_targets.py` renderiza `plugins/chuli/agents/*.md` con `model` válido para AGY (`pro`, `flash`, `flash_lite` o `inherit`) manteniendo `settings.json` con los nombres de display; `agy plugin validate ... --strict --json` pasa con `0` errores y los `6` subagentes cargados.
   - En instalaciones reales de AGY (`~/.gemini/config` con `~/.gemini/antigravity-cli` existente), el modelo activo del CLI se sincroniza de forma segura con `~/.gemini/antigravity-cli/settings.json` (con respaldo y rollback), preservando sandboxes aislados en `home/settings.json`.
   - `scripts/profile.ps1`, `scripts/install.ps1`, `scripts/verify.ps1`, `scripts/rollback.ps1` y `scripts/install-manager.ps1` calculan SHA-256 mediante `[Security.Cryptography.SHA256]` de .NET sin depender de `Get-FileHash`.
2. **Motor Manager (autodetección, releases, versiones y aplicación local):**
   - `internal/manager` expone autodetección de clientes instalados (`codex`, `claude`, `agy`: presencia de directorio/binario y versión registrada en `state.json`).
   - `Latest` y `ListReleases` obtienen las releases estables de GitHub incluyendo su resumen de novedades (`body`) y fecha, permitiendo avisar si hay una release más nueva para el perfil y/o para el binario de `chuli`.
   - Soporta instalar o cambiar a una versión específica (incluido downgrade explícito con respaldo previo) y aplicar cambios de perfil/skills localmente usando la caché verificada o `--source` sin obligar a descargar `latest`.
3. **Perfiles de AI, MCPs multi-cliente y gestión de Skills:**
   - Perfiles de AI propios nombrados por proveedor (`codex`, `claude`, `agy`), guardados en `UserConfigDir()/chuli/ai-profiles.json`, editables y aplicables localmente.
   - `internal/integrations` soporta `codex` (`config.toml`), `claude` (`~/.claude.json`) y `agy` (`mcp_config.json`), manteniendo preservación atómica, registro de reversión y soporte para agregar/quitar MCPs personalizados (URL remota o comando local) sin exponer secretos.
   - Gestión de Skills por plataforma: listar skills instaladas/disponibles con su descripción (`SKILL.md`), activar/desactivar skills por plataforma registrando `disabled_skills` en el manifiesto sin romper `verify.ps1`/`doctor`, y agregar/quitar skills propias.
4. **Nueva TUI CRT guiada (identidad preservada, instalación sin ruido visual):**
   - Mantiene intacto `renderChuliLogo` (PC retro + `CHULI`) y la responsividad en `40×24`, `80×24` y `120×32`.
   - Muestra notificación en la cabecera cuando el chequeo en segundo plano detecta una nueva release.
   - Menú principal de 6 opciones: **Instalar**, **Actualizar**, **Gestionar versiones**, **Perfiles de AI**, **MCPs** y **Skills** (con accesos secundarios `D` verificar, `?` acerca de, `q` salir).
   - **Instalar** guía en 4 pasos: selección múltiple de plataformas (`AGY`, `Claude`, `Codex`) → preset o personalización por rol → confirmación humana limpia → **animación CRT por etapas que oculta el log crudo** (con detalle técnico opcional ante error o tecla explícita) → pantalla de éxito con próximos pasos.
   - **Actualizar** muestra las versiones instaladas, resumen de novedades de la release y aviso si hay nuevo binario Manager antes de aplicar con animación limpia.
5. **Documentación y validación:**
   - Documentos (`README.md`, `docs/TUI.md`, `docs/INSTALLATION.md`, `docs/AGENTS.md`) actualizados y suites Go, Python, PowerShell y `validate_repository.py` en verde.
