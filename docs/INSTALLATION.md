# Instalación y actualización de Chuli (antes Codex Engineering Harness)

La TUI Go es opcional. El instalador PowerShell conserva compatibilidad con todas las instalaciones actuales. Ver [Chuli Manager](TUI.md).

## Requisitos

- Codex CLI/Desktop instalado y funcionando.
- Windows PowerShell 5.1+ o PowerShell 7.
- **Python 3.11+ requerido para instalar y actualizar**: se utiliza `tomllib` para validar el TOML completo antes de escribir. Si no está disponible, el instalador rechaza el cambio.
- Los modelos sugeridos son opcionales y pueden no estar disponibles en todas las cuentas.
- Engram, GitHub MCP, Context7, Exa y Ego(lite) son integraciones opcionales con instalación separada.
- El checklist nativo requiere que `update_plan` exista realmente en la versión/cliente de Codex; habilitarlo no garantiza un widget en la app.

## Procedimiento

1. Clonar el repositorio y revisar `scripts/install.ps1`.
2. Ejecutar sin `-Apply` para ver las acciones por archivo y las personalizaciones que serán conservadas.
3. Ejecutar `scripts/install.ps1 -Apply`; opcionalmente `-ConfigureModels` establece Sol. Para **sobrescribir una personalización detectada**, primero revisar los cambios y después usar `-Apply -ReplaceCustomized` de forma explícita.
4. Ejecutar `scripts/verify.ps1`; reiniciar Codex Desktop y abrir una sesión nueva.
5. Probar los escenarios descritos en [TESTING.md](TESTING.md).

`CODEX_HOME` prevalece sobre el directorio por defecto `~/.codex`. No se requieren permisos administrativos ni cambios globales de la política de ejecución de PowerShell.

El instalador gestiona **solo** `config/AGENTS.md`, cuatro agentes TOML, cinco skills propias, los tres recursos TTP y entradas `[agents.*]` / `[tools.update_plan]` del TOML. Para preservar un ajuste de progreso desactivado expresamente, usar `-SkipTaskProgress`. No borra otros agentes, skills, hooks, plugins ni memoria. Guarda copia local previa y registra la última instalación en `harness-install/state.json`.

Desde v0.2.1, el manifiesto separa los archivos que necesitan backup para rollback (`files`) del inventario completo de recursos administrados (`managed_files`). Por eso, una reinstalación sin cambios ya no pierde los hashes que permiten reconocer archivos oficiales en la siguiente actualización. El estado también declara `schema_version` y `rollback_schema_version`.

### Migración de instalaciones anteriores

Si el manifiesto de una instalación v0.1.3 o v0.2.0 quedó con `files: []` por la regresión de reinstalación, v0.2.1 compara los archivos actuales con baselines SHA-256 históricos incluidos en `scripts/legacy-managed-hashes.json`. Los archivos que coinciden con el contenido oficial anterior se actualizan normalmente; los que no coinciden se siguen tratando como personalizaciones y se preservan salvo que se autorice `-ReplaceCustomized`. Los baselines históricos solo cubren v0.1.3 y v0.2.0; cualquier release futura que necesite migración debe agregar su propio mapa antes de cambiar o retirar recursos.

**Advertencia:** los scripts son candidatos hasta que se apruebe la release. Probar en un `CODEX_HOME` temporal antes de modificar un entorno importante. Cuando una estructura TOML existente no puede actualizarse de manera conservadora, se aborta antes de escribir; no se hace una sustitución aproximada.

## Reversión

```powershell
.\scripts\rollback.ps1 -Apply
```

El rollback **valida la totalidad de los backups, sus hashes disponibles, las rutas y el JSON del manifiesto anterior antes de tocar los archivos**. Rechaza esquemas desconocidos, rutas inseguras, entradas duplicadas y respaldos incompletos sin realizar restauraciones parciales. El ejecutor debe provenir de v0.2.1 o posterior; los checkouts antiguos se rechazan para impedir que una versión conocida como insegura controle la recuperación. Mantiene los respaldos para recuperación manual. Ante un error imprevisto de escritura, intenta compensar y declara explícitamente cualquier fallo de recuperación.

## Verificación de recursos (v0.2.1)

`scripts/verify.ps1` comprueba que existan los 13 recursos obligatorios (instrucciones globales, cuatro agentes, cinco skills y tres recursos TTP). La ausencia de cualquiera incrementa el contador de errores y termina con código de salida distinto de cero. Las diferencias de hash se informan como deriva o personalización; no se sobrescriben durante la verificación. La verificación no certifica modelos ejecutados, MCP remoto ni aislamiento real.

## Seguimiento de tareas (v0.1.1)

Al ejecutar R2 desde Codex, fuera de Plan Mode, el orquestador puede actualizar `update_plan` y mostrar hitos en la interfaz que admita esa herramienta. Las tareas detalladas residen en `odd/tasks/<feature>.md` del proyecto. Copiá la plantilla de `templates/feature-tasks.md` y consultá `docs/protocols/TTP.md`. Con Python 3.11+, podés comprobar ese archivo con `python ~/.codex/harness/bin/taskboard.py check <archivo>` (ajustá la ruta en Windows).

La app y el CLI no necesariamente comparten el mismo panel de progreso; si el widget no aparece, usá el documento y pedile a Codex un resumen del estado persistente. No exige Python para que el agente mantenga el archivo.

## MCP

No se descargan MCP automáticamente ni se alteran secretos. Revisar `config/mcp.example.json` para las herramientas recomendadas y configurarlas de manera separada según la documentación de su proveedor.
