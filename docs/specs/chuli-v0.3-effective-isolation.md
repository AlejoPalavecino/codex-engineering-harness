# Chuli v0.3 — Aislamiento efectivo de ejecución

> **Estado:** propuesta técnica revisada tras investigación M0 del 2026-11-07.  
> **Base:** main, commit 7c15a83 (v0.2.1 integrado; tag/release todavía pendiente).  
> **Aprobación de arquitectura:** pendiente.  
> **Implementación:** no iniciada.

## 1. Objetivo

Pasar de restricciones declarativas en prompts y agentes a permisos efectivos por proceso. La propuesta combina Permission Profiles nativos de Codex, un CODEX_HOME aislado, procesos principales separados, Git Worktrees, tests congelados y verificación independiente.

No vamos a crear un sandbox propio ni a reescribir Chuli. Se conservan EWP, MRP, ASEP, TKA, CES, TTP, las cinco skills, los agentes, el instalador, rollback, Manager TUI y Evaluation Foundation.

El primer adaptador será Codex CLI. Antigravity y otros proveedores quedan fuera de v0.3.0.

## 2. Hallazgos M0 confirmados

La investigación documental del 2026-11-07 está registrada en [2026-11-07-v0.3-m0-research.md](../evaluations/2026-11-07-v0.3-m0-research.md).

- Permission Profiles son beta, están documentados para Windows, Linux, WSL y macOS y requieren Codex CLI 0.138.0 o posterior. La release más reciente observada es 0.167.0.
- Codex incluye :read-only, :workspace y :danger-full-access. Chuli solo utilizará :read-only y :workspace, más perfiles propios.
- Permission Profiles no componen con sandbox_mode / sandbox_workspace_write. Si sandbox_mode aparece en alguna capa cargada o se pasa --sandbox, Codex utiliza el sistema heredado. Chuli debe detectar el conflicto y fallar cerrado.
- CODEX_HOME contiene configuración, autenticación, sesiones, logs y skills. --profile nombre carga CODEX_HOME/nombre.config.toml.
- codex sandbox es un comando estable para ejecutar probes bajo las políticas locales de Codex.
- Los subagentes heredan sandbox, approval policy, permisos y MCP del proceso padre. Un custom agent TOML no puede reemplazarlos.
- Git Worktrees tienen árboles de archivos separados, pero comparten metadata Git. Son aislamiento de trabajo, no una frontera de seguridad.
- macOS utiliza Seatbelt; Linux/WSL usa bubblewrap y seccomp; Windows puede usar MXC o elevated. Unelevated no soporta deny-read y tiene limitaciones de red.
- network.enabled no inicia el proxy. Para reglas de dominio se requiere features.network_proxy=true.
- MCP, búsqueda web, conectores, navegador, Computer Use y Codex Cloud son superficies independientes del sandbox local.
- apply_patch consulta la política efectiva de filesystem y puede requerir permisos adicionales. Debe probarse en una sesión real, no solo con comandos shell.

