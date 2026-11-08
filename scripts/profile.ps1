#Requires -Version 5.1
function Get-Python {
    foreach ($candidate in @('python', 'python3', 'py')) {
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

function Get-ProfileInventory($Runtime, [string]$Action, [string]$ProfileHome) {
    $tool = Join-Path $PSScriptRoot 'profile_resources.py'
    $json = Invoke-ProfilePython $Runtime @($tool,$Action,'--home',$ProfileHome)
    return ($json | ConvertFrom-Json)
}

function New-PrivateDirectory($Runtime, [string]$Path) {
    # 0700 for staging and backups on POSIX. Windows inherits the user's directory ACL.
    $null = Invoke-ProfilePython $Runtime @('-c','import pathlib,sys; pathlib.Path(sys.argv[1]).mkdir(mode=448,parents=True)', $Path)
}

function Resolve-ProfileHome([string]$Requested) {
    $value = if ($Requested) { $Requested } elseif ($env:CODEX_HOME) { $env:CODEX_HOME } else { Join-Path $HOME '.codex' }
    if ($value -match '(^|[\\/])\.\.([\\/]|$)') { throw 'CODEX_HOME contains path traversal.' }
    return [IO.Path]::GetFullPath($value)
}
