# TKA v1.0 — Tooling & Knowledge Architecture

Una responsabilidad por componente:

- **Instrucciones globales:** `config/AGENTS.md`, política de orquestación y estilo.
- **Agentes:** `agents/*.toml`, trabajos delimitados.
- **Skills:** `skills/*/SKILL.md`, procedimientos cargados según necesidad.
- **MCP:** herramientas externas autorizadas, configuradas localmente y con permisos propios.
- **Engram:** memoria histórica y continuidad por proyecto.
- **Git:** especificaciones, código y decisiones aprobadas, fuente de verdad vigente.
- **CI:** evidencia reproducible, tests, controles de integridad.

## Orden de autoridad

Decisión humana vigente > especificación y ADR aprobados/versionados > comportamiento comprobado existente > memoria histórica Engram > hipótesis.

## Integraciones recomendadas

| Componente | Tipo | Instalación automática |
| --- | --- | --- |
| Engram | MCP/plugin opcional de memoria | No, para no tocar memorias ni hooks |
| Context7 | MCP opcional para librerías | No, requiere herramienta externa |
| Exa | MCP opcional de investigación | No, requiere configuración y/o clave |
| GitHub | Conector/MCP, según Codex | No, requiere autenticación |
| Ego(lite) | Skill/browser opcional | No, se distribuye separadamente |
| CodeGraph, DB | Opcionales por proyecto | No |

El instalador **preserva todos los MCP y plugins existentes**, no inventa credenciales y no copia skills de terceros. Para activar herramientas opcionales seguir las instrucciones de su proveedor. No se debe leer ni publicar `~/.codex/config.toml` real con secretos.

Para trabajos sustanciales, conservar specs/ADR y tareas recuperables en el repositorio. Evitar sistemas paralelos de documentación y memorias con información contradictoria.
