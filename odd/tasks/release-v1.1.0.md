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
- **State:** DONE
- **Depends on:** TASK-001
- **Owner:** orchestrator
- **Acceptance:** Go/Python/estructura y CI del SHA final pasan; perfil y tres Managers nativos versionados, sin secretos.
- **Evidence:** Validación local: go test ./..., go vet ./..., go mod tidy -diff y 53 pruebas Python OK (2 skip por symlinks). Estructura OK en copia con fixture histórica canónica; no se modificó hash esperado. CI 42698161035 OK para 8e93ce4c5aece7b478bbf0b64028ff8bfbf9daf7: estructura, Windows/Linux/macOS e instalador en ambos PowerShell. Cuatro artefactos descargados de ese run: perfil de 370 recursos, 24 skills, 293 originales/licencia upstream intactos; archivos y cabeceras nativas correctos.
- **Blocker:** none

### TASK-003 — Publicar y comprobar actualización y reversión
- **State:** DONE
- **Depends on:** TASK-002
- **Owner:** orchestrator
- **Acceptance:** Release estable/latest v1.1.0 ligada al SHA validado; cuatro paquetes y SHA256SUMS correctos; upgrade 1.0.1→1.1.0 y rollback con perfil temporal.
- **Evidence:** Publicador oficial 42698910525 OK; release estable v1.1.0, tag 8e93ce4c5aece7b478bbf0b64028ff8bfbf9daf7. Cuatro paquetes publicados coinciden byte por byte con los artefactos del CI y SHA256SUMS; cinco assets esperados. Prueba previa con artefactos y prueba posterior por canal público: Manager oficial 1.0.1 actualiza perfil 1.0.1→1.1.0 (24 skills), doctor pasa con ambos Managers y rollback restaura hashes activos y managed_files previos. Modelo, MCP, instrucciones y archivo personalizados preservados. Todo sobre perfil y caché temporales; no se modificó el perfil del usuario.
- **Blocker:** none

### TASK-004 — Confirmar disponibilidad para actualizar
- **State:** DONE
- **Depends on:** TASK-003
- **Owner:** orchestrator
- **Acceptance:** API latest resuelve v1.1.0; instrucciones distinguen actualización de perfil desde TUI y descarga del nuevo ejecutable.
- **Evidence:** API releases/latest devuelve v1.1.0 sin draft/prerelease. Actualización real por red con Manager 1.0.1 OK. Usuario informado: actualizar perfil desde la TUI actual; descargar el Manager v1.1.0 para obtener CRT e Integraciones y abrir nueva sesión de Codex. Release: https://github.com/AlejoPalavecino/codex-engineering-harness/releases/tag/v1.1.0 . No se afirma autoactualización del ejecutable ni uso real de proveedores con cuentas.
- **Blocker:** none
