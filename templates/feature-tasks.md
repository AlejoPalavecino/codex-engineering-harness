# Feature: Reservas de turnos (ejemplo)

- **Spec:** #especificacion
- **Design:** #diseno-tecnico
- **Approval:** pending
- **Evidence mode:** candidate-v1
- **Trace mode:** requirements-v1

> Esta plantilla es un ejemplo, no un registro activo. Adaptá referencias,
> criterios y Scope al proyecto. Cambiá Approval a approved solamente después
> de aprobar alcance/diseño y autorizar implementación. Las referencias locales
> apuntan al ejemplo de abajo; podés reemplazarlas por documentos aprobados.

## Especificacion

- **REQ-001:** Un usuario autenticado puede reservar un turno disponible y consultar su disponibilidad.
- **REQ-002:** Puede cancelar una reserva propia.
- **REQ-003:** Dos solicitudes concurrentes para el mismo turno solo pueden confirmar una reserva.

Se excluyen pagos y notificaciones.

## Diseno tecnico

Conservar el stack y el flujo de turnos existentes. Cada operación devuelve
confirmación o un error concreto. La reserva es atómica y la cancelación comprueba
propiedad; un conflicto conserva la reserva preexistente. En un proyecto real,
aprobar los contratos y la experiencia aplicables antes de ejecutar.

## Tasks

### TASK-001 — Reservar un turno disponible
- **State:** PLANNED
- **Depends on:** none
- **Owner:** chuli_worker
- **Requirements:** REQ-001
- **Verification:** tests/test_reservas.py::test_reservar_turno
- **Acceptance:** Un usuario confirma un turno libre y al consultarlo aparece ocupado; un turno ocupado devuelve conflicto sin alterar la reserva existente.
- **Scope:** ["src/reservas.py", "tests/test_reservas.py"]
- **Evidence:** none
- **Blocker:** none

### TASK-002 — Cancelar una reserva propia
- **State:** PLANNED
- **Depends on:** TASK-001
- **Owner:** chuli_worker
- **Requirements:** REQ-002
- **Verification:** tests/test_reservas.py::test_cancelar_reserva_propia
- **Acceptance:** Cancelar una reserva propia libera el turno; otro usuario no puede cancelarla ni modificarla.
- **Scope:** ["src/reservas.py", "tests/test_reservas.py"]
- **Evidence:** none
- **Blocker:** none

### TASK-003 — Resolver dos reservas concurrentes del mismo turno
- **State:** PLANNED
- **Depends on:** TASK-001
- **Owner:** chuli_worker
- **Requirements:** REQ-003
- **Verification:** tests/test_reservas.py::test_reservas_concurrentes
- **Acceptance:** Dos solicitudes concurrentes producen una sola reserva confirmada y un conflicto; la consulta conserva un único propietario.
- **Scope:** ["src/reservas.py", "tests/test_reservas.py"]
- **Evidence:** none
- **Blocker:** none
