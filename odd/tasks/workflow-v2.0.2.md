# Feature: Flujo de trabajo de Chuli 2.0.2

- **Spec:** docs/releases/v2.0.2.md
- **Design:** docs/protocols/ODD.md
- **Approval:** approved
- **Evidence mode:** candidate-v1

El usuario autorizó implementar las cuatro mejoras, actualizar el README y publicar en main y GitHub Releases después de pasar las pruebas. La entrega incluye el catálogo y las correcciones de instalación del mismo trabajo. Para revertir una instalación se usa el respaldo del Manager. Una corrección posterior de la distribución debe tener su propia versión.

## Tasks

### TASK-001 — Retomar una unidad por el comportamiento y reconocer evidencia vencida
- **State:** DONE
- **Depends on:** none
- **Owner:** orchestrator
- **Acceptance:** El perfil y la plantilla orientan el corte vertical y la delegación justificada. El cierre de un tablero nuevo detecta cambios en los archivos declarados y admite registros históricos.
- **Scope:** ["config/AGENTS.md", "docs/protocols/ODD.md", "docs/protocols/MRP.md", "docs/protocols/TTP.md", "scripts/taskboard.py", "templates/feature-tasks.md", "tests/test_taskboard.py", "tests/test_ttp_resources.py", "skills/harness-specification/SKILL.md"]
- **Evidence:** {"command": "python -m unittest discover -s tests -p test_taskboard.py / python -m unittest discover -s tests -p test_ttp_resources.py", "result": "13 pruebas de tablero (1 omitida en Windows), 3 contratos de recursos OK.", "candidate": "sha256:83b3c4cf8c984d57edc3d17767d9ec900fbacba16bb36f7849c002c822627293"}
- **Blocker:** none
- **Reopen reason:** Revalidar dependencias transitivas y explicaciones del README en el candidato final

### TASK-002 — Comparar tres recorridos y entender el flujo desde el README
- **State:** DONE
- **Depends on:** none
- **Owner:** orchestrator
- **Acceptance:** El piloto distingue observaciones faltantes y calidad en un arreglo chico, una interrupción y una decisión de permisos. El README explica cómo usar el flujo.
- **Scope:** ["scripts/workflow_friction.py", "tests/test_workflow_friction.py", "tests/fixtures/workflow-friction/baseline.json", "tests/fixtures/workflow-friction/candidate.json", "docs/evaluations/workflow-friction.md", "docs/EVALUATION.md", "README.md", "assets/diagrams/chuli-odd.html", "assets/diagrams/chuli-odd.svg", "assets/diagrams/chuli-architecture.html", "assets/diagrams/chuli-architecture.svg"]
- **Evidence:** {"command": "python -m unittest discover -s tests -p test_workflow_friction.py / python scripts/workflow_friction.py tests/fixtures/workflow-friction/baseline.json tests/fixtures/workflow-friction/candidate.json", "result": "13 pruebas y tres escenarios sinteticos OK. SVG parseables y enlaces locales presentes.", "candidate": "sha256:c92c8190874e9bfcc2b2cc67cc3033413c056c7d59c9ac0c9371035c998ba284"}
- **Blocker:** none
- **Reopen reason:** Revalidar dependencias transitivas y explicaciones del README en el candidato final

### TASK-003 — Obtener una distribución instalable y comprobada
- **State:** DONE
- **Depends on:** TASK-001, TASK-002
- **Owner:** orchestrator
- **Acceptance:** El candidato pasa las pruebas locales de Go, Python y estructura y permite empaquetar la distribución. Para publicar, GitHub exige además CI exitosa sobre el commit exacto de main. El run y el tag documentan esa operación externa.
- **Scope:** [".github/workflows/release.yml", ".github/workflows/validate.yml", "CHANGELOG.md", "README.md", "VERSION", "assets/diagrams/chuli-architecture.html", "assets/diagrams/chuli-architecture.svg", "assets/diagrams/chuli-odd.html", "assets/diagrams/chuli-odd.svg", "cmd/chuli/main.go", "cmd/chuli/models_discovery_test.go", "cmd/chuli/status_test.go", "cmd/chuli/update_models_test.go", "config/AGENTS.md", "config/models.json", "docs/EVALUATION.md", "docs/INSTALLATION.md", "docs/TUI.md", "docs/evaluations/workflow-friction.md", "docs/protocols/MRP.md", "docs/protocols/ODD.md", "docs/protocols/TTP.md", "docs/releases/v2.0.2.md", "internal/manager/agy_models.go", "internal/manager/agy_models_test.go", "internal/manager/archive_test.go", "internal/manager/catalog.go", "internal/manager/claude_models.go", "internal/manager/claude_models_test.go", "internal/manager/manager.go", "internal/manager/model_discovery.go", "internal/manager/profiles.go", "internal/tui/model_discovery_test.go", "internal/tui/models.go", "internal/tui/presentation.go", "internal/tui/tui.go", "internal/tui/wizard.go", "internal/tui/wizard_batch_test.go", "scripts/client_targets.py", "scripts/install.ps1", "scripts/package_release.py", "scripts/profile_resources.py", "scripts/taskboard.py", "scripts/workflow_friction.py", "skills/harness-specification/SKILL.md", "templates/feature-tasks.md", "tests/fixtures/workflow-friction/baseline.json", "tests/fixtures/workflow-friction/candidate.json", "tests/test_client_targets.py", "tests/test_taskboard.py", "tests/test_toml_config.py", "tests/test_workflow_friction.py", "tests/validate_repository.py", "tests/windows-ttp-resources.ps1"]
- **Evidence:** {"command": "go test -timeout 15m ./... / go vet ./... / python -m unittest discover -s tests / python tests/validate_repository.py / pwsh -NoProfile -File tests/windows-ttp-resources.ps1 / python scripts/package_release.py profile <temp>", "result": "Cinco paquetes Go OK; vet y mod tidy -diff OK; 86 pruebas Python, 83 aprobadas y 3 omitidas; recursos, instalacion, reintentos y rollback en perfiles temporales OK; ZIP validado con tres recorridos sinteticos comparables.", "candidate": "sha256:0f528af5ee4fa021966eb7a7d39fddd07a2853d88e7d12810041e5872591ea48"}
- **Blocker:** none
