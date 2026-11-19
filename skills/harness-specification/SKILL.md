---
name: harness-specification
description: "Convertir alcance y diseño aprobados en especificaciones verificables y tareas recuperables. Aplicar a trabajo EWP R2 cuando un diff y contexto breve no alcanzan."
---

# Specification & Planning

1. Confirmar aprobaciones vigentes y las fuentes originales de alcance, UX, arquitectura y contratos.
2. Describir comportamientos por unidad de valor: actor, disparador, entradas, validaciones, permisos, efectos, resultados y errores.
3. Escribir criterios observables y trazarlos a pruebas o inspecciones reproducibles. Cuando se requiera analizar el impacto de cambios, conservá IDs estables de requisito y usá el modo opcional `requirements-v1` del tablero TTP. Si la especificación ya define el requisito, enlazá su sección.
4. Identificar dependencias y superficies de edición para evitar colisiones entre escritores.
5. Planificar tareas por resultado de la persona usuaria. Reunir en cada unidad la entrada, el efecto observable y su comprobación. Preguntarse qué necesitaría otra sesión para retomar desde el diff.
6. Incluir riesgos, supuestos y decisiones pendientes; no inventar detalles bloqueantes.
7. Crear documentos de ODD/ADR/UML únicamente cuando aporten continuidad y comprensión.
8. En mantenimiento del harness, especificar migraciones entre versiones, integridad del manifiesto y condiciones de rollback antes de dar por segura una actualización.

No convertir un cambio R1 trivial en una especificación extensa.

## Revisar la aceptación antes de planificar

Para cada criterio, intentá describir una prueba que pase y un resultado que deba fallar. Comprobá si la situación inicial, el actor cuando importa y el efecto esperado permiten distinguirlos. Si dos interpretaciones razonables cambian el veredicto, citá el fragmento y explicá cuál decisión falta. Revisá primero los contratos y ejemplos aprobados; preguntá a la persona responsable solo por lo que siga abierto.

Por ejemplo, para «la búsqueda debe ser rápida», recuperá del contrato el tiempo, la carga y la medición acordados. Si falta el umbral, preguntá «Con la carga acordada, ¿qué tiempo máximo aceptarías?». Para «el administrador puede borrar los registros del equipo», comprobá de qué equipo se trata y qué ocurre con un registro ajeno. Este segundo caso requiere contexto aunque todas las palabras parezcan precisas.

El CLI instalado ofrece `harness/bin/taskboard.py review-criteria "criterio uno" "criterio dos"` como ayuda de lectura en español. Contrastá cada aviso con la especificación. Registrá la respuesta en el criterio original y retomá la planificación de las tareas dependientes cuando sea comprobable. Las unidades independientes con aceptación resuelta pueden avanzar.

## Seguimiento TTP para implementación R2

Tras aprobar alcance, decisiones de arquitectura/diseño y autorizar implementación, dividí la especificación en tareas pequeñas por resultado verificable. Crear o actualizar un solo registro `odd/tasks/<feature>.md` por funcionalidad, usando IDs estables, dependencias, rol, criterio de aceptación, evidencia y bloqueo. No generar tareas para R0 o cambios triviales R1.
El orquestador es dueño del registro. Usá `update_plan` fuera de Plan Mode si está disponible: 3–6 hitos visibles, reflejando cambios de tareas reales. No marcar tareas DONE hasta que hayan sido verificadas y exista evidencia; recuperá el registro desde Git al reanudar.

Antes de crear un registro nuevo, leé bajo `CODEX_HOME` (por defecto `~/.codex`) la plantilla `harness/templates/feature-tasks.md` y la guía `harness/docs/protocols/TTP.md`. Conservá `Spec`, `Design`, `Approval`, encabezados `### TASK-001 — ...` y los campos `State`, `Depends on`, `Owner`, `Acceptance`, `Evidence`, `Blocker`. No traduzcas las claves ni las reemplaces por títulos T1/T2; adaptá solo contenido, IDs estables y dependencias. Ejecutá `python <CODEX_HOME>/harness/bin/taskboard.py check odd/tasks/<feature>.md` antes de anunciar el cierre y corregí el registro si falla. Si no hay Python o faltan recursos, informá el límite sin inventar una validación. El check valida estructura, no certifica las pruebas declaradas. No agregar tableros para R0/R1.
