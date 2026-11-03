#Requires -Version 5.1
Set-StrictMode -Version Latest
$ErrorActionPreference = 'Stop'
$repo = [IO.Path]::GetFullPath((Join-Path $PSScriptRoot '..'))
$homePath = Join-Path ([IO.Path]::GetTempPath()) ('codex-harness-regression-' + [guid]::NewGuid().ToString('N'))
function Assert-That([bool]$condition, [string]$message) {
    if (-not $condition) { throw $message }
}
function Hash([string]$path) {
    return (Get-FileHash -LiteralPath $path -Algorithm SHA256).Hash
}
$install = Join-Path $repo 'scripts/install.ps1'
$rollback = Join-Path $repo 'scripts/rollback.ps1'
$verify = Join-Path $repo 'scripts/verify.ps1'
try {
    New-Item -ItemType Directory -Path $homePath -Force | Out-Null
    $configFile = Join-Path $homePath 'config.toml'
    $agentFile = Join-Path $homePath 'AGENTS.md'
    $nl = [Environment]::NewLine
    $original = '[tools.update_plan] # valid comment' + $nl +
        'enabled = false' + $nl +
        '[mcp_servers.engram]' + $nl +
        'command = "engram"' + $nl
    [IO.File]::WriteAllText($configFile, $original)
    [IO.File]::WriteAllText($agentFile, 'CES argentino personalizado')
    $originalAgentHash = Hash $agentFile

    # P1-1: a commented section header must not generate duplicate tables.
    & $install -CodexHome $homePath -Apply
    $config = [IO.File]::ReadAllText($configFile)
    Assert-That (([regex]::Matches($config,'(?m)^\[tools\.update_plan\]')).Count -eq 1) 'Duplicated TOML table.'
    Assert-That ($config.Contains('enabled = false')) 'Explicit local override was lost.'
    & python -c 'import tomllib,sys; tomllib.load(open(sys.argv[1],"rb"))' $configFile
    Assert-That ($LASTEXITCODE -eq 0) 'Candidate TOML invalid.'
    Assert-That ((Hash $agentFile) -eq $originalAgentHash) 'P1-3: update overwrote customized AGENTS.'

    # Valid unrelated config changes must be accepted.
    [IO.File]::AppendAllText($configFile, '[custom]' + $nl + "owner = 'fixture'" + $nl)
    & $verify -CodexHome $homePath
    Assert-That ($LASTEXITCODE -eq 0) 'Unrelated valid TOML change should not fail verification.'

    # Explicit replacement updates managed values and backs up personalizations.
    & $install -CodexHome $homePath -Apply -ReplaceCustomized
    Assert-That ((Hash $agentFile) -ne $originalAgentHash) 'Explicit replacement ignored.'
    $config = [IO.File]::ReadAllText($configFile)
    Assert-That (([regex]::Matches($config,'(?m)^\[tools\.update_plan\]')).Count -eq 1) 'Replacement duplicated section.'
    Assert-That ($config.Contains('enabled = true')) 'Explicit replacement did not enable plan.'
    Assert-That ($config.Contains('[custom]')) 'Unrelated config lost.'
    & $verify -CodexHome $homePath
    Assert-That ($LASTEXITCODE -eq 0) 'Verification failed on valid installation.'

    # P1-2: missing previous state must abort before restoring anything.
    $statePath = Join-Path $homePath 'harness-install/state.json'
    $state = Get-Content $statePath -Raw | ConvertFrom-Json
    Assert-That ([bool]$state.previous_state_existed) 'Expected previous installation.'
    $previousFile = Join-Path $state.backup_dir '_previous_state.json'
    $previousContents = [IO.File]::ReadAllText($previousFile)
    Remove-Item -LiteralPath $previousFile -Force
    $agentBefore = Hash $agentFile
    $configBefore = Hash $configFile
    $stateBefore = Hash $statePath
    & pwsh -NoProfile -File $rollback -CodexHome $homePath -Apply *> $null
    Assert-That ($LASTEXITCODE -ne 0) 'Incomplete backup should have prevented rollback.'
    Assert-That ((Hash $agentFile) -eq $agentBefore) 'Rollback changed AGENTS before preflight completed.'
    Assert-That ((Hash $configFile) -eq $configBefore) 'Rollback changed config before preflight completed.'
    Assert-That ((Hash $statePath) -eq $stateBefore) 'Rollback changed state before preflight completed.'
    [IO.File]::WriteAllText($previousFile,$previousContents)
    & $rollback -CodexHome $homePath -Apply
    Assert-That ((Hash $agentFile) -eq $originalAgentHash) 'Rollback failed to restore customization.'

    # Invalid TOML must fail validation even if file hash is accepted.
    $invalid = '[tools.update_plan]' + $nl + 'enabled = true' + $nl +
        '[tools.update_plan]' + $nl + 'enabled = false'
    [IO.File]::WriteAllText($configFile,$invalid)
    & pwsh -NoProfile -File $verify -CodexHome $homePath *> $null
    Assert-That ($LASTEXITCODE -ne 0) 'Invalid TOML was not rejected.'
    Write-Output 'SAFETY REGRESSION TESTS OK'
} finally {
    Remove-Item -LiteralPath $homePath -Recurse -Force -ErrorAction SilentlyContinue
}
