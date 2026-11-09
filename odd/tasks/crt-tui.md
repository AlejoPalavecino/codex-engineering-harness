# Consola CRT

- **Spec:** docs/specs/crt-tui.md
- **Design:** docs/specs/crt-tui.md#diseño-aprobado
- **Approval:** approved

El usuario autorizó integrar y publicar la presentación en la release v1.1.0 el 2026-11-09. El corte mantiene juntos renderer, paginación, sanitización, pruebas y conexión a View; supera la guía de 400 líneas por commit porque separarlos dejaría un render sin su contrato de altura. No se publica una PR fuera del límite del repositorio.

### TASK-001 — Presentación retro accesible y adaptable
- **State:** DONE
- **Depends on:** none
- **Owner:** chuli_worker
- **Acceptance:** CRT-01 y CRT-02; marca textual, jerarquía, anchos y atajos claros.
- **Evidence:** chuli_worker entregó presentation.go y pruebas con RED por renderCRT ausente y GREEN focalizado. Orquestador comprobó anchos 32/60/99/100/140 y altura de pantallas paginadas en 40×24, 80×24 y 120×32, ajustó explicación contextual y destino desplazable. Tests del render y teclado OK; sin dependencias nuevas, x/ansi existente pasa a uso directo.
- **Blocker:** none

### TASK-002 — Integración y comprobación de comportamiento
- **State:** DONE
- **Depends on:** TASK-001
- **Owner:** orchestrator
- **Acceptance:** CRT-03 y CRT-04; render integrado, flujos preservados y terminal temporal comprobada.
- **Evidence:** View conectado a renderCRT; pruebas existentes de navegación/confirmación/cancelación y nuevas de ancho/altura/sanitización OK. Se normalizan ANSI y padding en la aserción de atajos, sin cambiar el criterio. `go test ./...`, `go vet ./...`, `go mod tidy -diff`, build y suite Python (53 pruebas, 2 skip de symlink) OK. Validador estructural OK en copia con fixture histórica canónica. PTY real 80×24 con perfil temporal: inicio, catálogo, Exa, error por conexión ausente, vista previa y cancelación sin config creada; salida normal. Anchos/alturas 40×24, 80×24 y 120×32 comprobados por render en todas las pantallas. No se probaron terminales nativas de Linux/macOS, sesiones con proveedores reales ni colores en todos los emuladores. PRs #52–#56 pertenecen al asistente y están integradas; la publicación conjunta se verifica en odd/tasks/release-v1.1.0.md.
- **Blocker:** none
