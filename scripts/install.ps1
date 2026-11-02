#Requires -Version 5.1
<#
.SYNOPSIS
  Instala o actualiza los archivos administrados por Codex Engineering Harness.
.DESCRIPTION
  Sin -Apply es un dry-run. Respalda antes de modificar, preserva MCP,
  plugins, hooks, memorias y agentes no administrados.
#>
[CmdletBinding()]
param(
    [switch]$Apply,
    [switch]$ConfigureModels,
    [switch]$RemoveLegacyAgents,
    [string]$CodexHome
)

Set-StrictMode -Version Latest
$ErrorActionPreference = 'Stop'

function Resolve-Home {
    param([string]$Explicit)
    if ($Explicit) { return [IO.Path]::GetFullPath($Explicit) }
    if ($env:CODEX_HOME) { return [IO.Path]::GetFullPath($env:CODEX_HOME) }
    return Join-Path $HOME '.codex'
}
function Read-Text {
    param([string]$Path)
    if (-not (Test-Path -LiteralPath $Path -PathType Leaf)) { return '' }
    return [IO.File]::ReadAllText($Path)
}
function Save-Text {
    param([string]$Path, [string]$Content)
    $parent = Split-Path -Parent $Path
    if (-not (Test-Path -LiteralPath $parent)) {
        New-Item -ItemType Directory -Path $parent -Force | Out-Null
    }
    $encoding = New-Object System.Text.UTF8Encoding($false)
    [IO.File]::WriteAllText($Path, $Content, $encoding)
}
function Set-TomlValue {
    param([string]$Text, [string]$Section, [string]$Key, [string]$Value)
    $lines = New-Object 'System.Collections.Generic.List[string]'
    foreach ($line in [regex]::Split($Text, '\r?\n')) { [void]$lines.Add($line) }
    $current = ''
    $sectionStart = -1
    $foundKey = $false
    $keyPattern = '^\s*' + [regex]::Escape($Key) + '\s*='
    for ($i = 0; $i -lt $lines.Count; $i++) {
        if ($lines[$i] -match '^\s*\[\[?([^\]]+)\]\]?\s*$') {
            $current = $Matches[1]
            if ($current -eq $Section) { $sectionStart = $i }
        }
        if ($current -eq $Section -and $lines[$i] -match $keyPattern) {
            $lines[$i] = $Key + ' = ' + $Value
            $foundKey = $true
            break
        }
    }
    if (-not $foundKey) {
        $newLine = $Key + ' = ' + $Value
        if (-not $Section) {
            $lines.Insert(0, $newLine)
        } elseif ($sectionStart -ge 0) {
            $lines.Insert($sectionStart + 1, $newLine)
        } else {
            [void]$lines.Add('')
            [void]$lines.Add('[' + $Section + ']')
            [void]$lines.Add($newLine)
        }
    }
    return [string]::Join([Environment]::NewLine, $lines).TrimEnd() + [Environment]::NewLine
}
function Get-Hash {
    param([string]$Path)
    if (-not (Test-Path -LiteralPath $Path -PathType Leaf)) { return $null }
    return (Get-FileHash -LiteralPath $Path -Algorithm SHA256).Hash
}

$homePath = Resolve-Home $CodexHome
$repoRoot = [IO.Path]::GetFullPath((Join-Path $PSScriptRoot '..'))
$configPath = Join-Path $homePath 'config.toml'
$statePath = Join-Path $homePath 'harness-install/state.json'
$managed = New-Object 'System.Collections.Generic.List[object]'
$managed.Add(@{ relative='AGENTS.md'; source='config/AGENTS.md'; operation='copy' })
foreach ($name in @('solution_architect','product_designer','test_engineer','quality_reviewer')) {
    $managed.Add(@{ relative=('agents/' + $name + '.toml'); source=('agents/' + $name + '.toml'); operation='copy' })
}
foreach ($name in @('harness-discovery','harness-architecture','harness-ux-ui','harness-specification','harness-tdd-qa')) {
    $managed.Add(@{ relative=('skills/' + $name + '/SKILL.md'); source=('skills/' + $name + '/SKILL.md'); operation='copy' })
}
if ($RemoveLegacyAgents) {
    foreach ($name in @('architect','explorer','reviewer','slice_worker','test_runner')) {
        $managed.Add(@{ relative=('agents/' + $name + '.toml'); source=''; operation='remove' })
    }
}
$managed.Add(@{ relative='config.toml'; source=''; operation='config' })

foreach ($item in $managed) {
    if ($item.operation -ne 'copy') { continue }
    $src = Join-Path $repoRoot $item.source
    if (-not (Test-Path -LiteralPath $src -PathType Leaf)) { throw "Falta archivo del harness: $($item.source)" }
}

