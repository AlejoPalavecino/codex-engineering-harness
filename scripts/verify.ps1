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
    if ($null -ne $State.managed_files) {
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
$homePath = if ($CodexHome) { [IO.Path]::GetFullPath($CodexHome) } elseif ($env:CODEX_HOME) { [IO.Path]::GetFullPath($env:CODEX_HOME) } else { Join-Path $HOME '.codex' }
$root = [IO.Path]::GetFullPath((Join-Path $PSScriptRoot '..'))
$statePath = Join-Path $homePath 'harness-install/state.json'
$configPath = Join-Path $homePath 'config.toml'
if (-not (Test-Path -LiteralPath $statePath -PathType Leaf)) { throw 'No se encontró state.json de una instalación del harness.' }
$state = Get-Content -LiteralPath $statePath -Raw | ConvertFrom-Json
if ($null -eq $state.files) { throw 'Manifiesto de instalación incompleto.' }

$errors = 0
$warnings = 0
$legacyBaselinePath = Join-Path $root 'scripts/legacy-managed-hashes.json'
if (-not (Test-Path -LiteralPath $legacyBaselinePath -PathType Leaf)) {
    throw 'Falta scripts/legacy-managed-hashes.json; no se puede validar el inventario administrado.'
}
$legacyBaselines = Get-Content -LiteralPath $legacyBaselinePath -Raw | ConvertFrom-Json
$requiredResources = @(
    @{path='AGENTS.md'; source='config/AGENTS.md'},
    @{path='agents/solution_architect.toml'; source='agents/solution_architect.toml'},
    @{path='agents/product_designer.toml'; source='agents/product_designer.toml'},
    @{path='agents/test_engineer.toml'; source='agents/test_engineer.toml'},
    @{path='agents/quality_reviewer.toml'; source='agents/quality_reviewer.toml'},
    @{path='skills/harness-discovery/SKILL.md'; source='skills/harness-discovery/SKILL.md'},
    @{path='skills/harness-architecture/SKILL.md'; source='skills/harness-architecture/SKILL.md'},
    @{path='skills/harness-ux-ui/SKILL.md'; source='skills/harness-ux-ui/SKILL.md'},
    @{path='skills/harness-specification/SKILL.md'; source='skills/harness-specification/SKILL.md'},
    @{path='skills/harness-tdd-qa/SKILL.md'; source='skills/harness-tdd-qa/SKILL.md'},
    @{path='harness/bin/taskboard.py'; source='scripts/taskboard.py'},
    @{path='harness/templates/feature-tasks.md'; source='templates/feature-tasks.md'},
    @{path='harness/docs/protocols/TTP.md'; source='docs/protocols/TTP.md'}
)
foreach ($resource in $requiredResources) {
    $target = Join-Path $homePath $resource.path
    $actual = Get-Hash $target
    if ($null -eq $actual) {
        Write-Warning ('Falta recurso obligatorio: ' + $resource.path)
        $errors++
        continue
    }
    $expected = Get-ExpectedManagedHash $state $legacyBaselines $resource.path
    if ($null -eq $expected) {
        $source = Join-Path $root $resource.source
        $sourceHash = Get-Hash $source
        if ($null -eq $sourceHash -or $actual -ne $sourceHash) {
            Write-Warning ('No existe una huella confiable para verificar: ' + $resource.path)
            $warnings++
        }
    } elseif ($actual -ne $expected) {
        Write-Warning ('Archivo administrado con deriva o personalización: ' + $resource.path)
        $warnings++
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
    Write-Output (($result | Out-String).Trim())
    if ($LASTEXITCODE -ne 0) { $errors++ }
    foreach ($name in @('solution_architect','product_designer','test_engineer','quality_reviewer')) {
        $agentPath = Join-Path $homePath ('agents/' + $name + '.toml')
        if (-not (Test-Path -LiteralPath $agentPath -PathType Leaf)) { continue }
        # Avoid nested quotes: Windows PowerShell 5.1 rewrites native -c arguments.
        & $command @($python.prefix) -c 'import tomllib,pathlib,sys; tomllib.loads(pathlib.Path(sys.argv[1]).read_bytes().decode())' $agentPath 2>$null
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
