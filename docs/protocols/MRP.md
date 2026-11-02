# MRP v1.0 — Model Routing & Reasoning Policy

**Objetivo:** mínimo consumo total para cumplir la calidad, no máximo razonamiento ni mayor número de subagentes.

| Rol | Modelo inicial sugerido | Esfuerzo |
| --- | --- | --- |
| Orchestrator | GPT-6.1 Sol | Medium |
| Explorer | GPT-6 Luna (si se permite elegir al delegar) | Medium |
| Solution Architect | GPT-6.1 Sol | High |
| Product Designer | GPT-6.1 Terra | Medium |
| Test Engineer | GPT-6.1 Sol | Medium |
| Implementation Worker | GPT-6.1 Sol | Medium |
| Quality Reviewer | GPT-6.1 Sol | High |

GPT-6 Astra es escalamiento **excepcional** por insuficiencia de razonamiento, no por tamaño del repo.

## Ruta de decisión

1. Determinar tipo de trabajo, dificultad, riesgo y necesidad de herramientas visuales.
2. Si faltan requisitos, consultar al usuario o investigar: cambiar de modelo no resuelve una definición pendiente.
3. Preferir trabajo inline en tareas pequeñas. Delegar al especialista solo si aporta valor.
4. Elegir esfuerzo low/medium/high compatible con el modelo y la versión real de Codex.
5. Ante fallo, distinguir falta de requisitos, contexto, herramientas o capacidad del modelo. Escalar solo en el último caso y con evidencia.

Los IDs de modelos de los archivos TOML son candidatos validados en un entorno, **no garantía de disponibilidad universal**. Codex Desktop puede no honrar una selección de subagente en todas las versiones. Verificar mediante sesión real. La suscripción no equivale necesariamente a un presupuesto de API por token.
