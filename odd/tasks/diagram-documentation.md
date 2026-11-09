# Documentación y diagram-design

- **Spec:** docs/specs/diagram-documentation.md
- **Design:** docs/specs/diagram-documentation.md#diseño-de-documentación-e-integración
- **Approval:** approved

Integración a main y publicación en el repositorio autorizadas el 2026-11-09.

### TASK-001 — Conservar alcance, diseño y procedencia
- **State:** DONE
- **Depends on:** none
- **Owner:** orchestrator
- **Acceptance:** DOC-01 a DOC-06 y origen fijado reflejan decisiones aprobadas, alcance y reversión.
- **Evidence:** Aprobación explícita del paquete completo y uso condicional; origen 6a7f19f, licencia MIT, diseño retro y criterios DOC-01 a DOC-06 registrados y contrastados con la solicitud.
- **Blocker:** none

### TASK-002 — Distribuir diagram-design con integridad
- **State:** DONE
- **Depends on:** TASK-001
- **Owner:** orchestrator
- **Acceptance:** Recursos y licencia intactos, integración condicional y pruebas de distribución y exportación pertinentes.
- **Evidence:** Commit 8da1cab. Importación completa desde blobs Git fijados: 293 originales + LICENSE/UPSTREAM.json; scripts intactos. RED por recursos ausentes observado, GREEN de integridad/exportación/rechazo de script y suite de 45 pruebas (1 skip). Ruta condicional documentada en cognitive-doc-design y catálogo.
- **Blocker:** none

### TASK-003 — Explicar Chuli y comprobar diagramas
- **State:** DONE
- **Depends on:** TASK-002
- **Owner:** orchestrator
- **Acceptance:** README claro, fuentes HTML y SVG exportados con texto equivalente; checks, distribución y QA disponibles documentados sin inventar pruebas.
- **Evidence:** README renderizado con GitHub Markdown: 7 tablas, enlaces locales y anclas válidos. Ambos HTML pasan `self_check.py --offline` y exportan SVG con `--system-fonts`. Egolite abrió ambos: title/desc, etiquetas, tipografía de sistema, cero textos fuera del viewBox y cero recursos externos detectados. Suite Python: 45 pruebas OK, 1 skip de symlink existente. Validador estructural OK en copia temporal usando bytes canónicos de la fixture histórica (checkout CRLF falla su hash preexistente). Paquete y blobs con autocrlf conservan los 293 SHA-256. Instalación temporal 23→24 skills, verify y rollback OK; perfil previo restaurado byte por byte. QA por captura pendiente: Page.captureScreenshot agotó tiempo; inspección DOM no acredita acabado visual ni usabilidad humana. Sin cambios al perfil personal ni publicación.
- **Blocker:** none
