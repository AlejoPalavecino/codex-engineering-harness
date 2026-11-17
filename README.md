<div align="center">

<img src="assets/brand/chuli-logo.png" width="740" alt="Chuli: una PC retro con monitor de fósforo verde y el nombre del proyecto en letras pixeladas">

# Chuli

**Un harness de ingeniería para Codex, Claude Code y AGY CLI.**

De la idea al cambio comprobado, con decisiones claras, modelos configurables por agente y un compañero que te explica mientras trabajan.

[Descargar](https://github.com/AlejoPalavecino/codex-engineering-harness/releases/latest) · [Qué es](#que-es) · [Instalación](#instalacion) · [Cómo trabaja](#como-trabaja) · [Componentes](#componentes) · [Documentación](#documentacion)

**6 especialistas · decisiones claras · documentación para compartir**

</div>

---

<a name="que-es"></a>

## 00 · Qué estás instalando

Chuli es un **perfil de ingeniería para Codex, Claude Code y AGY CLI**. Reúne instrucciones para coordinar el trabajo, especialistas para tareas concretas y procedimientos que el agente consulta cuando hacen falta. Está pensado para construir y mantener un proyecto con decisiones, cambios y comprobaciones que otra persona pueda entender.

![El Manager instala el perfil del cliente elegido; Codex, Claude Code o AGY CLI lo usan para trabajar sobre tu proyecto y conservar evidencia en Git.](assets/diagrams/chuli-architecture.svg)

Cada cliente usa su propio destino y formato de configuración. El Manager administra el perfil; el trabajo de ingeniería ocurre en la sesión del cliente elegido.

Abrís el Manager para instalar o actualizar el perfil del cliente que elegís. Después abrís ese cliente en tu proyecto. El orquestador corre dentro de esa sesión y te ayuda a decidir qué hace falta investigar, acordar, implementar o verificar.

| Parte | Dónde está | Qué hace |
| --- | --- | --- |
| **Chuli Manager** | El ejecutable `chuli` / `chuli.exe` | Descarga, instala, verifica, actualiza y revierte el perfil. |
| **Perfil de Chuli** | `~/.codex` (`CODEX_HOME`), `~/.claude` (`CLAUDE_CONFIG_DIR`) o `~/.gemini/config` (`AGY_HOME`) | Contiene instrucciones, configuración de roles, skills y recursos según el cliente. |
| **Tu sesión de ingeniería** | Codex, Claude Code o AGY CLI | Es donde conversás, el agente usa herramientas y se modifica el proyecto autorizado. |
| **Tu proyecto** | El repositorio donde estás trabajando | Conserva código, pruebas y documentación; en trabajo sustancial, decisiones y seguimiento para retomar. |

[Fuente del diagrama](assets/diagrams/chuli-architecture.html) · [Operaciones del Manager](docs/TUI.md) · [Responsabilidad de cada componente](docs/protocols/TKA.md)

<a name="instalacion"></a>

## 01 · Encendé el equipo

**Necesitás Python 3.11 o posterior y PowerShell para aplicar el perfil.**

### Con un comando (Go 1.26+, Windows / Linux / macOS)

```sh
# Stable channel: the latest release (v2.0.2)
go install github.com/AlejoPalavecino/codex-engineering-harness/v2/cmd/chuli@v2.0.2

# Or always install the latest tagged release
go install github.com/AlejoPalavecino/codex-engineering-harness/v2/cmd/chuli@latest
```

Después ejecutá `chuli`: la consola CRT te recibe con un menú directo de 6 acciones (**Instalar**, **Actualizar**, **Gestionar versiones**, **Perfiles de AI**, **MCPs** y **Skills**) y te avisa si salió una nueva release. En **Instalar**, elegís una o varias plataformas (**OpenAI Codex**, **Claude Code** o **AGY CLI**), seleccionás un preset o personalizás tus modelos, confirmás un resumen simple y Chuli aplica todo con respaldo automático SHA-256 sin llenarte la pantalla de texto técnico.

### Sin Go (instalador Windows o descarga portable)

1. **Descargá el Manager** desde la [última release estable](https://github.com/AlejoPalavecino/codex-engineering-harness/releases/latest) (o usá `install-chuli-windows.ps1` en Windows).
2. **Extraé y abrí el ejecutable:** `chuli.exe` en Windows o `chuli` en Linux/macOS.
3. **Elegí instalar desde la última release.** Revisá el destino y los cambios, y confirmá. El Manager descarga el perfil y verifica su SHA-256.
4. **Abrí una sesión nueva del cliente que elegiste** en tu proyecto y empezá con una tarea concreta.

| Tu equipo | Archivo del Manager |
| --- | --- |
| Windows x64 | `chuli_v<versión>_windows_amd64.zip` |
| Linux x64 | `chuli_v<versión>_linux_amd64.tar.gz` |
| macOS ARM64 | `chuli_v<versión>_darwin_arm64.tar.gz` |

La [guía de instalación](docs/INSTALLATION.md) cubre requisitos, personalizaciones, actualización, verificación y rollback. Desde v1.1.1, el perfil, los tres Managers y el instalador Windows se acompañan de `SHA256SUMS.txt` para comprobar las descargas.

> **¿Ya tenés un perfil personalizado?** Revisá la vista previa antes de aplicar. Chuli tiene reglas para preservar archivos ajenos y recursos personalizados; reemplazarlos requiere una decisión explícita. Consultá la guía antes de actualizar.

<a name="como-trabaja"></a>

## 02 · Del pedido al resultado

El orquestador es el punto de contacto: conversa con vos, revisa el contexto y elige cuánto proceso necesita el pedido. Habla en español argentino, explica primero en palabras cotidianas y suma el concepto técnico cuando te sirve para aprender.

![Una consulta se investiga; un cambio entendido y autorizado se implementa y prueba; una funcionalidad necesita acuerdos antes de construir. Todas las rutas entregan evidencia y límites.](assets/diagrams/chuli-workflow.svg)

| Vos necesitás… | Chuli trabaja así |
| --- | --- |
| Entender código o evaluar una idea | Investiga y explica con referencias. Una consulta mantiene el trabajo en lectura. |
| Corregir algo puntual | Inspecciona, hace el cambio autorizado y ejecuta una comprobación proporcionada. |
| Construir una funcionalidad | Aclara decisiones pendientes, propone el diseño y conserva tareas y evidencia cuando hace falta continuidad. |
| Diseñar una interfaz | Parte de la tarea de la persona, define estados y prepara un guion para comprobar apariencia, interacción y resultado. |

Pregunta por decisiones que cambian el resultado y resuelve detalles técnicos reversibles dentro del alcance autorizado. Esa autorización acompaña la implementación, las correcciones y la verificación. Si cambia el alcance, vuelve a consultar lo afectado.

El riesgo se evalúa aparte del tamaño: un cambio pequeño de permisos puede necesitar controles adicionales. Aprobar un diseño y autorizar su implementación son decisiones distintas. Publicar, hacer push, abrir una PR, integrar o desplegar requieren la autorización aplicable.

Usa TDD cuando hay una prueba adecuada y procura unidades revisables de unas 400 líneas, con código, pruebas y documentación del comportamiento juntos. Cada unidad atraviesa el recorrido necesario para entregar un resultado observable y deja contexto suficiente para retomarlo: qué cambió, qué se comprobó y cuál es el siguiente paso.

![Una unidad vertical conserva su criterio y su próximo paso; las pruebas se vinculan con los archivos del candidato y el cierre compara fricción y calidad.](assets/diagrams/chuli-odd.svg)

En un tablero nuevo, el orquestador captura la huella de los archivos de la tarea antes de probarlos. Al cerrarla registra el comando, el resultado y esa huella. Si editás uno de los archivos, podés detectar qué prueba hay que repetir con `taskboard.py check TABLERO --current`. [Ejemplo de captura y cierre](docs/protocols/TTP.md)

Para evaluar el flujo, podés comparar dos observaciones del mismo recorrido. Chuli muestra preguntas repetidas, unidades rehechas y tiempo hasta la primera prueba útil junto con la calidad obtenida. El piloto incluido usa datos sintéticos y se puede ejecutar sin una cuenta. [Cómo reproducir la comparación](docs/evaluations/workflow-friction.md)

Un cierre explica qué cambió, cómo se comprobó y qué quedó pendiente; escribir código o guardar una captura no alcanza para afirmar que todo funciona. [Fuente de la figura ODD](assets/diagrams/chuli-odd.html).

[Fuente del diagrama](assets/diagrams/chuli-workflow.html) · [Rutas de trabajo](docs/protocols/EWP.md) · [Unidades y evidencia](docs/protocols/ODD.md)

### Probalo con un pedido real

> «Explicame cómo funciona este módulo y dónde se comprueban los permisos. Por ahora, solo investigá».

> «Quiero agregar reservas de equipos. Primero definamos qué puede hacer cada persona y cómo se cancela una reserva».

> «Revisemos este formulario: qué necesita completar la persona, cómo se recupera de un error y cómo vamos a comprobarlo con egolite».

<a name="componentes"></a>

## 03 · Debajo del gabinete

### El orquestador coordina; los especialistas profundizan

No todos los pedidos necesitan un equipo de agentes. El orquestador trabaja en línea por defecto y delega cuando una exploración, una especialidad o una revisión independiente aporta valor. Vos conservás las decisiones de producto, alcance, diseño y entrega.

Antes de delegar, define la pregunta concreta que necesita resolver, qué trabajo independiente puede avanzar mientras tanto y qué prueba o evidencia debe devolver el especialista. La respuesta se contrasta con el candidato antes de dar la tarea por terminada.

| Especialista | Cuándo ayuda | Qué devuelve |
| --- | --- | --- |
| Exploración · `chuli_explorer` | Hace falta ubicar un contrato o entender código existente. | Un mapa con rutas, consumidores y referencias. |
| Implementación · `chuli_worker` | Hay una unidad de comportamiento definida y autorizada. | Cambios, comprobaciones, límites y forma de revertir. |
| Arquitectura · `solution_architect` | Hay alternativas técnicas que conviene comparar. | Opciones, consecuencias, riesgos y decisiones pendientes. |
| Producto · `product_designer` | Hay que definir una experiencia o rediseñar una interfaz. | Flujos, estados, dirección visual y criterios verificables. |
| Pruebas · `test_engineer` | Hace falta convertir aceptación y riesgos en pruebas. | Casos ejecutables, resultados y cobertura pendiente. |
| Revisión · `quality_reviewer` | El riesgo justifica verificar un candidato de forma independiente. | Hallazgos reproducibles y verificaciones realizadas/omitidas. |

Los modelos declarados y los contratos de cada rol están en [Especialistas](docs/AGENTS.md). Su disponibilidad depende del cliente y de la cuenta; una configuración no demuestra por sí sola qué modelo se ejecutó.

### Modelos según el cliente

Al configurar los modelos, el Manager muestra la fuente del catálogo que pudo obtener. La selección conserva los roles ya configurados hasta que confirmás los cambios.

| Cliente | Catálogo y selección |
| --- | --- |
| **Codex** | Lee la caché de modelos del cliente en su perfil local. |
| **Claude Code** | Lee una caché compatible y sin ambigüedad de cuenta; incluye `inherit` para heredar el modelo. |
| **AGY CLI** | Consulta `agy --output-format json models` para el orquestador; los especialistas eligen un tier. |

Si no obtiene un catálogo utilizable, ofrece el catálogo incluido en Chuli. Esa lista permite configurar el perfil; la disponibilidad efectiva se comprueba en el cliente y la cuenta que lo ejecutan. [Configuración de modelos](docs/TUI.md).

### Las skills aportan el procedimiento

Un rol define **de qué se ocupa un agente**. Una skill explica **cómo encarar una tarea**. Las herramientas son las que permiten leer archivos, ejecutar pruebas o usar un navegador, con los permisos efectivos del entorno.

| Skill propia | Pregunta que ayuda a resolver |
| --- | --- |
| [`harness-discovery`](skills/harness-discovery/SKILL.md) | ¿Qué necesita la persona y qué falta decidir? |
| [`harness-architecture`](skills/harness-architecture/SKILL.md) | ¿Qué solución técnica encaja con el proyecto y sus restricciones? |
| [`harness-ux-ui`](skills/harness-ux-ui/SKILL.md) | ¿Cómo completa su tarea una persona y cómo se recupera de errores? |
| [`harness-specification`](skills/harness-specification/SKILL.md) | ¿Qué significa terminado y cómo se divide el trabajo? |
| [`harness-tdd-qa`](skills/harness-tdd-qa/SKILL.md) | ¿Qué evidencia necesitamos para comprobar el cambio? |

El catálogo de este checkout suma **25 skills**: las cinco propias, 18 de Gentle AI v4.4.0, Diagram Design 2.6 y Frontend Design. Las importadas incluyen diseño frontend, documentación, revisiones, contribución y preparación de unidades de trabajo; algunas requieren un repositorio o runtime específico. [Catálogo y condiciones](docs/protocols/SKILLS.md) · [Procedencia y licencias](THIRD_PARTY_NOTICES.md).

> Diagram Design se incluye desde v1.1.0 (la v1.0.1 conserva su catálogo histórico de 23 skills); Frontend Design se incorporó sobre la base de `v2.0.1` sin modificar el flujo de aprobación de [`harness-ux-ui`](skills/harness-ux-ui/SKILL.md).

### Documentación que otra persona pueda seguir

[`cognitive-doc-design`](skills/cognitive-doc-design/SKILL.md) organiza la explicación y deriva a [`diagram-design`](skills/diagram-design/SKILL.md) cuando una figura aclara relaciones, decisiones o recorridos. El objetivo es poder compartir el proyecto y entender qué existe, cómo se conecta y qué falta resolver. Una lista o tabla sigue siendo suficiente cuando comunica lo mismo.

Los diagramas conservan fuente HTML, exportación SVG/PNG y explicación textual. En GitHub se muestran los SVG estáticos; el HTML queda disponible para revisar o editar. La skill incluye sus recursos y licencia, sin instalar el plugin o dependencias opcionales. [Guía de uso en Chuli](skills/cognitive-doc-design/references/diagrams.md).

### El seguimiento permite retomar

En un trabajo sustancial, Chuli mantiene un tablero por funcionalidad en `odd/tasks/<funcionalidad>.md`, conectado con la especificación y el diseño aprobados. Registra tareas, dependencias, estado, evidencia y bloqueos. Una tarea termina cuando hay comprobaciones que sostienen sus criterios.

Las decisiones y el tablero viven en Git, junto al proyecto. Al retomar se contrastan con el diff y las pruebas disponibles. En cambios pequeños alcanza el contexto y una comprobación focalizada; no se generan documentos por obligación. [Seguimiento TTP](docs/protocols/TTP.md).

### Las integraciones amplían lo que puede comprobar

| Integración opcional | Para qué se usa |
| --- | --- |
| **Egolite / ego-browser** | Inspeccionar interfaces y diagramas en un navegador y comprobar interacción real. |
| **Engram** | Recuperar decisiones duraderas del proyecto; complementa los archivos versionados. |
| **Context7** | Consultar documentación de librerías cuando la versión importa. |
| **GitHub** | Leer o gestionar issues, PR y entregas dentro de la autorización recibida. |

**El código actual del Manager incluye un asistente de Integraciones:** elegís una opción, revisás los requisitos y confirmás la conexión. Engram y Egolite tienen descarga guiada; las claves se configuran en tu entorno y se conservan conexiones existentes. Disponible desde el Manager v1.1.0: actualizar sólo el perfil no reemplaza tu ejecutable. [Cómo usarlo y qué comprueba](docs/TUI.md#asistente-de-integraciones) · [Herramientas y conocimiento](docs/protocols/TKA.md).

<a name="documentacion"></a>

## 04 · El manual

| Quiero… | Abrí… |
| --- | --- |
| Instalar, actualizar o volver atrás | [Instalación](docs/INSTALLATION.md) |
| Operar el Manager y su CLI | [Manager](docs/TUI.md) |
| Conocer los roles y las skills | [Especialistas](docs/AGENTS.md) · [Catálogo](docs/protocols/SKILLS.md) |
| Entender las decisiones del producto | [Trayectoria y roadmap](docs/ROADMAP.md) |
| Ver qué cambió | [Release 2.0.2](docs/releases/v2.0.2.md) · [Changelog](CHANGELOG.md) |
| Revisar pruebas y límites de evidencia | [Verificación](docs/TESTING.md) · [Evaluación conductual](docs/EVALUATION.md) |
| Contribuir o informar un problema | [Contribuciones](CONTRIBUTING.md) · [Seguridad](SECURITY.md) |

---

<p align="center"><strong>Estética de otra época. Decisiones y evidencia en tu proyecto.</strong><br><a href="docs/BRAND.md">Identidad de Chuli</a> · <a href="LICENSE">Licencia</a></p>
