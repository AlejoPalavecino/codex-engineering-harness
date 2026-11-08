# Roles y contratos del perfil Codex

El agente principal lee el bloque de Chuli en el `AGENTS.md` del perfil. Puede trabajar en línea y delegar en seis roles declarados en TOML. Codex determina qué roles y modelos están efectivamente disponibles en la sesión.

| Rol | Modelo declarado | Esfuerzo | Resultado |
| --- | --- | --- | --- |
| Orquestador | Perfil elegido por el usuario | Según sesión | Enrutamiento EWP, autoridad, continuidad y cierre |
| [chuli_explorer](../agents/chuli_explorer.toml) | gpt-6-luna | medium | Mapa de código y contratos con referencias; solicita sandbox de lectura |
| [chuli_worker](../agents/chuli_worker.toml) | gpt-6.1-sol | medium | Unidad autorizada con regresiones, verificación y reversión acotada |
| [solution_architect](../agents/solution_architect.toml) | gpt-6.1-sol | high | Alternativas, restricciones y decisiones pendientes |
| [product_designer](../agents/product_designer.toml) | gpt-6.1-terra | medium | Flujos, estados, identidad y accesibilidad |
| [test_engineer](../agents/test_engineer.toml) | gpt-6.1-sol | medium | Pruebas vinculadas con aceptación y riesgo |
| [quality_reviewer](../agents/quality_reviewer.toml) | gpt-6.1-sol | high | Revisión de un candidato fijo con evidencia independiente |

## Contrato de una delegación

Indicá objetivo, tarea, criterios, rutas de contexto y skills concretas, herramientas requeridas, superficie de escritura, restricciones, evidencia y condición de parada. El orquestador conserva la conversación de producto y escribe el tablero. Si el cliente carece de delegación, puede hacer trabajo autorizado en línea y dejar pendiente lo que requiera independencia.

No cambies los nombres de los dos roles `chuli_*` por `explorer` o `worker` nativos. Esos alias pueden corresponder a otros modelos o no existir. Los archivos de rol expresan los valores pedidos. Para afirmar que el cliente los honró hay que observar una sesión.

## Procedimientos

El [catálogo de skills](protocols/SKILLS.md) identifica las cinco nativas y las 18 importadas, su runtime y requisitos. El [protocolo ODD](protocols/ODD.md) describe cómo combinar las decisiones del orquestador, TDD y la revisión según riesgo.
