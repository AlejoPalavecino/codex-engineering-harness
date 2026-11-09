# Catálogo de skills

Seleccioná por objetivo y por capacidad real del cliente. El perfil completo tiene 5 skills de Chuli y 18 tomadas de Gentle AI™ v4.4.0. Cargar una skill consiste en leer su procedimiento cuando hace falta. No instala las herramientas que ese procedimiento describe.

Las cinco nativas son `harness-discovery` (alcance), `harness-architecture` (contratos y alternativas), `harness-ux-ui` (flujos e interacción), `harness-specification` (aceptación y tareas), y `harness-tdd-qa` (pruebas y evidencia). Necesitan el contexto del proyecto. Un runner de pruebas o navegador sólo corresponde cuando hay una verificación concreta que lo requiere.

Para interfaces, `harness-ux-ui` prepara tarea, estados y guion; `harness-tdd-qa` comprueba apariencia, interacción y resultado. Ambas incluyen referencias locales instalables. Ego-browser/egolite es la preferencia de revisión web si está disponible y autorizado; se distribuye por separado. Bibliotecas como ui-ux-pro-max o design-taste-frontend son apoyos opcionales revisados, no parte de este paquete: elegir una fuente de dirección visual por tarea, respetando el diseño aprobado. No se instalan automáticamente ni se impone un stack.

| Skill importada | Objetivo | Runtime o destino | Requisitos |
| --- | --- | --- | --- |
| `branch-pr` | Preparar PR según la política del destino | Codex | Git; para operaciones remotas, GitHub autorizado |
| `chained-pr` | Dividir una entrega en unidades revisables | Codex | Git; publicar ramas o PR requiere autorización |
| `cognitive-doc-design` | Diseñar documentación legible | Codex | Documento y audiencia |
| `comment-writer` | Redactar comentarios colaborativos | Codex | Contexto del intercambio |
| `gentle-ai-bench` | Autoría y verificación de journeys de Gentle AI | Checkout de Gentle AI | bench/, compilador Go y CI del upstream |
| `gentle-ai-collab-perfect` | Colaborar externamente con Gentle AI | Checkout de Gentle AI | Política de contribución de Gentleman-Programming/gentle-ai |
| `go-testing` | Probar Go y TUI según comportamiento | Codex | Go; teatest sólo si el proyecto lo usa |
| `hermes-ephemeral-delegation` | Delegar tareas efímeras en Hermes | Hermes | Herramienta Hermes delegate_task |
| `issue-creation` | Preparar o clasificar issues del destino | Codex | Política del repositorio; acceso GitHub para mutar |
| `issue-root-resolution` | Investigar una familia de incidencias | Codex | Issues, historial y capacidad de reproducir |
| `judgment-day` | Doble revisión ciega solicitada | Codex | Dos sesiones independientes con acceso de lectura |
| `rdd-advisory-transport` | Mantener el transporte de revisores RDD upstream | Checkout de Gentle AI | Contrato Go y journeys de ese proyecto |
| `rdd-defect-workflow` | Diagnosticar defectos del flujo de recibos RDD | Proyecto que implemente RDD de Gentle AI | Productor RDD y política explícita del destino |
| `skill-creator` | Crear una skill reutilizable | Codex | Convenciones del destino y guía incluida |
| `skill-improver` | Revisar o mejorar skills existentes | Codex | Skills existentes y guía incluida |
| `skill-registry` | Resolver nombres, triggers y rutas | Codex | Metadatos locales; CLI Gentle AI y Engram opcionales |
| `systemic-issue-triage` | Agrupar incidencias por causa comprobable | Codex | Issues o reportes; GitHub para acciones remotas |
| `work-unit-commits` | Agrupar código, pruebas y docs por comportamiento | Codex | Git y política local de entrega |

## Origen y resolución

El catálogo conserva las 14 skills seleccionables del instalador upstream, la skill específica de Hermes y tres de su repositorio de contribución: `gentle-ai-collab-perfect`, `issue-root-resolution` y `rdd-advisory-transport`. Para las ocho duplicadas en origen se eligió la edición de distribución. Cada directorio conserva sus recursos. Las siete referencias comunes desplegables se encuentran en `skills/_shared/`; el fragmento que Gentle AI usa para construir su propio orquestador se conserva en `skills/_shared/upstream/` para consulta.

En la instalación, resolvé los archivos por la ruta efectiva de la skill. Los avisos y el pin verificable viajan en `skills/_shared/NOTICES.md`. Un procedimiento instalado sólo se activa si coincide el objetivo, el runtime y las dependencias. Si el host ya proporciona otra skill con igual nombre, verificá su origen antes de invocarla.
