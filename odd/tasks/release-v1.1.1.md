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
- **State:** DONE
- **Depends on:** TASK-001
- **Owner:** orchestrator
- **Acceptance:** Go, Python, estructura, instalador y CI del SHA final pasan; cinco artefactos oficiales completos.
- **Evidence:** Go test/vet/tidy OK; Python 54 pruebas OK (2 skip de symlink); instalador en PowerShell 7/5.1 OK; estructura y tablero OK (fixture histórica canónica sólo en copia temporal). CI 42705706402 OK para befd975e794373a77a7272b549e71a47a503837d: estructura, tres sistemas e instalador. Cinco artefactos descargados y verificados: 370 recursos del perfil, 24 skills, 293 originales/licencia intactos, cabeceras nativas y script exacto al blob Git revisado.
- **Blocker:** none

### TASK-003 — Publicar y comprobar la release
- **State:** DONE
- **Depends on:** TASK-002
- **Owner:** orchestrator
- **Acceptance:** Release estable/latest v1.1.1, tag correcto, cinco archivos/checksums; instalador descargado resuelve chuli en un entorno temporal.
- **Evidence:** Publicador 42706653822 OK. Release estable/latest v1.1.1 y tag befd975e794373a77a7272b549e71a47a503837d confirmados; cinco archivos más SHA256SUMS coinciden con los bytes del CI. Pruebas previa y posterior por canal público: Manager oficial 1.1.0 actualiza perfil a 1.1.1, ambos doctors pasan y rollback restaura hashes activos/manifest anterior. Modelo, MCP, instrucciones y archivo personalizado conservados. Instalador publicado descargado/verificado: selecciona latest 1.1.1 y resuelve chuli version en PowerShell 7/5.1 con PATH de proceso temporal; PATH permanente sin cambios.
- **Blocker:** none

### TASK-004 — Confirmar disponibilidad y cerrar entrega
- **State:** DONE
- **Depends on:** TASK-003
- **Owner:** orchestrator
- **Acceptance:** Usuario informado con enlaces de instalación y límites; registro cerrado y releases previas conservadas.
- **Evidence:** API latest y tag comprobados; usuario informado del merge de #58, cierre de #57 y release v1.1.1 disponible con instalador descargable. Versiones anteriores intactas. No se instaló en el usuario ni se probó escritura real de su PATH. Release: https://github.com/AlejoPalavecino/codex-engineering-harness/releases/tag/v1.1.1 .
- **Blocker:** none
