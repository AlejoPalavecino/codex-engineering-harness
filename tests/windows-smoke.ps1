#Requires -Version 5.1
Set-StrictMode -Version Latest
$ErrorActionPreference = 'Stop'
$repo = [IO.Path]::GetFullPath((Join-Path $PSScriptRoot '..'))
$homePath = Join-Path ([IO.Path]::GetTempPath()) ('codex-harness-smoke-' + [guid]::NewGuid().ToString('N'))
function Assert-That([bool]$ok,[string]$message) {
    if (-not $ok) { throw $message }
}
try {
    New-Item -ItemType Directory -Path $homePath -Force | Out-Null
    $originalConfig = '[mcp_servers.engram]' + [Environment]::NewLine + 'command = "engram"' + [Environment]::NewLine
    [IO.File]::WriteAllText((Join-Path $homePath 'config.toml'),$originalConfig)
    [IO.File]::WriteAllText((Join-Path $homePath 'AGENTS.md'),'anterior')
    New-Item -ItemType Directory -Path (Join-Path $homePath 'agents') | Out-Null
    [IO.File]::WriteAllText((Join-Path $homePath 'agents/other_agent.toml'),'custom = true')
    $before = (Get-FileHash -LiteralPath (Join-Path $homePath 'config.toml') -Algorithm SHA256).Hash
    & (Join-Path $repo 'scripts/install.ps1') -CodexHome $homePath
    Assert-That (((Get-FileHash -LiteralPath (Join-Path $homePath 'config.toml') -Algorithm SHA256).Hash) -eq $before) 'Dry-run cambio config.'
    & (Join-Path $repo 'scripts/install.ps1') -CodexHome $homePath -Apply -ConfigureModels
    Assert-That (Test-Path (Join-Path $homePath 'agents/product_designer.toml')) 'Falta agente.'
    Assert-That (Test-Path (Join-Path $homePath 'skills/harness-discovery/SKILL.md')) 'Falta skill.'
    Assert-That (Test-Path (Join-Path $homePath 'agents/other_agent.toml')) 'Se borro agente ajeno.'
    $config = [IO.File]::ReadAllText((Join-Path $homePath 'config.toml'))
    Assert-That ($config.Contains('[mcp_servers.engram]')) 'Se perdio el MCP.'
    Assert-That ($config.Contains('gpt-6.1-sol')) 'No se configuro el modelo.'
    Assert-That ($config.Contains('[tools.update_plan]')) 'Falta configuracion update_plan.'
    Assert-That ($config.Contains('enabled = true')) 'update_plan no se habilito.'
    Assert-That (Test-Path (Join-Path $homePath 'harness/bin/taskboard.py')) 'Falta validador de tareas.'
    $installedState = (Get-Content -LiteralPath (Join-Path $homePath 'harness-install/state.json') -Raw | ConvertFrom-Json)
    $expectedVersion = [IO.File]::ReadAllText((Join-Path $repo 'VERSION')).Trim()
    Assert-That ($installedState.version -eq $expectedVersion) 'El instalador no registró VERSION correctamente.'
    & (Join-Path $repo 'scripts/verify.ps1') -CodexHome $homePath
    & (Join-Path $repo 'scripts/rollback.ps1') -CodexHome $homePath -Apply
    Assert-That (([IO.File]::ReadAllText((Join-Path $homePath 'AGENTS.md'))) -eq 'anterior') 'Rollback no restauro AGENTS.'
    Assert-That (([IO.File]::ReadAllText((Join-Path $homePath 'config.toml'))) -eq $originalConfig) 'Rollback no restauro config.'
    Assert-That (Test-Path (Join-Path $homePath 'agents/other_agent.toml')) 'Rollback perdio agente ajeno.'
    Assert-That (-not (Test-Path (Join-Path $homePath 'agents/product_designer.toml'))) 'Rollback no retiro agente nuevo.'
    Assert-That (-not (Test-Path (Join-Path $homePath 'harness/bin/taskboard.py'))) 'Rollback no retiro validador nuevo.'
    & (Join-Path $repo 'scripts/install.ps1') -CodexHome $homePath -Apply -RemoveLegacyAgents
    & (Join-Path $repo 'scripts/install.ps1') -CodexHome $homePath -Apply
    $repeatConfig = [IO.File]::ReadAllText((Join-Path $homePath 'config.toml'))
    Assert-That (([regex]::Matches($repeatConfig,'(?m)^\[tools\.update_plan\]')).Count -eq 1) 'Se duplico tools.update_plan.'
    & (Join-Path $repo 'scripts/rollback.ps1') -CodexHome $homePath -Apply
    Assert-That (Test-Path (Join-Path $homePath 'agents/solution_architect.toml')) 'Rollback de actualizacion no restauro version previa.'
    Write-Output 'SMOKE TESTS OK'
} finally {
    if (Test-Path -LiteralPath $homePath) {
        Remove-Item -LiteralPath $homePath -Recurse -Force
    }
}
