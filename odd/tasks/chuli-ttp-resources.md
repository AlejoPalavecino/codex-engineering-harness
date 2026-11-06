# Feature: Distribución y uso de recursos TTP

- **Spec:** docs/specs/ttp-resources.md
- **Design:** docs/specs/ttp-resources.md
- **Approval:** approved

> Usuario autorizó diseño e implementación el 2026-11-05. Orquestador único escritor. No instalar en el home principal ni cambiar modelos, memorias reales o release.

## Tasks

### TASK-001 — Fijar contrato verificable del paquete TTP
- **State:** DONE
- **Depends on:** none
- **Owner:** test_engineer
- **Acceptance:** Rutas, preservación, rollback y formato canónico definidos; pruebas RED reproducen ausencia de recursos sin alterar producción.
- **Evidence:** Test engineer: RED Python (3 tests, dos subtests de referencias ausentes) y RED PowerShell (6 escenarios por recursos ausentes). Contrato revisado; test PS tuvo corrección por su autor de captura stderr5.1 sin debilitar exit/hash/snapshot. Hashes inicial y corregido registrados.
- **Blocker:** none

### TASK-002 — Entregar recursos canónicos localizables
- **State:** DONE
- **Depends on:** TASK-001
- **Owner:** orchestrator
- **Acceptance:** Instalador usa fuentes únicas y gestión existente; instrucciones/skill/guía apuntan a rutas instaladas y exigen check sin imponer tablero a R0/R1.
- **Evidence:** Entradas de instalador y referencias locales implementadas sin nuevas fuentes ni esquema. Python32 PASS; seis escenarios en PS5.1 y7 PASS. Revisión independiente VERIFIED sobre cd89872; recursos e instrucciones instalados por script real y copias exactas comprobadas.
- **Blocker:** none

### TASK-003 — Verificar instalación, rollback y comportamiento real
- **State:** DONE
- **Depends on:** TASK-002
- **Owner:** orchestrator
- **Acceptance:** Casos de instalación/personalización/rollback y una generación real canónica comprobados; R0/R1 proporcionales y fixtures/criterios originales conservados.
- **Evidence:** Seis escenarios en PS5.1/7 PASS; upgrade real desde edf2d4f y rollback restauran hashes/manifiesto y preservan archivo ajeno. Sesión R2 genera TASK-001/002/003 canónicos; dos checks del agente y uno observador exit0, seis tests PASS. R0 sin mutaciones; R1 solo literal autorizado y cuatro tests del observador PASS.
- **Blocker:** none

### TASK-004 — Cerrar revisión, CI y entrega
- **State:** DONE
- **Depends on:** TASK-003
- **Owner:** orchestrator
- **Acceptance:** Revisión independiente y checks pertinentes verdes; evidencia privada y resumen sanitizado; credenciales temporales retiradas y hashes personales intactos.
- **Evidence:** Revisión independiente VERIFIED sobre cd89872, sin bloqueantes; checks locales completos exit0 y CI 42141105008 con cinco jobs SUCCESS. Informe docs/evaluations/2026-11-05-ttp-resources.md; evidencia privada y comparación conservadas. Cero auth temporales restantes; doce hashes personales sin cambios. Tablero propio validado por taskboard.py check.
- **Blocker:** none
