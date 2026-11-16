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
    [switch]$AdoptGentleRouting,
    [string]$CodexHome,
    [ValidateSet('codex','claude','agy')][string]$Client = 'codex',
    [string]$Preset = '',
    [string]$RoleModelsJson = '',
    [string]$DisabledSkillsFile = ''
)
Set-StrictMode -Version Latest
$ErrorActionPreference = 'Stop'

function Get-Hash([string]$Path) {
    return Get-ChuliFileSha256 $Path
}
function Save-Text([string]$Path, [string]$Value) {
    $parent = Split-Path -Parent $Path
    New-Item -ItemType Directory -Force -Path $parent | Out-Null
    [IO.File]::WriteAllText($Path, $Value, (New-Object System.Text.UTF8Encoding($false)))
}

. (Join-Path $PSScriptRoot 'profile.ps1')
$homePath = Resolve-ProfileHome $CodexHome $Client
$instrRel = Get-ClientInstructionPath $Client
$configRel = Get-ClientConfigPath $Client
$root = [IO.Path]::GetFullPath((Join-Path $PSScriptRoot '..'))
$harnessVersion = [IO.File]::ReadAllText((Join-Path $root 'VERSION')).Trim()
if ($harnessVersion -notmatch '^[0-9]+\.[0-9]+\.[0-9]+$') {
    throw 'VERSION inválido; instalación cancelada antes de modificar Codex.'
}
$statePath = Join-Path $homePath 'harness-install/state.json'
$configPath = Join-Path $homePath $configRel
$legacyBaselinePath = Join-Path $root 'scripts/legacy-managed-hashes.json'
if (-not (Test-Path -LiteralPath $legacyBaselinePath -PathType Leaf)) {
    throw 'Falta scripts/legacy-managed-hashes.json; instalación cancelada antes de modificar Codex.'
}
$legacyBaselines = Get-Content -LiteralPath $legacyBaselinePath -Raw | ConvertFrom-Json
$python = Get-Python
$previous = $null
$managedFiles = @{}
if (Test-Path -LiteralPath $statePath -PathType Leaf) {
    $previous = Get-Content -LiteralPath $statePath -Encoding UTF8 -Raw | ConvertFrom-Json
    if ($null -eq $previous.files) { throw 'Manifiesto anterior sin lista de archivos. No se modificó Codex.' }
    if ($previous.PSObject.Properties.Name -contains 'managed_files' -and $null -ne $previous.managed_files) {
        foreach ($entry in $previous.managed_files.PSObject.Properties) {
            if ([string]$entry.Value -match '^[A-Fa-f0-9]{64}$') {
                $managedFiles[$entry.Name] = [string]$entry.Value
            }
        }
    }
    foreach ($record in @($previous.files)) {
        if ($record.path -and $record.path -notin @('config.toml','settings.json') -and $record.installed_sha256 -and -not $managedFiles.ContainsKey([string]$record.path)) {
            $managedFiles[[string]$record.path] = [string]$record.installed_sha256
        }
    }
    # Recover official baselines from releases that predate managed_files.
    foreach ($versionEntry in $legacyBaselines.PSObject.Properties) {
        if ($versionEntry.Name -ne [string]$previous.version) { continue }
        foreach ($entry in $versionEntry.Value.PSObject.Properties) {
            if (-not $managedFiles.ContainsKey($entry.Name)) {
                $managedFiles[$entry.Name] = [string]$entry.Value
            }
        }
    }
    # config.toml / settings.json is tracked separately through managed_config.
    $managedFiles.Remove('config.toml') | Out-Null
    $managedFiles.Remove('settings.json') | Out-Null
}
$staging = Join-Path ([IO.Path]::GetTempPath()) ('harness-stage-' + [guid]::NewGuid().ToString('N'))
$preservedFiles = @{}
$changes = New-Object 'System.Collections.Generic.List[object]'
$preserved = New-Object 'System.Collections.Generic.List[string]'
$modified = New-Object 'System.Collections.Generic.List[object]'
$applied = $false
$cliApplied = $false
$backup = $null

