# Documentación comprensible con diagramas

## Alcance y decisiones aprobadas

El usuario solicitó ampliar las explicaciones del README y usar diagram-design; aprobó incluir su skill completa, fijada a la revisión auditada, en el paquete de Chuli el 2026-11-09. Se conserva la identidad retro aprobada y el logo existente.

Se distribuye diagram-design 2.6 desde cathrynlavery/diagram-design, commit 6a7f19fcc132b39b126bea90167679b46b14bcc4: 293 archivos originales, licencia MIT, procedencia e integridad comprobables. No se instala el plugin ni se modifica el perfil personal. No se alteran los scripts upstream. PNG requiere Playwright y navegador compatibles, opcionales y no instalados por este cambio; HTML, comprobación y SVG usan Python estándar y navegador disponible según corresponda.

## Diseño de documentación e integración

README: explicar qué es un perfil, dónde trabaja la persona, qué hace el Manager, cómo decide el orquestador, para qué sirven agentes/skills y qué permanece en Git. Conservar instalación y enlaces; sumar diagramas de arquitectura y flujo con explicación textual equivalente.

Diagramas: fuentes HTML estáticas, exportación SVG para GitHub, tipografía de sistema y paleta verde/beige derivada de la marca. No animación, fuentes externas ni instrucciones incrustadas en imágenes. La precisión de relaciones y estados importa más que dibujar todas las piezas. Tipos/alcance se anuncian antes de dibujar; una tabla basta para el catálogo de roles.

cognitive-doc-design deriva a diagram-design cuando mejora la comprensión, no en todos los documentos. La guía de Chuli mantiene aprobación humana, fuentes verificadas, proporcionalidad y egolite como navegador preferido disponible/autorizado. No se prometen slash commands del plugin con la copia de la skill.

## Aceptación

| ID | Resultado observable |
| --- | --- |
| DOC-01 | README distingue Manager, perfil, Codex y proyecto con lenguaje y diagramas coherentes con el código/protocolos. |
| DOC-02 | Agentes, skills, seguimiento y herramientas externas tienen responsabilidades y ejemplos claros. |
| DOC-03 | HTML/SVG accesibles, sin dependencias remotas, mantienen texto equivalente, referencias y fuentes regenerables. |
| DOC-04 | Los 293 archivos y licencia conservan bytes upstream fijados; inventario, paquete y copia instalada contienen sus recursos. |
| DOC-05 | Pruebas comprueban integridad, exportación y rechazo de SVG inseguro; QA visual y límites se informan por separado. |
| DOC-06 | Activación condicional documentada, catálogo actual de 24 skills y atribución completos; sin cambiar release publicada, modelos o permisos. |

## Riesgos y reversión

El contenido importado agrega unos 5,5 MB y puede contener referencias a herramientas opcionales. Se mantiene fijado, sin actualizaciones automáticas ni instalación de dependencias. La copia de una skill no garantiza su descubrimiento o conducta en todos los clientes. Una imagen del diagrama no impone permisos efectivos.

Revertir solo esta unidad: importación y manifiesto, guía de conexión, pruebas, cambios del catálogo/atribución y ampliación del README/diagramas. Conservar logo y README retro ya publicados en su rama. Sin migración de formato; instalación futura y rollback usan el mecanismo existente. Evidencia y tareas: odd/tasks/diagram-documentation.md. El usuario autorizó integrar todo a main y publicarlo en el repositorio el 2026-11-09; no se modifica la release publicada.
