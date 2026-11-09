# Integrar y publicar Chuli v1.0.1

- **Spec:** docs/releases/v1.0.1.md
- **Design:** .github/workflows/release.yml
- **Approval:** approved

El usuario autorizó integrar las PR #45, #46 y #50 y publicar v1.0.1 el 2026-11-09. La autorización cubre preparación de versión, resolución de conflictos, verificaciones, tag y publicación por el workflow existente. No modifica el perfil personal. Los estados describen hechos comprobados, no la sola finalización de una herramienta.

### TASK-001 — Integrar las tres mejoras aprobadas
- **State:** DONE
- **Depends on:** none
- **Owner:** orchestrator
- **Acceptance:** PR #45, #46 y #50 integradas con checks de sus candidatos; conflictos resueltos conservando las tres mejoras.
- **Evidence:** PR #45 integrada en 416bd75; #46 en 3b19438 tras CI 42648686701/42648674610 exitosos; #50 integrada tras CI 42649322623/42649312093 exitosos. Conflictos de docs/EVALUATION.md resueltos conservando las tres referencias.
- **Blocker:** none

### TASK-002 — Preparar metadatos y notas de versión
- **State:** DONE
- **Depends on:** none
- **Owner:** orchestrator
- **Acceptance:** VERSION, Manager de desarrollo, documentación y validación coherentes con 1.0.1; notas honestas sobre pruebas y límites.
- **Evidence:** VERSION y validación fijados en 1.0.1; RED de versión anterior observado y GREEN estructural en copia temporal canónica; Manager compilado informa 1.0.1. Notas, changelog y enlaces revisados.
- **Blocker:** none

### TASK-003 — Validar el candidato final de main
- **State:** VERIFYING
- **Depends on:** TASK-001, TASK-002
- **Owner:** orchestrator
- **Acceptance:** Checks locales pertinentes y Validate harness de push sobre el SHA final de main exitosos; artefactos ligados a ese run y sin cambios de head.
- **Evidence:** none
- **Blocker:** none

### TASK-004 — Publicar y comprobar la release estable
- **State:** PLANNED
- **Depends on:** TASK-003
- **Owner:** orchestrator
- **Acceptance:** Workflow de publicación exitoso; v1.0.1 apunta al SHA validado, es estable/latest y contiene cuatro archivos y SHA256SUMS comprobados.
- **Evidence:** none
- **Blocker:** none

## Reversión y límites

Antes de publicar, revertir solo la preparación de versión si falla la aceptación; conservar las mejoras integradas salvo defecto comprobado. No mover tags publicados. Ante defectos posteriores, conservar trazabilidad y preparar una corrección. No afirmar aislamiento, comprensión humana o campañas del modelo que no se ejecutaron.

## Comprobaciones locales previas a publicar

- `python -X utf8 -B -m unittest discover -s tests -p 'test_*.py'`: 43 pruebas, OK, 1 omitida por symlinks no disponibles; composición local con las tres mejoras y versión 1.0.1.
- `go test ./...`, `go vet ./...` y `go mod tidy -diff`: exit 0. Manager con ldflags de release: «Chuli Manager 1.0.1».
- `go run ./cmd/chuli --source . eval check`: 11 escenarios válidos; huella cc0d2af804909ee1e10c99826b5de8979591b048c3d6e984f2de5dcfd0a286f1, sin cambios al catálogo histórico.
- Versión: RED observado al exigir 1.0.1 antes de modificar VERSION; GREEN estructural en copia temporal usando los bytes canónicos del fixture histórico de HEAD, evitando la conversión local CRLF. No se alteró el fixture versionado.
- Las campañas conductuales complementarias, comprensión humana y accesibilidad completa permanecen pendientes. La prueba manual de egolite y sus límites están en el tablero UX/UI; no se repitió ni se atribuye una campaña nueva al preparar la release.