try {
    New-PrivateDirectory $python $staging
    $disabledSkillsList = @()
    $effectiveDisabledFile = ''
    if ($DisabledSkillsFile) {
        $rawDisabled = [IO.File]::ReadAllText($DisabledSkillsFile)
        if ($rawDisabled.Trim()) {
            $parsedDisabled = $rawDisabled | ConvertFrom-Json
            $disabledSkillsList = @($parsedDisabled | ForEach-Object { [string]$_ })
        }
        $effectiveDisabledFile = $DisabledSkillsFile
    } elseif ($null -ne $previous -and ($previous.PSObject.Properties.Name -contains 'disabled_skills') -and $null -ne $previous.disabled_skills) {
        $disabledSkillsList = @($previous.disabled_skills | ForEach-Object { [string]$_ })
        if ($disabledSkillsList.Count -gt 0) {
            $effectiveDisabledFile = Join-Path $staging 'disabled-skills.json'
            Save-Text $effectiveDisabledFile (ConvertTo-Json -InputObject @($disabledSkillsList) -Compress)
        }
    }
    $inventory = Get-ProfileInventory $python 'install' $homePath -AdoptGentleRouting:$AdoptGentleRouting -Client $Client -DisabledSkillsFile $effectiveDisabledFile
    $fullInventory = if ($disabledSkillsList.Count -gt 0) {
        Get-ProfileInventory $python 'install' $homePath -AdoptGentleRouting:$AdoptGentleRouting -Client $Client
    } else {
        $inventory
    }
    $roleModelsFile = $null
    if ($RoleModelsJson) {
        $roleModelsFile = Join-Path $staging 'role-models.json'
        Save-Text $roleModelsFile $RoleModelsJson
    }
    $items = New-Object 'System.Collections.Generic.List[object]'
    $instructionsPath = Join-Path $staging 'AGENTS.md'
    $instructionArgs = @((Join-Path $PSScriptRoot 'profile_resources.py'), 'instructions', '--home', $homePath, '--output', $instructionsPath, '--client', $Client)
    if ($Preset) { $instructionArgs += @('--preset', $Preset) }
    if ($roleModelsFile) { $instructionArgs += @('--role-models-file', $roleModelsFile) }
    if ($ReplaceCustomized) { $instructionArgs += '--replace' }
    if ($AdoptGentleRouting) { $instructionArgs += '--adopt-gentle-routing' }
    $instructions = (Invoke-ProfilePython $python $instructionArgs) | ConvertFrom-Json
    $stagedMap = @{}
    $orchModel = 'gpt-6.1-sol'
    $orchEffort = 'medium'
    $subModel = 'gpt-6.1-sol'
    $resolvedPreset = 'codex-balanced'
    $resolvedRoleModels = [ordered]@{
        orchestrator       = @{ model = 'gpt-6.1-sol';   effort = 'medium' }
        chuli_explorer     = @{ model = 'gpt-6-luna';    effort = 'medium' }
        chuli_worker       = @{ model = 'gpt-6.1-sol';   effort = 'medium' }
        solution_architect = @{ model = 'gpt-6.1-sol';   effort = 'high'   }
        product_designer   = @{ model = 'gpt-6.1-terra'; effort = 'medium' }
        test_engineer      = @{ model = 'gpt-6.1-sol';   effort = 'medium' }
        quality_reviewer   = @{ model = 'gpt-6.1-sol';   effort = 'high'   }
    }
    if ($Client -ne 'codex' -or $Preset -or $RoleModelsJson) {
        $stagedAgentsDir = Join-Path $staging 'staged-agents'
        $stageArgs = @((Join-Path $PSScriptRoot 'client_targets.py'), 'stage-agents', '--client', $Client, '--output-dir', $stagedAgentsDir)
        if ($Preset) { $stageArgs += @('--preset', $Preset) }
        if ($roleModelsFile) { $stageArgs += @('--role-models-file', $roleModelsFile) }
        $stageInfo = (Invoke-ProfilePython $python $stageArgs) | ConvertFrom-Json
        foreach ($prop in $stageInfo.staged.PSObject.Properties) { $stagedMap[$prop.Name] = [string]$prop.Value }
        $orchModel = [string]$stageInfo.role_models.orchestrator.model
        $orchEffort = [string]$stageInfo.role_models.orchestrator.effort
        $subModel = [string]$stageInfo.role_models.chuli_worker.model
        $resolvedPreset = [string]$stageInfo.preset
        $resolvedRoleModels = $stageInfo.role_models
    }
    foreach ($entry in $inventory.PSObject.Properties) {
        $items.Add(@{path=$entry.Name; source=[string]$entry.Value; kind='copy'})
    }
    if ($disabledSkillsList.Count -gt 0) {
        $activeNames = @{}
        foreach ($entry in $inventory.PSObject.Properties) { $activeNames[$entry.Name] = $true }
        foreach ($entry in $fullInventory.PSObject.Properties) {
            if (-not $activeNames.ContainsKey($entry.Name) -and $managedFiles.ContainsKey($entry.Name)) {
                $items.Add(@{path=$entry.Name; source=''; kind='remove'})
            }
        }
    }
    if ($RemoveLegacyAgents -and $Client -eq 'codex') {
        foreach ($name in @('architect','explorer','reviewer','slice_worker','test_runner')) {
            $items.Add(@{path=('agents/' + $name + '.toml'); source=''; kind='remove'})
        }
    }

    foreach ($item in $items) {
        $target = Join-Path $homePath $item.path
        $oldHash = Get-Hash $target
        $source = $null
        $sourceHash = $null
        if ($item.kind -eq 'copy') {
            $source = Join-Path $root $item.source
            if (-not (Test-Path -LiteralPath $source -PathType Leaf)) { throw ('Falta archivo del repositorio: ' + $item.source) }
            if ($item.path -eq $instrRel) { $source = $instructionsPath }
            elseif ($stagedMap.ContainsKey([string]$item.path)) { $source = $stagedMap[[string]$item.path] }
            $sourceHash = Get-Hash $source
            $knownManagedHash = $null
            if ($managedFiles.ContainsKey([string]$item.path)) {
                $knownManagedHash = [string]$managedFiles[[string]$item.path]
            }
            $customModelUpdate = ($Preset -ne '' -or $RoleModelsJson -ne '') -and $stagedMap.ContainsKey([string]$item.path)
            if ($item.path -eq $instrRel -and $instructions.preserved -and @($instructions.adopted_gentle).Count -eq 0) {
                $preserved.Add(($instrRel + ' (customized Chuli block)'))
                $action = 'PRESERVE-CUSTOMIZED'
            } elseif ($sourceHash -eq $oldHash -and ($knownManagedHash -or $item.path -eq $instrRel -or $ReplaceCustomized -or $customModelUpdate)) {
                $action = 'UNCHANGED'
            } elseif ($null -eq $oldHash) {
                $action = 'CREATE'
            } elseif ($item.path -eq $instrRel -or $knownManagedHash -eq $oldHash -or ($customModelUpdate -and $knownManagedHash)) {
                $action = 'UPDATE'
            } elseif ($ReplaceCustomized) {
                $action = 'REPLACE-APPROVED'
            } else {
                $action = 'PRESERVE-CUSTOMIZED'
                $preserved.Add([string]$item.path)
            }
            if ($action -eq 'PRESERVE-CUSTOMIZED') {
                $preservedFiles[[string]$item.path] = $oldHash
            }
            if ($action -in @('UNCHANGED','CREATE','UPDATE','REPLACE-APPROVED')) {
                $managedFiles[[string]$item.path] = $sourceHash
            }
        } else {
            $action = if ($null -ne $oldHash) { 'REMOVE-APPROVED' } else { 'UNCHANGED' }
            if ($item.kind -eq 'remove') {
                $managedFiles.Remove([string]$item.path) | Out-Null
            }
        }
        $changes.Add([ordered]@{path=[string]$item.path; kind=[string]$item.kind; action=$action; original=$oldHash; source=$source; source_hash=$sourceHash})
    }

    $inputPath = Join-Path $staging ('original-' + $configRel)
    $candidatePath = Join-Path $staging ('candidate-' + $configRel)
    $reportPath = Join-Path $staging 'config-report.json'
    if (Test-Path -LiteralPath $configPath -PathType Leaf) {
        Copy-Item -LiteralPath $configPath -Destination $inputPath
    } else {
        Save-Text $inputPath ''
    }
    $wantModelUpdate = [bool]($ConfigureModels -or $Preset -or $RoleModelsJson)
    if ($Client -eq 'codex') {
        $args = @((Join-Path $root 'scripts/toml_config.py'), 'prepare', '--input', $inputPath,
            '--output', $candidatePath, '--report', $reportPath,
            '--orchestrator-model', $orchModel, '--orchestrator-effort', $orchEffort, '--subagent-model', $subModel)
        if ($wantModelUpdate) { $args += '--configure-models' }
        if ($SkipTaskProgress) { $args += '--skip-task-progress' }
        if ($ReplaceCustomized) { $args += '--replace-managed' }
        $reportOutput = Invoke-ProfilePython $python $args
    } else {
        $args = @((Join-Path $root 'scripts/client_targets.py'), 'json-config', 'prepare', '--client', $Client,
            '--input', $inputPath, '--output', $candidatePath, '--report', $reportPath)
        if ($Preset) { $args += @('--preset', $Preset) }
        if ($roleModelsFile) { $args += @('--role-models-file', $roleModelsFile) }
        if ($wantModelUpdate) { $args += '--configure-models' }
        if ($ReplaceCustomized) { $args += '--replace-managed' }
        $reportOutput = Invoke-ProfilePython $python $args
    }
    $report = Get-Content -LiteralPath $reportPath -Raw | ConvertFrom-Json
    foreach ($key in @($report.preserved)) { $preserved.Add(($configRel + ' -> ' + $key)) }
    $oldConfigHash = Get-Hash $configPath
    $newConfigHash = Get-Hash $candidatePath
    $configAction = if ($oldConfigHash -eq $newConfigHash) { 'UNCHANGED' } elseif ($null -eq $oldConfigHash) { 'CREATE' } else { 'UPDATE-VALIDATED' }
    $changes.Add([ordered]@{path=$configRel;kind='config';action=$configAction;original=$oldConfigHash;source=$candidatePath})

    $cliSettingsPath = Resolve-AgyCliSettingsPath $homePath $Client
    $cliInputPath = Join-Path $staging 'original-agy-cli-settings.json'
    $cliCandidatePath = Join-Path $staging 'candidate-agy-cli-settings.json'
    $cliReportPath = Join-Path $staging 'agy-cli-config-report.json'
    $cliReport = $null
    $oldCliHash = $null
    $newCliHash = $null
    $cliConfigAction = $null
    if ($null -ne $cliSettingsPath) {
        Assert-ChuliSafeRegularTarget $cliSettingsPath
        if (Test-Path -LiteralPath $cliSettingsPath -PathType Leaf) {
            Copy-Item -LiteralPath $cliSettingsPath -Destination $cliInputPath
        } else {
            Save-Text $cliInputPath ''
        }
        $cliArgs = @((Join-Path $root 'scripts/client_targets.py'), 'json-config', 'prepare', '--client', 'agy',
            '--input', $cliInputPath, '--output', $cliCandidatePath, '--report', $cliReportPath)
        if ($Preset) { $cliArgs += @('--preset', $Preset) }
        if ($roleModelsFile) { $cliArgs += @('--role-models-file', $roleModelsFile) }
        if ($wantModelUpdate) { $cliArgs += '--configure-models' }
        if ($ReplaceCustomized) { $cliArgs += '--replace-managed' }
        $null = Invoke-ProfilePython $python $cliArgs
        $cliReport = Get-Content -LiteralPath $cliReportPath -Raw | ConvertFrom-Json
        foreach ($key in @($cliReport.preserved)) { $preserved.Add(('antigravity-cli/settings.json -> ' + $key)) }
        $oldCliHash = Get-Hash $cliSettingsPath
        $newCliHash = Get-Hash $cliCandidatePath
        $cliConfigAction = if ($oldCliHash -eq $newCliHash) { 'UNCHANGED' } elseif ($null -eq $oldCliHash) { 'CREATE' } else { 'UPDATE-VALIDATED' }
    }

    Write-Output ('Codex home: ' + $homePath)
    Write-Output ('Client: ' + $Client + ' (preset: ' + $resolvedPreset + ')')
    Write-Output ('Orchestrator: Chuli, ' + (Join-Path $homePath $instrRel))
    if (@($instructions.adopted_gentle).Count -gt 0) {
        Write-Output ('Gentle AI -> Chuli: ' + ($instructions.adopted_gentle -join ', ') + ' in ' + $instrRel + '. Original included in backup on apply.')
    }
    Write-Output ('Version candidate: ' + $harnessVersion)
    Write-Output ('Python TOML validation: PASS')
    Write-Output ('ReplaceCustomized requested: ' + [bool]$ReplaceCustomized)
    foreach ($entry in $changes) {
        Write-Output ('  ' + $entry.action.PadRight(20) + ' ' + $entry.path)
    }
    if ($null -ne $cliConfigAction) {
        Write-Output ('  ' + $cliConfigAction.PadRight(20) + ' ../antigravity-cli/settings.json')
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
    New-PrivateDirectory $python $backup
    if ($null -ne $previous) {
        Copy-Item -LiteralPath $statePath -Destination (Join-Path $backup '_previous_state.json')
    }
    $agyCliStatePath = Join-Path $homePath 'harness-install/agy-cli-settings.json'
    $previousAgyCliStateExisted = Test-Path -LiteralPath $agyCliStatePath -PathType Leaf
    if ($previousAgyCliStateExisted) {
        Copy-Item -LiteralPath $agyCliStatePath -Destination (Join-Path $backup '_previous_agy_cli_settings.json')
    }
    $cliBackupPath = Join-Path $backup '_agy_cli_settings.json'
    $cliExisted = [bool]($null -ne $oldCliHash)
    $cliChanged = ($null -ne $cliSettingsPath -and $cliConfigAction -ne 'UNCHANGED')
    if ($cliChanged) {
        if ((Get-Hash $cliSettingsPath) -ne $oldCliHash) {
            throw 'antigravity-cli/settings.json cambió durante el preflight'
        }
        if ($cliExisted) {
            Copy-Item -LiteralPath $cliSettingsPath -Destination $cliBackupPath
            if ((Get-Hash $cliBackupPath) -ne $oldCliHash) {
                throw 'Backup corrupto antes de aplicar: antigravity-cli/settings.json'
            }
        }
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
            New-Item -ItemType Directory -Path (Split-Path -Parent $target) -Force | Out-Null
            if ([IO.File]::Exists($target)) {
                # In-place write preserves the user's existing permissions.
                [IO.File]::WriteAllBytes($target, [IO.File]::ReadAllBytes($entry.source))
            } else {
                [IO.File]::Copy($entry.source, $target)
            }
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
        client=$Client
        preset=$resolvedPreset
        role_models=$resolvedRoleModels
        schema_version=2
        rollback_schema_version=2
        installed_at=(Get-Date).ToString('o')
        backup_dir=$backup
        previous_state_existed=[bool]($null -ne $previous)
        previous_state_sha256=(Get-Hash (Join-Path $backup '_previous_state.json'))
        adopted_gentle=@($instructions.adopted_gentle)
        orchestrator_sha256=$instructions.orchestrator_sha256
        orchestrator_baseline_sha256=$instructions.baseline_sha256
        files=@($modified.ToArray())
        managed_files=$managedFiles
        preserved_files=$preservedFiles
        preserved_customizations=@($preserved.ToArray())
        managed_config=$tracked
        disabled_skills=@($disabledSkillsList)
    }
    Save-Text $statePath (ConvertTo-Json -InputObject $state -Depth 12)
    if ($null -ne $cliSettingsPath) {
        if ($cliChanged) {
            if ([IO.File]::Exists($cliSettingsPath)) {
                [IO.File]::WriteAllBytes($cliSettingsPath, [IO.File]::ReadAllBytes($cliCandidatePath))
            } else {
                [IO.File]::Copy($cliCandidatePath, $cliSettingsPath)
            }
            $cliApplied = $true
        }
        $cliTracked = @{}
        foreach ($property in $cliReport.tracked.PSObject.Properties) { $cliTracked[$property.Name] = $property.Value }
        $cliMeta = [ordered]@{
            target=$cliSettingsPath
            existed=$cliExisted
            changed=$cliChanged
            backup_sha256=$oldCliHash
            installed_sha256=(Get-Hash $cliSettingsPath)
            previous_state_existed=$previousAgyCliStateExisted
            managed_config=$cliTracked
        }
        Save-Text $agyCliStatePath (ConvertTo-Json -InputObject $cliMeta -Depth 8)
    }
    Write-Output 'INSTALACION OK. Respaldo creado y configuración validada.'
    Write-Output ('Backup: ' + $backup)
    Write-Output 'Ejecutá chuli doctor y abrí una sesión de Codex con este perfil.'
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
    if ($cliApplied) {
        try {
            if ($cliExisted) {
                if ((Get-Hash $cliBackupPath) -ne $oldCliHash) { throw 'Backup no coincide con SHA esperado' }
                Copy-Item -LiteralPath $cliBackupPath -Destination $cliSettingsPath -Force
            } elseif (Test-Path -LiteralPath $cliSettingsPath) {
                Remove-Item -LiteralPath $cliSettingsPath -Force
            }
            $oldCliState = Join-Path $backup '_previous_agy_cli_settings.json'
            if ($previousAgyCliStateExisted -and (Test-Path -LiteralPath $oldCliState -PathType Leaf)) {
                Copy-Item -LiteralPath $oldCliState -Destination $agyCliStatePath -Force
            } elseif (-not $previousAgyCliStateExisted -and (Test-Path -LiteralPath $agyCliStatePath -PathType Leaf)) {
                Remove-Item -LiteralPath $agyCliStatePath -Force
            }
        } catch { $recoveryErrors.Add(('antigravity-cli/settings.json: ' + $_.Exception.Message)) }
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
