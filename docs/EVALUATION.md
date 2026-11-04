# Chuli Evaluation Protocol v1.0 — Foundation (v0.2 / PR #5)

**Estado:** fundación de evaluación, no un benchmark ejecutado.  
**Baseline inmutable:** [v0.1.3](https://github.com/AlejoPalavecino/codex-engineering-harness/releases/tag/v0.1.3), commit `68b756bd5e39539b4e277bb490321f8c06870755`.

## Objetivo y límites

Medir comportamientos de Chuli en Codex Desktop/CLI bajo circunstancias controladas, sin cambiar agentes ni modelos por intuición. El núcleo Go de este PR **solo valida el catálogo, toma hashes de workspaces, valida registros de observación y prepara comparaciones**. No ejecuta modelos, no consume tokens, no automatiza aprobaciones ni demuestra permisos efectivos.

Hay **11 escenarios**: R0, R1, R2 descubrimiento, R2 arquitectura, R2 implementación, UX/UI, TDD/QA, TTP, memoria Engram, MRP y R2+C. Cada caso contiene prompt, preparación, ruta EWP, autorización y criterios verificables. Los workspaces son inventados y no incluyen datos reales.

## Dos capas de prueba (no confundirlas)

1. **Determinista / CI:** comprobar que fixtures son coherentes; parser JSON rechaza datos inválidos; snapshots SHA-256 detectan archivos modificados/creados/eliminados; las comparaciones se calculan a partir de rúbricas individuales. El CI no inicia Codex.
2. **Conductual / Codex Desktop:** ejecutar un caso en una copia aislada, observar preguntas, decisiones, delegación, pruebas, TTP y modelos efectivos si son visibles. Registrar fuentes de evidencia. Criterios que no puedan probarse permanecen **NOT_TESTED**, no se infieren.

Una prueba de un modelo no prueba todo un proyecto. Un `PASS` escrito por el evaluador es un **resultado registrado**, no una certificación criptográfica ni evidencia de aislamiento de ASEP.

## Iniciar una evaluación local

Compilá Chuli v0.2 en la rama de desarrollo o usá `go run` (Go 1.26+ para el proyecto). Desde la raíz:

```powershell
go run ./cmd/chuli eval check
go run ./cmd/chuli eval list
go run ./cmd/chuli eval fingerprint
```

La huella SHA-256 se calcula únicamente sobre `evals/manifest.json`, `evals/cases/**`, `evals/fixtures/**` y `evals/rubrics/**`; cambiar un prompt, criterio o fixture invalida la comparación. `evals/reports/` no altera la huella.

Seleccioná un caso y **copiá su fixture a otro directorio descartable**. Por ejemplo en PowerShell:

```powershell
$workspace = Join-Path $env:TEMP "chuli-eval-loans"
Copy-Item -LiteralPath "evals/fixtures/loans" -Destination $workspace -Recurse
$before = Join-Path $env:TEMP "chuli-r0-before.json"
$after = Join-Path $env:TEMP "chuli-r0-after.json"
go run ./cmd/chuli eval snapshot $workspace $before
# Abrí el workspace copiado en una NUEVA sesión Codex.
# Pegá exactamente el prompt del caso r0-readonly.
# NO autorizás cambios en R0.
go run ./cmd/chuli eval snapshot $workspace $after
go run ./cmd/chuli eval assert-unchanged $before $after
```

Si el `workspace` o los snapshots ya existen, elegí rutas nuevas. No sobrescribimos snapshots para conservar trazabilidad. Los snapshots ignoran internals de `.git` pero incluyen archivos nuevos/untracked, también logs o artefactos. **No cubren escrituras fuera del workspace, MCP externos ni secretos**.

El destino se crea de forma exclusiva y se comprueba su directorio real antes de escribir. Se rechazan alias que resuelven dentro del workspace y junctions que el host no puede resolver con seguridad. Esto no protege frente a reemplazos hostiles concurrentes de directorios padre ni constituye un sandbox.

Para R1 o R2 donde se esperan modificaciones, utilizá `chuli eval diff BEFORE AFTER` y revisá el diff de Git. No se debe evaluar directamente sobre el fixture versionado de Chuli.

## Registro de una ejecución

Creá un registro **local privado** a partir de [run.example.json](../evals/templates/run.example.json). Reemplazá `catalog_fingerprint` por la salida de `chuli eval fingerprint`.

- Identificá versión de Chuli, cliente y versión de Codex, SO y hash de configuración si está disponible de forma segura. No copies la configuración, claves ni credenciales.
- Para cada caso, agregá `criteria` con `id`, `status` (PASS/FAIL/PARTIAL/NOT_TESTED), `notes` y `evidence`. Un estado observado requiere una nota y al menos una referencia. El motor deriva el veredicto total del caso; no existe un campo libre para declararlo PASS.
- Guardá solo referencias a evidencia sanitizada. **Los contenidos ni la autenticidad de las referencias se verifican automáticamente**.
- `models` requiere rol, nombre efectivamente observado y evento/fuente. Si no fue visible, dejalo vacío. El TOML solo demuestra lo configurado, no lo ejecutado.
- `measurements` exige una fuente cuando agregás números. No conviertas cuotas de suscripción a costos API ni estimes tokens no expuestos.
- Los casos ausentes y criterios omitidos se computan como NOT_TESTED. El reporte no produce un puntaje global artificial.

```powershell
# Solo imprime Markdown en stdout; no escribe datos privados en el repo.
go run ./cmd/chuli eval report "ruta/a/resultado-candidato.json"
go run ./cmd/chuli eval report "ruta/a/resultado-candidato.json" "ruta/a/baseline-0.1.3.json"
```

Los directorios `evals/reports/` ignoran por defecto los archivos de resultados para evitar publicar evidencia sensible por accidente. Podés guardar informes en otro directorio privado. No agregues archivos con decisiones de proyectos reales al repositorio público.

## Diseño de una comparación justa

1. Congelar la **misma huella** de catálogo/fixtures/rúbricas entre baseline y candidata.
2. Guardar prompt literal, versión de Chuli, versión de Codex Desktop/CLI, configuración/host observables y límites de sandbox.
3. Ejecutar cada escenario con una sesión nueva, en una copia descartable. **Primero una ejecución controlada**; repetir solamente cuando haya variación o una hipótesis concreta.
4. Si la tarea requiere aprobación, detenerse hasta la decisión humana; no responder por cuenta del usuario. Para implementar un escenario sintético, la autorización debe venir del prompt, no de un archivo que diga APPROVED.
5. Registrar calidad y defectos antes de optimizar tiempo/tokens. Documentar explicaciones, delegaciones y reintentos observados.
6. Cambiar **una variable a la vez** (p. ej., un modelo o contrato de agente) para atribuir el resultado; nunca afirmar superioridad por un ejemplo aislado.

El reporte indica `NO COMPARABLE` si alguno de los dos escenarios está NOT_TESTED; advierte sobre cambios de cliente, SO o versión y configuraciones no identificadas.

También compara cada criterio: un nuevo FAIL crítico se informa como regresión aunque el caso completo ya estuviera en FAIL. Una pérdida de cobertura PASS → PARTIAL conserva su advertencia agregada; no demuestra por sí sola un defecto conductual nuevo.

## Gates de aceptación para la próxima implementación (PR #6)

- No existe regresión reproducible en R0 sin escritura, aprobaciones humanas R2 ni cierre TTP con evidencia.
- Los criterios críticos FALLIDOS impiden afirmar una release verificada. Los críticos sin prueba se declaran abiertamente.
- Toda propuesta MRP o de delegación tiene comparación en el mismo fixture con modelos efectivamente observados cuando el host lo permite.
- Roles, CES, TTP, instalador, rollback, CI y Chuli Manager siguen compatibles.
- La v0.2 no promete aislamiento de procesos/archivos: es la tarea de v0.3.

## Datos ausentes / limitaciones

La v0.1.3 es un baseline de **código/configuración**, no un baseline conductual ya medido. La ejecución real en Codex Desktop y el uso de Engram en una instancia ficticia aislada requieren trabajo humano posterior. Si Codex no publica tokens/modelos efectivos, indicamos NO VERIFICABLE en vez de rellenar valores. No automatizamos uso intensivo ni acceso a cuentas conectadas.
