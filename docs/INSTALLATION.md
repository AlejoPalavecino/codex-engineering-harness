# Instalar y actualizar el perfil

## Requisitos y destino

Instalá Codex CLI o la app de escritorio con Codex local, Python 3.11 o posterior y PowerShell. En Windows se admite PowerShell 7 o Windows PowerShell 5.1. En Linux y macOS se usa `pwsh`. La release distribuye el Manager para Windows x64, Linux x64 y macOS ARM64. Para la instalación desde la release basta con el ejecutable de Chuli y estos requisitos.

El perfil se instala bajo `CODEX_HOME`, con `~/.codex` como valor habitual. En Windows suele corresponder a `%USERPROFILE%\.codex`. Las reglas del proyecto y los permisos efectivos que administre el cliente también aplican a la sesión.

## Primera instalación

### Comando chuli en Windows

El checkout actual incluye un instalador del Manager para el usuario de Windows. Desde la carpeta del repositorio, en PowerShell:

```powershell
.\scripts\install-manager.ps1
.\scripts\install-manager.ps1 -Apply
chuli
```

La primera llamada muestra el destino y la versión; `-Apply` descarga el Manager estable oficial, comprueba SHA-256 y lo instala en `%LOCALAPPDATA%\Programs\Chuli`. Agrega esa carpeta al PATH del usuario conservando sus entradas, variables y tipo de registro; no requiere administrador ni cambia el PATH del sistema. Después podés abrir otra ventana de PowerShell y ejecutar `chuli` desde cualquier carpeta. Si ejecutaste el script desde otro proceso, cerrá y reabrí PowerShell para recibir el nuevo entorno.

Para fijar una versión, agregá `-Version 1.1.0`; para elegir destino, `-InstallDir RUTA_ABSOLUTA`. Repetir el instalador permite actualizar el Manager y no duplica el PATH. Rechaza un comando `chuli` que pertenezca a otra instalación, archivos modificados, enlaces/junctions, paquetes inesperados y versiones anteriores. Si falla una escritura, intenta restaurar el ejecutable, registro y PATH anteriores; si la recuperación queda incompleta, conserva sus archivos `.chuli-backup` para revisión manual. Cerrá el Manager antes de reemplazarlo y evitá editar el PATH durante la instalación.

Este instalador todavía no forma parte de los archivos de la release v1.1.0 publicada; por ahora se ejecuta desde un checkout revisado. Instala únicamente el Manager: la instalación/actualización del perfil de Codex se inicia después desde su TUI. No instala Python o PowerShell ni modifica su política de ejecución. La comparación SHA-256 comprueba integridad respecto de la release oficial, no una firma independiente del editor.

### Descarga portable

Descargá y extraé el Manager de [GitHub Releases](https://github.com/AlejoPalavecino/codex-engineering-harness/releases/latest). Si verificás la descarga manualmente, compará el resultado de `Get-FileHash -Algorithm SHA256` o `sha256sum` con la línea del mismo nombre en `SHA256SUMS.txt`. Abrí la TUI, elegí la release estable, revisá el destino y confirmá el plan. El perfil descargado se llama `chuli-profile-v1.1.0.zip` en esta versión.

Reiniciá el cliente o abrí una sesión nueva. En el selector de skills buscá `harness-tdd-qa`, `cognitive-doc-design` y `judgment-day`. Podés pedir al orquestador que delegue una exploración a `chuli_explorer` y mirar el agente usado. La disponibilidad de un modelo depende del cliente y la cuenta. El chat web o móvil no lee la carpeta local del equipo.

## Personalizaciones y recuperación

Cuando Codex muestre dos entradas con el nombre `skill-creator`, la edición importada de Gentle AI está en `CODEX_HOME/skills/skill-creator/SKILL.md`. Podés indicar esa ruta en el pedido para distinguirla de la skill interna de Codex.

Revisá el plan si ya tenés agentes, skills del mismo nombre, un modelo elegido o un orquestador global de otro proveedor. El instalador conserva recursos personalizados y distingue el inventario administrado del registro de la operación que permite revertirla. El contenido ajeno en el archivo de instrucciones requiere convivencia explícita. Las credenciales y las memorias personales quedan fuera del paquete.

Si ya tenés Gentle AI en ese mismo perfil, la acción normal se detiene al detectar su orquestador. Elegí en la TUI la opción de adoptar Chuli y revisá la nueva vista previa. Se respaldan los bloques Gentle de orquestación y routing; otros bloques separados, como persona o memoria, permanecen en el archivo. El programa rechaza marcas ambiguas o incompletas. En la CLI, previsualizá con `chuli install --adopt-gentle-routing` y añadí `--apply` únicamente después de revisar el resultado. La decisión afecta ese perfil de Codex. Tu instalación Gentle en Pi tiene su propio directorio.

Para actualizar, volvés a abrir la TUI. Para comprobar o revertir, usá las acciones de diagnóstico y rollback. Si el destino cambió después de instalar, consultá el diagnóstico antes de revertir. La reversión se aplica a la última operación registrada y conserva evidencia en el directorio de respaldos del perfil.

## Desarrollo desde fuentes

Desde un checkout revisado podés ejecutar `./scripts/install.ps1` para ver el plan y agregar `-Apply` para aplicarlo. Usá `-CodexHome RUTA` para trabajar sobre una carpeta temporal. `-ConfigureModels` expresa la intención de sustituir valores de modelo ya configurados. `-ReplaceCustomized` es una opción explícita que requiere revisar primero los archivos afectados. El instalador no necesita credenciales ni añade proveedores MCP automáticamente. El Manager v1.1.0 ofrece un [asistente de integraciones voluntario](TUI.md#asistente-de-integraciones), separado de instalar el perfil. Para obtenerlo, descargá el nuevo ejecutable; la actualización del perfil no reemplaza el Manager.
