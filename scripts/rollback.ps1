#Requires -Version 5.1
<#
.SYNOPSIS
  Reverts the most recent managed Codex Engineering Harness installation.
.DESCRIPTION
  Prevalidates every backup and the previous manifest BEFORE restoring anything.
  Refuses rollback if installed files have changed since installation.
#>
[CmdletBinding()]
param([switch]$Apply, [string]$CodexHome)
Set-StrictMode -Version Latest
$ErrorActionPreference = 'Stop'

function Get-Hash([string]$Path) {
    if (-not (Test-Path -LiteralPath $Path -PathType Leaf)) { return $null }
    return (Get-FileHash -LiteralPath $Path -Algorithm SHA256).Hash
}
function Save-Text([string]$Path, [string]$Value) {
    New-Item -ItemType Directory -Path (Split-Path -Parent $Path) -Force | Out-Null
    [IO.File]::WriteAllText($Path,$Value,(New-Object System.Text.UTF8Encoding($false)))
}
function Assert-Within([string]$Root,[string]$Candidate) {
    $rootPath = [IO.Path]::GetFullPath($Root).TrimEnd('\','/') + [IO.Path]::DirectorySeparatorChar
    $itemPath = [IO.Path]::GetFullPath($Candidate)
    if (-not $itemPath.StartsWith($rootPath,[StringComparison]::OrdinalIgnoreCase)) {
        throw 'Invalid relative path in rollback manifest; no files changed.'
    }
}
$homePath = if ($CodexHome) { [IO.Path]::GetFullPath($CodexHome) } elseif ($env:CODEX_HOME) { [IO.Path]::GetFullPath($env:CODEX_HOME) } else { Join-Path $HOME '.codex' }
$statePath = Join-Path $homePath 'harness-install/state.json'
if (-not (Test-Path -LiteralPath $statePath -PathType Leaf)) { throw 'No existe un manifiesto de instalación.' }
$rawState = [IO.File]::ReadAllText($statePath)
$state = $rawState | ConvertFrom-Json
if ($null -eq $state.files -or -not $state.backup_dir) { throw 'Manifiesto incompleto; se rechazó rollback.' }

$backup = [string]$state.backup_dir
Assert-Within (Join-Path $homePath 'backups/harness') (Join-Path $backup 'fixture')
if (-not (Test-Path -LiteralPath $backup -PathType Container)) { throw 'Falta carpeta de backup; sin modificaciones.' }

# P1-2: validate the previous manifest before restoring ANY targets.
$previous = Join-Path $backup '_previous_state.json'
if ($state.previous_state_existed) {
    if (-not (Test-Path -LiteralPath $previous -PathType Leaf)) {
        throw 'Falta _previous_state.json. Rollback cancelado SIN modificar archivos.'
    }
    $previousText = [IO.File]::ReadAllText($previous)
    $previousState = $previousText | ConvertFrom-Json
    if ($null -eq $previousState.files) { throw 'Manifiesto anterior inválido; no se realizó rollback.' }
}

$restore = New-Object 'System.Collections.Generic.List[object]'
foreach ($item in @($state.files)) {
    if (-not $item.path) { throw 'Ruta faltante en manifiesto; no se realizó rollback.' }
    $target = Join-Path $homePath ([string]$item.path)
    Assert-Within $homePath $target
    $currentHash = Get-Hash $target
    if ($currentHash -ne $item.installed_sha256) {
        throw ('Archivo administrado modificado o ausente: ' + $item.path + '. Sin cambios.')
    }
    $saved = Join-Path $backup ([string]$item.path)
    Assert-Within $backup $saved
    if ($item.existed) {
        if (-not (Test-Path -LiteralPath $saved -PathType Leaf)) {
            throw ('Falta backup: ' + $item.path + '. Rollback cancelado SIN modificar archivos.')
        }
        if ($item.PSObject.Properties.Name -contains 'backup_sha256') {
            if ((Get-Hash $saved) -ne $item.backup_sha256) {
                throw ('Hash de backup incorrecto: ' + $item.path + '. Sin cambios.')
            }
        }
    }
    $restore.Add(@{path=[string]$item.path;target=$target;saved=$saved;existed=[bool]$item.existed;current=$currentHash})
}

Write-Output ('Rollback preflight: PASS; ' + $restore.Count + ' archivos afectados.')
Write-Output ('Backup: ' + $backup)
if (-not $Apply) {
    Write-Output 'DRY-RUN. Usá -Apply para restaurar la instalación anterior.'
    return
}

# For an unexpected I/O failure during rollback, preserve the current state for compensation.
$staging = Join-Path ([IO.Path]::GetTempPath()) ('harness-rollback-' + [guid]::NewGuid().ToString('N'))
$mutating = $false
try {
    New-Item -ItemType Directory -Path $staging -Force | Out-Null
    foreach ($item in $restore) {
        if ($null -ne $item.current) {
            $snapshot = Join-Path $staging $item.path
            New-Item -ItemType Directory -Path (Split-Path -Parent $snapshot) -Force | Out-Null
            Copy-Item -LiteralPath $item.target -Destination $snapshot
            if ((Get-Hash $snapshot) -ne $item.current) { throw ('Snapshot inválido: ' + $item.path) }
        }
    }
    $mutating = $true
    foreach ($item in $restore) {
        if ($item.existed) {
            New-Item -ItemType Directory -Path (Split-Path -Parent $item.target) -Force | Out-Null
            Copy-Item -LiteralPath $item.saved -Destination $item.target -Force
        } elseif (Test-Path -LiteralPath $item.target) {
            Remove-Item -LiteralPath $item.target -Force
        }
    }
    if ($state.previous_state_existed) {
        Copy-Item -LiteralPath $previous -Destination $statePath -Force
    } else {
        Remove-Item -LiteralPath $statePath -Force
    }
    Write-Output 'ROLLBACK OK. Respaldo preservado. Reiniciá Codex.'
} catch {
    $cause = $_.Exception.Message
    $problems = New-Object 'System.Collections.Generic.List[string]'
    if ($mutating) {
        foreach ($item in $restore) {
            try {
                $snapshot = Join-Path $staging $item.path
                if ($null -eq $item.current) {
                    if (Test-Path -LiteralPath $item.target) { Remove-Item -LiteralPath $item.target -Force }
                } else {
                    if ((Get-Hash $snapshot) -ne $item.current) { throw 'Snapshot de recuperación no coincide' }
                    Copy-Item -LiteralPath $snapshot -Destination $item.target -Force
                }
            } catch {
                $problems.Add(($item.path + ': ' + $_.Exception.Message))
            }
        }
        try { Save-Text $statePath $rawState } catch { $problems.Add(('state.json: ' + $_.Exception.Message)) }
    }
    if ($problems.Count -gt 0) {
        throw ('ROLLBACK FALLÓ: ' + $cause + '. Recuperación incompleta: ' + ($problems -join '; ') + '. Guardá los backups.')
    }
    throw ('ROLLBACK ABORTADO: ' + $cause + '. Estado anterior preservado o recuperado.')
} finally {
    Remove-Item -LiteralPath $staging -Recurse -Force -ErrorAction SilentlyContinue
}
