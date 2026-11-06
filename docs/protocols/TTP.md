# TTP v1.0 — Task Tracking Protocol (Codex Harness v0.1.1)

TTP conecta las decisiones y especificaciones aprobadas con **tareas pequeñas, verificables y recuperables**. Permite seguir el avance en Codex Desktop o CLI, sin crear un nuevo agente, MCP ni tablero paralelo.

## Dos capas, una fuente de verdad

1. **Estado persistente:** un archivo por funcionalidad sustancial, `odd/tasks/<feature>.md`, versionado en el repositorio del proyecto. Conserva referencias a especificación/diseño, dependencias, estado, responsable, aceptación, evidencia y bloqueos.
2. **Progreso nativo de sesión:** la herramienta `update_plan`, **cuando esté disponible**, refleja unos pocos hitos significativos. No es el historial oficial ni un panel Kanban permanente.

En Codex CLI 0.152.0+ se documenta `[tools.update_plan]\nenabled = true` como habilitación explícita. El instalador la configura con backup; es posible que tu versión de **Codex Desktop** no represente el checklist igual que CLI. Reiniciá la app, abrí una sesión nueva y comprobá una llamada real. Si la herramienta no está disponible, se sigue trabajando con el registro persistente y un resumen legible en el chat; **no se simula** un widget.

Plan Mode sirve para discutir y aprobar el enfoque; TTP no escribe archivos ni fuerza `update_plan` durante Plan Mode si este es de solo lectura. El progreso se registra **después** de aprobar el plan y autorizar la implementación.

## En qué trabajos aplica

- **R0:** análisis, propuesta o planificación sin autorización de escritura. No crear archivo persistente; mostrar borrador en el chat.
- **R1:** cambio acotado y recuperable desde el diff. Sin tablero ni documento obligatorio.
- **R2:** tras aprobación del alcance, arquitectura/diseño aplicables, criterios y autorización de ejecución, crear o actualizar el archivo de tareas **una sola vez por funcionalidad**.
- **C:** reforzar evidencias y validaciones necesarias; no declarar DONE porque un subagente dice que terminó.

## Recursos disponibles después de instalar

El instalador distribuye las mismas fuentes canónicas del repositorio bajo `CODEX_HOME` (por defecto `~/.codex`):

| Recurso | Ruta instalada |
| --- | --- |
| Plantilla | `harness/templates/feature-tasks.md` |
| Esta guía | `harness/docs/protocols/TTP.md` |
| Validador | `harness/bin/taskboard.py` |

Leé la [plantilla canónica](../../templates/feature-tasks.md) antes de crear un registro nuevo; su enlace también resuelve desde la guía instalada. No sobrescribas un tablero existente ni cambies Approval a approved sin aprobación y autorización reales. Conservá los encabezados, metadatos y claves de la plantilla en inglés para que el validador pueda leerlos; el contenido puede estar en español.

Desde el proyecto, en PowerShell:

```powershell
$ttpHome = if ($env:CODEX_HOME) { $env:CODEX_HOME } else { Join-Path $HOME '.codex' }
python (Join-Path $ttpHome 'harness/bin/taskboard.py') check odd/tasks/mi-funcionalidad.md
```

En otras shells resolvé ese mismo directorio desde `CODEX_HOME` o su valor por defecto. Los comandos `scripts/taskboard.py` de esta guía son para una copia del repositorio del harness; en otro proyecto usá la ruta instalada. Comprobá criterios y pruebas antes de DONE; después validá el tablero antes de anunciar el cierre. Un check exitoso no demuestra que se ejecutaron las pruebas declaradas ni ofrece aislamiento de escritura.

Plantilla y guía participan del mismo plan, manifiesto, hashes, backup, preservación de personalizaciones y rollback que el validador. Un recurso personalizado se conserva por defecto; reemplazarlo exige `-ReplaceCustomized`. No se modifica el formato TTP ni se crea otro sistema de seguimiento.

## Unidad de trabajo

Una tarea debe representar **comportamiento o entregable verificable**, no operaciones como «abrir un archivo», «crear carpeta» o «ejecutar comando». Su contrato contiene:

- **ID estable:** `TASK-001`, `TASK-002`, etc., sin reutilizar IDs.
- **Fuente:** referencias a la spec/diseño aprobados de la funcionalidad.
- **Dependencias:** IDs que deben estar DONE antes de activarla.
- **Owner:** rol que ejecuta; el orquestador es el **único editor** del tablero.
- **Acceptance:** criterio observable acordado.
- **Evidence:** comandos y resultados, pruebas, diff/commit o comprobaciones que demuestran el cierre.
- **Blocker:** razón concreta y la decisión que falta, si está bloqueada.

