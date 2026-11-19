# Feature: Flujo Git y GitHub en tres clientes

- **Spec:** https://github.com/AlejoPalavecino/codex-engineering-harness/issues/64
- **Design:** #diseno
- **Approval:** approved
- **Evidence mode:** candidate-v1

## Diseno

Distribuir el protocolo desde los recursos compartidos existentes. El usuario eligió aviso y justificación para este repo. Refs #64, base origin/main, rama feat/64-git-github-workflow. Un PR cohesivo previsto. El [PR #65](https://github.com/AlejoPalavecino/codex-engineering-harness/pull/65) se publicó como borrador con Refs #64. El [CI del head d4b4e78](https://github.com/AlejoPalavecino/codex-engineering-harness/actions/runs/44620447559) pasó. La ejecución en sesiones quedó no disponible en este entorno. Pendientes revisión humana, etiquetas, CI del próximo candidato e integración.

### TASK-001 — Recuperar y aplicar el flujo desde cada perfil
- **State:** DONE
- **Depends on:** none
- **Owner:** orchestrator
- **Acceptance:** Los tres perfiles resuelven el protocolo, sus seis escenarios y las cuatro skills. La guía de cortes usa autoría y bases inmediatas.
- **Scope:** ["config/AGENTS.md","docs/protocols/GIT-GITHUB.md","docs/protocols/SKILLS.md","skills/branch-pr/SKILL.md","skills/chained-pr/SKILL.md","skills/chained-pr/references/chaining-details.md","skills/issue-creation/SKILL.md","skills/work-unit-commits/SKILL.md","templates/feature-tasks.md","tests/test_git_workflow.py"]
- **Evidence:** {"command":"python -m unittest discover -s tests -p test_git_workflow.py","result":"OK: dos pruebas verifican referencias, seis escenarios, cuatro skills y coherencia de autoría y bases inmediatas en tres perfiles.","candidate":"sha256:8fa9fb99f7833ddd4a5c3056e17b723c721b0e83a45953415234acb1deb7543b"}
- **Blocker:** none

### TASK-002 — Avisar cuando la revisión exceda el presupuesto
- **State:** DONE
- **Depends on:** none
- **Owner:** orchestrator
- **Acceptance:** El check conserva la verificación del candidato y superar 400 produce aviso; la plantilla pide justificar la estrategia.
- **Scope:** [".github/workflows/validate.yml",".github/PULL_REQUEST_TEMPLATE.md","CONTRIBUTING.md","docs/ISSUE_WORKFLOW.md","tests/test_review_budget.py"]
- **Evidence:** {"command":"python -m unittest discover -s tests -p test_review_budget.py","result":"OK: 400, 401, SHA distinto y estadísticas incompletas; API simulada.","candidate":"sha256:1d52bb3893dac601c6d43ac13130a7d6f7cd58680cdaf11cd05ad138e63e82f0"}
- **Blocker:** none

### TASK-003 — Comparar escenarios y registrar disponibilidad de cada cliente
- **State:** DONE
- **Depends on:** TASK-001
- **Owner:** orchestrator
- **Acceptance:** La batería común aporta evidencia de estado Git local por perfil. El informe registra la disponibilidad de ejecución de cada cliente y los límites que impone el host.
- **Scope:** ["docs/evaluations/git-github-parity.md","tests/git_workflow_fixtures.py"]
- **Evidence:** {"command":"python tests/git_workflow_fixtures.py --codex-home <perfil-codex> --claude-home <perfil-claude> --agy-home <perfil-agy>","result":"OK: 18 casos Git locales. Informe de disponibilidad y limites de sesiones en docs/evaluations/git-github-parity.md.","candidate":"sha256:c4a7032a4f011aae185176bf3fcc50d6082a5c803fa7132d2e3c5553e7899629"}
- **Blocker:** none
