# Entrega v1.1.2

- **Spec:** docs/releases/v1.1.2.md
- **Design:** docs/INSTALLATION.md
- **Approval:** approved

El usuario autorizó preparar la rama, los commits y lanzar la release v1.1.2 en GitHub el 2026-11-16 con soporte de `go install` y las tres correcciones de auditoría. No se modifica el perfil personal del usuario.

### TASK-001 — Incorporar correcciones de auditoría y soporte go install
- **State:** DONE
- **Depends on:** none
- **Owner:** orchestrator
- **Acceptance:** Fixtures `tests/fixtures/**` fijadas a `LF` y validadas en CI Windows, reintentos y detalle en `Restore-ManagerFile`, limpieza de `release.yml`, lectura de versión vía `debug.ReadBuildInfo()` y documentación actualizada.
- **Evidence:** Cambios aplicados en 14 archivos + `docs/releases/v1.1.2.md` y validados localmente con `TestEffectiveVersion` y prueba `locked-rollback`.
- **Blocker:** none

### TASK-002 — Validar suites locales y empaquetado v1.1.2
- **State:** DONE
- **Depends on:** TASK-001
- **Owner:** orchestrator
- **Acceptance:** Validación estructural, Go (`test`, `vet`, `gofmt`, `tidy`), Python (`54` tests) y scripts PowerShell 7 / 5.1 verdes en entornos temporales.
- **Evidence:** `python tests/validate_repository.py` OK; `unittest` 54 tests (52 OK, 2 skip symlinks Windows); `go test -count=1 ./...` OK en los 5 paquetes; `go vet`, `gofmt`, `go mod tidy -diff`, `eval check` y `package_release.py` (perfil + instalador) OK; `windows-smoke.ps1`, `windows-regression.ps1`, `windows-manager-install.ps1` (PS7/PS5.1) y `windows-ttp-resources.ps1` (15/15 PS7/PS5.1) OK; `go install ./cmd/chuli` a `GOBIN` temporal OK.
- **Blocker:** none

### TASK-003 — Integrar en main, validar CI y publicar v1.1.2
- **State:** DONE
- **Depends on:** TASK-002
- **Owner:** orchestrator
- **Acceptance:** `Validate harness` verde en `main`, `release.yml` exitoso para `1.1.2`, release `v1.1.2` publicada como `latest` con 5 artefactos y `SHA256SUMS.txt`, y resolución verificada con `go install ...@v1.1.2`.
- **Evidence:** Rama `feat/release-v1.1.2` integrada en `main` (`922dcc9`). CI `Validate harness` push `43809647921` OK en `structure`, `windows-installer` (incluyendo `validate_repository.py` en Windows) y `chuli-go` (Ubuntu/Windows/macOS con `go install ./cmd/chuli`). Publicador `43810038695` OK; release `v1.1.2` publicada como `latest` (`https://github.com/AlejoPalavecino/codex-engineering-harness/releases/tag/v1.1.2`) con 5 artefactos + `SHA256SUMS.txt` y tag `v1.1.2` apuntando a `922dcc9`. Verificado remotamente con `go install ...@v1.1.2` y `@latest` en `GOBIN` temporal (`Chuli Manager 1.1.2`), seguido de `chuli install --apply` y `chuli doctor` en `CODEX_HOME` y `APPDATA` temporales (`Versión registrada: 1.1.2`, `Errores: 0`).
- **Blocker:** none
