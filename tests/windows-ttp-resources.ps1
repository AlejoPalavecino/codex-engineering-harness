#Requires -Version 5.1
# Integration contract: real installer/rollback, synthetic homes only; no network.
Set-StrictMode -Version Latest
$ErrorActionPreference = 'Stop'
$repo = [IO.Path]::GetFullPath((Join-Path $PSScriptRoot '..'))
$tempParent = [IO.Path]::GetFullPath([IO.Path]::GetTempPath())
$sandbox = Join-Path $tempParent ('codex-ttp-resources-' + [guid]::NewGuid().ToString('N'))
$engine = (Get-Process -Id $PID).Path
$resources = @(
    @{source='templates/feature-tasks.md'; target='harness/templates/feature-tasks.md'},
    @{source='docs/protocols/TTP.md'; target='harness/docs/protocols/TTP.md'},
    @{source='scripts/taskboard.py'; target='harness/bin/taskboard.py'}
)
$failures = New-Object 'System.Collections.Generic.List[string]'

function Assert-That([bool]$condition, [string]$message) {
    if (-not $condition) { throw $message }
}
function Write-Fixture([string]$path, [string]$text) {
    New-Item -ItemType Directory -Path (Split-Path -Parent $path) -Force | Out-Null
    [IO.File]::WriteAllText($path, $text, (New-Object Text.UTF8Encoding($false)))
}
function Hash([string]$path) {
    Assert-That (Test-Path -LiteralPath $path -PathType Leaf) ('Falta recurso: ' + $path)
    return (Get-FileHash -LiteralPath $path -Algorithm SHA256).Hash
}
function Snapshot([string]$directory) {
    if (-not (Test-Path -LiteralPath $directory)) { return '<absent>' }
    $base = [IO.Path]::GetFullPath($directory).TrimEnd('\','/')
    $entries = @(Get-ChildItem -LiteralPath $directory -Force -Recurse | Sort-Object FullName | ForEach-Object {
        $relative = $_.FullName.Substring($base.Length)
        if ($_.PSIsContainer) { 'D ' + $relative } else { 'F ' + $relative + ' ' + (Hash $_.FullName) }
    })
    return ($entries -join "`n")
}
function Run-Script([string]$name, [string]$homePath, [string[]]$options = @(), [bool]$expectSuccess = $true) {
    $previousPreference = $ErrorActionPreference
    try {
        # Windows PowerShell 5.1 wraps redirected native stderr as ErrorRecord.
        # Capture it before checking the child exit code, including expected failures.
        $ErrorActionPreference = 'Continue'
        $output = & $engine -NoProfile -File (Join-Path $repo ('scripts/' + $name + '.ps1')) -CodexHome $homePath @options 2>&1
        $exitCode = $LASTEXITCODE
    } finally {
        $ErrorActionPreference = $previousPreference
    }
    if ($expectSuccess) {
        Assert-That ($exitCode -eq 0) ($name + ' falló: ' + ($output -join "`n"))
    } else {
        Assert-That ($exitCode -ne 0) ($name + ' debía rechazar la operación')
    }
    return ($output -join "`n")
}
function Read-State([string]$homePath) {
    return (Get-Content -LiteralPath (Join-Path $homePath 'harness-install/state.json') -Raw | ConvertFrom-Json)
}
function Assert-Installed([string]$homePath) {
    $state = Read-State $homePath
    foreach ($resource in $resources) {
        $actual = Hash (Join-Path $homePath $resource.target)
        Assert-That ($actual -eq (Hash (Join-Path $repo $resource.source))) ('Copia no exacta: ' + $resource.target)
        $records = @($state.files | Where-Object { $_.path -eq $resource.target })
        Assert-That ($records.Count -eq 1) ('Recurso no administrado una única vez: ' + $resource.target)
        Assert-That ($records[0].installed_sha256 -eq $actual) ('Hash no registrado: ' + $resource.target)
    }
}
function Case([string]$name, [scriptblock]$body) {
    $homePath = Join-Path $sandbox $name
    try {
        & $body $homePath
        Write-Output ('PASS: ' + $name)
    } catch {
        $failures.Add(($name + ': ' + $_.Exception.Message))
        Write-Output ('FAIL: ' + $name + ': ' + $_.Exception.Message)
    }
}

try {
    New-Item -ItemType Directory -Path $sandbox | Out-Null
    Case 'dry-run' {
        param($homePath)
        $before = Snapshot $homePath
        $plan = Run-Script 'install' $homePath
        Assert-That ((Snapshot $homePath) -eq $before) 'Dry-run creó archivos o carpetas en home ausente.'
        foreach ($resource in $resources) {
            Assert-That ($plan.Contains($resource.target)) ('Dry-run omitió recurso del plan: ' + $resource.target)
        }
        Write-Fixture (Join-Path $homePath 'foreign.txt') 'archivo ajeno'
        Write-Fixture (Join-Path $homePath $resources[0].target) 'plantilla personal'
        $before = Snapshot $homePath
        $null = Run-Script 'install' $homePath
        Assert-That ((Snapshot $homePath) -eq $before) 'Dry-run modificó el home existente.'
    }
    Case 'fresh-install-and-rollback' {
        param($homePath)
        $foreign = Join-Path $homePath 'harness/templates/foreign.md'
        Write-Fixture $foreign 'documento ajeno'
        $foreignHash = Hash $foreign
        $null = Run-Script 'install' $homePath @('-Apply')
        Assert-Installed $homePath
        $state = Read-State $homePath
        foreach ($resource in $resources) {
            $record = @($state.files | Where-Object { $_.path -eq $resource.target })[0]
            Assert-That (-not $record.existed) ('Recurso nuevo marcado como previo: ' + $resource.target)
        }
        $null = Run-Script 'verify' $homePath
        $project = Join-Path $sandbox 'synthetic-project/odd/tasks/feature.md'
        $template = [IO.File]::ReadAllText((Join-Path $homePath 'harness/templates/feature-tasks.md'))
        Write-Fixture $project ($template.Replace('- **Approval:** pending', '- **Approval:** approved'))
        $output = & python (Join-Path $homePath 'harness/bin/taskboard.py') check $project 2>&1
        Assert-That ($LASTEXITCODE -eq 0) ('Plantilla instalada rechazada por validador instalado: ' + ($output -join "`n"))
        $before = Snapshot $homePath
        $null = Run-Script 'rollback' $homePath
        Assert-That ((Snapshot $homePath) -eq $before) 'Dry-run rollback modificó el home.'
        $null = Run-Script 'rollback' $homePath @('-Apply')
        foreach ($resource in $resources) {
            Assert-That (-not (Test-Path -LiteralPath (Join-Path $homePath $resource.target))) ('Rollback no retiró ' + $resource.target)
        }
        Assert-That ((Hash $foreign) -eq $foreignHash) 'Rollback modificó archivo ajeno.'
        Assert-That (-not (Test-Path -LiteralPath (Join-Path $homePath 'harness-install/state.json'))) 'Rollback dejó manifiesto nuevo activo.'
    }
    Case 'repeat-install' {
        param($homePath)
        $null = Run-Script 'install' $homePath @('-Apply')
        Assert-Installed $homePath
        $firstStateHash = Hash (Join-Path $homePath 'harness-install/state.json')
        $null = Run-Script 'install' $homePath @('-Apply')
        foreach ($resource in $resources) {
            Assert-That ((Hash (Join-Path $homePath $resource.target)) -eq (Hash (Join-Path $repo $resource.source))) ('Reinstalación cambió ' + $resource.target)
        }
        $null = Run-Script 'verify' $homePath
        $null = Run-Script 'rollback' $homePath @('-Apply')
        Assert-That ((Hash (Join-Path $homePath 'harness-install/state.json')) -eq $firstStateHash) 'Rollback de reinstalación no restauró manifiesto anterior.'
        Assert-Installed $homePath
        $null = Run-Script 'rollback' $homePath @('-Apply')
        foreach ($resource in $resources) {
            Assert-That (-not (Test-Path -LiteralPath (Join-Path $homePath $resource.target))) ('Segundo rollback no retiró ' + $resource.target)
        }
    }
    Case 'preserve-customized' {
        param($homePath)
        foreach ($resource in $resources[0..1]) {
            Write-Fixture (Join-Path $homePath $resource.target) ('personalización: ' + $resource.target)
        }
        $before = @($resources[0..1] | ForEach-Object { Hash (Join-Path $homePath $_.target) })
        $null = Run-Script 'install' $homePath @('-Apply')
        $state = Read-State $homePath
        for ($index = 0; $index -lt 2; $index++) {
            $resource = $resources[$index]
            Assert-That ((Hash (Join-Path $homePath $resource.target)) -eq $before[$index]) ('Personalización sobrescrita: ' + $resource.target)
            Assert-That ($state.preserved_customizations -contains $resource.target) ('Personalización no informada: ' + $resource.target)
        }
        $null = Run-Script 'rollback' $homePath @('-Apply')
        for ($index = 0; $index -lt 2; $index++) {
            Assert-That ((Hash (Join-Path $homePath $resources[$index].target)) -eq $before[$index]) 'Rollback perdió personalización conservada.'
        }
    }
    Case 'replace-customized-and-restore' {
        param($homePath)
        foreach ($resource in $resources[0..1]) {
            Write-Fixture (Join-Path $homePath $resource.target) ('original personalizado: ' + $resource.target)
        }
        $before = @($resources[0..1] | ForEach-Object { Hash (Join-Path $homePath $_.target) })
        $null = Run-Script 'install' $homePath @('-Apply', '-ReplaceCustomized')
        Assert-Installed $homePath
        $state = Read-State $homePath
        for ($index = 0; $index -lt 2; $index++) {
            $resource = $resources[$index]
            $record = @($state.files | Where-Object { $_.path -eq $resource.target })[0]
            Assert-That ([bool]$record.existed) 'Falta registro del original personalizado.'
            Assert-That ($record.backup_sha256 -eq $before[$index]) 'Manifiesto no registró hash del original.'
            Assert-That ((Hash (Join-Path $state.backup_dir $resource.target)) -eq $before[$index]) 'Backup no conserva original exacto.'
        }
        $null = Run-Script 'rollback' $homePath @('-Apply')
        for ($index = 0; $index -lt 2; $index++) {
            Assert-That ((Hash (Join-Path $homePath $resources[$index].target)) -eq $before[$index]) 'Rollback no restauró original personalizado.'
        }
    }
    Case 'reject-modified-resource-without-partial-rollback' {
        param($homePath)
        foreach ($resource in $resources[0..1]) {
            $isolatedHome = Join-Path $homePath ([IO.Path]::GetFileName($resource.target))
            $null = Run-Script 'install' $isolatedHome @('-Apply')
            Assert-Installed $isolatedHome
            [IO.File]::AppendAllText((Join-Path $isolatedHome $resource.target), "`nCambio después de instalar")
            $before = Snapshot $isolatedHome
            $null = Run-Script 'rollback' $isolatedHome @('-Apply') $false
            Assert-That ((Snapshot $isolatedHome) -eq $before) ('Rollback parcial al detectar deriva en ' + $resource.target)
        }
    }
    if ($failures.Count -gt 0) { throw ('TTP RESOURCE TESTS FAILED: ' + ($failures -join '; ')) }
    $global:LASTEXITCODE = 0
    Write-Output 'TTP RESOURCE TESTS OK'
} finally {
    # Resolve and check the absolute deletion target immediately before recursive cleanup.
    if (Test-Path -LiteralPath $sandbox) {
        $resolved = (Resolve-Path -LiteralPath $sandbox).ProviderPath
        $allowed = $tempParent.TrimEnd('\','/') + [IO.Path]::DirectorySeparatorChar
        Assert-That ($resolved.StartsWith($allowed, [StringComparison]::OrdinalIgnoreCase)) 'Cleanup fuera de TEMP rechazado.'
        Assert-That ((Split-Path -Leaf $resolved) -match '^codex-ttp-resources-[a-f0-9]{32}$') 'Cleanup sin prefijo propio rechazado.'
        Remove-Item -LiteralPath $resolved -Recurse -Force
    }
}
