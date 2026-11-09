$ErrorActionPreference = 'Stop'
. (Join-Path $PSScriptRoot '../scripts/install-manager.ps1')

$managerTestRoot = Join-Path ([IO.Path]::GetTempPath()) ('chuli-command-test-' + [guid]::NewGuid().ToString('N'))
$managerOriginalPath = $env:Path
New-Item -ItemType Directory -Path $managerTestRoot | Out-Null
try {
    Add-Type -AssemblyName System.IO.Compression.FileSystem
    $managerPackageDir = Join-Path $managerTestRoot 'package'
    New-Item -ItemType Directory -Path $managerPackageDir | Out-Null
    [IO.File]::WriteAllText((Join-Path $managerPackageDir 'chuli.exe'),'MZtest-payload-one')
    $managerArchive = Join-Path $managerTestRoot 'manager.zip'
    [IO.Compression.ZipFile]::CreateFromDirectory($managerPackageDir,$managerArchive)
    $managerDigest = (Get-FileHash -LiteralPath $managerArchive -Algorithm SHA256).Hash.ToLowerInvariant()
    $managerDestination = Join-Path $managerTestRoot 'installed'
    $managerArguments = @{Archive=$managerArchive;Sha256=$managerDigest;ReleaseVersion='1.1.0';Destination=$managerDestination;PathTarget='Process'}
    Install-ChuliManagerPackage @managerArguments | Out-Null
    if (Test-Path -LiteralPath $managerDestination) { throw 'Preview created persistent files.' }
    Install-ChuliManagerPackage @managerArguments -Apply | Out-Null
    if (-not (Test-Path -LiteralPath (Join-Path $managerDestination 'chuli.exe'))) { throw 'Binary missing.' }
    if ($env:Path -ne ($managerOriginalPath.TrimEnd(';') + ';' + $managerDestination)) { throw 'Existing PATH changed.' }
    $managerFirstPath = $env:Path
    Install-ChuliManagerPackage @managerArguments -Apply | Out-Null
    if ($env:Path -ne $managerFirstPath) { throw 'Repeated installation duplicated PATH.' }
    $managerBad = @{} + $managerArguments
    $managerBad.Sha256 = '0' * 64
    $managerFailed = $false
    try { Install-ChuliManagerPackage @managerBad -Apply | Out-Null } catch { $managerFailed = $true }
    if (-not $managerFailed) { throw 'Checksum mismatch accepted.' }
    [IO.File]::WriteAllText((Join-Path $managerDestination 'chuli.exe'),'custom-user-binary')
    $managerFailed = $false
    try { Install-ChuliManagerPackage @managerArguments -Apply | Out-Null } catch { $managerFailed = $true }
    if (-not $managerFailed -or [IO.File]::ReadAllText((Join-Path $managerDestination 'chuli.exe')) -ne 'custom-user-binary') { throw 'Customized binary overwritten.' }
    $managerUnmanaged = @{} + $managerArguments
    $managerUnmanaged.Destination = Join-Path $managerTestRoot 'unmanaged'
    New-Item -ItemType Directory -Path $managerUnmanaged.Destination | Out-Null
    [IO.File]::WriteAllText((Join-Path $managerUnmanaged.Destination 'chuli.exe'),'unmanaged')
    $managerFailed = $false
    try { Install-ChuliManagerPackage @managerUnmanaged -Apply | Out-Null } catch { $managerFailed = $true }
    if (-not $managerFailed) { throw 'Unmanaged binary overwritten.' }
    $managerExtra = Join-Path $managerPackageDir 'unexpected.txt'
    [IO.File]::WriteAllText($managerExtra,'unexpected')
    $managerBadArchive = Join-Path $managerTestRoot 'extra.zip'
    [IO.Compression.ZipFile]::CreateFromDirectory($managerPackageDir,$managerBadArchive)
    $managerBad.Archive = $managerBadArchive
    $managerBad.Sha256 = (Get-FileHash -LiteralPath $managerBadArchive -Algorithm SHA256).Hash.ToLowerInvariant()
    $managerFailed = $false
    try { Install-ChuliManagerPackage @managerBad -Apply | Out-Null } catch { $managerFailed = $true }
    if (-not $managerFailed) { throw 'Unexpected archive files accepted.' }
    $managerBadLocation = @{} + $managerArguments
    $managerBadLocation.Destination = Join-Path $managerTestRoot 'invalid;path'
    $managerFailed = $false
    try { Install-ChuliManagerPackage @managerBadLocation -Apply | Out-Null } catch { $managerFailed = $true }
    if (-not $managerFailed -or (Test-Path -LiteralPath $managerBadLocation.Destination)) { throw 'PATH separator in destination accepted.' }
    foreach ($managerRelative in @('C:relative-install','\relative-install','relative-install')) {
        $managerBadLocation.Destination = $managerRelative
        $managerFailed = $false
        try { Install-ChuliManagerPackage @managerBadLocation | Out-Null } catch { $managerFailed = $true }
        if (-not $managerFailed) { throw 'Partially relative destination accepted.' }
    }
    $managerRollback = @{} + $managerArguments
    $managerRollback.Destination = Join-Path $managerTestRoot 'rollback'
    Install-ChuliManagerPackage @managerRollback -Apply | Out-Null
    $managerBeforeExe = [IO.File]::ReadAllText((Join-Path $managerRollback.Destination 'chuli.exe'))
    $managerBeforeRecord = [IO.File]::ReadAllText((Join-Path $managerRollback.Destination 'manager-install.json'))
    $managerUpgradeDir = Join-Path $managerTestRoot 'upgrade-package'
    New-Item -ItemType Directory -Path $managerUpgradeDir | Out-Null
    [IO.File]::WriteAllText((Join-Path $managerUpgradeDir 'chuli.exe'),'MZtest-payload-two')
    $managerUpgradeArchive = Join-Path $managerTestRoot 'upgrade.zip'
    [IO.Compression.ZipFile]::CreateFromDirectory($managerUpgradeDir,$managerUpgradeArchive)
    $managerRollback.Archive = $managerUpgradeArchive
    $managerRollback.Sha256 = (Get-FileHash -LiteralPath $managerUpgradeArchive -Algorithm SHA256).Hash.ToLowerInvariant()
    $managerRollback.ReleaseVersion = '1.1.1'
    $env:Path = $managerOriginalPath
    $managerSavedSetter = (Get-Item Function:Set-ChuliPath).ScriptBlock
    function Set-ChuliPath { param($Value,$Target); throw 'Simulated PATH failure' }
    try {
        $managerFailed = $false
        try { Install-ChuliManagerPackage @managerRollback -Apply | Out-Null } catch { $managerFailed = $true }
        if (-not $managerFailed) { throw 'PATH failure was ignored.' }
    } finally { Set-Item Function:Set-ChuliPath $managerSavedSetter }
    if ([IO.File]::ReadAllText((Join-Path $managerRollback.Destination 'chuli.exe')) -cne $managerBeforeExe -or
        [IO.File]::ReadAllText((Join-Path $managerRollback.Destination 'manager-install.json')) -cne $managerBeforeRecord -or
        $env:Path -cne $managerOriginalPath) { throw 'Failed installation did not restore previous state.' }
    $managerRawPath = 'C:\Existing;%LOCALAPPDATA%\Programs\Chuli'
    if ((Get-ChuliPathPlan $managerRawPath (Join-Path $env:LOCALAPPDATA 'Programs/Chuli')) -cne $managerRawPath) { throw 'Expanded PATH entry duplicated.' }
    Write-Output 'OK: preview, install, idempotent PATH, checksums, archive, preservation and rollback after PATH failure.'
} finally {
    $env:Path = $managerOriginalPath
    $managerResolved = [IO.Path]::GetFullPath($managerTestRoot)
    $managerTemp = [IO.Path]::GetFullPath([IO.Path]::GetTempPath())
    if (-not $managerResolved.StartsWith($managerTemp,[StringComparison]::OrdinalIgnoreCase) -or (Split-Path $managerResolved -Leaf) -notlike 'chuli-command-test-*') { throw 'Cleanup outside temporary workspace.' }
    Remove-Item -LiteralPath $managerResolved -Recurse -Force
}
