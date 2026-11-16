#Requires -Version 5.1
[CmdletBinding()]
param(
    [string]$CodexHome,
    [ValidateSet('','codex','claude','agy')][string]$Client = ''
)
Set-StrictMode -Version Latest
$ErrorActionPreference = 'Stop'

function Get-Hash([string]$Path) {
    return Get-ChuliFileSha256 $Path
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
$effectiveClient = if ($Client) { $Client } else { 'codex' }
$homePath = Resolve-ProfileHome $CodexHome $effectiveClient
$statePath = Join-Path $homePath 'harness-install/state.json'
if (-not (Test-Path -LiteralPath $statePath -PathType Leaf)) { throw 'No se encontró state.json de una instalación del harness.' }
$state = Get-Content -LiteralPath $statePath -Encoding UTF8 -Raw | ConvertFrom-Json
if ($null -eq $state.files) { throw 'Manifiesto de instalación incompleto.' }
if (-not $Client -and $null -ne $state.PSObject.Properties['client'] -and $state.client) {
    $effectiveClient = [string]$state.client
}
$instrRel = Get-ClientInstructionPath $effectiveClient
$configRel = Get-ClientConfigPath $effectiveClient
$root = [IO.Path]::GetFullPath((Join-Path $PSScriptRoot '..'))
$python = Get-Python
$disabledTempFile = ''
try {
    if (($state.PSObject.Properties.Name -contains 'disabled_skills') -and $null -ne $state.disabled_skills -and @($state.disabled_skills).Count -gt 0) {
        $disabledTempFile = Join-Path ([IO.Path]::GetTempPath()) ('chuli-verify-disabled-' + [guid]::NewGuid().ToString('N') + '.json')
        [IO.File]::WriteAllText($disabledTempFile, (ConvertTo-Json -InputObject @($state.disabled_skills) -Compress), (New-Object System.Text.UTF8Encoding($false)))
    }
    $inventory = Get-ProfileInventory $python 'verify' $homePath -Client $effectiveClient -DisabledSkillsFile $disabledTempFile
} finally {
    if ($disabledTempFile -and (Test-Path -LiteralPath $disabledTempFile -PathType Leaf)) {
        Remove-Item -LiteralPath $disabledTempFile -Force -ErrorAction SilentlyContinue
    }
}
Write-Output ('Orchestrator: Chuli, ' + (Join-Path $homePath $instrRel))
$configPath = Join-Path $homePath $configRel

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
    if ($resource.path -eq $instrRel) { continue }
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
    if ($item.path -ne $configRel) { continue }
    $actual = Get-Hash (Join-Path $homePath $item.path)
    if ($actual -ne $item.installed_sha256) {
        Write-Warning ($configRel + ' cambió después de instalarse; se validarán sintaxis y claves administradas.')
        $warnings++
    }
}
foreach ($custom in @($state.preserved_customizations)) {
    Write-Output ('Personalización preservada durante instalación: ' + $custom)
}

if ($effectiveClient -eq 'codex') {
    $tool = Join-Path $root 'scripts/toml_config.py'
    $params = @($tool,'validate','--input',$configPath,'--manifest',$statePath)
} else {
    $tool = Join-Path $root 'scripts/client_targets.py'
    $params = @($tool,'json-config','validate','--client',$effectiveClient,'--input',$configPath,'--manifest',$statePath)
}
$command = [string]$python.command
$allArgs = @($python.prefix) + $params
$result = & $command @allArgs 2>&1
Write-Output (($result | Out-String).Trim())
if ($LASTEXITCODE -ne 0) { $errors++ }
if ($effectiveClient -eq 'agy') {
    $agyCliStatePath = Join-Path $homePath 'harness-install/agy-cli-settings.json'
    if (Test-Path -LiteralPath $agyCliStatePath -PathType Leaf) {
        $cliMeta = Get-Content -LiteralPath $agyCliStatePath -Encoding UTF8 -Raw | ConvertFrom-Json
        $expectedCliTarget = Resolve-AgyCliSettingsPath $homePath 'agy'
        if (-not $expectedCliTarget -or $cliMeta.target -ne $expectedCliTarget) {
            Write-Warning 'Meta de antigravity-cli/settings.json fuera del destino esperado.'
            $errors++
        } else {
            Assert-ChuliSafeRegularTarget $expectedCliTarget
            $cliValParams = @((Join-Path $root 'scripts/client_targets.py'),'json-config','validate','--client','agy','--input',$expectedCliTarget,'--manifest',$agyCliStatePath)
            $cliValArgs = @($python.prefix) + $cliValParams
            $cliValOut = & $command @cliValArgs 2>&1
            Write-Output ('antigravity-cli/settings.json: ' + (($cliValOut | Out-String).Trim()))
            if ($LASTEXITCODE -ne 0) { $errors++ }
        }
    }
}
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
