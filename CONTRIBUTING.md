# Cómo contribuir

Gracias por ayudar a construir un harness útil, no uno innecesariamente complejo.

## Antes de escribir código

Revisá [EWP](docs/protocols/EWP.md) y [la arquitectura](docs/AGENTS.md). Buscá un issue existente y describí el comportamiento que querés mejorar. Para cambios grandes, proponé primero su alcance.

## Pull requests

Para proponer y aprobar cambios, seguí el [flujo de issues y PRs](docs/ISSUE_WORKFLOW.md), su formulario y catálogo de etiquetas. Los PRs deben vincular un issue aprobado, tener exactamente un tipo y mostrar evidencia; dividí diffs mayores de 400 líneas en unidades revisables contra su padre inmediato. No confundas aprobación del alcance con autorización de merge.

1. Crear una rama descriptiva.
2. Cambiar el mínimo conjunto de archivos necesario.
3. Documentar decisiones y efectos sobre el instalador o la compatibilidad.
4. Ejecutar `python tests/validate_repository.py` (Python 3.11+) y, en Windows, los scripts de verificación.
5. Incluir evidencia, escenarios probados, riesgos y pasos de reversión.
6. No introducir secretos, capturas con datos reales ni código de terceros sin licencia compatible.

## Criterios de aceptación

Las adaptaciones de Gentle AI deben indicar el commit de origen, la licencia aplicable y las dependencias del runtime en [el inventario de terceros](THIRD_PARTY_NOTICES.md). Incluí archivos auxiliares en el paquete y verificá que Codex descubra la skill desde un perfil temporal.

Una nueva skill debe resolver una necesidad distinta. Un nuevo agente necesita justificar el costo de delegación. Una nueva dependencia o MCP debe ser opcional salvo necesidad técnica demostrada. La compatibilidad de modelos y mecanismos de permiso debe probarse, no suponerse.

## Idioma

La documentación principal está en español. Se aceptan contribuciones en español o inglés; los cambios de comportamiento deben ser comprensibles para ambos públicos.

## Código de conducta

Trato respetuoso, discusión técnica basada en evidencia y críticas centradas en el trabajo, no en las personas.