Usá la [plantilla](../../templates/feature-tasks.md). No dupliques el mismo plan en Engram: guardá allí solo decisiones y resúmenes de continuidad.

## Estados persistentes

| Estado | Significado |
| --- | --- |
| `PLANNED` | Identificada; no necesariamente preparada |
| `READY` | Plan aprobado, dependencias verificadas, autorizada |
| `IN_PROGRESS` | Implementación o producción del entregable |
| `BLOCKED` | Necesita dato, permiso, corrección o dependencia |
| `VERIFYING` | Implementación terminada; verificaciones en marcha |
| `DONE` | Criterios satisfechos **con evidencia** |
| `CANCELLED` | Descartada con motivo documentado |

Flujo normal: `PLANNED → READY → IN_PROGRESS → VERIFYING → DONE`. Permitir `BLOCKED` desde trabajo iniciado y volver a `READY` cuando el bloqueo se resuelva. Una tarea DONE solo se reabre cuando cambia su alcance/aceptación, dejando motivo explícito y nueva evidencia pendiente.

Una tarea no puede pasar a READY/IN_PROGRESS/VERIFYING/DONE si tiene dependencias no terminadas o falta aprobación del plan. El verificador independiente **no modifica el tablero**; entrega hallazgos al orquestador.

## Actualización de la UI

En una ejecución R2 autorizada, el orquestador:
1. Lee `odd/tasks/<feature>.md` y contrasta estados con el repo y las verificaciones disponibles.
2. Si `update_plan` está disponible y no está en Plan Mode, publica **3–6 hitos de la ejecución actual**, no 30 microtareas. Usa `pending`, `in_progress` y `completed`; solo un hito `in_progress` en el plan nativo.
3. Actualiza el registro **al iniciar una tarea, al entrar en VERIFYING, al bloquearse y al cerrarse con evidencia**. En cada transición significativa actualiza el checklist visible y muestra qué terminó y qué viene después.
4. Para múltiples subagentes en paralelo, el hito visible agrupa la etapa; las tareas reales conservan estados independientes. No marcar un hito `completed` si sus tareas críticas no están DONE.
5. Si `update_plan` falta o falla, lo explica brevemente y presenta un resumen derivado del archivo. No afirma haber mostrado una tarjeta nativa si no la hubo.

## Recuperación y contradicciones

Al retomar una sesión, **no confiar ciegamente** en la última conversación ni en Engram. Leer archivo de tareas + especificaciones + `git status` / diff y evidencia relevante; comprobar que DONE sigue respaldado, identificar IN_PROGRESS posiblemente interrumpidas y preguntar solo si existe una contradicción sustancial. Reconstruir `update_plan` desde los hitos vigentes cuando se pueda. No reabrir tareas verificadas sin un cambio justificado.

Para una nueva versión de spec, registrar impacto, conservar tareas no afectadas y replanificar únicamente las que cambien. Un cambio crítico puede regresar a Discovery/Design conforme al EWP.

## Verificador ligero (opcional)

Si hay Python 3.11+, el repo distribuye `scripts/taskboard.py`, sin dependencias:

```sh
python scripts/taskboard.py check odd/tasks/mi-funcionalidad.md
python scripts/taskboard.py show odd/tasks/mi-funcionalidad.md
python scripts/taskboard.py set odd/tasks/mi-funcionalidad.md TASK-001 IN_PROGRESS --expect READY
python scripts/taskboard.py set odd/tasks/mi-funcionalidad.md TASK-001 VERIFYING --expect IN_PROGRESS
python scripts/taskboard.py set odd/tasks/mi-funcionalidad.md TASK-001 DONE --expect VERIFYING --evidence "pytest -q: 14 passed; commit abc123"
```

El instalador coloca también una copia en `~/.codex/harness/bin/taskboard.py`. Si no hay Python instalado, Codex puede editar el Markdown respetando estas reglas y el usuario puede inspeccionar el diff. El script valida **estructura y transiciones**, no puede garantizar que una prueba se ejecutó ni impedir escrituras fuera del sandbox.

El registro de tareas debe vivir **en el repositorio del proyecto**, no en la carpeta global del harness. No versionar secretos ni datos de clientes.

## No objetivos de v0.1.1

- No implementar panel custom dentro de la app cerrada de Codex.
- No crear un nuevo MCP, servidor, base de datos o agente administrador.
- No sincronizar automáticamente con GitHub Issues/Projects.
- No prometer consistencia transaccional entre procesos o sesiones simultáneas; **un orquestador es propietario del registro**.
- No incluir cronometraje ni un porcentaje de esfuerzo inventado. El avance puede expresarse como tareas DONE sobre tareas activas, sin confundirlo con duración.
