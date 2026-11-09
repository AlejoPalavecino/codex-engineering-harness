<div align="center">

<img src="assets/brand/chuli-logo.png" width="740" alt="Chuli: una PC retro con monitor de fósforo verde y el nombre del proyecto en letras pixeladas">

# Chuli

**Un perfil de ingeniería para el Codex que ya usás.**

De la idea al cambio comprobado, con decisiones claras y un compañero que te explica mientras trabajan.

[Descargar](https://github.com/AlejoPalavecino/codex-engineering-harness/releases/latest) · [Instalación](#instalacion) · [Cómo trabaja](#como-trabaja) · [Documentación](#documentacion)

**6 especialistas · 23 skills · instalación con respaldo y rollback**

</div>

---

```text
                 CHULI / PERFIL PERSONAL DE INGENIERÍA

       IDEA  ───►  DECISIONES  ───►  CAMBIO  ───►  EVIDENCIA
```

Chuli instala instrucciones, roles y herramientas de trabajo en tu perfil local de Codex. El **Manager** administra la instalación y las versiones desde la terminal. Después seguís programando desde Codex CLI o desde el modo Codex local de la app de escritorio de ChatGPT.

El orquestador te ayuda a definir lo que falta, resolver cambios y comprobar el resultado. Habla en español argentino, explica primero con palabras cotidianas y suma el concepto técnico cuando te sirve para aprender.

<a name="instalacion"></a>

## 01 · Encendé el equipo

**Necesitás Python 3.11 o posterior y PowerShell para aplicar el perfil.**

1. **Descargá el Manager** desde la [última release estable](https://github.com/AlejoPalavecino/codex-engineering-harness/releases/latest).
2. **Extraé y abrí el ejecutable:** `chuli.exe` en Windows o `chuli` en Linux/macOS.
3. **Elegí instalar desde la última release.** Revisá el destino y los cambios, y confirmá. El Manager descarga el perfil y verifica su SHA-256.
4. **Abrí una sesión nueva de Codex** en tu proyecto y empezá con una tarea concreta.

| Tu equipo | Archivo del Manager |
| --- | --- |
| Windows x64 | `chuli_v<versión>_windows_amd64.zip` |
| Linux x64 | `chuli_v<versión>_linux_amd64.tar.gz` |
| macOS ARM64 | `chuli_v<versión>_darwin_arm64.tar.gz` |

La [guía de instalación](docs/INSTALLATION.md) cubre requisitos, personalizaciones, actualización, verificación y rollback. Los cuatro artefactos de cada release se acompañan de `SHA256SUMS.txt` para comprobar las descargas.

> **¿Ya tenés un perfil personalizado?** Revisá la vista previa antes de aplicar. Chuli tiene reglas para preservar archivos ajenos y recursos personalizados; reemplazarlos requiere una decisión explícita. Consultá la guía antes de actualizar.

<a name="como-trabaja"></a>

## 02 · Del pedido al resultado

| Vos necesitás… | Chuli trabaja así |
| --- | --- |
| Entender código o evaluar una idea | Investiga y explica con referencias. Una consulta mantiene el trabajo en lectura. |
| Corregir algo puntual | Inspecciona, hace el cambio autorizado y ejecuta una comprobación proporcionada. |
| Construir una funcionalidad | Aclara decisiones pendientes, propone el diseño y conserva tareas y evidencia cuando hace falta continuidad. |
| Diseñar una interfaz | Parte de la tarea de la persona, define estados y prepara un guion para comprobar apariencia, interacción y resultado. |

Pregunta por decisiones que cambian el resultado y resuelve detalles técnicos reversibles dentro del alcance autorizado. Esa autorización acompaña la implementación, las correcciones y la verificación. Si cambia el alcance, vuelve a consultar lo afectado.

Delega cuando una exploración, un especialista o una revisión independiente aporta valor. Usa TDD cuando hay una prueba adecuada y procura unidades revisables de unas 400 líneas, con código, pruebas y documentación del comportamiento juntos. Las reglas de tu proyecto siguen vigentes.

### Probalo con un pedido real

> «Explicame cómo funciona este módulo y dónde se comprueban los permisos. Por ahora, solo investigá».

> «Quiero agregar reservas de equipos. Primero definamos qué puede hacer cada persona y cómo se cancela una reserva».

> «Revisemos este formulario: qué necesita completar la persona, cómo se recupera de un error y cómo vamos a comprobarlo con egolite».

## 03 · Debajo del gabinete

| Pieza | Para qué sirve |
| --- | --- |
| **Orquestador** | Coordina el trabajo y conserva las decisiones de producto, diseño y entrega. |
| **6 especialistas** | Exploración, implementación, arquitectura, diseño de producto, pruebas y revisión. [Roles y modelos](docs/AGENTS.md). |
| **5 skills de Chuli** | Relevamiento, arquitectura, UX/UI, especificación y pruebas. |
| **18 skills de Gentle AI v4.4.0** | Procedimientos de documentación, revisión y unidades de trabajo, entre otros; conservan sus recursos y atribución. |
| **Seguimiento por funcionalidad** | Guías, plantilla y validador para retomar trabajos sustanciales desde decisiones y evidencia en Git. |
| **Manager** | Instalación, actualización, diagnóstico y rollback del perfil. |

El perfil usa `CODEX_HOME/AGENTS.md`, configuración de roles y modelos en `config.toml`, y recursos instalados. El destino predeterminado suele ser `~/.codex`. El [catálogo](docs/protocols/SKILLS.md) explica el objetivo y las condiciones de cada skill; [licencias y procedencia](THIRD_PARTY_NOTICES.md) identifica los recursos de terceros.

**Las integraciones externas se configuran por separado.** Egolite es la preferencia de revisión web cuando está disponible y autorizado. Engram y otros MCP son opcionales; el paquete de Chuli no los instala automáticamente. Los modelos declarados dependen de lo que el cliente y la cuenta permitan ejecutar.

<a name="documentacion"></a>

## 04 · El manual

| Quiero… | Abrí… |
| --- | --- |
| Instalar, actualizar o volver atrás | [Instalación](docs/INSTALLATION.md) |
| Operar el Manager y su CLI | [Manager](docs/TUI.md) |
| Conocer los roles y las skills | [Especialistas](docs/AGENTS.md) · [Catálogo](docs/protocols/SKILLS.md) |
| Entender las decisiones del producto | [Trayectoria y roadmap](docs/ROADMAP.md) |
| Ver qué cambió | [Release 1.0.1](docs/releases/v1.0.1.md) · [Changelog](CHANGELOG.md) |
| Revisar pruebas y límites de evidencia | [Verificación](docs/TESTING.md) · [Evaluación conductual](docs/EVALUATION.md) |
| Contribuir o informar un problema | [Contribuciones](CONTRIBUTING.md) · [Seguridad](SECURITY.md) |

---

<p align="center"><strong>Estética de otra época. Decisiones y evidencia en tu proyecto.</strong><br><a href="docs/BRAND.md">Identidad de Chuli</a> · <a href="LICENSE">Licencia</a></p>
