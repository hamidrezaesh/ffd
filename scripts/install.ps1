$REPO = "hamidrezaesh/ffd"
$INSTALL_DIR = "$env:LOCALAPPDATA\FFD"

$policy = Get-ExecutionPolicy

if ($policy -eq "Restricted") {
    Write-Host "PowerShell scripts are blocked."
    Write-Host "Run:"
    Write-Host "Set-ExecutionPolicy -Scope CurrentUser RemoteSigned"
    exit 1
}

Write-Host "Installing ffd..."

# Detect architecture
switch ($env:PROCESSOR_ARCHITECTURE.ToLower()) {
    "amd64" { $GOARCH = "amd64" }
    "arm64" { $GOARCH = "arm64" }
    default {
        Write-Host "Error: unsupported architecture"
        exit 1
    }
}

# Get latest release
$LATEST_TAG = (
    Invoke-RestMethod "https://api.github.com/repos/$REPO/releases/latest"
).tag_name

if (-not $LATEST_TAG) {
    Write-Host "Error: could not determine latest release."
    exit 1
}

$VERSION = $LATEST_TAG.TrimStart("v")

# Download URL
$ARCHIVE = "ffd_${VERSION}_windows_${GOARCH}.zip"
$URL = "https://github.com/$REPO/releases/download/$LATEST_TAG/$ARCHIVE"

# Temp directory
$TMP_DIR = Join-Path $env:TEMP ([System.Guid]::NewGuid())
New-Item -ItemType Directory -Path $TMP_DIR | Out-Null

function Cleanup {
    if (Test-Path $TMP_DIR) {
        Remove-Item -Path $TMP_DIR -Recurse -Force
    }
}

try {
    Write-Host "Downloading ffd $VERSION..."

    Invoke-WebRequest `
        -Uri $URL `
        -OutFile "$TMP_DIR\ffd.zip"

    Write-Host "Extracting..."

    Expand-Archive -Path "$TMP_DIR\ffd.zip" -DestinationPath "$TMP_DIR"

    Write-Host "Installing to $INSTALL_DIR..."

    New-Item -ItemType Directory -Force -Path $INSTALL_DIR | Out-Null

    Copy-Item `
        "$TMP_DIR\ffd.exe" `
        "$INSTALL_DIR\ffd.exe" `
        -Force

    # Add to PATH
    $currentPath = [Environment]::GetEnvironmentVariable("Path", "User")

    if ($currentPath -notlike "*$INSTALL_DIR*") {
        [Environment]::SetEnvironmentVariable(
            "Path",
            "$currentPath;$INSTALL_DIR",
            "User"
        )

        Write-Host "Added $INSTALL_DIR to PATH"
    }

    Write-Host ""
    Write-Host "ffd v$VERSION installed successfully!"
    Write-Host ""
    Write-Host "Restart your terminal and run:"
    Write-Host "  ffd --help"
}
finally {
    Cleanup
}