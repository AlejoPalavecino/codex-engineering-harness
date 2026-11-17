# MRP v1.0 — Model Routing & Reasoning Policy

**Objetivo:** mínimo consumo total para cumplir la calidad, no máximo razonamiento ni mayor número de subagentes.

| Rol | Modelo inicial sugerido | Esfuerzo |
| --- | --- | --- |
| Orchestrator | GPT-6.1 Sol | Medium |
| `chuli_explorer` | GPT-6 Luna | Medium |
| Solution Architect | GPT-6.1 Sol | High |
| Product Designer | GPT-6.1 Terra | Medium |
| Test Engineer | GPT-6.1 Sol | Medium |
| `chuli_worker` | GPT-6.1 Sol | Medium |
| Quality Reviewer | GPT-6.1 Sol | High |

GPT-6 Astra es escalamiento **excepcional** por insuficiencia de razonamiento, no por tamaño del repo.

## Ruta de decisión

Antes de despachar un especialista, escribí en el encargo la pregunta o resultado concreto, la superficie que le corresponde y la prueba que habilita el próximo paso. Identificá qué podés resolver vos en paralelo. Si se necesita independencia, nombrá el riesgo y el paso que depende de la revisión. Si responder el encargo cuesta lo mismo que resolverlo en línea y no aporta independencia, resolvelo en la sesión actual.

En una instalación multicliente, una persona puede probar que reintentar Claude conserva el perfil de Codex mientras otra verifica qué modelos admite el selector. Ambos resultados se pueden integrar con pruebas independientes. En cambio, corregir la codificación de lectura de un archivo de prueba y volver a ejecutarla cabe en un solo turno de trabajo.

1. Determinar tipo de trabajo, dificultad, riesgo y necesidad de herramientas visuales.
2. Si faltan requisitos, consultar al usuario o investigar: cambiar de modelo no resuelve una definición pendiente.
3. Preferir trabajo inline en tareas pequeñas. Delegar al especialista solo si aporta valor.
4. Elegir esfuerzo low/medium/high compatible con el modelo y la versión real de Codex.
5. Ante fallo, distinguir falta de requisitos, contexto, herramientas o capacidad del modelo. Escalar solo en el último caso y con evidencia.

Los IDs de modelos de los archivos TOML son valores solicitados por el perfil, **no garantía de disponibilidad universal**. Codex Desktop puede no honrar una selección de subagente en todas las versiones. Verificar mediante sesión real. La suscripción no equivale necesariamente a un presupuesto de API por token.
