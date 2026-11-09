# Preguntas y continuidad de la autorización

- **Spec:** docs/specs/questions-and-authorization.md
- **Design:** docs/specs/questions-and-authorization.md#reglas-y-diseño-de-interacción
- **Approval:** approved

### TASK-001 — Conservar reglas y aceptación aprobadas
- **State:** DONE
- **Depends on:** none
- **Owner:** orchestrator
- **Acceptance:** Las reglas aprobadas y AC-01 a AC-08 tienen alcance, diseño, riesgos y reversión recuperables.
- **Evidence:** Inspección contra la aprobación del 2026-11-09: reglas y AC-01 a AC-08 conservados en docs/specs/questions-and-authorization.md; alcance y reversión explícitos.
- **Blocker:** none

### TASK-002 — Distribuir las reglas y preparar su evaluación
- **State:** DONE
- **Depends on:** TASK-001
- **Owner:** orchestrator
- **Acceptance:** Orquestador, Discovery y protocolos son coherentes; ocho escenarios trazan la aceptación sin alterar el catálogo histórico.
- **Evidence:** Inspección AC-01 a AC-08: config/AGENTS.md, Discovery, EWP y ODD coherentes; guía Q-01 a Q-08 preparada y trazada a la spec. Escenarios conductuales NOT_TESTED; catálogo histórico sin cambios.
- **Blocker:** none

### TASK-003 — Verificar la entrega y declarar sus límites
- **State:** DONE
- **Depends on:** TASK-002
- **Owner:** orchestrator
- **Acceptance:** Pasan checks estructurales, regresiones pertinentes, integridad del catálogo y tablero; se informa por separado el estado de la evaluación conductual.
- **Evidence:** Entrega documental comprobada: suite Python 42 pruebas (1 skip), go test ./..., go vet ./..., eval check y enlaces correctos; validador estructural PASS con fixture canónico en copia temporal. Límites de CRLF/UTF-8 y campaña NOT_TESTED detallados en este registro.
- **Blocker:** none

## Evidencia de verificación de la entrega

- `python -X utf8 -B -m unittest discover -s tests -p 'test_*.py'`: 42 pruebas, OK, 1 omitida por symlinks no disponibles. Sin UTF-8 explícito falla una comparación de acentos en test_toml_config; no se modificó esa prueba.
- `python -B tests/validate_repository.py`: falla en el checkout por el hash del fixture histórico convertido a CRLF. Se comprobó que el blob HEAD coincide con el hash esperado y que la diferencia del checkout es exclusivamente CRLF.
- En una copia temporal de HEAD, reponiendo ese fixture con los bytes exactos de `git show HEAD:tests/fixtures/legacy-v0.2.0-harness-specification.md`, el validador pasa tanto antes como después de superponer todos los archivos de esta unidad. No se modificó el fixture del checkout ni su hash esperado.
- `go test ./...`: PASS en cmd/chuli, internal/eval, internal/manager e internal/tui; `go vet ./...`: exit 0.
- `go run ./cmd/chuli --source . eval check`: OK, 11 escenarios; huella cc0d2af804909ee1e10c99826b5de8979591b048c3d6e984f2de5dcfd0a286f1. `git diff HEAD -- evals` vacío.
- Inspección de AC-01 a AC-08 contra instrucciones y guion; enlaces Markdown relativos modificados resuelven; `git diff --check` sin errores.
- No se ejecutó campaña conductual de modelos: Q-01 a Q-08 y Q-07A/B siguen NOT_TESTED. No se instaló en el perfil personal ni se ejecutaron scripts Windows adicionales; las regresiones Go incluyen el Manager.
- No se agregaron pruebas que busquen frases en prompts como sustituto de evidencia conductual. No hay ciclo RED/GREEN del modelo observado; esta unidad cambia instrucciones y documentación.
