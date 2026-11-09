# Asistente de integraciones del Manager

## Autoridad y alcance

El usuario aprobó el catálogo de Context7, Engram, Exa, GitHub y Egolite y autorizó diseñar, implementar y verificar el asistente el 2026-11-09. Se conserva Go/Bubble Tea, Python 3.11+ y la identidad de terminal existente. La propuesta aprobada establece elegir, detectar, previsualizar, confirmar, guiar credenciales, comprobar y deshacer. No autoriza publicación ni usar cuentas reales durante pruebas.

## Diseño del recorrido

Menú principal → Integraciones → catálogo de cinco opciones → ficha con propósito, estado y requisitos. ↑↓/j/k seleccionan, Enter abre, Esc vuelve un nivel. En la ficha: C prepara configuración, V comprueba, G muestra guía/descarga oficial, D prepara deshacer. S/Y confirma una vista previa; N/Esc cancela sin escritura. Toda operación larga muestra estado ocupado; resultados desplazables distinguen error, configuración y conexión comprobada. Se conservan colores y tipografía del Manager; no se agrega interfaz web.

Context7 y Exa usan HTTP remoto; GitHub usa su endpoint remoto y referencia a GITHUB_PAT_TOKEN. Context7 referencia CONTEXT7_API_KEY. Las claves se cargan fuera del Manager, en el entorno local; nunca se piden ni guardan valores en el catálogo, historial o logs. Engram registra `engram mcp` solamente si el ejecutable está disponible. Egolite ofrece instalación/onboarding oficial y detección de `ego-browser`; no se registra como MCP. Descargas locales permanecen guiadas y no ejecutan scripts remotos.

Un servidor homónimo, un endpoint equivalente o un plugin relacionado existente se preserva; no se duplica ni se adopta como propiedad de Chuli. Ante configuración personalizada, se indica verificarla desde Codex. Los cambios requieren un hash de previsualización vigente; se rechazan symlinks, TOML inválido, cambios concurrentes detectados y recursos ajenos. El registro de reversión contiene sólo el bloque agregado, no copia credenciales. Deshacer elimina sólo un bloque propio intacto, preservando cambios ajenos; rechaza modificaciones en esa conexión. Se serializan escritores de este asistente; no se promete exclusión de aplicaciones externas.

## Criterios verificables

- INT-01: cinco fichas útiles, navegación por teclado, cancelación sin cambios y confirmación explícita de configuración y reversión.
- INT-02: detectar conexiones/plugin relacionados sin sobrescribir, duplicar ni mostrar secretos; configuración inválida no se modifica.
- INT-03: aplicar sólo al CODEX_HOME elegido, conservar bytes ajenos, verificar previsualización vigente y permitir reversión conservadora de cambios propios.
- INT-04: guías oficiales y requisitos para las cinco opciones; credenciales por variables de entorno, sin instaladores automáticos ni cambios a hooks/memoria.
- INT-05: comprobación remota mediante handshake MCP y listado de herramientas, con timeout, errores de autenticación/red seguros; comprobación local distingue disponibilidad de conexión efectiva.
- INT-06: pruebas con perfiles temporales, servidores simulados y comandos controlados; suite Go/Python, formato, vet y build. No usar credenciales o memorias personales para verificar.

## Entrega y límites

El asistente se incorpora al ejecutable Manager; instalar sólo el perfil no actualiza un Manager anterior. Los plugins de la app pueden tener configuración fuera de config.toml: la detección de plugins declarados es conservadora, sin prometer inventariar conectores administrados por la app. La prueba HTTP usa sólo el endpoint del catálogo y la variable declarada, no extrae credenciales de configuraciones personalizadas. Un handshake no acredita operaciones de producto ni permisos para escribir.

Reversión de desarrollo: quitar el paquete internal/integrations, su conexión al TUI/CLI, pruebas y documentación de esta funcionalidad. La reversión de una conexión instalada es independiente del rollback del perfil. El usuario autorizó validar, crear commits y abrir PRs apiladas hacia main el 2026-11-09; issue #51 aprobado y cierre desde la última PR. Cortes: configuración conservadora; catálogo y servicio; diagnóstico MCP; comandos y adaptación a terminal; flujo de la TUI. Cada PR se compara contra su padre inmediato y debe tener hasta 400 líneas. Sin autorización de merge ni release.

## Fuentes consultadas

Revisadas el 2026-11-09: [Codex MCP](https://learn.chatgpt.com/docs/extend/mcp?surface=cli), [Context7](https://github.com/upstash/context7), [Engram](https://github.com/Gentleman-Programming/engram/blob/main/docs/AGENT-SETUP.md), [Exa](https://exa.ai/docs/get-started/exa-mcp), [GitHub para Codex](https://github.com/github/github-mcp-server/blob/main/docs/installation-guides/install-codex.md), [Egolite](https://lite.ego.app/).
