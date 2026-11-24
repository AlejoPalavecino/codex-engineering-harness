#Requires -Version 5.1
Set-StrictMode -Version Latest
$ErrorActionPreference = 'Stop'
$repo = [IO.Path]::GetFullPath((Join-Path $PSScriptRoot '..'))
$sandbox = Join-Path ([IO.Path]::GetTempPath()) ('chuli-multiclient-' + [guid]::NewGuid().ToString('N'))

function Assert-That([bool]$ok, [string]$message) {
    if (-not $ok) { throw $message }
}

$install = Join-Path $repo 'scripts/install.ps1'
$verify = Join-Path $repo 'scripts/verify.ps1'
$rollback = Join-Path $repo 'scripts/rollback.ps1'

try {
    New-Item -ItemType Directory -Path $sandbox -Force | Out-Null

    # 1. Claude Code installation with preset, custom role override, verify and rollback
    $claudeHome = Join-Path $sandbox 'claude'
    New-Item -ItemType Directory -Path $claudeHome -Force | Out-Null
    $claudeMd = "<!-- gentle-ai:engram-protocol -->`nmemoria`n<!-- /gentle-ai:engram-protocol -->`n"
    $claudeSettings = '{"outputStyle":"concise","permissions":{"allow":["Bash(git status)"]}}'
    [IO.File]::WriteAllText((Join-Path $claudeHome 'CLAUDE.md'), $claudeMd)
    [IO.File]::WriteAllText((Join-Path $claudeHome 'settings.json'), $claudeSettings)

    & $install -CodexHome $claudeHome -Client 'claude' -Preset 'claude-balanced' -RoleModelsJson '{"orchestrator":{"model":"claude-fable-5-2","effort":"high"}}' -Apply | Out-Null
    Assert-That (Test-Path (Join-Path $claudeHome 'agents/solution_architect.md')) 'Missing Claude agent markdown.'
    Assert-That (Test-Path (Join-Path $claudeHome 'skills/harness-discovery/SKILL.md')) 'Missing Claude skill.'
    $installedClaudeMd = [IO.File]::ReadAllText((Join-Path $claudeHome 'CLAUDE.md'))
    Assert-That ($installedClaudeMd.Contains('<!-- gentle-ai:engram-protocol -->') -and $installedClaudeMd.Contains('<!-- chuli:orchestrator -->')) 'Claude instructions merge failed.'
    $installedClaudeSettings = [IO.File]::ReadAllText((Join-Path $claudeHome 'settings.json')) | ConvertFrom-Json
    Assert-That ($installedClaudeSettings.model -eq 'claude-fable-5-2') 'Claude orchestrator model override not applied.'
    Assert-That ($installedClaudeSettings.outputStyle -eq 'concise') 'Claude unrelated setting lost.'
    $claudeVerify = (& $verify -CodexHome $claudeHome -Client 'claude' | Out-String)
    Assert-That ($LASTEXITCODE -eq 0) 'Claude verify failed with an active outputStyle.'
    Assert-That ($claudeVerify.Contains('WARNING: outputStyle "concise"')) 'Claude verify did not warn about the active outputStyle.'
    Assert-That ($claudeVerify -match 'advertencias \(deriva no bloqueante\): [1-9]') 'Claude verify did not count the outputStyle warning.'
    & $rollback -CodexHome $claudeHome -Client 'claude' -Apply | Out-Null
    Assert-That ([IO.File]::ReadAllText((Join-Path $claudeHome 'CLAUDE.md')) -eq $claudeMd) 'Claude rollback did not restore CLAUDE.md.'
    Assert-That (-not (Test-Path (Join-Path $claudeHome 'agents/solution_architect.md'))) 'Claude rollback did not remove installed agent.'

    # 2. AGY CLI installation with preset, verify and rollback
    $agyHome = Join-Path $sandbox 'agy'
    New-Item -ItemType Directory -Path $agyHome -Force | Out-Null
    $agySettings = '{"colorScheme":"dark","toolPermission":"ask"}'
    [IO.File]::WriteAllText((Join-Path $agyHome 'settings.json'), $agySettings)

    & $install -CodexHome $agyHome -Client 'agy' -Preset 'agy-balanced' -Apply | Out-Null
    Assert-That (Test-Path (Join-Path $agyHome 'plugins/chuli/plugin.json')) 'Missing AGY plugin.json.'
    Assert-That (Test-Path (Join-Path $agyHome 'plugins/chuli/rules/AGENTS.md')) 'Missing AGY plugin rules.'
    Assert-That (Test-Path (Join-Path $agyHome 'plugins/chuli/agents/solution_architect.md')) 'Missing AGY subagent.'
    $installedAgySettings = [IO.File]::ReadAllText((Join-Path $agyHome 'settings.json')) | ConvertFrom-Json
    Assert-That ($installedAgySettings.model -eq 'Orchestrator (Argon)') 'AGY model not set to Orchestrator (Argon).'
    Assert-That ($installedAgySettings.colorScheme -eq 'dark') 'AGY unrelated setting lost.'
    & $verify -CodexHome $agyHome -Client 'agy' | Out-Null
    & $rollback -CodexHome $agyHome -Client 'agy' -Apply | Out-Null
    Assert-That (-not (Test-Path (Join-Path $agyHome 'plugins/chuli/plugin.json'))) 'AGY rollback did not remove plugin.json.'
    Assert-That ([IO.File]::ReadAllText((Join-Path $agyHome 'settings.json')) -eq $agySettings) 'AGY rollback did not restore settings.json.'

    # 2b. AGY real layout (.gemini/config + .gemini/antigravity-cli/settings.json) and valid agent tiers
    $geminiRoot = Join-Path $sandbox '.gemini'
    $agyConfigHome = Join-Path $geminiRoot 'config'
    $agyCliDir = Join-Path $geminiRoot 'antigravity-cli'
    New-Item -ItemType Directory -Path $agyConfigHome -Force | Out-Null
    New-Item -ItemType Directory -Path $agyCliDir -Force | Out-Null
    $agyCliOriginal = '{"colorScheme":"tokyo night","trustedWorkspaces":["C:\\work"]}'
    $agyCliSettings = Join-Path $agyCliDir 'settings.json'
    [IO.File]::WriteAllText($agyCliSettings, $agyCliOriginal)
    & $install -CodexHome $agyConfigHome -Client 'agy' -Preset 'agy-balanced' -Apply | Out-Null
    $updatedAgyCli = [IO.File]::ReadAllText($agyCliSettings) | ConvertFrom-Json
    Assert-That ($updatedAgyCli.model -eq 'Orchestrator (Argon)') 'AGY CLI settings.json model not updated.'
    Assert-That ($updatedAgyCli.colorScheme -eq 'tokyo night') 'AGY CLI colorScheme lost.'
    $archAgent = [IO.File]::ReadAllText((Join-Path $agyConfigHome 'plugins/chuli/agents/solution_architect.md'))
    Assert-That ($archAgent.Contains('model: pro')) 'AGY agent tier must be valid (pro).'
    & $verify -CodexHome $agyConfigHome -Client 'agy' | Out-Null
    if (Get-Command agy -ErrorAction SilentlyContinue) {
        $agyJson = (& agy plugin validate (Join-Path $agyConfigHome 'plugins/chuli') --json | Out-String) | ConvertFrom-Json
        Assert-That ($agyJson.components.agents -eq 6) 'agy plugin validate did not load all 6 Chuli agents.'
        $agyErrors = @($agyJson.findings | Where-Object { $_.severity -eq 'error' })
        Assert-That ($agyErrors.Count -eq 0) ('agy plugin validate returned errors: ' + ($agyErrors | ConvertTo-Json -Compress))
    }
    & $rollback -CodexHome $agyConfigHome -Client 'agy' -Apply | Out-Null
    Assert-That ([IO.File]::ReadAllText($agyCliSettings) -eq $agyCliOriginal) 'AGY rollback did not restore antigravity-cli/settings.json.'

    # 3. Codex installation with preset and custom role model
    $codexHome = Join-Path $sandbox 'codex'
    New-Item -ItemType Directory -Path $codexHome -Force | Out-Null
    & $install -CodexHome $codexHome -Client 'codex' -Preset 'codex-astra-max' -Apply | Out-Null
    $codexConfig = [IO.File]::ReadAllText((Join-Path $codexHome 'config.toml'))
    Assert-That ($codexConfig.Contains('model = "gpt-6-astra"')) 'Codex preset model not applied to config.toml.'
    $explorerToml = [IO.File]::ReadAllText((Join-Path $codexHome 'agents/chuli_explorer.toml'))
    Assert-That ($explorerToml.Contains('model = "gpt-6.1-sol"')) 'Codex preset model not applied to chuli_explorer.toml.'
    & $verify -CodexHome $codexHome -Client 'codex' | Out-Null
    & $rollback -CodexHome $codexHome -Client 'codex' -Apply | Out-Null
    Assert-That (-not (Test-Path (Join-Path $codexHome 'agents/chuli_explorer.toml'))) 'Codex rollback did not remove agent.'

    Write-Output 'MULTICLIENT INSTALL, VERIFY AND ROLLBACK TESTS OK'
} finally {
    if (Test-Path -LiteralPath $sandbox) {
        Remove-Item -LiteralPath $sandbox -Recurse -Force -ErrorAction SilentlyContinue
    }
}
