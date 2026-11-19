# Feature: Roadmap del ciclo de vida de software

- **Spec:** https://github.com/AlejoPalavecino/codex-engineering-harness/issues/63
- **Design:** #entrega
- **Approval:** approved
- **Evidence mode:** candidate-v1

## Entrega

Implementar las seis issues abiertas del roadmap en los protocolos, skills, CLI y evaluaciones existentes. La persona responsable pidió terminar todas las issues. Las operaciones remotas se rigen por la autorización vigente para el repositorio. Sin despliegue de productos ni publicación de release dentro de este alcance.

La publicación y el cierre del roadmap en GitHub fueron autorizados. La entrega de cada corte utiliza su PR, su SHA y sus checks. Los resultados de TASK-001 a TASK-005 se cotejan sobre el candidato conjunto, con su huella, antes de la comprobación de main y del estado remoto en TASK-006. Las tareas siguientes recogen la aceptación de las fases. La validación estructural del tablero no certifica CI remoto ni comportamiento de los clientes.

## Criterios y evidencia de entrega

| Issue | Entrega y comprobación | Límite |
| --- | --- | --- |
| #64 | #65 y #79. Protocolo y recursos en los tres perfiles. Seis escenarios Git por perfil, medición de autoría, preservación de cambios, base inmediata y detección de CI obsoleto. | El informe de paridad registra qué sesiones de cliente no se pudieron comprobar y por qué. |
| #59 | #66, #67, #74 y #75. Trazabilidad por ID, impacto por archivo, preguntas para criterios ambiguos y ADR con pruebas de unicidad y rollback. Sustituir la descripción nueva por la previa en una copia del importador hizo fallar la nueva prueba mientras pasaban las anteriores. | Los umbrales de rendimiento y los conflictos de significado requieren una decisión en contexto. |
| #60 | #76. Fixture de riesgo crítico, estados por caso y rechazo de evidencia vencida tras revalidar una dependencia. Las verificaciones declaradas se consultan sin ejecutar el texto del registro. | La huella depende del Scope y el registro de ejecución requiere corroboración. |
| #61 | #77, #80 y #81. Análisis defensivo en CI, fallos y recuperación con fixtures, verificación de paquetes y preparación para despacho manual por run, SHA e ID de artefacto. | Los binarios de Linux y macOS se prueban en sus runners. La publicación con permisos de escritura permanece separada de la preparación. |
| #62 | #78. Guía de operación, incidente sintético con regresión de dos devoluciones, evidencia ausente o no comparable en métricas y verificación del recurso instalado. | La campaña sintética no aporta métricas de usuarios reales ni de producción. |

### TASK-001 — Flujo Git/GitHub en los tres clientes (#64)
- **State:** DONE
- **Depends on:** none
- **Owner:** orchestrator
- **Acceptance:** Recuperar referencias, contrastar escenarios equivalentes por cliente y verificar integración y CI de la entrega.
- **Scope:** ["docs/protocols/GIT-GITHUB.md", "skills/chained-pr/SKILL.md", "skills/chained-pr/references/chaining-details.md", "tests/git_workflow_fixtures.py", "docs/evaluations/git-github-parity.md", ".github/workflows/validate.yml"]
- **Evidence:** {"command":"python -m unittest discover -s tests -p 'test_git_workflow.py'","result":"Distribución verificada. 18 escenarios locales en perfiles aislados, límites de sesiones documentados en el informe. #65 y #79 integrados tras CI.","candidate":"sha256:4d806af85d02bd7c7514701cc75f6935baa55a8282302c2ba9d508d50789a203"}
- **Blocker:** none

### TASK-002 — Fundamentos y arquitectura (#59)
- **State:** DONE
- **Depends on:** none
- **Owner:** orchestrator
- **Acceptance:** Trazar requisitos, consultar impacto, revisar ambigüedades y justificar arquitectura con un contrato ejecutable.
- **Scope:** ["scripts/taskboard.py", "config/AGENTS.md", "skills/harness-architecture/SKILL.md", "skills/harness-discovery/SKILL.md", "skills/harness-specification/SKILL.md", "templates/feature-tasks.md", "docs/protocols/TTP.md", "evals/fixtures/importer/ADR-001.md", "evals/fixtures/importer/tests/test_adr_inventory.py", "tests/test_acceptance_review.py", "tests/test_taskboard_trace.py", "tests/test_taskboard_impact.py", "evals/cases/r2-architecture.json"]
- **Evidence:** {"command":"python -m unittest discover -s tests -p 'test_acceptance_review.py'","candidate":"sha256:f60d1c27912adbbd0d4eec9cf12294ab2b92a619d8cac007ac141df87cc2577a","result":"3 pruebas verdes; PR #66/#67/#74/#75 integrados. Además se verificaron 4 pruebas del importador y se compararon check/show/impact sobre 21 tableros."}
- **Blocker:** none

