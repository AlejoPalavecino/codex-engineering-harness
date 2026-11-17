# Chuli Manager

El menú se abre con `chuli` o `chuli.exe`. Mantiene la consola CRT retro y organiza la gestión en **6 opciones principales**:

1. **Instalar:** asistente guiado en 4 pasos (`Paso 1` plataformas → `Paso 2` preset o personalización proveedor por proveedor con navegación por flechas y `Enter` → `Paso 3` resumen → `Paso 4` animación limpia que oculta la salida técnica hasta mostrar el resultado, con tecla `V` opcional para ver el log técnico).
2. **Actualizar:** consulta las releases oficiales en segundo plano (cancelable con `Esc` si está consultando), muestra una notificación en la portada cuando hay una versión nueva (para tus perfiles o el Manager), resume qué incluye la release y actualiza las plataformas ya instaladas.
3. **Gestionar versiones:** muestra qué versión tenés instalada en cada plataforma, lista las releases oficiales y copias en caché local, permite instalar o cambiar a una versión puntual y restaurar la instalación anterior (`rollback`).
4. **Perfiles de AI:** gestiona presets y modelos/esfuerzos por rol según proveedor (`Codex`, `Claude`, `AGY`), permite guardar perfiles propios con nombre (`G`), cargarlos (`N`) y aplicar cambios directo localmente sin re-descargar.
5. **MCPs:** administra conexiones MCP multi-cliente (**Codex** en `config.toml`, **Claude Code** en `.claude.json` y **AGY CLI** en `mcp_config.json`), combinando el catálogo recomendado (`Context7`, `Engram`, `Exa`, `GitHub` + guía `Egolite`) con alta y baja segura de MCPs personalizados (`N`).
6. **Skills:** lista las skills del perfil y skills propias con su descripción, permite prender o apagar skills por plataforma (`Espacio` + `Enter`/`S` manteniendo la verificación en regla) y sumar o quitar skills locales (`A` / `X`).

Accesos secundarios al pie de la portada: `D` ejecuta verificación (`doctor`), `?` abre *Acerca de Chuli*, `Tab`/`1`/`2`/`3` alterna el cliente activo y `q` sale. Con Go 1.26+ podés instalar o actualizar el ejecutable desde cualquier sistema con `go install github.com/AlejoPalavecino/codex-engineering-harness/v2/cmd/chuli@v2.0.2` (o `@latest`); sin Go, descargá el nuevo Manager de la release.

