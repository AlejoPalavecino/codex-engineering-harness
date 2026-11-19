# Fixture de verificación de reservas

- **Spec:** docs/reservas.md
- **Design:** docs/diseño.md
- **Approval:** approved
- **Evidence mode:** quality-v1
- **Trace mode:** requirements-v1
- **REQ-001:** Un usuario puede reservar un turno libre.

### TASK-001 — Verificar reserva
- **State:** VERIFYING
- **Depends on:** none
- **Owner:** quality_reviewer
- **Requirements:** REQ-001
- **Acceptance:** Dos intentos simultáneos confirman una sola reserva.
- **Scope:** ["app.py", "tests/test_app.py"]
- **Risk:** critical
- **Verification:** {"concurrencia":{"kind":"test","target":"tests/test_app.py::test_reserva_unica"},"dueño":{"kind":"negative","target":"tests/test_app.py::test_reserva_ajena"},"aislamiento":{"kind":"review","target":"Revisar la escritura atómica sobre el candidato"},"lint":{"kind":"static","target":"ruff check app.py"}}
- **Evidence:** none
- **Blocker:** none
