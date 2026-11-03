#Requires -Version 5.1
[CmdletBinding()]
param([string]$CodexHome)
Set-StrictMode -Version Latest
$ErrorActionPreference = 'Stop'

function Get-Hash([string]$Path) {
    if (-not (Test-Path -LiteralPath $Path -PathType Leaf)) { return $null }
    return (Get-FileHash -LiteralPath $Path -Algorithm SHA256).Hash
}
$homePath = if ($CodexHome) { [IO.Path]::GetFullPath($CodexHome) } elseif ($env:CODEX_HOME) { [IO.Path]::GetFullPath($env:CODEX_HOME) } else { Join-Path $HOME '.codex' }
$root = [IO.Path]::GetFullPath((Join-Path $PSScriptRoot '..'))
$statePath = Join-Path $homePath 'harness-install/state.json'
$configPath = Join-Path $homePath 'config.toml'
if (-not (Test-Path -LiteralPath $statePath -PathType Leaf)) { throw 'No se encontró state.json de una instalación del harness.' }
$state = Get-Content -LiteralPath $statePath -Raw | ConvertFrom-Json
if ($null -eq $state.files) { throw 'Manifiesto de instalación incompleto.' }

$errors = 0
$warnings = 0
foreach ($item in @($state.files)) {
    $target = Join-Path $homePath $item.path
    $actual = Get-Hash $target
    if ($actual -eq $item.installed_sha256) { continue }
    if ($item.path -eq 'config.toml' -and $null -ne $actual) {
        Write-Warning 'config.toml cambió después de instalarse; se validarán sintaxis y claves administradas.'
        $warnings++
    } else {
        Write-Warning ('Archivo administrado cambió después de instalar: ' + $item.path)
        $warnings++
    }
}
foreach ($name in @('solution_architect','product_designer','test_engineer','quality_reviewer')) {
    if (-not (Test-Path -LiteralPath (Join-Path $homePath ('agents/' + $name + '.toml')) -PathType Leaf)) {
        Write-Warning ('Falta agente: ' + $name)
        $errors++
    }
}
foreach ($custom in @($state.preserved_customizations)) {
    Write-Output ('Personalización preservada durante instalación: ' + $custom)
}

$tool = Join-Path $root 'scripts/toml_config.py'
$python = $null
foreach ($candidate in @('python','py')) {
    if (-not (Get-Command $candidate -ErrorAction SilentlyContinue)) { continue }
    $prefix = @()
    if ($candidate -eq 'py') { $prefix = @('-3.11') }
    try {
        & $candidate @prefix -c 'import tomllib; import sys; sys.exit(0 if sys.version_info >= (3, 11) else 1)' 2>$null
        if ($LASTEXITCODE -eq 0) { $python = @{command=$candidate;prefix=$prefix}; break }
    } catch {}
}
if ($null -eq $python) {
    Write-Warning 'Falta Python 3.11+: no se pudo validar la sintaxis de TOML.'
    $errors++
} else {
    $params = @($tool,'validate','--input',$configPath,'--manifest',$statePath)
    $command = [string]$python.command
    $allArgs = @($python.prefix) + $params
    $result = & $command @allArgs 2>&1
    Write-Output ($result | Out-String).Trim()
    if ($LASTEXITCODE -ne 0) { $errors++ }
    foreach ($name in @('solution_architect','product_designer','test_engineer','quality_reviewer')) {
        $agentPath = Join-Path $homePath ('agents/' + $name + '.toml')
        if (-not (Test-Path -LiteralPath $agentPath -PathType Leaf)) { continue }
        & $command @($python.prefix) -c 'import tomllib,sys; tomllib.load(open(sys.argv[1], "rb"))' $agentPath 2>$null
        if ($LASTEXITCODE -ne 0) {
            Write-Warning ('Agente con TOML inválido: ' + $name)
            $errors++
        }
    }
}
if (Get-Command codex -ErrorAction SilentlyContinue) {
    Write-Output ('Codex CLI: ' + (& codex --version))
} else {
    Write-Output 'Codex CLI no disponible en PATH; ejecución de agentes no verificada.'
}
Write-Output ('Versión registrada: ' + $state.version)
Write-Output ('Errores: ' + $errors + '; advertencias (deriva no bloqueante): ' + $warnings)
Write-Output 'La verificación NO certifica modelos ejecutados, MCP remoto, sandbox real ni comportamiento del agente.'
if ($errors -gt 0) { exit 1 }
