# Especificación: Chuli v2.0 · Soporte Multi-Cliente, Selector de Modelos y TUI Visual

## Objetivo
Extender Chuli a **OpenAI Codex (`codex`)**, **Claude Code (`claude`)** y **AGY CLI (`agy`)** desde una única fuente canónica, permitiendo elegir presets rápidos o personalizar el modelo y nivel de esfuerzo de cada uno de los 7 roles (`orchestrator` + 6 especialistas), con una consola CRT más visual y expresiva.

## Criterios de aceptación
1. **Catálogo `config/models.json`:** Define los 7 roles y los 3 clientes (`codex`, `claude`, `agy`) con sus modelos actuales (`GPT-6.1 Sol/Terra/Luna/Astra`, `Claude 5.5 Opus/Sonnet/Haiku + Fable 5.2`, `Gemini 4 Carbon/Argon/Hydrogen + Orchestrator`), niveles de esfuerzo y 3 presets por cliente (`balanced`, `max`, `fast`).
2. **Adaptación de archivos por cliente:**
   - `codex` (`~/.codex`): `AGENTS.md`, `config.toml`, `agents/*.toml`, `skills/**` y `harness/**`.
   - `claude` (`~/.claude`): `CLAUDE.md`, `settings.json` (preservando UTF-8 BOM y claves existentes), `agents/*.md` (frontmatter YAML con `model`, `effort`, `disallowedTools`) y `skills/**`.
   - `agy` (`~/.gemini/config`): `plugins/chuli/plugin.json`, `plugins/chuli/rules/AGENTS.md`, `plugins/chuli/agents/*.md` (`subagent: true`, `mainAgent: false`, `model`, `effort`, `tools`), `plugins/chuli/skills/**` y `settings.json` con formato de modelo AGY.
3. **Preservación, verificación y rollback:** `install.ps1`, `verify.ps1` y `rollback.ps1` operan sobre `-Client codex|claude|agy`, `-Preset` y `-RoleModelsJson`, respaldando con SHA-256 y restaurando byte a byte.
4. **TUI CRT visual y CLI v2.0:** Selector de cliente (`1`/`2`/`3` o `Tab`), pantalla interactiva de modelos y presets por rol (`chuli models` y menú TUI), estética CRT mejorada que respeta los límites de `40×24`, `80×24` y `120×32`.