Fuentes oficiales y límites: [Permissions](https://learn.chatgpt.com/docs/permissions), [Subagents](https://learn.chatgpt.com/docs/agent-configuration/subagents), [Worktrees](https://learn.chatgpt.com/docs/environments/git-worktrees), [Windows sandbox](https://learn.chatgpt.com/docs/windows/windows-sandbox), [CLI commands](https://learn.chatgpt.com/docs/developer-commands?surface=cli), [Codex 0.167.0](https://github.com/openai/codex/releases/tag/rust-v0.167.0).

## 3. Arquitectura propuesta

### 3.1 CODEX_HOME aislado

Crear un runtime independiente, por defecto ~/.chuli/codex-home, configurable por el usuario. Chuli no copia auth.json, tokens, configuración de proveedores ni MCP personales. El usuario inicia sesión una vez en este home. El Codex personal permanece intacto.

Cada rol tendrá un config profile en ese home, por ejemplo chuli-implementation.config.toml. Chuli lanzará Codex con --profile y con el CODEX_HOME configurado únicamente en el entorno del proceso hijo.

El detector revisará también la configuración del proyecto y los requisitos administrados. Si sandbox_mode o sandbox_workspace_write impiden usar Permission Profiles, o el perfil efectivo no coincide con el solicitado, la sesión queda BLOCKED.

### 3.2 Procesos principales separados

Los perfiles de confianza diferentes requieren procesos principales separados. Los subagentes pueden utilizarse dentro de cada etapa, pero heredan el mismo nivel de permisos.

| Rol | Perfil | Superficie |
|---|---|---|
| planner | :read-only | Discovery y planificación |
| test-author | :read-only + write en paths de tests aprobados | Crear tests antes del freeze |
| implementation | :workspace + read en tests congelados | Escribir código sin alterar aceptación |
| reviewer | :read-only | Revisar diff y evidencia |
| verifier | :workspace sobre una copia temporal | Ejecutar pruebas sin tocar el candidato |
| integrator | Operaciones Git de Chuli | Integrar tras aprobación humana |

Los paths protegidos se configuran por proyecto. No se asume que todos usan tests/ o una estructura fija.

### 3.3 Ejemplo de perfil implementation

Ejemplo ilustrativo. El generador debe insertar los paths reales y validar el TOML contra la versión de Codex detectada.

~~~toml
approval_policy = "never"
default_permissions = "chuli-implementation"

[permissions.chuli-implementation]
extends = ":workspace"

[permissions.chuli-implementation.filesystem]
":root" = "deny"
":minimal" = "read"

[permissions.chuli-implementation.filesystem.":workspace_roots"]
"." = "write"
"tests" = "read"
"**/*.env" = "deny"
~~~

read permite consultar y ejecutar pruebas sin modificarlas. deny se reserva para secretos o rutas que tampoco deben leerse. Los globs read/write no se usarán como única protección portable.

### 3.4 Flujo de trabajo

1. Prepare: detectar versión, sistema, backend, capas de configuración y capacidades.
2. Plan: mostrar roles, paths, red, MCP y worktrees; no escribir.
3. Test-author: trabajar en su worktree, crear pruebas y esperar aprobación.
4. Freeze: validar diff, registrar commit y SHA-256 de cada test aprobado.
5. Implementation: proceso separado y worktree propio; código writable, tests congelados read-only.
6. Review: proceso independiente read-only.
7. Verify: ejecutar pruebas sobre una copia temporal; comparar el candidato original y los hashes protegidos.
8. Integrate: Chuli realiza operaciones Git solo después de pruebas, revisión y aprobación humana.

Los agentes no escriben .git. Chuli administra commits, branches y merge. Si el repo tiene cambios locales que no pueden preservarse de forma segura, la operación se bloquea; no se ejecutan stash, reset o clean automáticamente.

### 3.5 Red y MCP

La red de comandos locales queda deshabilitada por defecto. Los dominios solo se habilitan si el proxy está activo y los probes verifican el resultado.

El CODEX_HOME aislado comienza sin servidores MCP. Cada servidor requiere allowlist, credenciales y permisos externos declarados. El sandbox local no se presenta como control de conectores, búsqueda web, browser, Computer Use o servicios cloud.

### 3.6 Backends por plataforma

- Windows: primera plataforma de aceptación. MXC o elevated se aceptan solo si pasan la suite completa. Para MXC, la documentación indica Windows 11 24H2 build 26100.9278 o posterior, 25H2 build 26200.9278 o posterior y Codex CLI 0.162.0 o posterior. Unelevated no se acepta para perfiles que requieren deny-read.
- Linux/WSL: bubblewrap y seccomp; validar requisitos del host y evitar el backend legacy Landlock para perfiles de filesystem restringidos.
- macOS: Seatbelt.

Linux y macOS solo se declaran soportados tras superar probes de filesystem, procesos hijos, red y sesión real. Compilar el binario no certifica el backend.

## 4. Threat model y límites

- Prompts y archivos TOML de agentes no son permisos efectivos.
- Approval policy y sandbox son controles diferentes. Los perfiles estrictos no conceden escaladas automáticas.
- Un usuario/administrador local que modifica deliberadamente su configuración está fuera del threat model.
- Worktrees separan archivos de trabajo, pero comparten metadata Git.
- Los hashes detectan alteraciones; no impiden escrituras.
- Symlinks, junctions, hardlinks, traversal, procesos hijos y archivos creados durante una sesión requieren probes específicos.
- Si una restricción no se puede aplicar o verificar, la ejecución no puede quedar READY.

## 5. Verificación

Comandos propuestos:
- chuli isolate check
- chuli isolate plan
- chuli isolate probe
- chuli isolate freeze-tests
- chuli isolate launch
- chuli isolate verify
- chuli isolate report

Estados:
- READY: perfil/backend identificados y todos los probes obligatorios aprobados.
- BLOCKED: una restricción fue violada o hay configuración incompatible.
- NOT_VERIFIABLE: falta backend, versión o evidencia.

### Pruebas obligatorias

| ID | Prueba | Resultado esperado |
|---|---|---|
| FS-01 | Escribir en ruta permitida | Permitido |
| FS-02 | Modificar, borrar o renombrar test congelado | Denegado |
| FS-03 | Escribir fuera del workspace | Denegado |
| FS-04 | Traversal con .. | Denegado |
| FS-05 | Symlink/junction/hardlink a ruta protegida | Denegado o NOT_VERIFIABLE |
| PROC-01 | Proceso hijo intenta escribir en ruta denegada | Denegado |
| PATCH-01 | apply_patch intenta modificar test congelado | Denegado sin escalada |
| CODEX-01 | Sesión implementation intenta editar test congelado | Denegado y hash intacto |
| CODEX-02 | Subagente intenta editar test congelado | Hereda la denegación |
| REVIEW-01 | Reviewer intenta modificar un archivo | Denegado |
| VERIFY-01 | Verifier altera su copia temporal | Permitido solo en la copia |
| NET-01 | Comando intenta conexión externa con red deshabilitada | Denegado |
| NET-02 | Perfil con proxy allowlist | Solo destinos permitidos |
| MCP-01 | Perfil sin MCP intenta invocar un servidor | No disponible |

Cada probe se ejecuta con fixtures descartables. CODEX-01 y PATCH-01 son gates de release: no se reemplazan por tests unitarios del generador TOML.

## 6. Fases

| Fase | Entregable | Dependencia |
|---|---|---|
| M0 | Investigación, ADR y decisiones aprobadas | Aprobación del usuario |
| M1 | Policy core neutral y normalización de paths | M0 |
| M2 | CODEX_HOME, perfiles y capability detector | M1 |
| M3 | Worktrees, launcher y freeze de tests | M1/M2 |
| M4 | Probes adversariales y sesión real | M2/M3 |
| M5 | CLI, TUI e informes | M2/M3/M4 |
| M6 | CI, matriz de compatibilidad y release candidate | M0–M5 |

## 7. Criterios de aceptación v0.3.0

1. Detectar Codex CLI, backend y perfil efectivo antes de iniciar una sesión estricta.
2. Ejecutar etapas de distinto nivel de confianza en procesos principales separados.
3. Mantener subagentes dentro de la política heredada del proceso padre.
4. Proteger tests congelados frente a shell y apply_patch.
5. Reviewer read-only y verifier en copia temporal.
6. Bloquear traversal, enlaces y procesos hijos que intenten evadir rutas denegadas.
7. Red local deshabilitada por defecto; proxy verificado cuando se habilite.
8. MCP ausente por defecto y allowlist explícita.
9. Sin fallback silencioso ni READY sin probes.
10. CI para policy core y evidencia real por backend/plataforma declarada.

## 8. Decisiones pendientes

1. Aceptar CODEX_HOME separado y login independiente.
2. Elegir handoff de TUI a Codex CLI en la misma terminal o preparación manual de perfiles/worktrees.
3. Confirmar Windows como primera plataforma de aceptación.
4. Confirmar red deshabilitada y cero MCP por defecto.
5. Confirmar Codex CLI como único adaptador inicial.

Hasta aprobar estas decisiones, las tareas de implementación siguen PLANNED.
