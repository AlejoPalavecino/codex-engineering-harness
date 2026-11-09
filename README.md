# Chuli

Chuli instala un perfil personal de ingeniería en el Codex que ya usás. Abrís el Manager para instalar o actualizar el perfil. Después trabajás desde Codex CLI o desde el modo Codex local de la app de escritorio de ChatGPT.

## Instalación

1. Descargá el Manager para tu plataforma desde [la release estable](https://github.com/AlejoPalavecino/codex-engineering-harness/releases/latest). Windows x64 usa el archivo que termina en `windows_amd64.zip`.
2. Extraé y abrí `chuli.exe` en Windows o `chuli` en Linux/macOS. Para aplicar el perfil, el equipo necesita PowerShell y Python 3.11 o posterior.
3. Elegí la instalación desde la última release. Leé el destino y los cambios, y confirmá. El Manager obtiene el ZIP del perfil y verifica su SHA-256.
4. Abrí una sesión nueva de Codex en tu proyecto. Probá «Usá cognitive-doc-design para revisar este README» o pedile al orquestador que derive una tarea acotada al agente de implementación.

La [guía de instalación](docs/INSTALLATION.md) explica el perfil existente, la verificación y el rollback. Para actualizar una instalación anterior conviene usar el ejecutable 1.0, que incluye el nuevo formato de distribución y la recuperación sin checkout.

## Qué instala

- Un orquestador global en `CODEX_HOME/AGENTS.md` y configuración de modelos y roles en `config.toml`. El destino predeterminado suele ser `~/.codex`.
- Seis especialistas para exploración, implementación, arquitectura, diseño, pruebas y revisión. Los modelos y esfuerzos están en el [catálogo de roles](docs/AGENTS.md).
- Cinco skills propias y 18 de Gentle AI v4.4.0, adaptadas o condicionadas a su entorno. Incluye `cognitive-doc-design`, `judgment-day`, `work-unit-commits`, `branch-pr` y `chained-pr`. Cada entrada conserva sus recursos. Consultá las [condiciones de uso por skill](docs/protocols/SKILLS.md) y la [procedencia y licencias](THIRD_PARTY_NOTICES.md).
- Guías de trabajo, plantilla y validador del tablero por funcionalidad. Engram y otros MCP se configuran por separado.

## Cómo trabaja el perfil

Una consulta se resuelve leyendo. Un cambio entendido y reversible se implementa y prueba directamente. Para una funcionalidad que otra sesión tendría dificultad en retomar, el orquestador conserva criterios, tareas, decisiones y evidencia en Git. Delega cuando necesita un mapa del código, un especialista, trabajo paralelo independiente o una revisión por riesgo.

El flujo integra TDD cuando hay una prueba adecuada. Los cambios se dividen por comportamiento, manteniendo pruebas y documentación junto al código. El presupuesto orientativo es de 400 líneas de autoría por unidad revisable. Las reglas del repositorio destino siguen vigentes.

## Proyecto y mantenimiento

| Necesidad | Referencia |
| --- | --- |
| Operar la TUI y la CLI | [Manager](docs/TUI.md) |
| Entender el producto y las decisiones | [Trayectoria y roadmap](docs/ROADMAP.md) |
| Revisar novedades y cambios históricos | [Release 1.0.1](docs/releases/v1.0.1.md) · [Changelog](CHANGELOG.md) |
| Verificar el perfil y reproducir las pruebas | [Pruebas](docs/TESTING.md) · [Evaluación conductual](docs/EVALUATION.md) |
| Contribuir o informar un problema | [Contribuciones](CONTRIBUTING.md) · [Seguridad](SECURITY.md) |
