#Requires -Version 5.1
[CmdletBinding()]
param([string]$CodexHome)
Set-StrictMode -Version Latest
$ErrorActionPreference = 'Stop'

$homePath = if ($CodexHome) { [IO.Path]::GetFullPath($CodexHome) } elseif ($env:CODEX_HOME) { [IO.Path]::GetFullPath($env:CODEX_HOME) } else { Join-Path $HOME '.codex' }
$statePath = Join-Path $homePath 'harness-install/state.json'
if (-not (Test-Path -LiteralPath $statePath -PathType Leaf)) {
    throw 'No hay state.json del harness en este CODEX_HOME.'
}
$state = Get-Content -LiteralPath $statePath -Raw | ConvertFrom-Json
$failures = 0
foreach ($item in @($state.files)) {
    $target = Join-Path $homePath $item.path
    $present = Test-Path -LiteralPath $target -PathType Leaf
    if (-not $item.installed_sha256) {
        if ($present) { Write-Warning ('Archivo retirado reaparecio: ' + $item.path); $failures++ }
        continue
    }
    if (-not $present) {
        Write-Warning ('Falta archivo: ' + $item.path)
        $failures++
        continue
    }
    $actual = (Get-FileHash -LiteralPath $target -Algorithm SHA256).Hash
    if ($actual -ne $item.installed_sha256) {
        Write-Warning ('Cambio posterior a la instalacion: ' + $item.path)
        $failures++
    }
}
foreach ($name in @('solution_architect','product_designer','test_engineer','quality_reviewer')) {
    if (-not (Test-Path -LiteralPath (Join-Path $homePath ('agents/' + $name + '.toml')))) { $failures++ }
}
if (Get-Command codex -ErrorAction SilentlyContinue) {
    Write-Output 'Codex detectado:'
    & codex --version
} else {
    Write-Warning 'Codex CLI no esta en PATH: no se puede validar la ejecucion de agentes/modelos.'
}
if (Get-Command engram -ErrorAction SilentlyContinue) {
    Write-Output 'Engram detectado; su base no fue modificada por el instalador.'
} else {
    Write-Output 'Engram no instalado (integracion opcional).'
}
Write-Output ('Version registrada: ' + $state.version)
Write-Output ('Comprobaciones de integridad fallidas: ' + $failures)
Write-Output 'Este check no prueba sandbox real, MCP remoto, modelos ejecutados ni comportamiento de agentes.'
if ($failures -gt 0) { exit 1 }
