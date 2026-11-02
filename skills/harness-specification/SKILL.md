---
name: harness-specification
description: "Convertir alcance y diseño aprobados en especificaciones verificables y tareas recuperables. Aplicar a trabajo EWP R2 cuando un diff y contexto breve no alcanzan."
---

# Specification & Planning

1. Confirmar aprobaciones vigentes y las fuentes originales de alcance, UX, arquitectura y contratos.
2. Describir comportamientos por unidad de valor: actor, disparador, entradas, validaciones, permisos, efectos, resultados y errores.
3. Escribir criterios observables y trazarlos a pruebas o inspecciones reproducibles.
4. Identificar dependencias y superficies de edición para evitar colisiones entre escritores.
5. Planificar tareas pequeñas que puedan implementarse, verificar y retomar.
6. Incluir riesgos, supuestos y decisiones pendientes; no inventar detalles bloqueantes.
7. Crear documentos de ODD/ADR/UML únicamente cuando aporten continuidad y comprensión.

No convertir un cambio R1 trivial en una especificación extensa.

## Seguimiento TTP para implementación R2

Tras aprobar alcance, decisiones de arquitectura/diseño y autorizar implementación, dividí la especificación en tareas pequeñas por resultado verificable. Crear o actualizar un solo registro `odd/tasks/<feature>.md` por funcionalidad, usando IDs estables, dependencias, rol, criterio de aceptación, evidencia y bloqueo. No generar tareas para R0 o cambios triviales R1.
El orquestador es dueño del registro. Usá `update_plan` fuera de Plan Mode si está disponible: 3–6 hitos visibles, reflejando cambios de tareas reales. No marcar tareas DONE hasta que hayan sido verificadas y exista evidencia; recuperá el registro desde Git al reanudar. Protocolo completo: `docs/protocols/TTP.md` del harness.
