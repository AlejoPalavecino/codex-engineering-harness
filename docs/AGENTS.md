# Catálogo de roles y contratos

El harness tiene **siete roles conceptuales**, que no tienen por qué ejecutarse todos.

| Rol | Implementación | Resultado |
| --- | --- | --- |
| Engineering Orchestrator | Agente principal + `config/AGENTS.md` | Selecciona ruta EWP, consulta al humano, delega y valida |
| Explorer | Nativo `explorer` (si disponible) | Referencias precisas a código, evidencia sin escritura |
| Implementation Worker | Nativo `worker` (si disponible) | Implementación acotada de lo aprobado |
| Solution Architect | `agents/solution_architect.toml` | Alternativas, trade-offs, riesgos y decisiones humanas |
| Product Designer | `agents/product_designer.toml` | Flujos, interfaces, estados y dirección visual |
| Test Engineer | `agents/test_engineer.toml` | Pruebas vinculadas a criterios de aceptación |
| Quality Reviewer | `agents/quality_reviewer.toml` | Hallazgos verificables y resultado independiente |

## Contrato mínimo de delegación

Toda delegación indica: objetivo; referencias de contexto; alcance autorizado; herramientas necesarias; superficies de escritura; restricciones; salida esperada; criterios de aceptación; evidencia y condiciones de parada.

Delegar por exploración extensa, especialidad diferenciada, paralelismo sin conflictos o verificación de riesgo. No delegar cambios triviales. Evitar dos escritores sobre los mismos archivos. **Los permisos declarados en TOML no restringen por sí solos el sistema operativo.**

## Skills

- `harness-discovery`: alcance, actores, procesos y permisos.
- `harness-architecture`: elección fundamentada de solución y stack.
- `harness-ux-ui`: flujos, experiencia, identidad visual y accesibilidad.
- `harness-specification`: criterios de aceptación, contratos y plan.
- `harness-tdd-qa`: test-first y pruebas según riesgo.

Para incorporar habilidades externas se requiere respetar su licencia y evitar duplicación.
