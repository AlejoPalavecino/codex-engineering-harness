# ADR-0001 — Runtime aislado y perfiles para Chuli v0.3

- **Estado:** propuesto, pendiente de aprobación.
- **Fecha:** 2026-11-07.

## Contexto

Chuli v0.2.1 conserva agentes, skills, instalador, TTP, evaluación y Manager, pero las instrucciones de agentes no imponen por sí solas permisos de filesystem. Codex ofrece Permission Profiles nativos y los subagentes heredan la política del proceso padre.

## Decisión propuesta

1. Usar Permission Profiles nativos; no crear un sandbox propio.
2. Crear CODEX_HOME aislado en ~/.chuli/codex-home. No copiar auth.json, tokens, proveedores ni MCP personales. El usuario inicia sesión una vez en el runtime Chuli.
3. Ejecutar planner, test-author, implementation, reviewer y verifier en procesos principales separados cuando sus permisos difieran.
4. Utilizar Git Worktrees por etapa. Los agentes no escriben .git; el Manager integra después de aprobación.
5. Congelar tests aprobados por commit y SHA-256. Implementation obtiene read sobre esos paths.
6. Reviewer es read-only. Verifier ejecuta en una copia temporal, no sobre el candidato.
7. Red local deshabilitada y cero MCP por defecto.
8. Windows es la primera plataforma de aceptación; MXC o elevated solo se aceptan si pasan los probes. Unelevated no se acepta para perfiles que requieren deny-read.

## Consecuencias

**Beneficios:** permisos efectivos, separación de sesiones, preservación del Codex personal y evidencia verificable.

**Costos:** login independiente, espacio para worktrees, mantenimiento de perfiles beta y pruebas reales por backend.

## Alternativas descartadas

- Sandbox propio: duplicaría controles nativos y aumentaría la superficie de mantenimiento.
- Solo prompts/TOML por agente: los subagentes heredan la política del padre.
- Un proceso con todos los roles: no permite separar niveles de confianza.
- Solo hashes: detectan cambios, no impiden escrituras.

## Aprobación pendiente

- CODEX_HOME separado y login independiente.
- Modo de lanzamiento del TUI.
- Defaults de red/MCP y plataformas iniciales.
