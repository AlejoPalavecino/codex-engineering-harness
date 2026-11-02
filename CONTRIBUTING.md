# Cómo contribuir

Gracias por ayudar a construir un harness útil, no uno innecesariamente complejo.

## Antes de escribir código

Revisá [EWP](docs/protocols/EWP.md) y [la arquitectura](docs/AGENTS.md). Buscá un issue existente y describí el comportamiento que querés mejorar. Para cambios grandes, proponé primero su alcance.

## Pull requests

1. Crear una rama descriptiva.
2. Cambiar el mínimo conjunto de archivos necesario.
3. Documentar decisiones y efectos sobre el instalador o la compatibilidad.
4. Ejecutar `python tests/validate_repository.py` (Python 3.11+) y, en Windows, los scripts de verificación.
5. Incluir evidencia, escenarios probados, riesgos y pasos de reversión.
6. No introducir secretos, capturas con datos reales ni código de terceros sin licencia compatible.

## Criterios de aceptación

Una nueva skill debe resolver una necesidad distinta. Un nuevo agente necesita justificar el costo de delegación. Una nueva dependencia o MCP debe ser opcional salvo necesidad técnica demostrada. La compatibilidad de modelos y mecanismos de permiso debe probarse, no suponerse.

## Idioma

La documentación principal está en español. Se aceptan contribuciones en español o inglés; los cambios de comportamiento deben ser comprensibles para ambos públicos.

## Código de conducta

Trato respetuoso, discusión técnica basada en evidencia y críticas centradas en el trabajo, no en las personas.
