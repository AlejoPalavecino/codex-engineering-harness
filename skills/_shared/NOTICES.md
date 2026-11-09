# Material de terceros

Además del catálogo de Gentle AI descrito abajo, se distribuye [Diagram Design 2.6](../diagram-design/SKILL.md), © 2025 Cathryn Lavery, bajo [MIT](../diagram-design/LICENSE), fijado en `6a7f19fcc132b39b126bea90167679b46b14bcc4`. Su manifiesto UPSTREAM.json conserva la procedencia y hashes. La integración de documentación de Chuli está separada de los archivos upstream, que permanecen intactos.

Chuli integra procedimientos de **Gentle AI™ v4.4.0**, de Gentleman Programming, revisión `55240afb100ad6d2ef5fbe16f7be33a7693f77ad`. Las marcas Gentle AI y Engram pertenecen a Alan Buscaglia. Los nombres identifican la procedencia y los runtimes compatibles. [Código fuente](https://github.com/Gentleman-Programming/gentle-ai/tree/55240afb100ad6d2ef5fbe16f7be33a7693f77ad) · [Política de marcas](https://github.com/Gentleman-Programming/gentle-ai/blob/55240afb100ad6d2ef5fbe16f7be33a7693f77ad/TRADEMARKS.md).

## Alcance y licencias

Las cinco skills `harness-*` se rigen por la licencia MIT de Chuli. El material importado conserva autores y licencias declaradas en cada encabezado. Sus archivos de referencia se distribuyen bajo la licencia de su skill. Las referencias comunes y el fuente ODD sin licencia específica usan la licencia MIT del proyecto upstream, © 2025 Gentleman Programming. Se incluyen los textos completos en `licenses/`, además del MIT de Chuli.

| Skill upstream | Licencia | Origen fijado |
| --- | --- | --- |
| `branch-pr` | Apache-2.0 | [fuente](https://github.com/Gentleman-Programming/gentle-ai/tree/55240afb100ad6d2ef5fbe16f7be33a7693f77ad/internal/assets/skills/branch-pr) |
| `chained-pr` | Apache-2.0 | [fuente](https://github.com/Gentleman-Programming/gentle-ai/tree/55240afb100ad6d2ef5fbe16f7be33a7693f77ad/internal/assets/skills/chained-pr) |
| `cognitive-doc-design` | Apache-2.0 | [fuente](https://github.com/Gentleman-Programming/gentle-ai/tree/55240afb100ad6d2ef5fbe16f7be33a7693f77ad/internal/assets/skills/cognitive-doc-design) |
| `comment-writer` | Apache-2.0 | [fuente](https://github.com/Gentleman-Programming/gentle-ai/tree/55240afb100ad6d2ef5fbe16f7be33a7693f77ad/internal/assets/skills/comment-writer) |
| `gentle-ai-bench` | Apache-2.0 | [fuente](https://github.com/Gentleman-Programming/gentle-ai/tree/55240afb100ad6d2ef5fbe16f7be33a7693f77ad/internal/assets/skills/gentle-ai-bench) |
| `gentle-ai-collab-perfect` | Apache-2.0 | [fuente](https://github.com/Gentleman-Programming/gentle-ai/tree/55240afb100ad6d2ef5fbe16f7be33a7693f77ad/skills/gentle-ai-collab-perfect) |
| `go-testing` | Apache-2.0 | [fuente](https://github.com/Gentleman-Programming/gentle-ai/tree/55240afb100ad6d2ef5fbe16f7be33a7693f77ad/internal/assets/skills/go-testing) |
| `hermes-ephemeral-delegation` | Apache-2.0 | [fuente](https://github.com/Gentleman-Programming/gentle-ai/tree/55240afb100ad6d2ef5fbe16f7be33a7693f77ad/internal/assets/skills/hermes-ephemeral-delegation) |
| `issue-creation` | Apache-2.0 | [fuente](https://github.com/Gentleman-Programming/gentle-ai/tree/55240afb100ad6d2ef5fbe16f7be33a7693f77ad/internal/assets/skills/issue-creation) |
| `issue-root-resolution` | Apache-2.0 | [fuente](https://github.com/Gentleman-Programming/gentle-ai/tree/55240afb100ad6d2ef5fbe16f7be33a7693f77ad/skills/issue-root-resolution) |
| `judgment-day` | Apache-2.0 | [fuente](https://github.com/Gentleman-Programming/gentle-ai/tree/55240afb100ad6d2ef5fbe16f7be33a7693f77ad/internal/assets/skills/judgment-day) |
| `rdd-advisory-transport` | Apache-2.0 | [fuente](https://github.com/Gentleman-Programming/gentle-ai/tree/55240afb100ad6d2ef5fbe16f7be33a7693f77ad/skills/rdd-advisory-transport) |
| `rdd-defect-workflow` | Apache-2.0 | [fuente](https://github.com/Gentleman-Programming/gentle-ai/tree/55240afb100ad6d2ef5fbe16f7be33a7693f77ad/internal/assets/skills/rdd-defect-workflow) |
| `skill-creator` | Apache-2.0 | [fuente](https://github.com/Gentleman-Programming/gentle-ai/tree/55240afb100ad6d2ef5fbe16f7be33a7693f77ad/internal/assets/skills/skill-creator) |
| `skill-improver` | Apache-2.0 | [fuente](https://github.com/Gentleman-Programming/gentle-ai/tree/55240afb100ad6d2ef5fbe16f7be33a7693f77ad/internal/assets/skills/skill-improver) |
| `skill-registry` | MIT | [fuente](https://github.com/Gentleman-Programming/gentle-ai/tree/55240afb100ad6d2ef5fbe16f7be33a7693f77ad/internal/assets/skills/skill-registry) |
| `systemic-issue-triage` | Apache-2.0 | [fuente](https://github.com/Gentleman-Programming/gentle-ai/tree/55240afb100ad6d2ef5fbe16f7be33a7693f77ad/internal/assets/skills/systemic-issue-triage) |
| `work-unit-commits` | Apache-2.0 | [fuente](https://github.com/Gentleman-Programming/gentle-ai/tree/55240afb100ad6d2ef5fbe16f7be33a7693f77ad/internal/assets/skills/work-unit-commits) |

## Cambios de Chuli

Se agregaron metadatos de runtime, dependencias y condiciones de activación. El registro admite resolución local en memoria y el espejo completo en Engram requiere adhesión explícita por proyecto. La división de PR y el triage siguen las políticas del destino. Se acotaron los disparadores de Hermes, bench y transporte RDD y se mapearon los jueces a sesiones independientes de Codex. Los recursos propios de cada skill conservan su contenido original. Las referencias comunes llevan notas de alcance y la referencia del contrato de revisión general está ligada a Codex.

Las variantes de `branch-pr`, `chained-pr` y `work-unit-commits` de la raíz del upstream difieren de los assets instalables. Se usa la edición distribuida como fuente única. Otros cinco duplicados de la raíz coinciden con los assets. Las tres skills exclusivas de la raíz se distribuyen con el origen que figura en la tabla. Las once claves heredadas `sdd-*` del catálogo Go apuntan al flujo retirado y carecen de assets instalables en esta versión.

[Selección actual de presets](https://github.com/Gentleman-Programming/gentle-ai/blob/55240afb100ad6d2ef5fbe16f7be33a7693f77ad/internal/components/skills/presets.go) · [Retiro de SDD](https://github.com/Gentleman-Programming/gentle-ai/blob/55240afb100ad6d2ef5fbe16f7be33a7693f77ad/internal/assets/assets_test.go) · [Material compartido y exclusión de la fuente de construcción ODD](https://github.com/Gentleman-Programming/gentle-ai/blob/55240afb100ad6d2ef5fbe16f7be33a7693f77ad/internal/components/skills/support.go).
