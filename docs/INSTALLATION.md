# Instalar y actualizar el perfil

## Requisitos y destino

Instalá Codex CLI o la app de escritorio con Codex local, Python 3.11 o posterior y PowerShell. En Windows se admite PowerShell 7 o Windows PowerShell 5.1. En Linux y macOS se usa `pwsh`. La release distribuye el Manager para Windows x64, Linux x64 y macOS ARM64. Para la instalación desde la release basta con el ejecutable de Chuli y estos requisitos.

El perfil se instala bajo `CODEX_HOME`, con `~/.codex` como valor habitual. En Windows suele corresponder a `%USERPROFILE%\.codex`. Las reglas del proyecto y los permisos efectivos que administre el cliente también aplican a la sesión.

## Primera instalación

Descargá y extraé el Manager de [GitHub Releases](https://github.com/AlejoPalavecino/codex-engineering-harness/releases/latest). Si verificás la descarga manualmente, compará el resultado de `Get-FileHash -Algorithm SHA256` o `sha256sum` con la línea del mismo nombre en `SHA256SUMS.txt`. Abrí la TUI, elegí la release estable, revisá el destino y confirmá el plan. El perfil descargado se llama `chuli-profile-v1.0.1.zip` en esta versión.

Reiniciá el cliente o abrí una sesión nueva. En el selector de skills buscá `harness-tdd-qa`, `cognitive-doc-design` y `judgment-day`. Podés pedir al orquestador que delegue una exploración a `chuli_explorer` y mirar el agente usado. La disponibilidad de un modelo depende del cliente y la cuenta. El chat web o móvil no lee la carpeta local del equipo.

## Personalizaciones y recuperación

Cuando Codex muestre dos entradas con el nombre `skill-creator`, la edición importada de Gentle AI está en `CODEX_HOME/skills/skill-creator/SKILL.md`. Podés indicar esa ruta en el pedido para distinguirla de la skill interna de Codex.

Revisá el plan si ya tenés agentes, skills del mismo nombre, un modelo elegido o un orquestador global de otro proveedor. El instalador conserva recursos personalizados y distingue el inventario administrado del registro de la operación que permite revertirla. El contenido ajeno en el archivo de instrucciones requiere convivencia explícita. Las credenciales y las memorias personales quedan fuera del paquete.

Si ya tenés Gentle AI en ese mismo perfil, la acción normal se detiene al detectar su orquestador. Elegí en la TUI la opción de adoptar Chuli y revisá la nueva vista previa. Se respaldan los bloques Gentle de orquestación y routing; otros bloques separados, como persona o memoria, permanecen en el archivo. El programa rechaza marcas ambiguas o incompletas. En la CLI, previsualizá con `chuli install --adopt-gentle-routing` y añadí `--apply` únicamente después de revisar el resultado. La decisión afecta ese perfil de Codex. Tu instalación Gentle en Pi tiene su propio directorio.

Para actualizar, volvés a abrir la TUI. Para comprobar o revertir, usá las acciones de diagnóstico y rollback. Si el destino cambió después de instalar, consultá el diagnóstico antes de revertir. La reversión se aplica a la última operación registrada y conserva evidencia en el directorio de respaldos del perfil.

## Desarrollo desde fuentes

Desde un checkout revisado podés ejecutar `./scripts/install.ps1` para ver el plan y agregar `-Apply` para aplicarlo. Usá `-CodexHome RUTA` para trabajar sobre una carpeta temporal. `-ConfigureModels` expresa la intención de sustituir valores de modelo ya configurados. `-ReplaceCustomized` es una opción explícita que requiere revisar primero los archivos afectados. El instalador no necesita credenciales ni añade proveedores MCP automáticamente. El código actual del Manager ofrece un [asistente de integraciones voluntario](TUI.md#asistente-de-integraciones), separado de instalar el perfil; llegará en un próximo ejecutable.
