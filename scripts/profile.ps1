#Requires -Version 5.1
if (-not (Get-Variable -Name ChuliPythonRuntime -Scope Script -ErrorAction SilentlyContinue)) {
    $script:ChuliPythonRuntime = $null
}
function Get-Python {
    if ($script:ChuliPythonRuntime) { return $script:ChuliPythonRuntime }
    foreach ($candidate in @('python', 'python3', 'py')) {
        if (-not (Get-Command $candidate -ErrorAction SilentlyContinue)) { continue }
        $prefix = @()
        if ($candidate -eq 'py') { $prefix = @('-3.11') }
        try {
            & $candidate @prefix -c 'import tomllib; import sys; sys.exit(0 if sys.version_info >= (3, 11) else 1)' 2>$null
            if ($LASTEXITCODE -eq 0) {
                $script:ChuliPythonRuntime = @{ command = $candidate; prefix = @($prefix) }
                return $script:ChuliPythonRuntime
            }
        } catch { }
    }
    throw 'Python 3.11+ requerido para validar TOML antes de modificar Codex. No se realizaron cambios.'
}
function Invoke-ProfilePython {
    param($Runtime, [string[]]$Arguments)
    $command = [string]$Runtime.command
    $allArgs = @($Runtime.prefix) + $Arguments
    $output = & $command @allArgs 2>&1
    if ($LASTEXITCODE -ne 0) {
        throw ('Preflight failed: ' + (($output | Out-String).Trim()))
    }
    return $output
}

function Get-ProfileInventory($Runtime, [string]$Action, [string]$ProfileHome, [switch]$AdoptGentleRouting, [string]$Client = 'codex') {
    $tool = Join-Path $PSScriptRoot 'profile_resources.py'
    $arguments = @($tool,$Action,'--home',$ProfileHome,'--client',$Client)
    if ($AdoptGentleRouting) { $arguments += '--adopt-gentle-routing' }
    $json = Invoke-ProfilePython $Runtime $arguments
    return ($json | ConvertFrom-Json)
}

function Get-ClientInstructionPath([string]$Client = 'codex') {
    if ($Client -eq 'claude') { return 'CLAUDE.md' }
    if ($Client -eq 'agy') { return 'plugins/chuli/rules/AGENTS.md' }
    return 'AGENTS.md'
}

function Get-ClientConfigPath([string]$Client = 'codex') {
    if ($Client -in @('claude','agy')) { return 'settings.json' }
    return 'config.toml'
}

function New-PrivateDirectory($Runtime, [string]$Path) {
    # 0700 for staging and backups on POSIX. Windows inherits the user's directory ACL.
    if ($env:OS -eq 'Windows_NT') {
        New-Item -ItemType Directory -Path $Path -Force | Out-Null
    } else {
        $null = Invoke-ProfilePython $Runtime @('-c','import pathlib,sys; pathlib.Path(sys.argv[1]).mkdir(mode=448,parents=True)', $Path)
    }
}

function Resolve-ProfileHome([string]$Requested, [string]$Client = 'codex') {
    if ($Requested) {
        $value = $Requested
    } elseif ($Client -eq 'claude') {
        $value = if ($env:CLAUDE_CONFIG_DIR) { $env:CLAUDE_CONFIG_DIR } else { Join-Path $HOME '.claude' }
    } elseif ($Client -eq 'agy') {
        $value = if ($env:AGY_HOME) { $env:AGY_HOME } else { Join-Path $HOME '.gemini/config' }
    } else {
        $value = if ($env:CODEX_HOME) { $env:CODEX_HOME } else { Join-Path $HOME '.codex' }
    }
    if ($value -match '(^|[\\/])\.\.([\\/]|$)') { throw 'CODEX_HOME contains path traversal.' }
    return [IO.Path]::GetFullPath($value)
}
