# Diagramas que ayudan a compartir un proyecto

Primero definir audiencia, pregunta que debe responder la figura y fuentes verificadas. Usar tabla para atributos/listas y texto para una relación obvia. Mermaid puede ser suficiente para esquemas simples que cambian seguido; usar [diagram-design](../../diagram-design/SKILL.md) para diagramas editoriales que se van a compartir.

1. Leer la skill completa y solo las referencias del tipo/patrón necesario. Anunciar tipo, tamaño, contenido y simplificaciones; dejar oportunidad de corregir decisiones sustanciales antes de dibujar.
2. Resolver la identidad desde decisiones aprobadas y fuentes del proyecto; respetar el onboarding y perfiles de la skill. No sobrescribir la guía compartida ni escribir perfiles personales sin autorización. Declarar fuentes/fuentes tipográficas sustitutas.
3. Generar HTML autocontenido como fuente. Para README de GitHub, exportar SVG estático o PNG cuando sea el formato pedido; el HTML no se incrusta directamente. Conservar fuente y exportación juntas, con texto equivalente, título y descripción accesibles. No agregar animación o servicios remotos por defecto.
4. Comprobar el archivo con `python <skill>/scripts/self_check.py <diagrama.html> --offline` si se requiere salida sin red. Exportar SVG mediante `python <skill>/scripts/export_svg.py <diagrama.html> <diagrama.svg> --system-fonts` para tipografías de sistema. Las dependencias adicionales de PNG son opcionales; no instalarlas por inferencia.
5. Inspeccionar el resultado real con ego-browser/egolite disponible y autorizado, siguiendo su guía instalada. Comprobar legibilidad, conectores, recortes, anchos pertinentes y coherencia con el texto. Si no se puede inspeccionar, declarar el límite; un checker estructural no acredita calidad visual ni precisión del dominio.
6. Vincular cada figura a su explicación y a la fuente. Actualizarla cuando cambie el comportamiento representado; no inventar despliegues, servicios, aprobaciones o aislamiento para completar un dibujo.

La copia incluida conserva la skill y sus recursos, no los comandos de marketplace del plugin ni verificadores exclusivos del repositorio upstream. `self_check.py` y exportación SVG usan Python estándar; la exportación PNG upstream requiere Playwright y Chromium compatibles. Esta guía no configura herramientas, no actualiza versiones ni instala el plugin.
