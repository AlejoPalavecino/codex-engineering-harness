#Requires -Version 5.1
[CmdletBinding()]
param([switch]$Apply, [string]$CodexHome)
Set-StrictMode -Version Latest
$ErrorActionPreference = 'Stop'

function Save-Text {
    param([string]$Path,[string]$Content)
    $parent = Split-Path -Parent $Path
    New-Item -ItemType Directory -Force -Path $parent | Out-Null
    [IO.File]::WriteAllText($Path,$Content,(New-Object System.Text.UTF8Encoding($false)))
}
$homePath = if ($CodexHome) { [IO.Path]::GetFullPath($CodexHome) } elseif ($env:CODEX_HOME) { [IO.Path]::GetFullPath($env:CODEX_HOME) } else { Join-Path $HOME '.codex' }
$statePath = Join-Path $homePath 'harness-install/state.json'
if (-not (Test-Path -LiteralPath $statePath -PathType Leaf)) { throw 'No existe instalacion registrada para revertir.' }
$state = Get-Content -LiteralPath $statePath -Raw | ConvertFrom-Json
$backup = [string]$state.backup_dir
if (-not (Test-Path -LiteralPath $backup -PathType Container)) { throw 'No existe el respaldo local esperado.' }
foreach ($item in @($state.files)) {
    $target = Join-Path $homePath $item.path
    $present = Test-Path -LiteralPath $target -PathType Leaf
    if (-not $item.installed_sha256) {
        if ($present) { throw ('No se puede revertir: reaparecio un archivo administrado: ' + $item.path) }
    } else {
        if (-not $present) { throw ('No se puede revertir: falta un archivo administrado: ' + $item.path) }
        $actual = (Get-FileHash -LiteralPath $target -Algorithm SHA256).Hash
        if ($actual -ne $item.installed_sha256) {
            throw ('No se puede revertir: archivo editado desde la instalacion: ' + $item.path)
        }
    }
    if ($item.existed -and -not (Test-Path -LiteralPath (Join-Path $backup $item.path))) {
        throw ('Backup incompleto: ' + $item.path)
    }
}
Write-Output ('Reversion disponible: ' + @($state.files).Count + ' archivos; backup ' + $backup)
if (-not $Apply) {
    Write-Output 'DRY-RUN. Para revertir usa -Apply. Cambios posteriores bloquean la reversion.'
    return
}
foreach ($item in @($state.files)) {
    $target = Join-Path $homePath $item.path
    if ($item.existed) {
        $saved = Join-Path $backup $item.path
        Save-Text $target ([IO.File]::ReadAllText($saved))
    } elseif (Test-Path -LiteralPath $target -PathType Leaf) {
        Remove-Item -LiteralPath $target -Force
    }
}
$previous = Join-Path $backup '_previous_state.json'
if ($state.previous_state_existed) {
    if (-not (Test-Path -LiteralPath $previous)) { throw 'No existe state.json anterior respaldado.' }
    Save-Text $statePath ([IO.File]::ReadAllText($previous))
} else {
    Remove-Item -LiteralPath $statePath -Force
}
Write-Output 'ROLLBACK OK. Los backups no fueron borrados. Reiniciar Codex.'
