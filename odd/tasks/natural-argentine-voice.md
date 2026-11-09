# Una voz argentina que explica mientras se trabaja

- **Spec:** docs/specs/natural-argentine-voice.md
- **Design:** docs/specs/natural-argentine-voice.md#diseño-de-la-conversación
- **Approval:** approved

### TASK-001 — Conservar la dirección de voz aprobada
- **State:** DONE
- **Depends on:** none
- **Owner:** orchestrator
- **Acceptance:** Alcance, diseño, VOZ-01 a VOZ-07, riesgos y reversión reflejan la propuesta aprobada.
- **Evidence:** Comparación con la propuesta aprobada: dirección de voz, explicación cotidiana primero, adaptación, precisión, VOZ-01 a VOZ-07 y reversión conservados en la especificación.
- **Blocker:** none

### TASK-002 — Entregar instrucciones y escenarios coherentes
- **State:** DONE
- **Depends on:** TASK-001
- **Owner:** orchestrator
- **Acceptance:** CES y orquestador implementan la voz aprobada con ejemplos y escenarios trazables, conservando autoridad y precisión.
- **Evidence:** Inspección VOZ-01 a VOZ-07 contra CES y config/AGENTS.md: voz, ejemplos, adaptación, evidencia y límites coherentes. Guía V-01 a V-07 y variante V-04B preparada; conducta NOT_TESTED.
- **Blocker:** none

### TASK-003 — Comprobar recursos y declarar evidencia disponible
- **State:** DONE
- **Depends on:** TASK-002
- **Owner:** orchestrator
- **Acceptance:** Checks estructurales, regresiones relevantes, recursos instalables, enlaces y tablero comprobados; resultados conductuales separados de los documentales.
- **Evidence:** Suite Python OK (42 pruebas, 1 skip); empaquetado e inventario reales conservan CES y orquestador; validador estructural PASS en copia temporal con fixture histórico canónico; enlaces y diff correctos. Campaña conductual NOT_TESTED; límites detallados abajo.
- **Blocker:** none

## Evidencia y límites de esta entrega

- `python -X utf8 -B -m unittest discover -s tests -p 'test_*.py'`: 42 pruebas, OK; 1 omitida por symlinks no disponibles. Se usa UTF-8 explícito por la lectura sin encoding de una prueba existente.
- `python -X utf8 -B scripts/package_release.py profile <directorio-temporal>`: ZIP generado; comparación byte a byte confirma que contiene las versiones actuales de config/AGENTS.md y docs/protocols/CES.md.
- `scripts/profile_resources.py:inventory`: confirma destinos AGENTS.md y harness/docs/protocols/CES.md. No se instaló en el perfil personal.
- `python -X utf8 -B tests/validate_repository.py`: PASS en copia temporal del candidato con el fixture histórico canónico obtenido de HEAD. En el checkout Windows, ese fixture está convertido a CRLF y su hash difiere; se comprobó igualdad al normalizar solo CRLF. No se modificó el fixture ni su hash esperado.
- Enlaces relativos de los Markdown modificados y nuevos: resuelven. `git diff --check`: exit 0. Inspección de CES y orquestador contra VOZ-01 a VOZ-07: coherentes.
- `git diff HEAD -- evals agents scripts`: vacío; catálogo histórico, roles e instalador conservados. No se repitieron tests Go ni scripts de instalación Windows porque esta unidad modifica instrucciones y documentación, y la distribución se comprobó con el empaquetador real.
- V-01 a V-07, incluidas V-04A/B: NOT_TESTED contra un modelo. Los ejemplos y checks no demuestran el tono observado. No se atribuye RED/GREEN conductual a esta entrega documental.
