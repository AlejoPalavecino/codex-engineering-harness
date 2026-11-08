# Chuli 1.0 · Perfil de ingeniería

- **Spec:** docs/releases/v1.0.0.md
- **Design:** docs/ROADMAP.md
- **Approval:** approved

### TASK-001 — Distribuir el catálogo y los roles
- **State:** DONE
- **Depends on:** none
- **Owner:** solution_architect
- **Acceptance:** Las cinco skills originales, dieciocho de Gentle AI y seis roles tienen definiciones verificables, recursos completos y procedencia.
- **Evidence:** tests/validate_repository.py, tests/test_ttp_resources.py y THIRD_PARTY_NOTICES.md.
- **Blocker:** none

### TASK-002 — Instalar, actualizar y revertir el perfil
- **State:** DONE
- **Depends on:** TASK-001
- **Owner:** chuli_worker
- **Acceptance:** La vista previa conserva el perfil, el instalador protege archivos ajenos y el Manager verifica y recupera sin un checkout.
- **Evidence:** Go E2E con dos versiones temporales, 13 casos PowerShell y rollback de perfiles reales 0.2.0/0.2.1. Evidencia en docs/TESTING.md.
- **Blocker:** none

### TASK-003 — Preparar y vincular la publicación al CI
- **State:** DONE
- **Depends on:** TASK-002
- **Owner:** quality_reviewer
- **Acceptance:** El workflow obtiene binarios nativos y perfil del run validado, comprueba el SHA y publica el tag y sus sumas.
- **Evidence:** scripts/package_release.py, .github/workflows/release.yml y docs/releases/v1.0.0.md. La ejecución de la publicación queda registrada en el issue #17 y en el tag v1.0.0.
- **Blocker:** none
