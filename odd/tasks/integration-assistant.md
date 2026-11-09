# Asistente de integraciones

- **Spec:** docs/specs/integration-assistant.md
- **Design:** docs/specs/integration-assistant.md#diseño-del-recorrido
- **Approval:** approved

Entrega autorizada: cinco PRs apiladas hacia main; issue #51 con status:approved confirmado. Cortes 01 configuración, 02 catálogo, 03 diagnóstico MCP, 04 comandos, 05 flujo TUI. Los cuatro primeros referencian #51; el último lo cierra al integrarse.

### TASK-001 — Configurar y revertir sin afectar conexiones ajenas
- **State:** DONE
- **Depends on:** none
- **Owner:** orchestrator
- **Acceptance:** INT-02 e INT-03, configuración segura y detección conservadora.
- **Evidence:** RED por helper ausente; GREEN de preservación, duplicados, vista previa obsoleta, reversión conservadora, bloqueo, registro corrupto y reparse flag. Pruebas focalizadas Python/Go y revisión independiente sin nuevos fallos de preservación. No se usó perfil personal.
- **Blocker:** none

### TASK-002 — Guiar y comprobar las cinco integraciones
- **State:** DONE
- **Depends on:** TASK-001
- **Owner:** orchestrator
- **Acceptance:** INT-04 e INT-05, catálogo, guías y diagnósticos con límites explícitos.
- **Evidence:** Catálogo con cinco guías oficiales; pruebas HTTP JSON/SSE con initialize, notifications/initialized y tools/list, autenticación, errores, timeout y rechazo de redirecciones. Regresión SSE intermedio corregida. Comandos locales sólo detectados, sin ejecutar Engram ni usar memoria.
- **Blocker:** none

### TASK-003 — Completar el recorrido del Manager y verificar
- **State:** DONE
- **Depends on:** TASK-002
- **Owner:** orchestrator
- **Acceptance:** INT-01 e INT-06, interacción, regresiones y documentación.
- **Evidence:** RED del TUI sin estados de integración; GREEN de navegación, cancelación, confirmación y reversión, texto desplazable/ancho reducido y CLI con --apply explícito. Terminal real (PTY, perfil temporal): catálogo → Exa → cancelar (config inexistente) → confirmar (config creado) → deshacer (config inexistente), salida normal. `go test ./...`, `go vet ./...` y build del Manager OK. Python: 53 pruebas OK, 2 skip por symlinks. Validador estructural OK en copia con fixture histórica canónica para evitar el problema CRLF preexistente. Revisión independiente: junction real rechazado con API 3.11 simulada y regresiones SSE numérico/string-ID resueltas; test HTTP adicional verifica respuesta al ping. Sin servicios, cuentas o perfil personal reales; Python 3.11 real y plataformas Linux/macOS no ejecutados localmente. Entrega por PRs autorizada posteriormente; sin merge ni release.
- **Blocker:** none
