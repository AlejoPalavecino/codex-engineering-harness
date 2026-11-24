# Explicar para aprender mientras se trabaja

- **Spec:** docs/specs/natural-argentine-voice.md
- **Design:** docs/specs/natural-argentine-voice.md#ampliación-2026-11-23-explicar-para-aprender
- **Approval:** approved

### TASK-001 — Instrucciones que enseñan qué, cómo y por qué
- **State:** DONE
- **Depends on:** none
- **Owner:** orchestrator
- **Acceptance:** config/AGENTS.md y docs/protocols/CES.md piden no dar por sabido lo técnico, explicar qué se rompe si no se atiende, causas posibles de un error y por qué se elige la solución; VOZ-08 y V-08 trazables.
- **Evidence:** RED: `python -X utf8 -B tests/validate_repository.py` falló con `AssertionError: config/AGENTS.md: sin dar por sabido`. GREEN: mismo comando OK. Suite Python 126 pruebas OK (3 omitidas); `go test ./...` OK; `git diff --check` exit 0.
- **Blocker:** none

## Límites

- V-08 y el resto de los escenarios de voz: NOT_TESTED contra un modelo. Los checks comprueban el texto instalable, no la conducta observada.
- El perfil personal toma estas instrucciones recién al instalar una versión que incluya este cambio.
- Reversión: revertir el commit de esta unidad; no hay migración de formato.
