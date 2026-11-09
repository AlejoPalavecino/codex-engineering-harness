# Diseño por tareas y revisión con egolite

- **Spec:** docs/specs/task-based-ux-review.md
- **Design:** docs/specs/task-based-ux-review.md#diseño-del-procedimiento
- **Approval:** approved

### TASK-001 — Conservar alcance y aceptación
- **State:** DONE
- **Depends on:** none
- **Owner:** orchestrator
- **Acceptance:** UX-01 a UX-07, diseño, riesgos y reversión reflejan la recomendación aprobada.
- **Evidence:** Inspección de la propuesta aprobada: alcance, UX-01 a UX-07, procedimiento, límites y reversión conservados en docs/specs/task-based-ux-review.md.
- **Blocker:** none

### TASK-002 — Entregar diseño y revisión ejecutables
- **State:** DONE
- **Depends on:** TASK-001
- **Owner:** orchestrator
- **Acceptance:** Skills y guías coherentes, referencias instalables, escenario de formulario con errores y teclado.
- **Evidence:** UX-01 a UX-07 trazados: ambas skills, referencias portables, catálogo y escenario sintético implementados. RED por referencia faltante observado; prueba focalizada GREEN. No se agregan skills externas.
- **Blocker:** none

### TASK-003 — Comprobar distribución, regresiones y escenario
- **State:** DONE
- **Depends on:** TASK-002
- **Owner:** orchestrator
- **Acceptance:** Pruebas y recursos comprobados; intento real con egolite y límites de evidencia registrados sin equipararlo a investigación humana o campaña del modelo.
- **Evidence:** 43 pruebas Python OK (1 skip); RED/GREEN de distribución; ZIP e instalación/verificación/rollback con PowerShell 7 en perfil temporal; guion funcional real y capturas inspeccionadas con egolite. Limitaciones de PS5.1, iframe y campaña conductual documentadas abajo.
- **Blocker:** none

## Evidencia de la entrega

- RED observado: `python -X utf8 -B -m unittest discover -s tests -p test_ui_review_resources.py` falló por referencia distribuible ausente. GREEN: el mismo comando pasó al agregar guías y enlaces. La prueba copia el inventario a un perfil temporal y comprueba referencias portables; no busca frases en prompts para simular conducta.
- `python -X utf8 -B -m unittest discover -s tests -p 'test_*.py'`: 43 pruebas, OK, 1 omitida por falta de symlinks. `git diff --check` y enlaces locales: correctos.
- Validador estructural PASS en copia temporal del candidato con el fixture histórico canónico de HEAD. El checkout Windows convierte ese fixture a CRLF; no se modificó el fixture original ni su hash esperado.
- Empaquetador real en copia temporal con índice Git del candidato: ZIP contiene ambas referencias byte a byte. Scripts install.ps1 -Apply, verify.ps1 y rollback.ps1 -Apply con pwsh y CodexHome temporal: PASS; guías instaladas iguales al origen, instrucciones personales sintéticas restauradas y referencias nuevas retiradas por rollback.
- El primer intento con Windows PowerShell 5.1 no pudo completar instalación por Get-FileHash ausente en ese proceso. No se modificó el instalador ni se afirma soporte probado en ese motor para esta entrega. PowerShell 7 pasó los tres pasos.
- Guion manual con egolite: formulario vacío, código desconocido, error de red recuperable, éxito y repetición conservando contador; teclado y foco observados. Capturas inspeccionadas con view_image y anchos de documento 390/1280 comprobados dentro de iframe, sin emulación móvil real. Registro detallado en docs/evaluations/task-based-ux-review.md.
- Campaña conductual del orquestador, variante sin navegador, comprensión humana y accesibilidad completa siguen NOT_TESTED. No se modificaron evals/, roles o scripts; no se repitieron pruebas Go al no cambiar código Go ni contratos del Manager.
