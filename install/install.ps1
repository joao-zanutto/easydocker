Set-StrictMode -Version Latest
$ErrorActionPreference = "Stop"

$Repo = if ($env:EASYDOCKER_REPO) { $env:EASYDOCKER_REPO } else { "joao-zanutto/easydocker" }
$Binary = "easydocker.exe"
$InstallDir = if ($env:EASYDOCKER_INSTALL_DIR) { $env:EASYDOCKER_INSTALL_DIR } else { "$env:LOCALAPPDATA\Programs\easydocker" }
$VersionInput = if ($env:EASYDOCKER_VERSION) { $env:EASYDOCKER_VERSION } else { "latest" }

function Resolve-Tag {
    param([string]$RepoName, [string]$Version)

    if ($Version -eq "latest") {
        $api = "https://api.github.com/repos/$RepoName/releases/latest"
        $release = Invoke-RestMethod -Uri $api
        if (-not $release.tag_name) {
            throw "Could not resolve latest release tag from $api"
        }
        return [string]$release.tag_name
    }

    if ($Version.StartsWith("v")) {
        return $Version
    }

    return "v$Version"
}

function Resolve-Arch {
    switch ([System.Runtime.InteropServices.RuntimeInformation]::OSArchitecture) {
        "X64" { return "amd64" }
        "Arm64" { return "arm64" }
        default { throw "Unsupported architecture: $([System.Runtime.InteropServices.RuntimeInformation]::OSArchitecture)" }
    }
}

function Get-ExpectedChecksum {
    param([string]$ChecksumsPath, [string]$AssetName)

    $line = Get-Content -Path $ChecksumsPath | Where-Object { $_ -match "\s\s$([regex]::Escape($AssetName))$" } | Select-Object -First 1
    if (-not $line) {
        throw "Checksum not found for $AssetName"
    }

    return ($line -split '\s+')[0]
}

function Ensure-PathContainsInstallDir {
    param([string]$Dir)

    $userPath = [Environment]::GetEnvironmentVariable("Path", "User")
    $parts = @()
    if ($userPath) {
        $parts = $userPath -split ';'
    }

    if ($parts -contains $Dir) {
        return
    }

    $newPath = if ($userPath) { "$userPath;$Dir" } else { $Dir }
    [Environment]::SetEnvironmentVariable("Path", $newPath, "User")
    Write-Host "Added $Dir to your user PATH. Open a new terminal to use '$($Binary -replace '.exe$','')'."
}

$tag = Resolve-Tag -RepoName $Repo -Version $VersionInput
$versionNoV = $tag.TrimStart('v')
$arch = Resolve-Arch

$assetName = "easydocker`_v${versionNoV}`_windows_${arch}.zip"
$releaseBase = "https://github.com/$Repo/releases/download/$tag"

$tmpDir = Join-Path $env:TEMP ("easydocker-install-" + [guid]::NewGuid().ToString())
New-Item -ItemType Directory -Path $tmpDir | Out-Null

try {
    $archivePath = Join-Path $tmpDir $assetName
    $checksumsPath = Join-Path $tmpDir "checksums.txt"

    Write-Host "Installing easydocker $tag for windows/$arch..."
    Invoke-WebRequest -Uri "$releaseBase/$assetName" -OutFile $archivePath
    Invoke-WebRequest -Uri "$releaseBase/checksums.txt" -OutFile $checksumsPath

    $expected = Get-ExpectedChecksum -ChecksumsPath $checksumsPath -AssetName $assetName
    $actual = (Get-FileHash -Path $archivePath -Algorithm SHA256).Hash.ToLowerInvariant()
    if ($expected.ToLowerInvariant() -ne $actual) {
        throw "Checksum mismatch for $assetName"
    }

    Expand-Archive -Path $archivePath -DestinationPath $tmpDir -Force

    $exe = Get-ChildItem -Path $tmpDir -Recurse -File | Where-Object { $_.Name -eq "easydocker.exe" } | Select-Object -First 1
    if (-not $exe) {
        throw "Extracted binary 'easydocker.exe' not found"
    }

    New-Item -ItemType Directory -Path $InstallDir -Force | Out-Null
    $target = Join-Path $InstallDir "easydocker.exe"
    Copy-Item -Path $exe.FullName -Destination $target -Force

    Ensure-PathContainsInstallDir -Dir $InstallDir
    Write-Host "Installed: $target"
    Write-Host "Run: easydocker"
}
finally {
    if (Test-Path $tmpDir) {
        Remove-Item -Path $tmpDir -Recurse -Force
    }
}