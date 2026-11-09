#Requires -Version 5.1
<#
.SYNOPSIS
Installs Chuli Manager for the current Windows user. Preview unless -Apply.
#>
[CmdletBinding()]
param([switch]$Apply, [string]$InstallDir, [string]$Version)

function Test-ManagerAbsolutePath([string]$Destination) {
    if ($Destination.StartsWith('\\?\') -or $Destination.StartsWith('\\.\')) { return $false }
    return ($Destination -match '^[A-Za-z]:[\\/]' -or $Destination -match '^\\\\[^\\/:*?"<>|]+\\[^\\/:*?"<>|]+(?:[\\/]|$)')
}

function Assert-ManagerLocation([string]$Destination) {
    if ($Destination -match '[;\x00-\x1f]') { throw 'El destino no puede contener separadores de PATH ni caracteres de control.' }
    if (-not (Test-ManagerAbsolutePath $Destination) -or [IO.Path]::GetPathRoot($Destination).TrimEnd('\','/') -eq $Destination.TrimEnd('\','/')) {
        throw 'El destino debe ser una carpeta absoluta, no la raiz de un disco.'
    }
    $managerCursor = $Destination
    while ($managerCursor) {
        if (Test-Path -LiteralPath $managerCursor) {
            $managerItem = Get-Item -LiteralPath $managerCursor -Force
            if (-not $managerItem.PSIsContainer -or ($managerItem.Attributes -band [IO.FileAttributes]::ReparsePoint)) {
                throw 'El destino debe usar carpetas regulares, sin enlaces ni junctions.'
            }
        }
        $managerCursor = [IO.Path]::GetDirectoryName($managerCursor)
    }
}

function Get-ChuliPath([string]$Target) {
    if ($Target -eq 'Process') { return [Environment]::GetEnvironmentVariable('Path','Process') }
    $managerKey = [Microsoft.Win32.Registry]::CurrentUser.OpenSubKey('Environment')
    if ($managerKey) {
        try { return $managerKey.GetValue('Path',$null,[Microsoft.Win32.RegistryValueOptions]::DoNotExpandEnvironmentNames) }
        finally { $managerKey.Dispose() }
    }
    return $null
}

function Assert-ManagerFiles([string[]]$Files) {
    foreach ($managerFile in $Files) {
        if (Test-Path -LiteralPath $managerFile) {
            $managerInfo = Get-Item -LiteralPath $managerFile -Force
            if ($managerInfo.PSIsContainer -or ($managerInfo.Attributes -band [IO.FileAttributes]::ReparsePoint)) { throw 'Archivos de instalacion no regulares.' }
        }
    }
}

function Set-ChuliPath([AllowNull()][object]$Value, [string]$Target) {
    if ($Target -eq 'Process') { [Environment]::SetEnvironmentVariable('Path',$Value,'Process'); return }
    if (-not ('ChuliManagerEnvironment' -as [type])) {
        Add-Type -TypeDefinition @'
using System;
using System.Runtime.InteropServices;
public static class ChuliManagerEnvironment {
    [DllImport("user32.dll", CharSet=CharSet.Unicode, SetLastError=true)]
    public static extern IntPtr SendMessageTimeout(IntPtr window, uint message,
        UIntPtr parameter, string value, uint flags, uint timeout, out UIntPtr result);
}
'@
    }
    $managerKey = [Microsoft.Win32.Registry]::CurrentUser.CreateSubKey('Environment')
    try {
        if ($null -eq $Value) { $managerKey.DeleteValue('Path',$false) }
        else {
            $managerKind = [Microsoft.Win32.RegistryValueKind]::ExpandString
            if ($managerKey.GetValueNames() -contains 'Path') { $managerKind = $managerKey.GetValueKind('Path') }
            if ($managerKind -notin @([Microsoft.Win32.RegistryValueKind]::String,[Microsoft.Win32.RegistryValueKind]::ExpandString)) { throw 'Tipo de PATH no compatible.' }
            $managerKey.SetValue('Path',[string]$Value,$managerKind)
        }
    } finally { $managerKey.Dispose() }
    $managerNotice = [UIntPtr]::Zero
    [ChuliManagerEnvironment]::SendMessageTimeout([IntPtr]65535,26,[UIntPtr]::Zero,'Environment',2,2000,[ref]$managerNotice) | Out-Null
}

function Get-ChuliPathPlan([string]$Current, [string]$Destination) {
    foreach ($managerEntry in ($Current -split ';')) {
        $managerExpanded = [Environment]::ExpandEnvironmentVariables($managerEntry.Trim().Trim('"')).TrimEnd('\','/')
        if ($managerExpanded -eq $Destination.TrimEnd('\','/')) { return $Current }
    }
    if (-not $Current -or $Current.EndsWith(';')) { return $Current + $Destination }
    return $Current + ';' + $Destination
}

function Install-ChuliManagerPackage {
    [CmdletBinding()]
    param([string]$Archive, [string]$Sha256, [string]$ReleaseVersion, [string]$Destination,
          [ValidateSet('User','Process')][string]$PathTarget = 'User', [switch]$Apply)
    $ErrorActionPreference = 'Stop'
    if ($ReleaseVersion -notmatch '^\d+\.\d+\.\d+$' -or $Sha256 -notmatch '^[a-fA-F0-9]{64}$') { throw 'Version o SHA-256 invalido.' }
    if (-not (Test-ManagerAbsolutePath $Destination)) { throw 'El destino debe ser absoluto.' }
    $Destination = [IO.Path]::GetFullPath($Destination).TrimEnd('\','/')
    Assert-ManagerLocation $Destination
    $managerExe = Join-Path $Destination 'chuli.exe'
    $managerRecord = Join-Path $Destination 'manager-install.json'
    Assert-ManagerFiles @($managerExe,$managerRecord)
    if ((Get-FileHash -LiteralPath $Archive -Algorithm SHA256).Hash -ne $Sha256) { throw 'SHA-256 incorrecto; no se modifico la instalacion.' }
    Add-Type -AssemblyName System.IO.Compression.FileSystem
    $managerZip = [IO.Compression.ZipFile]::OpenRead($Archive)
    try {
        if ($managerZip.Entries.Count -ne 1 -or $managerZip.Entries[0].FullName -cne 'chuli.exe' -or
            $managerZip.Entries[0].Length -lt 2 -or $managerZip.Entries[0].Length -gt 64MB -or
            (($managerZip.Entries[0].ExternalAttributes -shr 16) -band 61440) -eq 40960) { throw 'El ZIP debe contener solamente chuli.exe regular.' }
        $managerStream = $managerZip.Entries[0].Open()
        try { if ($managerStream.ReadByte() -ne 77 -or $managerStream.ReadByte() -ne 90) { throw 'El paquete no contiene un ejecutable Windows.' } }
        finally { $managerStream.Dispose() }
    } finally { $managerZip.Dispose() }
    $managerPrevious = $null
    $managerPreviousRecordHash = $null
    if (Test-Path -LiteralPath $managerExe) {
        if (-not (Test-Path -LiteralPath $managerRecord)) { throw 'Existe un chuli.exe ajeno al instalador; no se reemplazo.' }
        $managerPrevious = [IO.File]::ReadAllText($managerRecord) | ConvertFrom-Json
        $managerPreviousRecordHash = (Get-FileHash -LiteralPath $managerRecord -Algorithm SHA256).Hash
        if ($managerPrevious.schema -ne 1 -or $managerPrevious.version -notmatch '^\d+\.\d+\.\d+$' -or
            $managerPrevious.sha256 -ne (Get-FileHash -LiteralPath $managerExe -Algorithm SHA256).Hash) { throw 'La instalacion fue modificada; revisala antes de actualizar.' }
        if ([version]$ReleaseVersion -lt [version]$managerPrevious.version) { throw 'No se admite bajar de version con este instalador.' }
    } elseif (Test-Path -LiteralPath $managerRecord) { throw 'Registro sin ejecutable; requiere revision manual.' }
    $managerOldPath = Get-ChuliPath $PathTarget
    $managerNewPath = Get-ChuliPathPlan $managerOldPath $Destination
    Write-Output "Manager $ReleaseVersion -> $Destination"
    Write-Output "PATH $PathTarget`: agregar solo esta carpeta si falta; conservar entradas existentes."
    if (-not $Apply) { Write-Output 'Vista previa. Agrega -Apply para instalar.'; return }
    $managerHasher = [Security.Cryptography.SHA256]::Create()
    try { $managerLockName = [BitConverter]::ToString($managerHasher.ComputeHash([Text.Encoding]::UTF8.GetBytes($Destination.ToLowerInvariant()))).Replace('-','') }
    finally { $managerHasher.Dispose() }
    $managerLock = New-Object Threading.Mutex($false,('Local\ChuliManager-' + $managerLockName))
    $managerAcquired = $false
    $managerToken = [guid]::NewGuid().ToString('N')
    $managerNewExe = Join-Path $Destination ('.chuli-new-' + $managerToken + '.exe')
    $managerNewRecord = Join-Path $Destination ('.chuli-new-' + $managerToken + '.json')
    $managerBackupExe = Join-Path $Destination ('.chuli-backup-' + $managerToken + '.exe')
    $managerBackupRecord = Join-Path $Destination ('.chuli-backup-' + $managerToken + '.json')
    $managerWroteExe = $false; $managerWroteRecord = $false; $managerRestored = $true
    try {
        try { $managerAcquired = $managerLock.WaitOne(0) } catch [Threading.AbandonedMutexException] { $managerAcquired = $true; throw 'Instalacion interrumpida; revisa sus respaldos antes de reintentar.' }
        if (-not $managerAcquired) { throw 'Otro instalador esta trabajando.' }
        Assert-ManagerLocation $Destination
        Assert-ManagerFiles @($managerExe,$managerRecord)
        New-Item -ItemType Directory -Path $Destination -Force | Out-Null
        if ($managerPrevious -and (Get-FileHash -LiteralPath $managerRecord -Algorithm SHA256).Hash -ne $managerPreviousRecordHash) { throw 'El registro cambio; revisa otra vista previa.' }
        $managerZip = [IO.Compression.ZipFile]::OpenRead($Archive)
        try {
            if ((Get-FileHash -LiteralPath $Archive -Algorithm SHA256).Hash -ne $Sha256) { throw 'El paquete cambio antes de extraerlo.' }
            [IO.Compression.ZipFileExtensions]::ExtractToFile($managerZip.Entries[0],$managerNewExe,$false)
        }
        finally { $managerZip.Dispose() }
        $managerNewDigest = (Get-FileHash -LiteralPath $managerNewExe -Algorithm SHA256).Hash.ToLowerInvariant()
        $managerJson = @{schema=1;version=$ReleaseVersion;sha256=$managerNewDigest} | ConvertTo-Json
        [IO.File]::WriteAllText($managerNewRecord,$managerJson,(New-Object Text.UTF8Encoding($false)))
        if ($managerPrevious) {
            if ((Get-FileHash -LiteralPath $managerExe -Algorithm SHA256).Hash -ne $managerPrevious.sha256) { throw 'El ejecutable cambio; revisa otra vista previa.' }
            [IO.File]::Replace($managerNewExe,$managerExe,$managerBackupExe)
        } else { [IO.File]::Move($managerNewExe,$managerExe) }
        $managerWroteExe = $true
        if ($managerPrevious) { [IO.File]::Replace($managerNewRecord,$managerRecord,$managerBackupRecord) }
        else { [IO.File]::Move($managerNewRecord,$managerRecord) }
        $managerWroteRecord = $true
        if ((Get-ChuliPath $PathTarget) -cne $managerOldPath) { throw 'El PATH cambio durante la instalacion; no se sobrescribio.' }
        if ($managerNewPath -cne $managerOldPath) { Set-ChuliPath $managerNewPath $PathTarget }
        if ($PathTarget -eq 'User') { $env:Path = Get-ChuliPathPlan $env:Path $Destination }
        Write-Output 'Instalado. En una terminal nueva podes ejecutar: chuli'
        Write-Output 'Solo se instalo el Manager; el perfil de Codex se administra desde su TUI.'
    } catch {
        $managerFailure = $_
        try {
            if ($managerWroteRecord) {
                if ($managerPrevious) { [IO.File]::Replace($managerBackupRecord,$managerRecord,[NullString]::Value) }
                else { [IO.File]::Delete($managerRecord) }
            }
            if ($managerWroteExe) {
                if ($managerPrevious) { [IO.File]::Replace($managerBackupExe,$managerExe,[NullString]::Value) }
                else { [IO.File]::Delete($managerExe) }
            }
            if ($managerNewPath -cne $managerOldPath -and (Get-ChuliPath $PathTarget) -ceq $managerNewPath) {
                Set-ChuliPath $managerOldPath $PathTarget
            }
        } catch { $managerRestored = $false; Write-Warning 'No se pudo restaurar todo. Conserva los archivos .chuli-backup para recuperacion.' }
        throw $managerFailure
    } finally {
        $managerCleanup = @($managerNewExe,$managerNewRecord)
        if ($managerRestored) { $managerCleanup += @($managerBackupExe,$managerBackupRecord) }
        foreach ($managerTemporary in $managerCleanup) { if (Test-Path -LiteralPath $managerTemporary) { [IO.File]::Delete($managerTemporary) } }
        if ($managerAcquired) { $managerLock.ReleaseMutex() }
        $managerLock.Dispose()
    }
}

function Install-ChuliManager([switch]$Apply,[string]$InstallDir,[string]$Version) {
    $ErrorActionPreference = 'Stop'
    if ($env:OS -ne 'Windows_NT' -or -not [Environment]::Is64BitOperatingSystem -or $env:PROCESSOR_ARCHITECTURE -eq 'ARM64') { throw 'Este instalador requiere Windows x64.' }
    if (-not $InstallDir) { $InstallDir = Join-Path $env:LOCALAPPDATA 'Programs/Chuli' }
    if (-not (Test-ManagerAbsolutePath $InstallDir)) { throw 'El destino debe ser absoluto.' }
    $InstallDir = [IO.Path]::GetFullPath($InstallDir)
    Assert-ManagerLocation $InstallDir
    $managerExistingCommand = Get-Command chuli -ErrorAction SilentlyContinue
    if ($managerExistingCommand -and ($managerExistingCommand.CommandType -ne 'Application' -or $managerExistingCommand.Source -ne (Join-Path $InstallDir 'chuli.exe'))) {
        throw 'Ya existe otro comando chuli. Revisalo antes de instalar para no cambiar su significado.'
    }
    if ($Version -and $Version -notmatch '^\d+\.\d+\.\d+$') { throw 'Usa una version estable como 1.1.0.' }
    [Net.ServicePointManager]::SecurityProtocol = [Net.ServicePointManager]::SecurityProtocol -bor [Net.SecurityProtocolType]::Tls12
    $managerApi = 'https://api.github.com/repos/AlejoPalavecino/codex-engineering-harness/releases/'
    if ($Version) { $managerApi += 'tags/v' + $Version } else { $managerApi += 'latest' }
    $managerRelease = Invoke-RestMethod -Uri $managerApi -Headers @{'User-Agent'='Chuli-Manager-Installer'}
    if ($managerRelease.draft -or $managerRelease.prerelease -or $managerRelease.tag_name -notmatch '^v(\d+\.\d+\.\d+)$') { throw 'No se encontro una release estable valida.' }
    $Version = $Matches[1]
    if (-not $Apply) { Write-Output "Instalar Manager $Version en $InstallDir y agregarlo al PATH de usuario. Agrega -Apply para confirmar."; return }
    $managerDownload = Join-Path ([IO.Path]::GetTempPath()) ('chuli-download-' + [guid]::NewGuid().ToString('N'))
    New-Item -ItemType Directory -Path $managerDownload | Out-Null
    try {
        $managerName = 'chuli_v' + $Version + '_windows_amd64.zip'
        $managerBase = 'https://github.com/AlejoPalavecino/codex-engineering-harness/releases/download/v' + $Version + '/'
        $managerSumsFile = Join-Path $managerDownload 'SHA256SUMS.txt'
        Invoke-WebRequest -UseBasicParsing -Uri ($managerBase + 'SHA256SUMS.txt') -OutFile $managerSumsFile
        $managerSums = [IO.File]::ReadAllText($managerSumsFile)
        $managerExpected = @($managerSums -split "`n" | Where-Object { $_ -match ('^([a-fA-F0-9]{64})\s+' + [regex]::Escape($managerName) + '\s*$') })
        if ($managerExpected.Count -ne 1) { throw 'No se encontro un checksum unico para el Manager.' }
        $managerExpected[0] -match '^([a-fA-F0-9]{64})' | Out-Null
        $managerDigest = $Matches[1]
        $managerArchive = Join-Path $managerDownload $managerName
        Invoke-WebRequest -UseBasicParsing -Uri ($managerBase + $managerName) -OutFile $managerArchive
        Install-ChuliManagerPackage -Archive $managerArchive -Sha256 $managerDigest -ReleaseVersion $Version -Destination $InstallDir -Apply
    } finally {
        foreach ($managerFile in (Get-ChildItem -LiteralPath $managerDownload -File)) { [IO.File]::Delete($managerFile.FullName) }
        [IO.Directory]::Delete($managerDownload)
    }
}

if ($MyInvocation.InvocationName -ne '.') { Install-ChuliManager -Apply:$Apply -InstallDir $InstallDir -Version $Version }
