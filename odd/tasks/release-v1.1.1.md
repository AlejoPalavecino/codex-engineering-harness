# Entrega v1.1.1

- **Spec:** docs/releases/v1.1.1.md
- **Design:** docs/INSTALLATION.md#comando-chuli-en-windows
- **Approval:** approved

El usuario autorizó integrar la PR #58 y publicar v1.1.1 el 2026-11-09. Se incluye el instalador revisado como descarga de release, sin instalarlo en la PC del usuario ni cambiar releases anteriores.

### TASK-001 — Integrar el instalador y preparar v1.1.1
- **State:** DONE
- **Depends on:** none
- **Owner:** orchestrator
- **Acceptance:** PR #58 integrada, versión/notas coherentes y cinco artefactos definidos con checksum.
- **Evidence:** PR #58 integrada con head d1a9fab y CI verde; merge c545738. Empaquetado del instalador: RED por tipo ausente y GREEN de copia exacta del script. Revisión independiente: payload exacto, cinco assets/checksums, selección de sólo dos tar y rechazo de instalador alterado/archivo extra. Versión y notas v1.1.1 coherentes.
- **Blocker:** none

### TASK-002 — Validar el candidato y generar artefactos oficiales
- **State:** IN_PROGRESS
- **Depends on:** TASK-001
- **Owner:** orchestrator
- **Acceptance:** Go, Python, estructura, instalador y CI del SHA final pasan; cinco artefactos oficiales completos.
- **Evidence:** Go test/vet/tidy OK; Python 54 pruebas OK (2 skip de symlink); instalador en PowerShell 7/5.1 OK; estructura y tablero OK (fixture histórica canónica sólo en copia temporal). CI del SHA final pendiente.
- **Blocker:** none

### TASK-003 — Publicar y comprobar la release
- **State:** PLANNED
- **Depends on:** TASK-002
- **Owner:** orchestrator
- **Acceptance:** Release estable/latest v1.1.1, tag correcto, cinco archivos/checksums; instalador descargado resuelve chuli en un entorno temporal.
- **Evidence:** none
- **Blocker:** none

### TASK-004 — Confirmar disponibilidad y cerrar entrega
- **State:** PLANNED
- **Depends on:** TASK-003
- **Owner:** orchestrator
- **Acceptance:** Usuario informado con enlaces de instalación y límites; registro cerrado y releases previas conservadas.
- **Evidence:** none
- **Blocker:** none
