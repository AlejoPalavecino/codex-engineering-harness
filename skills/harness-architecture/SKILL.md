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
