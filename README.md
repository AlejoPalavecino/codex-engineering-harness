# Chuli

**AI Engineering Harness — de la idea al código, sin perder el hilo.**

Chuli es un harness abierto, inicialmente para Codex, que combina análisis de requisitos, arquitectura agnóstica, diseño UX/UI, TDD, revisión independiente y memoria persistente, **sin montar una fábrica de agentes para cambiar un botón**.

> **Chuli v0.1.3 candidata:** identidad propia y primera interfaz de terminal en Go. Se conserva el repositorio y las rutas heredadas hasta una migración independiente; los roles y protocolos mantienen su comportamiento.

## Chuli Manager — primera TUI

Interfaz en Go y Bubble Tea para consultar el estado de Codex, preparar instalaciones y actualizaciones, revisar cambios y confirmar su aplicación, diagnosticar y realizar rollback. **Ningún cambio se aplica al iniciar la TUI.**

```powershell
# Requiere Go 1.26+ solo para compilar desde el código:
go run ./cmd/chuli --source .

# O compilar el ejecutable para Windows:
go build -o chuli.exe ./cmd/chuli
.\chuli.exe --source .
```

El instalador sigue siendo el de la v0.1.2, incluida su protección de personalizaciones. Ver [guía Chuli Manager](docs/TUI.md). La autoactualización del propio ejecutable y otros proveedores llegarán en versiones posteriores.

## Filosofía

- **Human-in-the-Loop:** el usuario aprueba alcance, stack, arquitectura importante, modelo de datos, experiencia visual y despliegue.
- **Sin arquitectura impuesta:** evaluar dominio y restricciones antes de elegir React, Node, DDD, microservicios o cualquier alternativa.
- **Trabajo proporcional:** investigación (R0), implementación directa (R1) o ingeniería guiada (R2), con controles críticos (C) independientes del tamaño.
- **Agentes bajo demanda:** el orquestador delega por necesidad, no por costumbre.
- **UX y UI primero:** flujos y dirección visual aprobados antes de implementar interfaces sustanciales.
- **Evidencia antes que promesas:** pruebas, contratos y validaciones objetivas, con límites de seguridad reconocidos.
- **Contexto selectivo:** skills cargadas por necesidad y Engram opcional como memoria del proyecto.
- **Seguimiento TTP:** tareas por funcionalidad en Git y checklist nativo de Codex durante ejecución R2, cuando esté disponible.

## Arquitectura

| Pieza | Responsabilidad |
| --- | --- |
| `config/AGENTS.md` | Instrucciones globales de orquestación, EWP y estilo CES |
| `agents/` | Cuatro especialistas personalizados: arquitectura, producto, pruebas y calidad |
| `explorer` y `worker` | Roles nativos de Codex cuando la versión los admita |
| `skills/` | Cinco procedimientos reutilizables de ingeniería |
| `docs/protocols/` | EWP, MRP, ASEP, TKA, CES y TTP documentados |
| `scripts/` | Instalación con respaldo, verificación y reversión |
| `cmd/chuli`, `internal/` | Manager TUI Go y ejecución segura de operaciones existentes |
| `tests/` | Validación estructural, instalador y transiciones TTP |
| `odd/tasks/<feature>.md` en proyectos | Fuente persistente de tareas, estados, dependencias y evidencias |
| Engram, MCP y Ego(lite) | Integraciones externas **opcionales**; sus memorias y credenciales no se incluyen |

## Instalación (Windows / PowerShell)

Requiere **Codex CLI o Desktop**, **Python 3.11+** para validar TOML y PowerShell 7 o Windows PowerShell 5.1.

```powershell
git clone https://github.com/AlejoPalavecino/codex-engineering-harness.git
cd codex-engineering-harness

# Inspeccionar el plan (no escribe)
.\scripts\install.ps1

# Aplicar los cambios propuestos con respaldo; conserva personalizaciones
.\scripts\install.ps1 -Apply

# Opcional, solo después de revisar las diferencias del dry-run:
# .\scripts\install.ps1 -Apply -ReplaceCustomized

# Opcional: también configurar Sol como modelo principal.
# Usar solo si la cuenta y versión de Codex ofrecen ese modelo:
.\scripts\install.ps1 -Apply -ConfigureModels

# Verificar la instalación
.\scripts\verify.ps1
```

El instalador habilita `[tools.update_plan] enabled = true` salvo `-SkipTaskProgress` o que exista una preferencia personalizada distinta. La actualización valida el TOML original y candidato antes de escribir, muestra las rutas afectadas y **preserva personalizaciones salvo reemplazo explícito** (`-ReplaceCustomized`). Esta opción no garantiza que Codex Desktop muestre una tarjeta visual en todas las versiones. También instala el validador opcional `harness/bin/taskboard.py` (requiere Python 3.11+ solo si se usa). **No ejecuta descargas de terceros**, no elimina otros agentes, no toca las memorias de Engram ni desinstala skills existentes. Si ya hay agentes/skills propios con los mismos nombres, los respalda antes de sustituirlos. Los MCP se configuran por separado: ver [herramientas y memoria](docs/protocols/TKA.md).

**Revertir la última instalación:**

```powershell
.\scripts\rollback.ps1 -Apply
```

Después de instalar, reiniciá Codex y abrí una sesión nueva. Consultá [instalación y actualización](docs/INSTALLATION.md).

## Documentación

- [Protocolo de ingeniería (EWP)](docs/protocols/EWP.md)
- [Catálogo de agentes](docs/AGENTS.md)
- [Selección de modelos (MRP)](docs/protocols/MRP.md)
- [Permisos y aislamiento (ASEP)](docs/protocols/ASEP.md)
- [Herramientas y memoria (TKA)](docs/protocols/TKA.md)
- [Comunicación argentina (CES)](docs/protocols/CES.md)
- [Tareas y progreso (TTP)](docs/protocols/TTP.md)
- [Pruebas y límites conocidos](docs/TESTING.md)
- [Hoja de ruta](docs/ROADMAP.md)
- [Identidad de Chuli](docs/BRAND.md)
- [Chuli Manager TUI](docs/TUI.md)

## Seguridad y privacidad

**Nunca subas** tu `~/.codex/config.toml` real, `~/.codex/auth.json`, `.env`, bases de Engram, hooks locales privados, conversaciones, backups ni archivos de clientes. Publicá solo plantillas revisadas. Las instrucciones de solo lectura para subagentes **no son** un sandbox real.

## Contribuir

Issues y PR son bienvenidos. Leé [CONTRIBUTING.md](CONTRIBUTING.md) y [SECURITY.md](SECURITY.md). El proyecto se distribuye bajo la licencia [MIT](LICENSE), con componentes externos opcionales sujetos a sus propias licencias.

---

**English:** Open-source Codex engineering harness featuring risk-proportional workflows, human approval gates, specialized agents, UX/UI-first design, optional persistent memory, and repeatable checks. The first candidate focuses on Windows; see the Spanish documentation above for installation, safety limitations, and contributions.