### TASK-003 — Calidad y evidencia (#60)
- **State:** DONE
- **Depends on:** none
- **Owner:** orchestrator
- **Acceptance:** Distinguir estados de ejecución y rechazar evidencia desactualizada de criterio o candidato en los escenarios del gate.
- **Scope:** ["scripts/taskboard.py", "docs/protocols/TTP.md", "skills/harness-tdd-qa/SKILL.md", "tests/test_taskboard_quality.py", "tests/fixtures/taskboard-quality.md"]
- **Evidence:** {"command":"python -m unittest discover -s tests -p 'test_taskboard_quality.py'","candidate":"sha256:4e22c47776b789f303e349df0b71718ece8b59cec834294d772fb5d2b582bbb6","result":"7 pruebas verdes sobre estados, drift transitivo, aprobación independiente y selección de regresiones. #76 integrado."}
- **Blocker:** none

### TASK-004 — Seguridad y entrega (#61)
- **State:** DONE
- **Depends on:** none
- **Owner:** orchestrator
- **Acceptance:** Validar preparación de release, rutas de rechazo y recuperación con permisos mínimos y aprobación de publicación.
- **Scope:** [".github/workflows/release.yml", ".github/workflows/validate.yml", "scripts/verify_release.py", "tests/test_verify_release.py", "docs/RELEASING.md", "docs/protocols/ASEP.md"]
- **Evidence:** {"command":"python -m unittest discover -s tests -p 'test_verify_release.py'","result":"12 pruebas. CI de #81 verificó seguridad y preparación de candidato; el primer intento de un job Ubuntu abortó en PowerShell, el segundo pasó sin cambiar el SHA. Despacho de publicación sin ejecutar.","candidate":"sha256:e8f362deb4ff2aa48411a82266dc048c7bf3efd24183d44dae13bda06dcad78f"}
- **Blocker:** none

### TASK-005 — Operación y evolución (#62)
- **State:** DONE
- **Depends on:** none
- **Owner:** orchestrator
- **Acceptance:** Relacionar incidentes, cambios, métricas y requisitos; distinguir evidencia de producto y comparación del harness.
- **Scope:** ["docs/protocols/OPERATIONS.md", "config/AGENTS.md", "docs/EVALUATION.md", "tests/test_profile_resources.py", "tests/windows-ttp-resources.ps1"]
- **Evidence:** {"command":"python -m unittest discover -s tests -p 'test_workflow_friction.py'","candidate":"sha256:c74ed8578efe317f3aacc2de6bbbce53e92eca8ff7435ee35c323466d0a7aa8b","result":"Comparador offline y casos inválidos comprobados. Guía instalada en tres perfiles; escenario sintético de incidente y dos motores de rollback verificados en #78."}
- **Blocker:** none

### TASK-006 — Integración y cierre (#63)
- **State:** DONE
- **Depends on:** TASK-001, TASK-002, TASK-003, TASK-004, TASK-005
- **Owner:** orchestrator
- **Acceptance:** Auditar las cinco fases contra main, comprobar CI y cierres de las cinco issues de fase. La issue #63 registra el resultado remoto del último PR después de integrar este tablero.
- **Scope:** ["docs/protocols/EWP.md"]
- **Evidence:** {"candidate":"sha256:804a8d3249ead01562290cf5a9115212c3026d17aa61fcb959ae377b8f1da0ce","result":"Cinco fases y seis issues de implementación cerradas como COMPLETED; once PR integrados con CI verde, todos ancestros de f098950. El cierre remoto de #63 se confirma al finalizar su PR de seguimiento.","command":"gh issue view N --json state,stateReason; gh pr view N --json state,mergeCommit,statusCheckRollup; git merge-base --is-ancestor MERGE origin/main"}
- **Blocker:** none
