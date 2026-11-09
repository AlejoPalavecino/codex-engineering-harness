# Chuli Manager 1.0

El menú se abre con `chuli` o `chuli.exe`. La acción de release obtiene el paquete del repositorio oficial, muestra el plan y espera confirmación para instalarlo. Una actualización del perfil puede incorporar skills, recursos o modelos. Para cambiar de versión del ejecutable descargá el nuevo Manager.

| Operación | Uso |
| --- | --- |
| Ver la versión del Manager | `chuli version` |
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

Disponible en el código actual del Manager; el ejecutable publicado en v1.0.1 todavía no incluye este asistente. Instalar o actualizar sólo el perfil no actualiza el ejecutable.

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
