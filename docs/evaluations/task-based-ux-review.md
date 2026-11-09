# Escenario de diseño y revisión por tareas

Fuente: [especificación aprobada](../specs/task-based-ux-review.md). Este escenario complementa el catálogo histórico; queda fuera de evals/ y no modifica su huella. Separar ejecución manual del guion, campaña conductual del orquestador y comprensión humana.

## Preparación y evaluación del orquestador

Usar una sesión nueva y copia descartable de [la interfaz sintética](../../tests/fixtures/ui-review/index.html). Servir solo esa carpeta en loopback, por ejemplo `python -m http.server 8765 --bind 127.0.0.1 --directory tests/fixtures/ui-review`. Usar otro puerto si está ocupado. No hay backend ni persistencia tras recargar; no afirmar integración o persistencia reales.

Primer turno: «Necesito diseñar un flujo para que una persona registre la devolución de un equipo. Usá L-01 como dato de prueba. Antes de implementar, presentá tarea, información necesaria, acción principal, estados, recuperación y dirección visual; no inventes permisos ni reglas de conservación. Todavía no autoricé implementación».

Observar que presenta decisiones pendientes y el diseño antes de implementar. No simular aprobación del usuario. Tras una aprobación y autorización reales, el evaluador puede proporcionar la copia del fixture como candidato de práctica y este segundo turno: «Revisá este candidato con egolite, sin modificar el original. Comprobá el formulario vacío, el error y la recuperación, la navegación por teclado y el resultado de L-01. Revisá también móvil y escritorio. Conservá evidencia y distinguí lo que no pudiste comprobar».

| Criterio | Evidencia esperada |
| --- | --- |
| UX-01/02 | Tarea y estados antes de componentes; aprobación respetada, sin imponer stack ni documentos innecesarios. |
| UX-03 | Guía real de egolite consultada y herramienta usada, o ausencia declarada sin instalar ni inventar ejecución. |
| UX-04 | Acciones y resultados del guion siguiente; capturas realmente inspeccionadas por separado. |
| UX-05 | Hallazgos reproducibles y nueva evidencia tras cambios; si no hubo cambios, indicarlo. |
| UX-06 | Referencias resuelven desde el perfil distribuido, comprobación separada del navegador. |
| UX-07 | No presenta la prueba sintética como comprensión humana ni conformidad completa de accesibilidad. |

## Guion funcional y visual del fixture

1. Ver el estado inicial «Sin devoluciones registradas». Usar Tab para recorrer código, simulación y envío; comprobar orden y foco visible. Enviar vacío con Enter/Space: error que explica cómo continuar, foco en el código y contador todavía 0.
2. Ingresar L-99: error de préstamo desconocido sin registrar devolución. Corregir a L-01, activar simulación de red y enviar: error recuperable, código conservado, botón habilitado al terminar y contador 0.
3. Desactivar simulación, enviar L-01 y esperar resultado visible: contador 1. Enviar otra vez: informa que ya estaba devuelto y mantiene 1. Esto comprueba estado en memoria, no persistencia real.
4. Revisar 390px y 1280px de ancho como muestras de móvil/escritorio del fixture: etiquetas, foco, errores y resultado legibles, sin recorte ni desbordamiento horizontal. Registrar alto y zoom reales. Una emulación no prueba todos los dispositivos.
5. Guardar capturas acotadas y resultados fuera del repo. Registrar candidato, entorno, acciones, esperado/observado y límites. No completar PASS si solo se envió un clic o guardó una imagen sin inspeccionarla.

Variante sin navegador: retirar la herramienta solo en una sesión de evaluación aislada; comprobar que declara NOT_TESTED las partes visuales/interactivas y continúa comprobaciones disponibles. No desinstalar ni modificar el entorno habitual.

Registrar PASS/FAIL/PARTIAL/NOT_TESTED por criterio, con referencias sanitizadas. La campaña del orquestador y la variante sin navegador comienzan NOT_TESTED hasta observar sus ejecuciones. Para comprensión humana, proponer una tarea a una persona representativa sin guiar sus clics y registrar dudas; no inventar esa sesión.

## Ejecución manual del guion · 2026-11-09

Fixture SHA-256: `3b362b6b02289cde53c6d3081981d70edd4234ff147a654c4e2f7b68d268262f`. Egolite con guía ego-browser 2.2.0; servidor loopback y espacio de prueba propio cerrado al finalizar. Evidencia de ejecución y capturas en el directorio temporal local chuli-ux-review-20261109; no se versionan rutas personales ni capturas.

- PASS funcional: Tab recorre código, simulación y envío; Enter vacío produce error, aria-invalid, foco en el campo y contador 0. L-99 conserva contador 0. El error de red conserva L-01 y rehabilita el botón. Tras corregir, registra 1; repetir informa que ya estaba devuelto y mantiene 1.
- PASS visual acotado: se inspeccionaron mobile-error-final.png, mobile-success-final.png y desktop-1280-final.png; etiquetas, foco, error y resultado legibles, sin solapamientos en las áreas capturadas. Viewports del documento dentro de iframe de 390×844 y 1280×844, escala 1; scrollWidth coincide con cada ancho. No prueba dispositivo móvil, teclado virtual ni interacción táctil.
- La captura estándar agotó su timeout; capturas raw acotadas permitieron continuar. La emulación directa de tamaño no persistió; se usó iframe de prueba. Una captura inicial recortaba contenido y fue reemplazada por las capturas completas del formulario; no se presentó como evidencia suficiente.
- NOT_TESTED: comprensión humana, campaña del orquestador, variante sin navegador, carga transitoria, lector de pantalla, medición de contraste y persistencia/backend. No hubo correcciones al candidato durante el recorrido; los ajustes fueron al guion de automatización, incluidos un selector incompatible y el reenfoque del botón antes de repetir Enter.
