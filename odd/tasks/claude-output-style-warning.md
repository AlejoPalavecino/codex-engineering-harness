# Avisar cuando un output style de Claude Code pisa el tono CES

- **Spec:** docs/INSTALLATION.md#personalizaciones-y-recuperación
- **Design:** aviso no bloqueante en `chuli --client claude doctor`; Chuli no modifica `outputStyle` (decisión del usuario, 2026-11-24)
- **Approval:** approved

### TASK-001 — Doctor informa un outputStyle activo sin tocar la configuración
- **State:** DONE
- **Depends on:** none
- **Owner:** orchestrator
- **Acceptance:** con `--client claude`, si `settings.json` tiene `outputStyle` distinto de vacío o `default`, `client_targets.py json-config validate` imprime `WARNING:` y sale 0; `verify.ps1` suma el aviso a las advertencias; `settings.json` queda byte a byte igual; sin aviso para AGY ni sin la clave.
- **Evidence:** RED unitario: `python -m unittest tests.test_client_targets -k output_style` → `TypeError: validate_json_settings() got an unexpected keyword argument 'client'`. RED de integración: `tests/windows-multiclient.ps1` sin el fix → `Claude verify did not warn about the active outputStyle.` GREEN: ambos OK (`MULTICLIENT INSTALL, VERIFY AND ROLLBACK TESTS OK`). Suite Python 127 pruebas OK (3 omitidas); `tests/validate_repository.py` OK; `go test -short ./...` OK; `git diff --check` exit 0.
- **Blocker:** none

## Límites

- Sólo revisa `settings.json` del perfil; no los `settings.json` de proyecto ni `settings.local.json`.
- Detecta cualquier estilo activo, no analiza su contenido: un estilo propio compatible con CES también genera el aviso.
- Reversión: revertir el commit; no hay migración ni cambios en el perfil instalado.
