---
name: harness-architecture
description: "Comparar stacks, patrones, arquitectura, dominio, persistencia e integraciones según restricciones; ADR opcional para decisiones duraderas. EWP R2 o cambios estructurales."
---

# Solution Architecture — opciones fundamentadas

1. Leer requisitos y límites aprobados, restricciones de equipo, presupuesto, rendimiento, disponibilidad, seguridad y operación.
2. Para un sistema existente, mapear contratos y estimar primero el costo de conservar la arquitectura.
3. Proponer dos o tres alternativas **viables**, no alternativas de relleno. Evaluar simplicidad, mantenibilidad, resiliencia, integración, coste y riesgos.
4. Considerar monolito, monolito modular, capas, hexagonal, DDD, sistemas distribuidos o eventos solo cuando el problema lo justifique.
5. Para cada alternativa: ventaja, desventaja, costo operativo, costos de cambio, riesgo y casos en que es apropiada.
6. Devolver recomendación argumentada, incertidumbres y decisiones para aprobación. Preparar ADR si existe decisión relevante y duradera.
7. No crear carpetas de arquitectura ni dependencias antes de autorización.

La arquitectura se elige para resolver el dominio, no para demostrar que conocemos patrones.

## Vincular la decisión con el comportamiento

Cuando una decisión sea costosa de revertir o cambie un contrato compartido, asignale un ID estable como `ADR-001`. Conservá contexto y requisitos afectados, restricciones, alternativas con sus costos, opción elegida, estado, quién decidió y consecuencias. Incluí una condición para reconsiderarla. Referenciá el ADR desde Design en TTP. Para una decisión sustituida, conservá el anterior y enlazá su sucesor.

Elegí la vista mínima que ayude a decidir: entidades y relaciones para el dominio, componentes para los límites, secuencia para el orden de mensajes, modelo de datos para las invariantes y flujo para las decisiones y errores. Si una línea explica la relación, alcanza con esa línea.

Antes de cerrar una tarea afectada, recorré requisito → decisión → punto de entrada real → prueba. Para un requisito funcional, ejercitá un resultado visible y un caso límite. Para un atributo de calidad, fijá escenario, unidad y criterio con el responsable. Registrá el comando, candidato y resultado en la evidencia existente. Si el código contradice el ADR, describí el comportamiento observado, evaluá el impacto y acordá corregir el código o sustituir el ADR. Una medición pendiente permanece abierta.
