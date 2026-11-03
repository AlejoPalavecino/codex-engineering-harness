#Requires -Version 5.1
<#
.SYNOPSIS
  Instala o actualiza Codex Engineering Harness. Sin -Apply es dry-run.
.DESCRIPTION
  v0.1.2: validación TOML, plan explícito, preservación de personalizaciones,
  backup completo antes de escribir y reversión ante fallos de instalación.
  Python 3.11+ se requiere para una actualización segura de config.toml.
#>
[CmdletBinding()]
param(
    [switch]$Apply,
    [switch]$ConfigureModels,
    [switch]$RemoveLegacyAgents,
    [switch]$SkipTaskProgress,
    [switch]$ReplaceCustomized,
    [string]$CodexHome
)
Set-StrictMode -Version Latest
$ErrorActionPreference = 'Stop'

function Get-Hash([string]$Path) {
    if (-not (Test-Path -LiteralPath $Path -PathType Leaf)) { return $null }
    return (Get-FileHash -LiteralPath $Path -Algorithm SHA256).Hash
}
function Save-Text([string]$Path, [string]$Value) {
    $parent = Split-Path -Parent $Path
    New-Item -ItemType Directory -Force -Path $parent | Out-Null
    [IO.File]::WriteAllText($Path, $Value, (New-Object System.Text.UTF8Encoding($false)))
}
function Get-Python {
    foreach ($candidate in @('python', 'py')) {
        if (-not (Get-Command $candidate -ErrorAction SilentlyContinue)) { continue }
        $prefix = @()
        if ($candidate -eq 'py') { $prefix = @('-3.11') }
        try {
            & $candidate @prefix -c 'import tomllib; import sys; sys.exit(0 if sys.version_info >= (3, 11) else 1)' 2>$null
            if ($LASTEXITCODE -eq 0) {
                return @{ command = $candidate; prefix = @($prefix) }
            }
        } catch { }
    }
    throw 'Python 3.11+ requerido para validar TOML antes de modificar Codex. No se realizaron cambios.'
}
function Invoke-Toml {
    param($Runtime, [string[]]$Arguments)
    $command = [string]$Runtime.command
    $allArgs = @($Runtime.prefix) + $Arguments
    $output = & $command @allArgs 2>&1
    if ($LASTEXITCODE -ne 0) {
        throw ('TOML preflight failed: ' + (($output | Out-String).Trim()))
    }
    return $output
}
function Record-Previous([object]$Previous, [string]$Relative) {
    if ($null -eq $Previous) { return $null }
    foreach ($record in @($Previous.files)) {
        if ($record.path -eq $Relative) { return $record }
    }
    return $null
}
function Is-Within([string]$Root, [string]$Path) {
    $rootFull = [IO.Path]::GetFullPath($Root).TrimEnd('\','/') + [IO.Path]::DirectorySeparatorChar
    $full = [IO.Path]::GetFullPath($Path)
    return $full.StartsWith($rootFull, [StringComparison]::OrdinalIgnoreCase)
}

$homePath = if ($CodexHome) { [IO.Path]::GetFullPath($CodexHome) } elseif ($env:CODEX_HOME) { [IO.Path]::GetFullPath($env:CODEX_HOME) } else { Join-Path $HOME '.codex' }
$root = [IO.Path]::GetFullPath((Join-Path $PSScriptRoot '..'))
$harnessVersion = [IO.File]::ReadAllText((Join-Path $root 'VERSION')).Trim()
if ($harnessVersion -notmatch '^[0-9]+\.[0-9]+\.[0-9]+$') {
    throw 'VERSION inválido; instalación cancelada antes de modificar Codex.'
}
$statePath = Join-Path $homePath 'harness-install/state.json'
$configPath = Join-Path $homePath 'config.toml'
$previous = $null
if (Test-Path -LiteralPath $statePath -PathType Leaf) {
    $previous = Get-Content -LiteralPath $statePath -Raw | ConvertFrom-Json
    if ($null -eq $previous.files) { throw 'Manifiesto anterior sin lista de archivos. No se modificó Codex.' }
}
$python = Get-Python
$staging = Join-Path ([IO.Path]::GetTempPath()) ('harness-stage-' + [guid]::NewGuid().ToString('N'))
$changes = New-Object 'System.Collections.Generic.List[object]'
$preserved = New-Object 'System.Collections.Generic.List[string]'
$modified = New-Object 'System.Collections.Generic.List[object]'
$applied = $false
$backup = $null

try {
    New-Item -ItemType Directory -Path $staging -Force | Out-Null
    $items = New-Object 'System.Collections.Generic.List[object]'
    $items.Add(@{path='AGENTS.md'; source='config/AGENTS.md'; kind='copy'})
    foreach ($name in @('solution_architect','product_designer','test_engineer','quality_reviewer')) {
        $items.Add(@{path=('agents/' + $name + '.toml'); source=('agents/' + $name + '.toml'); kind='copy'})
    }
    foreach ($name in @('harness-discovery','harness-architecture','harness-ux-ui','harness-specification','harness-tdd-qa')) {
        $items.Add(@{path=('skills/' + $name + '/SKILL.md'); source=('skills/' + $name + '/SKILL.md'); kind='copy'})
    }
    $items.Add(@{path='harness/bin/taskboard.py'; source='scripts/taskboard.py'; kind='copy'})
    if ($RemoveLegacyAgents) {
        foreach ($name in @('architect','explorer','reviewer','slice_worker','test_runner')) {
            $items.Add(@{path=('agents/' + $name + '.toml'); source=''; kind='remove'})
        }
    }

    foreach ($item in $items) {
        $target = Join-Path $homePath $item.path
        $oldHash = Get-Hash $target
        $source = $null
        if ($item.kind -eq 'copy') {
            $source = Join-Path $root $item.source
            if (-not (Test-Path -LiteralPath $source -PathType Leaf)) { throw ('Falta archivo del repositorio: ' + $item.source) }
            $sourceHash = Get-Hash $source
            if ($sourceHash -eq $oldHash) {
                $action = 'UNCHANGED'
            } elseif ($null -eq $oldHash) {
                $action = 'CREATE'
            } else {
                $oldRecord = Record-Previous $previous $item.path
                if ($null -ne $oldRecord -and $oldRecord.installed_sha256 -eq $oldHash) {
                    $action = 'UPDATE'
                } elseif ($ReplaceCustomized) {
                    $action = 'REPLACE-APPROVED'
                } else {
                    $action = 'PRESERVE-CUSTOMIZED'
                    $preserved.Add([string]$item.path)
                }
            }
        } else {
            $action = if ($null -ne $oldHash) { 'REMOVE-APPROVED' } else { 'UNCHANGED' }
        }
        $changes.Add([ordered]@{path=[string]$item.path; kind=[string]$item.kind; action=$action; original=$oldHash; source=$source})
    }

    $inputPath = Join-Path $staging 'original.toml'
    $candidatePath = Join-Path $staging 'candidate.toml'
    $reportPath = Join-Path $staging 'config-report.json'
    if (Test-Path -LiteralPath $configPath -PathType Leaf) {
        Copy-Item -LiteralPath $configPath -Destination $inputPath
    } else {
        Save-Text $inputPath ''
    }
    $args = @((Join-Path $root 'scripts/toml_config.py'), 'prepare', '--input', $inputPath,
        '--output', $candidatePath, '--report', $reportPath)
    if ($ConfigureModels) { $args += '--configure-models' }
    if ($SkipTaskProgress) { $args += '--skip-task-progress' }
    if ($ReplaceCustomized) { $args += '--replace-managed' }
    $reportOutput = Invoke-Toml $python $args
    $report = Get-Content -LiteralPath $reportPath -Raw | ConvertFrom-Json
    foreach ($key in @($report.preserved)) { $preserved.Add(('config.toml -> ' + $key)) }
    $oldConfigHash = Get-Hash $configPath
    $newConfigHash = Get-Hash $candidatePath
    $configAction = if ($oldConfigHash -eq $newConfigHash) { 'UNCHANGED' } elseif ($null -eq $oldConfigHash) { 'CREATE' } else { 'UPDATE-VALIDATED' }
    $changes.Add([ordered]@{path='config.toml';kind='config';action=$configAction;original=$oldConfigHash;source=$candidatePath})

    Write-Output ('Codex home: ' + $homePath)
    Write-Output ('Version candidate: 0.1.2')
    Write-Output ('Python TOML validation: PASS')
    Write-Output ('ReplaceCustomized requested: ' + [bool]$ReplaceCustomized)
    foreach ($entry in $changes) {
        Write-Output ('  ' + $entry.action.PadRight(20) + ' ' + $entry.path)
    }
    if ($preserved.Count -gt 0) {
        Write-Output 'Personalizaciones conservadas (no se sobrescriben):'
        foreach ($name in $preserved) { Write-Output ('  ' + $name) }
        Write-Output 'Para reemplazarlas explícitamente, revisá el diff y repetí con -ReplaceCustomized.'
    }
    if (-not $Apply) {
        Write-Output 'DRY-RUN OK. No se modificó CODEX_HOME. Usá -Apply para actualizar.'
        return
    }

    $id = (Get-Date -Format 'yyyyMMdd-HHmmss') + '-' + [guid]::NewGuid().ToString('N').Substring(0,8)
    $backup = Join-Path $homePath ('backups/harness/' + $id)
    New-Item -ItemType Directory -Path $backup -Force | Out-Null
    if ($null -ne $previous) {
        Copy-Item -LiteralPath $statePath -Destination (Join-Path $backup '_previous_state.json')
    }
    foreach ($entry in $changes) {
        if ($entry.action -in @('UNCHANGED','PRESERVE-CUSTOMIZED')) { continue }
        $target = Join-Path $homePath $entry.path
        if ((Get-Hash $target) -ne $entry.original) {
            throw ('Archivo cambió durante el preflight: ' + $entry.path)
        }
        $itemRecord = [ordered]@{
            path=$entry.path
            existed=[bool]($null -ne $entry.original)
            backup_sha256=$entry.original
            installed_sha256=$null
        }
        if ($null -ne $entry.original) {
            $destination = Join-Path $backup $entry.path
            New-Item -ItemType Directory -Path (Split-Path -Parent $destination) -Force | Out-Null
            Copy-Item -LiteralPath $target -Destination $destination
            if ((Get-Hash $destination) -ne $entry.original) {
                throw ('Backup corrupto antes de aplicar: ' + $entry.path)
            }
        }
        $modified.Add($itemRecord)
    }
    if ($null -ne $previous -and -not (Test-Path -LiteralPath (Join-Path $backup '_previous_state.json'))) {
        throw 'Backup del manifiesto anterior incompleto.'
    }
    $applied = $true
    foreach ($entry in $changes) {
        if ($entry.action -in @('UNCHANGED','PRESERVE-CUSTOMIZED')) { continue }
        $target = Join-Path $homePath $entry.path
        if ($entry.kind -eq 'remove') {
            if (Test-Path -LiteralPath $target) { Remove-Item -LiteralPath $target -Force }
        } else {
            Save-Text $target ([IO.File]::ReadAllText($entry.source))
        }
        foreach ($record in $modified) {
            if ($record.path -eq $entry.path) {
                $record.installed_sha256 = Get-Hash $target
                break
            }
        }
    }
    $tracked = @{}
    foreach ($property in $report.tracked.PSObject.Properties) { $tracked[$property.Name] = $property.Value }
    $state = [ordered]@{
        version=$harnessVersion
        installed_at=(Get-Date).ToString('o')
        backup_dir=$backup
        previous_state_existed=[bool]($null -ne $previous)
        files=@($modified.ToArray())
        preserved_customizations=@($preserved.ToArray())
        managed_config=$tracked
    }
    Save-Text $statePath (ConvertTo-Json -InputObject $state -Depth 12)
    Write-Output 'INSTALACION OK. Respaldo creado y configuración validada.'
    Write-Output ('Backup: ' + $backup)
    Write-Output 'Ejecutá scripts/verify.ps1 y reiniciá Codex Desktop.'
} catch {
    $errorText = $_.Exception.Message
    $recoveryErrors = New-Object 'System.Collections.Generic.List[string]'
    if ($applied) {
        foreach ($entry in $modified) {
            try {
                $target = Join-Path $homePath $entry.path
                if ($entry.existed) {
                    $saved = Join-Path $backup $entry.path
                    if ((Get-Hash $saved) -ne $entry.backup_sha256) { throw 'Backup no coincide con SHA esperado' }
                    New-Item -ItemType Directory -Path (Split-Path -Parent $target) -Force | Out-Null
                    Copy-Item -LiteralPath $saved -Destination $target -Force
                } elseif (Test-Path -LiteralPath $target) {
                    Remove-Item -LiteralPath $target -Force
                }
            } catch {
                $recoveryErrors.Add(($entry.path + ': ' + $_.Exception.Message))
            }
        }
    }
    if ($applied) {
        try {
            $oldState = Join-Path $backup '_previous_state.json'
            if ($null -ne $previous -and (Test-Path -LiteralPath $oldState -PathType Leaf)) {
                Copy-Item -LiteralPath $oldState -Destination $statePath -Force
            } elseif ($null -eq $previous -and (Test-Path -LiteralPath $statePath -PathType Leaf)) {
                Remove-Item -LiteralPath $statePath -Force
            }
        } catch { $recoveryErrors.Add(('state.json: ' + $_.Exception.Message)) }
    }
    if ($recoveryErrors.Count -gt 0) {
        throw ('Instalación fallida: ' + $errorText + '. Recuperación INCOMPLETA, revisar backup ' + $backup + ': ' + ($recoveryErrors -join '; '))
    }
    throw ('Instalación abortada: ' + $errorText + '. No se detectaron escrituras sin recuperar.')
} finally {
    if (Test-Path -LiteralPath $staging) {
        Remove-Item -LiteralPath $staging -Recurse -Force -ErrorAction SilentlyContinue
    }
}
