# Integración de frontend-design (Opción A)

- **Spec:** docs/specs/frontend-design-skill.md
- **Design:** docs/specs/frontend-design-skill.md#diseño-de-integración-y-límites
- **Approval:** approved

El usuario aprobó la Opción A (integrar `frontend-design` al catálogo oficial como skill #25 complementaria de `harness-ux-ui`) el 2026-11-16.

### TASK-001 — Escribir pruebas de integridad y distribución (RED)
- **State:** DONE
- **Depends on:** none
- **Owner:** test_engineer
- **Acceptance:** `tests/test_frontend_design_resources.py` y `tests/validate_repository.py` verifican el commit fijado `e4bb5d960d87126ff3b3a8af1a104b7db67560ba`, hashes SHA-256, inventario en Codex/Claude/AGY y enlace desde `harness-ux-ui`, fallando antes de importar la skill.
- **Evidence:** RED observado: `test_frontend_design_resources.py` falló con `FileNotFoundError` (`skills/frontend-design/UPSTREAM.json`) y `AssertionError` (`../frontend-design/SKILL.md` ausente en `harness-ux-ui/SKILL.md`); `validate_repository.py` falló en el set de skills (`AssertionError`).
- **Blocker:** none

### TASK-002 — Importar frontend-design y actualizar catálogo y atribución
- **State:** DONE
- **Depends on:** TASK-001
- **Owner:** chuli_worker
- **Acceptance:** `skills/frontend-design/` contiene los blobs exactos bajo `.gitattributes` `-text` más `UPSTREAM.json`; `harness-ux-ui`, avisos de terceros, catálogo (`25 skills`) y `client_targets.py` quedan actualizados y ponen `TASK-001` en verde.
- **Evidence:** Blobs fijados de `e4bb5d960d87126ff3b3a8af1a104b7db67560ba` guardados con `-text` (`LICENSE.txt` 10174 B, `SKILL.md` 9390 B + `UPSTREAM.json`). `test_frontend_design_resources.py`, `test_ui_review_resources.py` (`Ran 3 tests OK`) y `validate_repository.py` (`OK: 6 agents, 25 skills`) en verde.
- **Blocker:** none

### TASK-003 — Verificar suites completas y empaquetado
- **State:** DONE
- **Depends on:** TASK-002
- **Owner:** quality_reviewer
- **Acceptance:** Pasan `validate_repository.py`, tests Python (`unittest`), `taskboard.py check`, `gofmt`, `go vet`, `go test ./...`, empaquetado `package_release.py profile` y scripts PowerShell sin regresiones ni rutas personales.
- **Evidence:** Ejecutadas todas las suites tras la última edición funcional: `validate_repository.py` (`OK: 6 agents, 25 skills`), `unittest` (`Ran 64 tests — OK (skipped=2)` por symlinks en Windows), `package_release.py profile` con los 3 archivos de `skills/frontend-design/` en el zip, `gofmt`, `go vet ./...`, `go test ./...` (`internal/manager` ejecutado; resto en caché al no cambiar código Go) y las 5 suites PowerShell (`smoke`, `regression`, `ttp-resources` en PS 7/5.1, `manager-install` en PS 7/5.1 y `multiclient`).
- **Blocker:** none
