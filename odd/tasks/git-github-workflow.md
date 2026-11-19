# Feature: Flujo Git y GitHub en tres clientes

- **Spec:** https://github.com/AlejoPalavecino/codex-engineering-harness/issues/64
- **Design:** #diseno
- **Approval:** approved
- **Evidence mode:** candidate-v1

## Diseno

Distribuir el protocolo desde los recursos compartidos existentes. El usuario eligió aviso y justificación para este repo. Refs #64, base origin/main, rama feat/64-git-github-workflow. Un PR cohesivo previsto. Commit, PR, CI remoto e integración pendientes.

### TASK-001 — Recuperar y aplicar el flujo desde cada perfil
- **State:** DONE
- **Depends on:** none
- **Owner:** orchestrator
- **Acceptance:** Los tres perfiles resuelven el mismo protocolo y los escenarios de política, autorización, revisión y continuidad.
- **Scope:** ["config/AGENTS.md","docs/protocols/GIT-GITHUB.md","docs/protocols/SKILLS.md","skills/branch-pr/SKILL.md","skills/chained-pr/SKILL.md","skills/issue-creation/SKILL.md","skills/work-unit-commits/SKILL.md","templates/feature-tasks.md","tests/test_git_workflow.py"]
- **Evidence:** {"command":"python -m unittest discover -s tests -p test_git_workflow.py","result":"OK: referencias, seis escenarios y cuatro skills en tres perfiles. Conducta de sesiones reales pendiente.","candidate":"sha256:5479756afba15a433b5dd41ad9e31bd7c3b49582ad3130901e5a0404a2893289"}
- **Blocker:** none

### TASK-002 — Avisar cuando la revisión exceda el presupuesto
- **State:** DONE
- **Depends on:** none
- **Owner:** orchestrator
- **Acceptance:** El check conserva la verificación del candidato y superar 400 produce aviso; la plantilla pide justificar la estrategia.
- **Scope:** [".github/workflows/validate.yml",".github/PULL_REQUEST_TEMPLATE.md","CONTRIBUTING.md","docs/ISSUE_WORKFLOW.md","tests/test_review_budget.py"]
- **Evidence:** {"command":"python -m unittest discover -s tests -p test_review_budget.py","result":"OK: 400, 401, SHA distinto y estadísticas incompletas; API simulada.","candidate":"sha256:1d52bb3893dac601c6d43ac13130a7d6f7cd58680cdaf11cd05ad138e63e82f0"}
- **Blocker:** none
