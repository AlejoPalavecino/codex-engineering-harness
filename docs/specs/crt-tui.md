# TUI Consola CRT

El usuario aprobó e indicó implementar la dirección Consola CRT el 2026-11-09, separada del asistente de integraciones. Se conserva Go, Bubble Tea, Lipgloss y los contratos de operación. El logo PNG existente inspira una PC en texto y la palabra CHULI; no se requieren protocolos de imágenes ni nuevos recursos externos.

## Diseño aprobado

Inicio con marca, versión del Manager y estado del perfil diferenciados; acciones con explicación contextual. En terminal ancha, explicación lateral; en angosta, debajo. Paleta verde fósforo y acentos beige con contenido legible, selección por marcador además de color y controles visibles. Las confirmaciones conservan destino, cambios y aceptación explícita; errores, éxito y trabajo ocupado tienen texto propio. Mantener navegación, cancelación y desplazamiento actuales.

## Aceptación

- CRT-01: marca textual reconocible y propósito del Manager comprensible; acciones y estado disponibles sin depender del color.
- CRT-02: distribución adaptada a anchos de terminal; contenido largo desplazable y atajos disponibles.
- CRT-03: sin cambios a instalación, integración, autorización, confirmación o reversión; no ejecutar operaciones al navegar.
- CRT-04: pruebas del render y flujo, build/vet y comprobación en terminal real con perfil temporal; declarar plataformas y tamaños no ejecutados.

La implementación visual usa archivos nuevos y una conexión mínima a View. Se revierte esa conexión y los archivos de presentación para recuperar el render anterior. El usuario autorizó incluir y publicar esta mejora en v1.1.0 el 2026-11-09. Registro: odd/tasks/crt-tui.md.
