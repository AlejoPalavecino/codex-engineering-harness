# Roadmap (propuesto, no comprometido)

## v0.1 — Fundamentos reutilizables
- Protocolos, agentes, cinco skills originales, instalador Windows, verificación y documentación pública.
- Probar que la integración no toca Engram ni configura MCP con credenciales ficticias.

## v0.1.1 — Seguimiento de trabajo
- Registro persistente de tareas por funcionalidad, vinculadas a specs y diseño aprobados.
- Checklist nativo `update_plan` habilitado con backup, sujeto a compatibilidad de Codex Desktop.
- Protocolo TTP y validador de estados/dependencias/evidencias; pruebas de instalación y rollback.

## v0.1.2 — Instalación y rollback seguros (candidata)
- Validación TOML original/candidato antes de modificar Codex.
- Plan de actualización por archivo y preservación de personalizaciones sin reemplazos silenciosos.
- Prevalidación integral del rollback y pruebas de regresión de los tres problemas P1.
- Verificación por capas, distinguiendo sintaxis, integridad y modificaciones ajenas válidas.
- Publicación mediante autorización humana explícita del SHA probado.

## v0.2 — Routing medido
- Benchmarks reproducibles R0/R1/R2, modelos efectivamente usados, consumo observado.
- Mejorar estrategias de delegación según resultados, no por intuición.

## v0.3 — Aislamiento efectivo
- Ejecuciones separadas, permisos por ruta, pruebas protegidas y verificador de solo lectura real.
- Tests de seguridad que intenten violar restricciones.

## v0.4 — Mantenimiento
- Actualizaciones incrementales idempotentes y migraciones entre versiones.
- Compatibilidad con más versiones de Codex y sistemas operativos, si hay demanda.

## v1.0 — Release estable
- Instalación, rollback, documentación y compatibilidad comprobadas en entornos reales.
- Políticas de calidad y seguridad respaldadas por evidencia, no por prompts.
