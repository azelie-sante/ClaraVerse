# install.ps1 - ClaraVerse MCP Client Installer for Windows
# Usage: irm https://get.claraverse.com/mcp.ps1 | iex
#        irm https://get.claraverse.com/mcp.ps1 | iex -args install|update|uninstall

param(
    [Parameter(Position=0)]
    [ValidateSet('install', 'update', 'uninstall', 'help')]
    [string]$Action = 'install'
)

$ErrorActionPreference = "Stop"

# Configuration
$Version = if ($env:MCP_CLIENT_VERSION) { $env:MCP_CLIENT_VERSION } else { "latest" }
$InstallDir = if ($env:MCP_CLIENT_INSTALL_DIR) { $env:MCP_CLIENT_INSTALL_DIR } else { "$env:LOCALAPPDATA\ClaraVerse\bin" }
$Repo = "badboysm890/ClaraVerse-Scarlet"
$RemoveConfig = if ($env:REMOVE_CONFIG -eq "true") { $true } else { $false }

# Output functions
function Write-Info($msg) {
    Write-Host "info: " -ForegroundColor Blue -NoNewline
    Write-Host $msg
}

function Write-Success($msg) {
    Write-Host "success: " -ForegroundColor Green -NoNewline
    Write-Host $msg
}

function Write-Warn($msg) {
    Write-Host "warn: " -ForegroundColor Yellow -NoNewline
    Write-Host $msg
}

function Write-Error-Custom($msg) {
    Write-Host "error: " -ForegroundColor Red -NoNewline
    Write-Host $msg
    exit 1
}

function Write-Banner {
    Write-Host ""
    Write-Host "ClaraVerse MCP Client Installer" -ForegroundColor White
    Write-Host ""
}

# Get latest version from GitHub releases
function Get-LatestVersion {
    try {
        $response = Invoke-RestMethod "https://api.github.com/repos/$Repo/releases/latest" -UseBasicParsing
        $tag = $response.tag_name
        # Remove 'v' prefix if present
        return $tag -replace '^v', ''
    } catch {
        Write-Error-Custom "Failed to fetch latest version: $_"
    }
}

# Verify checksum
function Verify-Checksum {
    param(
        [string]$FilePath,
        [string]$Version
    )

    $checksumUrl = "https://github.com/$Repo/releases/download/v$Version/checksums.txt"

    try {
        Write-Info "Verifying checksum..."
        $checksums = Invoke-WebRequest -Uri $checksumUrl -UseBasicParsing
        $expectedLine = ($checksums.Content -split "`n" | Where-Object { $_ -match "mcp-client-windows-amd64" }) -split '\s+' | Select-Object -First 1

        if ($expectedLine) {
            $actual = (Get-FileHash -Path $FilePath -Algorithm SHA256).Hash.ToLower()
            $expected = $expectedLine.ToLower()

            if ($actual -ne $expected) {
                Remove-Item -Force $FilePath -ErrorAction SilentlyContinue
                Write-Error-Custom "Checksum verification failed! Expected: $expected, Got: $actual"
            }
            Write-Success "Checksum verified"
        } else {
            Write-Warn "Skipping checksum verification (checksums.txt not found)"
        }
    } catch {
        Write-Warn "Skipping checksum verification: $_"
    }
}

# Install action
function Install-MCPClient {
    Write-Banner

    if ($Version -eq "latest") {
        Write-Info "Fetching latest version..."
        $script:Version = Get-LatestVersion
    }

    $url = "https://github.com/$Repo/releases/download/v$Version/mcp-client-windows-amd64.exe"

    Write-Info "Installing mcp-client v$Version to $InstallDir"

    # Create directory
    if (-not (Test-Path $InstallDir)) {
        New-Item -ItemType Directory -Force -Path $InstallDir | Out-Null
    }

    # Download
    $tempFile = "$env:TEMP\mcp-client-download.exe"
    Write-Info "Downloading..."
    try {
        Invoke-WebRequest -Uri $url -OutFile $tempFile -UseBasicParsing
    } catch {
        Write-Error-Custom "Download failed: $_"
    }

    # Verify checksum
    Verify-Checksum -FilePath $tempFile -Version $Version

    # Move to install directory
    Move-Item -Force $tempFile "$InstallDir\mcp-client.exe"

    # Add to PATH
    $userPath = [Environment]::GetEnvironmentVariable("PATH", "User")
    if ($userPath -notlike "*$InstallDir*") {
        [Environment]::SetEnvironmentVariable("PATH", "$userPath;$InstallDir", "User")
        Write-Info "Added $InstallDir to PATH"
        Write-Host ""
        Write-Host "  Restart your terminal to use mcp-client" -ForegroundColor Yellow
        Write-Host ""
    }

    Write-Success "mcp-client v$Version installed!"

    # Print next steps
    Write-Host ""
    Write-Host "  " -NoNewline
    Write-Host "MCP Client installed successfully!" -ForegroundColor Green
    Write-Host ""
    Write-Host "  Get started:"
    Write-Host "    mcp-client login     # Authenticate with ClaraVerse"
    Write-Host ""
    Write-Host "  This will:"
    Write-Host "    1. Open browser for device authorization"
    Write-Host "    2. Let you select MCP servers"
    Write-Host "    3. Optionally install as background service"
    Write-Host ""
    Write-Host "  Other commands:"
    Write-Host "    mcp-client start     # Start manually"
    Write-Host "    mcp-client status    # Check connection"
    Write-Host "    mcp-client --help    # See all commands"
    Write-Host ""
    Write-Host "  Documentation: https://docs.claraverse.com/mcp-client"
    Write-Host ""
}

