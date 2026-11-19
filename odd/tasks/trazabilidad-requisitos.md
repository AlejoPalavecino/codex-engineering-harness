# Feature: Trazabilidad de requisitos (#59)

- **Spec:** #alcance
- **Design:** #diseno
- **Approval:** approved
- **Evidence mode:** candidate-v1
- **Trace mode:** requirements-v1

## Alcance

- **REQ-001:** Consultar qué criterio, tarea, código y prueba corresponden a un requisito aprobado.
- **REQ-002:** Conservar el contrato de los tableros existentes sin migración.

## Diseno

Extensión opcional del mismo Markdown y del CLI instalado. Una tarea referencia requisitos definidos en la especificación del tablero mediante IDs locales estables. El criterio y la evidencia mantienen sus campos actuales. La consulta solo lee. Para revertir, retirar la extensión del validador y de las guías; los tableros anteriores mantienen su estructura.

### TASK-001 — Consultar las tareas y verificaciones afectadas por un requisito
- **State:** DONE
- **Depends on:** none
- **Owner:** orchestrator
- **Requirements:** REQ-001, REQ-002
- **Acceptance:** La consulta filtra por ID, distingue prueba prevista de evidencia registrada y los tableros anteriores siguen funcionando.
- **Verification:** python -m unittest discover -s tests -p test_taskboard*.py
- **Scope:** ["scripts/taskboard.py","tests/test_taskboard_trace.py","docs/protocols/TTP.md","templates/feature-tasks.md","skills/harness-specification/SKILL.md"]
- **Evidence:** {"result":"89 pruebas, 3 omitidas, sin fallos. check y show idénticos a main en 20 tableros.","command":"python -m unittest discover -s tests","candidate":"sha256:e809745152522ae843b3dd7aae671f7472c9b46da7090be7a5bc0618d9f378e7"}
- **Blocker:** none
