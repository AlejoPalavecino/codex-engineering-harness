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

## v0.1.3 — Chuli: identidad y TUI
- Producto y comando `chuli`, preservando referencias existentes y el CODEX_HOME de usuarios.
- Manager en Go con menú interactivo, vista previa y confirmación, consulta de GitHub Releases, diagnóstico y rollback sobre scripts de v0.1.2.
- CI compilando y probando Go en Windows, Linux y macOS; instalación sobre Codex habilitada solo en Windows.
- Binarios oficiales para Windows, Linux y macOS con archivo SHA-256; autoactualización segura y adaptadores Antigravity quedan para próximas versiones.

## v0.2 — Evaluación y optimización medida (publicada)
- **v0.2.0:** 11 escenarios reproducibles, rúbricas, snapshots SHA-256, registros de observación, reportes comparativos, preflight offline y recursos TTP canónicos instalados.
- Baseline de código: tag v0.1.3 sin modificar agentes. La evidencia conductual publicada corresponde a Codex CLI en Windows y no certifica Codex Desktop ni aislamiento efectivo.
- No atribuir a CI pruebas de razonamiento que no se ejecutaron; el aislamiento efectivo queda para v0.3.

## v0.2.1 — Mantenimiento y confiabilidad (candidata)
- Preservar el inventario completo de recursos administrados entre reinstalaciones y migrar manifiestos v0.1.3/v0.2.0 que quedaron sin inventario.
- Verificar los 13 recursos obligatorios y fallar con código distinto de cero cuando falte alguno.
- Rechazar rollback con ejecutores antiguos, manifiestos corruptos, rutas inseguras, esquemas desconocidos o backups incompletos antes de escribir.
- Repetir regresiones en Windows PowerShell 5.1 y PowerShell 7; no sumar agentes, MCP ni cambios conductuales.

## v0.3 — Aislamiento efectivo
- Separar ejecución y superficies de escritura por tarea; permisos efectivos por ruta, no instrucciones declarativas.
- Proteger pruebas y criterios de aceptación fuera del alcance de escritura del implementador.
- Añadir un verificador independiente y de solo lectura, con evidencia de proceso y diff.
- Tests adversariales de escritura no autorizada, traversal/symlinks, subprocessos, red y alteración de pruebas.
- No declarar aislamiento real hasta demostrar que las restricciones sobreviven a intentos de evasión y a procesos hijos.

## v0.4 — Mantenimiento
- Actualizaciones incrementales idempotentes y migraciones entre versiones.
- Compatibilidad con más versiones de Codex y sistemas operativos, si hay demanda.

## v1.0 — Release estable
- Instalación, rollback, documentación y compatibilidad comprobadas en entornos reales.
- Políticas de calidad y seguridad respaldadas por evidencia, no por prompts.