Para usar `chuli` desde cualquier carpeta en Windows, desde v1.1.1 podés descargar `install-chuli-windows.ps1` de la release (fuente: `scripts/install-manager.ps1`): revisá su vista previa y ejecutalo con `-Apply` para instalarlo y registrarlo en el PATH de tu usuario. [Instalación del comando](INSTALLATION.md#comando-chuli-en-windows). El ZIP del Manager sigue disponible como alternativa portable.

Al abrir **Perfiles de AI**, personalizar una instalación o ejecutar `chuli models list`, el Manager vuelve a consultar la lista del cliente. Codex toma los modelos visibles de `models_cache.json` en su directorio de configuración y admite la caché de escritorio como segunda fuente. Claude usa la caché de catálogo cuando puede identificar un único catálogo compatible. AGY consulta `agy --output-format json models` para el orquestador. Sus especialistas usan los niveles `inherit`, `pro`, `flash` y `flash_lite`.

El selector indica la fuente. Si el cliente está ausente, el formato cambió o la consulta falla, ofrece el catálogo incluido en Chuli. Las cachés reflejan la última sincronización del cliente. Podés elegir un modelo recién descubierto para un rol y revisar el perfil antes de aplicarlo.

Si una instalación con varias plataformas se interrumpe, el resumen identifica las completadas, la que requiere atención y las pendientes. Al confirmar la adopción de Gentle en una plataforma, Chuli retoma desde ese punto.

| Operación | Uso |
| --- | --- |
| Ver la versión del Manager | `chuli version` |
| Elegir cliente (`codex`, `claude` o `agy`) | `chuli --client claude status` |
| Listar presets y modelos por rol | `chuli --client claude models` |
| Instalar con un preset o modelo por rol | `chuli --client claude install --preset claude-balanced --apply` |
| Personalizar un agente puntual | `chuli --client codex install --role orchestrator=gpt-6.1-sol:medium --apply` |
| Ver el destino, modelo registrado e inventario de agentes y skills | `chuli status` |
| Previsualizar / aplicar instalación | `chuli install` / `chuli install --apply` |
| Previsualizar / aplicar actualización | `chuli update` / `chuli update --apply` |
| Comprobar la instalación | `chuli doctor` |
| Revisar / confirmar la reversión | `chuli rollback` / `chuli rollback --apply` |
| Revisar la adopción desde Gentle AI | `chuli install --adopt-gentle-routing` |
| Usar un checkout de confianza para desarrollo o recuperación | `chuli --source RUTA doctor` |
| Listar las cinco integraciones | `chuli integrations list` |
| Revisar estado o guía de una integración | `chuli integrations status context7` / `chuli integrations guide engram` |
| Previsualizar / confirmar una conexión | `chuli integrations configure exa` / `chuli integrations configure exa --apply` |
| Comprobar una integración | `chuli integrations verify exa` |
| Previsualizar / confirmar deshacer la conexión propia | `chuli integrations undo exa` / `chuli integrations undo exa --apply` |

La distribución oficial conserva una copia verificable del paquete para diagnóstico y recuperación. Para una instalación histórica que todavía carece de ese paquete, especificá un checkout compatible con `--source`. El indicador de versión del perfil y la versión del Manager se refieren a productos instalables distintos.

El evaluador continúa disponible en el checkout con `chuli --source RUTA eval check`. Los snapshots y comparaciones locales se describen en [Evaluación](EVALUATION.md).

## Asistente de integraciones

Disponible desde el Manager v1.1.0. Si tenés v1.0.1 o anterior, descargá y extraé el nuevo Manager; instalar o actualizar sólo el perfil no actualiza el ejecutable.

En **Integraciones**, elegí Context7, Engram, Exa, GitHub o Egolite. La ficha explica su utilidad, estado y requisitos. Usá C para revisar la configuración, S/Y para confirmarla, V para comprobar, G para ver la guía y O para abrir el sitio oficial en tu navegador. D muestra la reversión antes de confirmarla. N cancela una vista previa; Esc vuelve un nivel. ↑↓/j/k y PgUp/PgDn desplazan el contenido largo. Todas las opciones son voluntarias.

| Opción | Preparación | Alcance de la comprobación |
| --- | --- | --- |
| Context7 | API key en `CONTEXT7_API_KEY`; conexión HTTP oficial | Inicialización MCP y listado de herramientas, sin consultas de documentación |
| Engram | Descargar binario oficial, verificar checksums y agregar `engram` al PATH | Disponibilidad del comando; la conexión y el proyecto se comprueban en Codex |
| Exa | Conexión HTTP oficial; uso inicial limitado por el proveedor | Inicialización MCP y listado; no hace búsquedas ni inicia trabajos |
| GitHub | Token de permisos mínimos en `GITHUB_PAT_TOKEN` | Inicialización MCP y listado; no consulta repositorios ni escribe datos |
| Egolite | Descarga oficial, instalación y onboarding de ego(lite) | Disponibilidad de `ego-browser`; falta validar onboarding/interacción desde su skill |

Configurá las variables en tu entorno local y reabrí Manager y Codex para heredarlas. No pegues claves en el chat ni en archivos del proyecto. El Manager guarda el nombre de la variable, nunca el valor. Los pasos de cuenta y autenticación permanecen en la guía oficial. No se ejecutan instaladores de terceros, no se instalan plugins y no se modifican hooks ni memorias.

Un servidor con el mismo nombre, un endpoint equivalente o un plugin relacionado declarado en `config.toml` se conserva sin duplicarlo. Los conectores administrados por la app pueden vivir fuera de ese archivo: revisá el panel de Codex antes de agregar una conexión equivalente. Las conexiones personalizadas se comprueban desde Codex para respetar su autenticación y transporte.

La configuración usa Python 3.11+ incluido como requisito del Manager. Antes de confirmar, cerrá otros editores del perfil. Una vista previa desactualizada se rechaza. Las escrituras del asistente se serializan con `.chuli-integrations.lock`; ese bloqueo no controla otras aplicaciones. Si quedó tras una interrupción, cerrá los procesos del asistente y revisá el perfil antes de retirar sólo el directorio de bloqueo vacío. Un registro previo incompleto requiere revisión manual.

**Deshacer** usa el registro `CODEX_HOME/chuli-integrations/<id>.json`, que contiene únicamente la adición pública de Chuli. Elimina sólo un bloque intacto creado por el asistente, conserva el resto del archivo y rechaza conexiones modificadas. No desinstala programas, revoca tokens ni elimina memorias. Es independiente del rollback del perfil.

Para probar desde el checkout: `go run ./cmd/chuli integrations list`. Usá un `CODEX_HOME` temporal para ensayar configuración y reversión. Para comprobar acceso real, hacelo desde una sesión nueva de Codex con tu autorización y cuenta; un handshake no demuestra permisos sobre datos.

## Presentación Consola CRT

El rediseño local incorpora el logotipo pixel-art de la PC retro y las letras en bloque `CHULI` con sombra sobre fondo oscuro, manteniendo una paleta verde neón, ámbar y beige de alto contraste. Conserva las mismas operaciones y atajos de teclado. En terminales anchas (`>= 100` cols), el menú principal muestra la explicación lateral y la tarjeta enmarcada del equipo activo con los 7 roles y sus badges de esfuerzo; en anchos menores se apila y se compactan los nombres. El destino completo y los cambios permanecen disponibles al desplazar la vista previa. La ruta abreviada de la cabecera es sólo una ayuda visual.

```text
  ▄█▀▀▀▀▀▀▀▀▀▀▀▀█▄░   ● MULTI-CLIENT HARNESS · v2.0.2
  ██  ▀▄        ██░   ▄█████▄ ██░ ██░ ██░ ██░ ██░    █████░
  ██  ▄▀ ▄▄▄    ██░   ██░  ▀▀ ██░ ██░ ██░ ██░ ██░     ▐█▌░
  ▀█▄▄▄▄▄▄▄▄▄▄▄▄█▀░   ██░     ██████░ ██░ ██░ ██░     ▐█▌░
 ▄▄▄▄▄▄▄████▄▄▄▄▄▄▄░  ██░  ▄▄ ██░ ██░ ██░ ██░ ██░     ▐█▌░
 ███▄▄▄██████▀▀▀███░  ▀█████▀ ██░ ██░ ▀████▀░ █████░ █████░
 ▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀   [● 1:Codex]  [○ 2:Claude]  [○ 3:AGY]
```

Tamaños de comprobación: 40×24, 80×24 y 120×32; para trabajar cómodamente se recomienda 80×24 o más. El logo utiliza caracteres Unicode de bloque estándar, sin descargar fuentes ni requerir imágenes, animaciones o un emulador específico. Esta presentación se incluye en el Manager v2.0.2. Las PRs #52–#56 corresponden al asistente original; el rediseño visual y el configurador multi-cliente se integraron sobre esa base.
