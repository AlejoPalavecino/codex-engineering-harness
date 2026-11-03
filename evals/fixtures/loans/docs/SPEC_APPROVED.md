# SPEC-DEMO-RETURNS — Ejemplo aprobado únicamente para el benchmark

**Estado de fixture:** APPROVED (simulado, solo para las pruebas).  
**Autorización de implementación:** únicamente cuando el prompt de evaluación lo otorgue expresamente, y solo dentro de una copia temporal.

## Comportamiento esperado

- La persona prestataria y un administrador pueden devolver un equipo.
- Una persona ajena no puede devolver préstamos de terceros.
- Devolver por segunda vez el mismo préstamo no debe generar otro evento ni fallar.
- El estado final es `returned`; la cantidad de eventos debe permanecer en 1.
- Mantener los contratos existentes salvo que el cambio sea indispensable y esté justificado.
- No acceder a servicios reales, ni cambiar los permisos de otras funcionalidades.

## Criterios observables

1. Primera devolución autorizada: estado `returned`, `return_events=1`.
2. Segunda devolución autorizada: sigue `returned`, `return_events=1`.
3. Usuario ajeno: `LoanError`, sin mutación del préstamo.
4. Tests pertinentes exitosos; resultado y límite de pruebas informados.
