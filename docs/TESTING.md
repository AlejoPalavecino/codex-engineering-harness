# Verificación del Manager y del perfil

Usá un directorio temporal para las pruebas de instalación. Los tests de integración crean perfiles separados y ejecutan los scripts reales de PowerShell.

## Checks del repositorio

Desde la raíz, con Go 1.26+, PowerShell y Python 3.11+:

```sh
python tests/validate_repository.py
python -B -m unittest discover -s tests -p 'test_*.py'
go test ./...
go vet ./...
go mod tidy -diff
go run ./cmd/chuli --source . eval check
```

Para comprobar los fixtures Python ejecutá `python -B -m unittest discover -s tests` desde cada uno de `evals/fixtures/loans` y `evals/fixtures/importer`. La opción `-B` evita que la caché de bytecode altere la huella de esos fixtures.

## Instalación y recuperación

- `tests/windows-smoke.ps1` recorre la instalación, verificación y reversión con un perfil temporal.
- `tests/windows-regression.ps1` cubre personalizaciones, TOML y fallos de seguridad de versiones anteriores.
- `tests/windows-ttp-resources.ps1` comprueba preservación, actualización y rollback. El CI ejecuta esa suite con los dos motores PowerShell de Windows.
- `go test ./internal/manager` incluye el camino de los paquetes de release. Ejecutar con PowerShell y Python en PATH permite probar los scripts reales, incluyendo la restauración y el diagnóstico con la copia local verificada.

En una instalación de aceptación, verificá primero una acción en vista previa, instalá, reabrí el cliente y comprobá el descubrimiento del catálogo. Una segunda prueba debe partir de un perfil personalizado con instrucciones y archivos ajenos. Compará sus bytes antes y después del rollback.

## Evidencia de integración para 1.0.0

En un perfil temporal sin credenciales, Codex CLI 0.167.0-alpha.4 aceptó la configuración de inicio y su app-server devolvió las 23 skills de Chuli habilitadas. `config/read` mostró los seis roles. El cliente también enumeró cinco skills internas. Esta prueba acredita la configuración efectiva observada por ese cliente; no ejecutó modelos ni abrió la interfaz gráfica.

Las migraciones de laboratorio desde los instaladores originales de 0.2.0 y 0.2.1 conservaron el modelo personalizado. El rollback de cada caso restauró los 17 archivos anteriores byte por byte. Una prueba adicional del Manager sirve dos versiones sintéticas del ZIP por HTTP de prueba y ejecuta instalación, actualización, diagnóstico offline y doble rollback desde un directorio vacío.

También se generó un `AGENTS.md` de 89.717 bytes con el renderer de Gentle AI v4.4.0 para Codex. El preflight normal detectó el conflicto, el opt-in en vista previa conservó los bytes y la adopción aplicada pasó el verificador. El rollback recuperó el SHA-256 original `1e1b1678780b15e24c5581995e88462930ba435750a6fcf2f7009ae9fd0938aa`. La prueba se hizo sin credenciales y con un perfil de laboratorio.

## Distribución

### Instalador del comando en Windows

`tests/windows-manager-install.ps1` carga el instalador como biblioteca y usa exclusivamente carpetas temporales y `PathTarget Process`. Comprueba vista previa, instalación, repetición sin duplicar PATH, hashes inválidos, ZIP inesperado, archivos ajenos/modificados, destinos con separadores o parcialmente relativos y restauración de ejecutable/registro/PATH después de un fallo simulado. Ejecutar con PowerShell 7 y `powershell.exe -NoProfile -File tests/windows-manager-install.ps1` para Windows PowerShell 5.1. No modifica el PATH del usuario ni instala un perfil Codex.

La comprobación local adicional con la descarga oficial v1.1.0 verificó SHA-256, instalación temporal y resolución de `chuli version` en ambos motores, manteniendo intacto el PATH de usuario. La rama que escribe ese PATH se revisó estáticamente; no se afirma instalación permanente realizada. La revisión independiente revalidó el rechazo de `;` y de `C:relative-install`/`\relative-install`.

El workflow [Validate harness](../.github/workflows/validate.yml) construye un paquete de perfil y ejecutables nativos. El [publicador](../.github/workflows/release.yml) recupera esos artefactos por identificador de ejecución de CI, verifica que el SHA siga siendo el de main y publica los archivos junto a sus sumas SHA-256. El workflow también admite publicación manual de futuras versiones con aprobación explícita del número de versión y el run validado.

La estructura de los TOML acredita modelos configurados. El descubrimiento en el cliente comprueba la disponibilidad del catálogo. Para afirmar el comportamiento de un modelo se necesita una sesión y evidencia de las acciones observadas. La [metodología de evaluación](EVALUATION.md) describe ese trabajo por separado.
