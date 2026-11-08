#Requires -Version 5.1
[CmdletBinding()]
param([string]$CodexHome)
Set-StrictMode -Version Latest
$ErrorActionPreference = 'Stop'

function Get-Hash([string]$Path) {
    if (-not (Test-Path -LiteralPath $Path -PathType Leaf)) { return $null }
    return (Get-FileHash -LiteralPath $Path -Algorithm SHA256).Hash
}
function Get-LegacyManagedHash([object]$Baselines, [string]$Version, [string]$Relative) {
    foreach ($versionEntry in $Baselines.PSObject.Properties) {
        if ($versionEntry.Name -ne $Version) { continue }
        foreach ($entry in $versionEntry.Value.PSObject.Properties) {
            if ($entry.Name -eq $Relative) { return [string]$entry.Value }
        }
    }
    return $null
}
function Get-ExpectedManagedHash([object]$State, [object]$Baselines, [string]$Relative) {
    if ($null -ne $State.PSObject.Properties['preserved_files'] -and $null -ne $State.preserved_files) {
        foreach ($entry in $State.preserved_files.PSObject.Properties) {
            if ($entry.Name -eq $Relative) { return [string]$entry.Value }
        }
    }
    if ($State.PSObject.Properties.Name -contains 'managed_files' -and $null -ne $State.managed_files) {
        foreach ($entry in $State.managed_files.PSObject.Properties) {
            if ($entry.Name -eq $Relative) { return [string]$entry.Value }
        }
    }
    foreach ($record in @($State.files)) {
        if ($record.path -eq $Relative -and $record.installed_sha256) {
            return [string]$record.installed_sha256
        }
    }
    return Get-LegacyManagedHash $Baselines ([string]$State.version) $Relative
}
. (Join-Path $PSScriptRoot 'profile.ps1')
$homePath = Resolve-ProfileHome $CodexHome
$root = [IO.Path]::GetFullPath((Join-Path $PSScriptRoot '..'))
$python = Get-Python
$inventory = Get-ProfileInventory $python 'verify' $homePath
Write-Output ('Orchestrator: Chuli, ' + (Join-Path $homePath 'AGENTS.md'))
$statePath = Join-Path $homePath 'harness-install/state.json'
$configPath = Join-Path $homePath 'config.toml'
if (-not (Test-Path -LiteralPath $statePath -PathType Leaf)) { throw 'No se encontró state.json de una instalación del harness.' }
$state = Get-Content -LiteralPath $statePath -Encoding UTF8 -Raw | ConvertFrom-Json
if ($null -eq $state.files) { throw 'Manifiesto de instalación incompleto.' }

$errors = 0
$warnings = 0
$legacyBaselinePath = Join-Path $root 'scripts/legacy-managed-hashes.json'
if (-not (Test-Path -LiteralPath $legacyBaselinePath -PathType Leaf)) {
    throw 'Falta scripts/legacy-managed-hashes.json; no se puede validar el inventario administrado.'
}
$legacyBaselines = Get-Content -LiteralPath $legacyBaselinePath -Raw | ConvertFrom-Json
foreach ($entry in $inventory.PSObject.Properties) {
    $resource = @{path=$entry.Name; source=[string]$entry.Value}
    $target = Join-Path $homePath $resource.path
    $actual = Get-Hash $target
    if ($null -eq $actual) {
        Write-Warning ('Falta recurso obligatorio: ' + $resource.path)
        $errors++
        continue
    }
    if ($resource.path -eq 'AGENTS.md') { continue }
    $expected = Get-ExpectedManagedHash $state $legacyBaselines $resource.path
    if ($null -eq $expected) {
        Write-Warning ('Falta una huella registrada: ' + $resource.path)
        $errors++
    } elseif ($actual -ne $expected) {
        Write-Warning ('El recurso cambió después de la instalación: ' + $resource.path)
        $errors++
    }
}
foreach ($item in @($state.files)) {
    if ($item.path -ne 'config.toml') { continue }
    $actual = Get-Hash (Join-Path $homePath $item.path)
    if ($actual -ne $item.installed_sha256) {
        Write-Warning 'config.toml cambió después de instalarse; se validarán sintaxis y claves administradas.'
        $warnings++
    }
}
foreach ($custom in @($state.preserved_customizations)) {
    Write-Output ('Personalización preservada durante instalación: ' + $custom)
}

$tool = Join-Path $root 'scripts/toml_config.py'
$params = @($tool,'validate','--input',$configPath,'--manifest',$statePath)
$command = [string]$python.command
$allArgs = @($python.prefix) + $params
$result = & $command @allArgs 2>&1
Write-Output (($result | Out-String).Trim())
if ($LASTEXITCODE -ne 0) { $errors++ }
foreach ($entry in $inventory.PSObject.Properties) {
    $name = $entry.Name
    if ($name -notmatch '^agents/[^/]+\.toml$') { continue }
    $agentPath = Join-Path $homePath $name
    if (-not (Test-Path -LiteralPath $agentPath -PathType Leaf)) { continue }
    # Avoid nested quotes: Windows PowerShell 5.1 rewrites native -c arguments.
    & $command @($python.prefix) -c 'import tomllib,pathlib,sys; tomllib.loads(pathlib.Path(sys.argv[1]).read_bytes().decode())' $agentPath 2>$null
    if ($LASTEXITCODE -ne 0) {
        Write-Warning ('Agente con TOML inválido: ' + $name)
        $errors++
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