# Update action
function Update-MCPClient {
    Write-Banner

    $current = Get-Command mcp-client -ErrorAction SilentlyContinue
    if (-not $current) {
        Write-Info "mcp-client not found, installing..."
        Install-MCPClient
        return
    }

    $currentPath = $current.Source
    $currentDir = Split-Path $currentPath

    # Get current version
    try {
        $currentVersion = & $currentPath --version 2>&1 | Select-String -Pattern '\d+\.\d+\.\d+' | ForEach-Object { $_.Matches[0].Value }
        if (-not $currentVersion) { $currentVersion = "unknown" }
    } catch {
        $currentVersion = "unknown"
    }
    Write-Info "Current version: $currentVersion"

    if ($Version -eq "latest") {
        Write-Info "Fetching latest version..."
        $script:Version = Get-LatestVersion
    }

    if ($currentVersion -eq $Version) {
        Write-Success "Already at version $Version"
        return
    }

    Write-Info "Updating to v$Version..."

    # Download to temp
    $tempFile = "$env:TEMP\mcp-client-update.exe"
    $url = "https://github.com/$Repo/releases/download/v$Version/mcp-client-windows-amd64.exe"

    try {
        Invoke-WebRequest -Uri $url -OutFile $tempFile -UseBasicParsing
    } catch {
        Write-Error-Custom "Download failed: $_"
    }

    # Verify checksum
    Verify-Checksum -FilePath $tempFile -Version $Version

    # Replace
    Move-Item -Force $tempFile $currentPath

    Write-Success "mcp-client updated to v$Version!"
    Write-Host ""
    Write-Host "  Run 'mcp-client --version' to verify"
    Write-Host ""
}

# Uninstall action
function Uninstall-MCPClient {
    Write-Banner

    $current = Get-Command mcp-client -ErrorAction SilentlyContinue
    if (-not $current) {
        Write-Error-Custom "mcp-client not found"
    }

    $currentPath = $current.Source

    Write-Info "Uninstalling from $currentPath..."

    # Stop service (if any)
    try {
        & $currentPath service stop 2>$null
        & $currentPath service uninstall 2>$null
    } catch {
        # Ignore errors
    }

    # Remove binary
    Remove-Item -Force $currentPath

    Write-Success "Removed $currentPath"

    # Optionally remove config
    $configDir = "$env:USERPROFILE\.claraverse"
    if ($RemoveConfig) {
        if (Test-Path $configDir) {
            Remove-Item -Recurse -Force $configDir
            Write-Info "Removed config directory: $configDir"
        }
    } else {
        Write-Host ""
        Write-Host "  Config directory preserved: $configDir"
        Write-Host "  To remove: Remove-Item -Recurse $configDir"
        Write-Host ""
    }

    Write-Success "mcp-client uninstalled"
}

# Show help
function Show-Help {
    Write-Host @"
ClaraVerse MCP Client Installer for Windows

Usage:
  irm https://get.claraverse.com/mcp.ps1 | iex
  irm https://get.claraverse.com/mcp.ps1 | iex -args install|update|uninstall

Commands:
  install     Install mcp-client (default)
  update      Update to latest version
  uninstall   Remove mcp-client

Environment Variables:
  MCP_CLIENT_VERSION      Specific version to install (default: latest)
  MCP_CLIENT_INSTALL_DIR  Custom installation directory
  REMOVE_CONFIG           Set to "true" to remove config on uninstall

Examples:
  # Install latest
  irm https://get.claraverse.com/mcp.ps1 | iex

  # Install specific version
  `$env:MCP_CLIENT_VERSION = "1.2.0"; irm https://get.claraverse.com/mcp.ps1 | iex

  # Install to custom directory
  `$env:MCP_CLIENT_INSTALL_DIR = "C:\bin"; irm https://get.claraverse.com/mcp.ps1 | iex

  # Update
  irm https://get.claraverse.com/mcp.ps1 | iex -args update

  # Uninstall with config removal
  `$env:REMOVE_CONFIG = "true"; irm https://get.claraverse.com/mcp.ps1 | iex -args uninstall

"@
}

# Main
switch ($Action) {
    'install' { Install-MCPClient }
    'update' { Update-MCPClient }
    'uninstall' { Uninstall-MCPClient }
    'help' { Show-Help }
    default {
        Write-Error-Custom "Unknown action: $Action. Use 'install', 'update', or 'uninstall'."
    }
}