$config = Read-Text $configPath
$roleDescriptions = [ordered]@{
    solution_architect = 'Evalua arquitectura, stacks y decisiones tecnicas'
    product_designer = 'Disena experiencias UX y UI'
    test_engineer = 'Prepara pruebas y criterios de aceptacion'
    quality_reviewer = 'Revisa resultados con evidencia independiente'
}
foreach ($name in $roleDescriptions.Keys) {
    $config = Set-TomlValue $config ('agents.' + $name) 'description' ('"' + $roleDescriptions[$name] + '"')
    $config = Set-TomlValue $config ('agents.' + $name) 'config_file' ('"./agents/' + $name + '.toml"')
}
if ($ConfigureModels) {
    $config = Set-TomlValue $config '' 'model' '"gpt-6.1-sol"'
    $config = Set-TomlValue $config '' 'model_reasoning_effort' '"medium"'
    $config = Set-TomlValue $config 'agents' 'default_subagent_model' '"gpt-6.1-sol"'
}
Write-Output ("Harness: directorio de Codex " + $homePath)
Write-Output ("Archivos administrados: " + $managed.Count)
Write-Output ("Configurar modelos principales: " + [bool]$ConfigureModels)
Write-Output ("Retirar cinco agentes heredados: " + [bool]$RemoveLegacyAgents)
Write-Output 'MCP, hooks, plugins, datos Engram y otros agentes permanecen intactos.'
if (-not $Apply) {
    Write-Output 'DRY-RUN: no se modifico nada. Para aplicar, usa -Apply.'
    return
}
$id = (Get-Date -Format 'yyyyMMdd-HHmmss') + '-' + [guid]::NewGuid().ToString('N').Substring(0,8)
$backup = Join-Path $homePath ('backups/harness/' + $id)
$records = New-Object 'System.Collections.Generic.List[object]'
$previousStateExisted = Test-Path -LiteralPath $statePath -PathType Leaf
$mutationsStarted = $false
try {
    New-Item -ItemType Directory -Path $backup -Force | Out-Null
    if ($previousStateExisted) {
        Copy-Item -LiteralPath $statePath -Destination (Join-Path $backup '_previous_state.json')
    }
    foreach ($item in $managed) {
        $relative = [string]$item.relative
        $target = Join-Path $homePath $relative
        $wasPresent = Test-Path -LiteralPath $target -PathType Leaf
        if ($wasPresent) {
            $saved = Join-Path $backup $relative
            $parent = Split-Path -Parent $saved
            New-Item -ItemType Directory -Path $parent -Force | Out-Null
            Copy-Item -LiteralPath $target -Destination $saved
        }
        $records.Add([ordered]@{
            path = $relative
            existed = [bool]$wasPresent
            operation = [string]$item.operation
            installed_sha256 = $null
        })
    }
    $mutationsStarted = $true
    for ($i = 0; $i -lt $managed.Count; $i++) {
        $item = $managed[$i]
        $target = Join-Path $homePath $item.relative
        switch ($item.operation) {
            'copy' { Save-Text $target (Read-Text (Join-Path $repoRoot $item.source)) }
            'config' { Save-Text $target $config }
            'remove' {
                if (Test-Path -LiteralPath $target -PathType Leaf) {
                    Remove-Item -LiteralPath $target -Force
                }
            }
        }
        $records[$i].installed_sha256 = Get-Hash $target
    }
    $info = [ordered]@{
        version='0.1.0-rc.1'
        installed_at=(Get-Date).ToString('o')
        backup_dir=$backup
        previous_state_existed=[bool]$previousStateExisted
        files=@($records.ToArray())
    }
    Save-Text $statePath (ConvertTo-Json $info -Depth 10)
    Write-Output 'INSTALACION OK. Reiniciar Codex y ejecutar scripts/verify.ps1.'
    Write-Output ('Backup: ' + $backup)
} catch {
    $failure = $_.Exception.Message
    if ($mutationsStarted) {
        foreach ($record in $records) {
            $target = Join-Path $homePath $record.path
            if ($record.existed) {
                $saved = Join-Path $backup $record.path
                Save-Text $target (Read-Text $saved)
            } elseif (Test-Path -LiteralPath $target -PathType Leaf) {
                Remove-Item -LiteralPath $target -Force
            }
        }
    }
    if ($previousStateExisted) {
        $oldState = Join-Path $backup '_previous_state.json'
        if (Test-Path -LiteralPath $oldState) { Save-Text $statePath (Read-Text $oldState) }
    } elseif (Test-Path -LiteralPath $statePath) {
        Remove-Item -LiteralPath $statePath -Force
    }
    throw "Instalacion fallida y recuperacion local intentada: $failure"
}
