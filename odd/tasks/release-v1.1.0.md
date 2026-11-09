# Entrega v1.1.0

- **Spec:** docs/releases/v1.1.0.md
- **Design:** docs/specs/crt-tui.md
- **Approval:** approved

El usuario autorizó lanzar una release con las implementaciones nuevas y aprobó explícitamente v1.1.0 el 2026-11-09. Incluye integrar los cambios necesarios, publicar los artefactos oficiales y comprobar cómo actualizar. No modifica releases anteriores ni el perfil personal del usuario.

### TASK-001 — Integrar el alcance aprobado y preparar el candidato
- **State:** DONE
- **Depends on:** none
- **Owner:** orchestrator
- **Acceptance:** PRs #52–#56 integradas, CRT incluido, versión y notas coherentes; preservación de alcances y reversión.
- **Evidence:** PRs #52–#56 integradas en orden con head esperado y CI verde; retarget validado por igualdad de árboles. Issue #51 cerrado por la última. CRT en commit e28f5ae, integración con main en 7771402. Versión y notas v1.1.0 coherentes. Revisión independiente local sin bloqueantes: TUI/CLI, build versionado, perfil de 370 recursos, ZIP Windows, documentación y publicador.
- **Blocker:** none

### TASK-002 — Validar main y generar paquetes oficiales
- **State:** IN_PROGRESS
- **Depends on:** TASK-001
- **Owner:** orchestrator
- **Acceptance:** Go/Python/estructura y CI del SHA final pasan; perfil y tres Managers nativos versionados, sin secretos.
- **Evidence:** Validación local: go test ./..., go vet ./..., go mod tidy -diff y 53 pruebas Python OK (2 skip por symlinks). Estructura OK en copia con fixture histórica canónica; no se modificó hash esperado. CI del SHA final todavía pendiente.
- **Blocker:** none

### TASK-003 — Publicar y comprobar actualización y reversión
- **State:** PLANNED
- **Depends on:** TASK-002
- **Owner:** orchestrator
- **Acceptance:** Release estable/latest v1.1.0 ligada al SHA validado; cuatro paquetes y SHA256SUMS correctos; upgrade 1.0.1→1.1.0 y rollback con perfil temporal.
- **Evidence:** none
- **Blocker:** none

### TASK-004 — Confirmar disponibilidad para actualizar
- **State:** PLANNED
- **Depends on:** TASK-003
- **Owner:** orchestrator
- **Acceptance:** API latest resuelve v1.1.0; instrucciones distinguen actualización de perfil desde TUI y descarga del nuevo ejecutable.
- **Evidence:** none
- **Blocker:** none
