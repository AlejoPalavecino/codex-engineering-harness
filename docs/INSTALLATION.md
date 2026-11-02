# Instalación y actualización

## Requisitos

- Codex CLI/Desktop instalado y funcionando.
- Windows PowerShell 5.1+ o PowerShell 7.
- Los modelos sugeridos son opcionales y pueden no estar disponibles en todas las cuentas.
- Engram, GitHub MCP, Context7, Exa y Ego(lite) son integraciones opcionales con instalación separada.
- El checklist nativo requiere que `update_plan` exista realmente en la versión/cliente de Codex; habilitarlo no garantiza un widget en la app.

## Procedimiento

1. Clonar el repositorio y revisar `scripts/install.ps1`.
2. Ejecutar sin `-Apply` para obtener el plan de cambios.
3. Ejecutar `scripts/install.ps1 -Apply`, opcionalmente `-ConfigureModels` para establecer Sol por defecto.
4. Ejecutar `scripts/verify.ps1`; reiniciar Codex Desktop y abrir una sesión nueva.
5. Probar los escenarios descritos en [TESTING.md](TESTING.md).

`CODEX_HOME` prevalece sobre el directorio por defecto `~/.codex`. No se requieren permisos administrativos ni cambios globales de la política de ejecución de PowerShell.

El instalador gestiona **solo** `config/AGENTS.md`, cuatro agentes TOML, cinco skills propias, el validador TTP y entradas `[agents.*]` / `[tools.update_plan]` del TOML. Para preservar un ajuste de progreso desactivado expresamente, usar `-SkipTaskProgress`. No borra otros agentes, skills, hooks, plugins ni memoria. Guarda copia local previa y registra la última instalación en `harness-install/state.json`.

**Advertencia:** los scripts son parte de una release candidata. Probar en un `CODEX_HOME` temporal antes de instalar en un entorno con configuración importante.

## Reversión

```powershell
.\scripts\rollback.ps1 -Apply
```

La reversión de la última instalación restaura únicamente archivos administrados y solo si no han sido modificados después de instalar; los backups quedan disponibles para recuperación manual.

## Seguimiento de tareas (v0.1.1)

Al ejecutar R2 desde Codex, fuera de Plan Mode, el orquestador puede actualizar `update_plan` y mostrar hitos en la interfaz que admita esa herramienta. Las tareas detalladas residen en `odd/tasks/<feature>.md` del proyecto. Copiá la plantilla de `templates/feature-tasks.md` y consultá `docs/protocols/TTP.md`. Con Python 3.11+, podés comprobar ese archivo con `python ~/.codex/harness/bin/taskboard.py check <archivo>` (ajustá la ruta en Windows).

La app y el CLI no necesariamente comparten el mismo panel de progreso; si el widget no aparece, usá el documento y pedile a Codex un resumen del estado persistente. No exige Python para que el agente mantenga el archivo.

## MCP

No se descargan MCP automáticamente ni se alteran secretos. Revisar `config/mcp.example.json` para las herramientas recomendadas y configurarlas de manera separada según la documentación de su proveedor.
