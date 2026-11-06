# Chuli Manager — TUI v0.2.0

Chuli Manager es una interfaz de terminal en Go. Se ejecuta en Windows Terminal, PowerShell o la terminal integrada de Codex; no es un widget del chat.

## Ejecutar desde un checkout

Para compilar: Go 1.26+, Git y Windows PowerShell. Para instalar el harness: también Python 3.11+ (el parser TOML seguro de v0.1.2). Los binarios compilados no necesitan que el usuario final instale Go.

```powershell
go run ./cmd/chuli --source .
go build -o chuli.exe ./cmd/chuli
.\chuli.exe --source .
```

La release oficial incluye binarios para Windows x64, Linux x64 y macOS arm64, además de `SHA256SUMS.txt`. En Windows, descargá el ZIP desde [GitHub Releases](https://github.com/AlejoPalavecino/codex-engineering-harness/releases), extraé `chuli.exe` y ejecutalo desde PowerShell. No hace falta tener Go instalado. La **instalación/rollback solo están habilitados en Windows** por ahora.

## Menú principal

- **Instalar o actualizar desde la última release estable:** consulta GitHub Releases del repositorio oficial. Rechaza prereleases, versiones iguales/inferiores y versiones anteriores a v0.1.2, cuyo instalador tenía fallos conocidos. Descarga el tag exacto en un directorio temporal, sin modificar el checkout original.
- **Instalar desde checkout local:** ejecuta `install.ps1` sin `-Apply`; muestra cambios y solo escribe tras confirmación `S/Y`.
- **Diagnóstico:** ejecuta `verify.ps1` del checkout local. No verifica modelos ejecutados ni aislamiento real.
- **Rollback:** utiliza el script existente del checkout local, con dry-run y confirmación. Impide revertir automáticamente instalaciones v0.1.1 que tenían un fallo de rollback conocido.
- **Acerca de:** presenta identidad y protocolos.

Las teclas `↑/↓` o `j/k` navegan; `Enter` selecciona; `S/Y` confirma; `N/Esc` cancela; `Q` sale. Durante una operación de escritura no se permite cerrar la TUI mediante estas teclas para evitar abandono de procesos.

## Comandos adicionales

```powershell
.\chuli.exe version
.\chuli.exe status
.\chuli.exe --source . doctor
```

La bandera `--source` debe preceder al subcomando. Si no se indica se utiliza el directorio actual como candidato a checkout.

## Seguridad y límites

1. Nunca se aplica automáticamente una actualización. Se muestra el plan y se solicita aprobación.
2. El programa reutiliza los scripts de instalación, verificación y rollback de v0.1.2. No sustituye sus controles de backup y preservación.
3. Por defecto no ofrece `-ReplaceCustomized`: las personalizaciones se conservan conforme al instalador. Para forzar un reemplazo hay que revisarlo manualmente.
4. Obtener una release requiere Internet y Git. Se cotejan el tag y el archivo `VERSION`, pero **esta primera versión no verifica firmas criptográficas de la release**. Los artefactos oficiales firmados y su comprobación quedan antes de un mecanismo de actualización sin supervisión.
5. Diagnóstico y rollback requieren un checkout local válido. Las instalaciones v0.1.1 se recuperan manualmente desde el backup por el fallo conocido de su script.
6. La opción remota consulta la última release estable. Los instaladores anteriores a v0.1.2 se rechazan por sus defectos conocidos.
7. El Manager todavía no autoactualiza su propio binario, no soporta Antigravity y no instala en macOS/Linux.

## Próximos pasos

Firmas verificables para ejecutables/releases, recuperación independiente del checkout, actualizaciones del Manager, adaptadores específicos y soporte en otros sistemas.
