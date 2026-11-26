# Feature: Resiliencia ante rate limit de GitHub API en releases

- **Spec:** #especificacion
- **Design:** #diseno-tecnico
- **Approval:** approved
- **Evidence mode:** candidate-v1
- **Trace mode:** requirements-v1

## Especificacion

- **REQ-001:** Si `api.github.com` responde HTTP 403 o 429 al consultar la última release, `Manager.Latest` resuelve el tag siguiendo de forma segura la redirección `https://github.com/<repo>/releases/latest`, rechazando destinos fuera de `github.com` o tags inválidos, y expone un mensaje claro con horario de reinicio y opción `GITHUB_TOKEN` si falla también el respaldo.
- **REQ-002:** Cuando `GITHUB_TOKEN` o `GH_TOKEN` están definidos en el entorno, `fetchJSON` envía el token únicamente a `https://api.github.com`, sin filtrarlo hacia redirecciones web.
- **REQ-003:** En instalaciones o actualizaciones por lote en `internal/tui/wizard.go`, la última release se resuelve una única vez antes de iterar por plataforma.

## Diseno tecnico

- `internal/manager/releases.go`: tipo `githubHTTPStatusError` con detección de rate limit (`429` o `403` con `X-RateLimit-Remaining: 0` o `"rate limit"` en el cuerpo) y `X-RateLimit-Reset`, helper `githubAPIToken()` acotado a `api.github.com`, y `latestViaWebRedirect(ctx)` sin seguir redirecciones externas ni enviar credenciales.
- `internal/tui/wizard.go`: `runInstallBatchRemainingCmd` resuelve `mgr.Latest(ctx)` una vez antes del bucle cuando hay más de una plataforma y fija `tag:<Tag>`.
- `internal/manager/releases_rate_limit_test.go`: pruebas unitarias de fallback, validación estricta de redirección, aislamiento de token y formato del error.

## Tasks

### TASK-001 — Fallback ante HTTP 403/429 de GitHub API y resolución única por lote
- **State:** DONE
- **Depends on:** none
- **Owner:** orchestrator
- **Requirements:** REQ-001, REQ-002, REQ-003
- **Verification:** go test -count=1 -short ./internal/manager ./internal/tui
- **Acceptance:** `Latest` resuelve el tag vía redirección HTTPS en `github.com` ante HTTP 403/429 y rechaza redirecciones fuera de `github.com`; `GITHUB_TOKEN` se adjunta solo a `api.github.com`; el asistente resuelve `latest` una vez por lote multi-plataforma y todas las pruebas unitarias pasan.
- **Scope:** ["internal/manager/releases.go", "internal/manager/releases_rate_limit_test.go", "internal/tui/wizard.go"]
- **Evidence:** {"command": "go vet ./...; go test -count=1 -short ./internal/manager ./internal/tui", "result": "internal/manager y internal/tui en verde; pruebas de fallback rate limit, aislamiento de token y redirecci\u00f3n segura OK", "candidate": "sha256:aba2eac0623b4a03b6123ef9c1e69539566a4312617c9c122b67cf106482371f"}
- **Blocker:** none
