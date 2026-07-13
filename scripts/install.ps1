# cx-cli Windows Installation Script
# Automated installer for PowerShell.

$ErrorActionPreference = "Stop"

# 1. Detect Architecture
$rawArch = $env:PROCESSOR_ARCHITECTURE.ToLower()
$arch = ""

if ($rawArch -eq "amd64" -or $rawArch -eq "ia64" -or $rawArch -eq "x86") {
    $arch = "amd64"
} elseif ($rawArch -eq "arm64") {
    $arch = "arm64"
} else {
    Write-Error "Unsupported architecture: $rawArch"
    exit 1
}

# 2. Determine download URLs and targets
$binaryName = "cx-windows-$arch.exe"
$downloadUrl = "https://github.com/guppshub/cx-cli/releases/latest/download/$binaryName"

$installDir = Join-Path $env:USERPROFILE ".cx\bin"
$targetPath = Join-Path $installDir "cx.exe"

Write-Host "========================================="
Write-Host " Installing cx-cli for Windows"
Write-Host " Detected Arch: $arch"
Write-Host "========================================="

# 3. Create target directory
if (!(Test-Path $installDir)) {
    New-Item -ItemType Directory -Force -Path $installDir | Out-Null
}

# 4. Fetch the latest release binary
Write-Host "Downloading $downloadUrl..."
try {
    # Using -UseBasicParsing to avoid dependency on Internet Explorer engines
    Invoke-WebRequest -Uri $downloadUrl -OutFile $targetPath -UseBasicParsing
    Write-Host "Download completed successfully."
} catch {
    Write-Error "Failed to download binary. Please make sure a release tag has been published on GitHub."
    exit 1
}

# 5. Persistently add directory to User PATH
$userPath = [System.Environment]::GetEnvironmentVariable("PATH", "User")
$normalizedInstallDir = $installDir.TrimEnd('\')

$paths = $userPath -split ';'
$alreadyInPath = $false
foreach ($p in $paths) {
    if ($p.TrimEnd('\') -eq $normalizedInstallDir) {
        $alreadyInPath = $true
        break
    }
}

if (!$alreadyInPath) {
    Write-Host "Adding $installDir to your User PATH..."
    # Ensure we don't prepend a stray semicolon if PATH was empty
    if ($userPath -and !$userPath.EndsWith(";")) {
        $userPath += ";"
    }
    $newPath = $userPath + $installDir
    [System.Environment]::SetEnvironmentVariable("PATH", $newPath, "User")
    Write-Host "PATH updated persistently."
}

Write-Host "========================================="
Write-Host " Success! cx-cli is now installed."
Write-Host " Please close and RESTART your PowerShell window to apply PATH updates."
Write-Host " Run 'cx init' to get started."
Write-Host "========================================="
