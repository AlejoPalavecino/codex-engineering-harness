# Feature: Reservas de turnos (ejemplo)

- **Spec:** #especificacion
- **Design:** #diseno-tecnico
- **Approval:** pending
- **Evidence mode:** candidate-v1

> Esta plantilla es un ejemplo, no un registro activo. Adaptá referencias,
> criterios y Scope al proyecto. Cambiá Approval a approved solamente después
> de aprobar alcance/diseño y autorizar implementación. Las referencias locales
> apuntan al ejemplo de abajo; podés reemplazarlas por documentos aprobados.

## Especificacion

Un usuario autenticado puede reservar un turno disponible, cancelar una reserva
propia y consultar su disponibilidad actual. Dos solicitudes concurrentes no
pueden confirmar el mismo turno. No incluye pagos ni notificaciones.

## Diseno tecnico

Conservar el stack y el flujo de turnos existentes. Cada operación devuelve
confirmación o un error concreto. La reserva es atómica y la cancelación comprueba
propiedad; un conflicto conserva la reserva preexistente. En un proyecto real,
aprobar los contratos y la experiencia aplicables antes de ejecutar.

## Seguimiento Git (cuando corresponda)

Issue y base acordada · rama · estrategia y límites de cada corte · commits y PRs con sus bases y dependencias · checks del candidato · próximo paso. Actualizá las referencias a medida que existan.

## Tasks

### TASK-001 — Reservar un turno disponible
- **State:** PLANNED
- **Depends on:** none
- **Owner:** chuli_worker
- **Acceptance:** Un usuario confirma un turno libre y al consultarlo aparece ocupado; un turno ocupado devuelve conflicto sin alterar la reserva existente.
- **Scope:** ["src/reservas.py", "tests/test_reservas.py"]
- **Evidence:** none
- **Blocker:** none

### TASK-002 — Cancelar una reserva propia
- **State:** PLANNED
- **Depends on:** TASK-001
- **Owner:** chuli_worker
- **Acceptance:** Cancelar una reserva propia libera el turno; otro usuario no puede cancelarla ni modificarla.
- **Scope:** ["src/reservas.py", "tests/test_reservas.py"]
- **Evidence:** none
- **Blocker:** none

### TASK-003 — Resolver dos reservas concurrentes del mismo turno
- **State:** PLANNED
- **Depends on:** TASK-001
- **Owner:** chuli_worker
- **Acceptance:** Dos solicitudes concurrentes producen una sola reserva confirmada y un conflicto; la consulta conserva un único propietario.
- **Scope:** ["src/reservas.py", "tests/test_reservas.py"]
- **Evidence:** none
- **Blocker:** none
