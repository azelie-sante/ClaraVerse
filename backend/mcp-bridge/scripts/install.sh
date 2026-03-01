#!/bin/bash
# install.sh - ClaraVerse MCP Client Installer
# Usage: curl -fsSL https://get.claraverse.com/mcp | sh
#        curl -fsSL https://get.claraverse.com/mcp | sh -s -- [install|update|uninstall]

set -euo pipefail

# Configuration
VERSION="${MCP_CLIENT_VERSION:-latest}"
INSTALL_DIR="${MCP_CLIENT_INSTALL_DIR:-}"
GITHUB_REPO="badboysm890/ClaraVerse-Scarlet"
BINARY_NAME="mcp-client"
REMOVE_CONFIG="${REMOVE_CONFIG:-false}"

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
BOLD='\033[1m'
NC='\033[0m' # No Color

# Output functions
error() {
  echo -e "${RED}${BOLD}error${NC}: $1" >&2
  exit 1
}

info() {
  echo -e "${BLUE}${BOLD}info${NC}: $1"
}

success() {
  echo -e "${GREEN}${BOLD}success${NC}: $1"
}

warn() {
  echo -e "${YELLOW}${BOLD}warn${NC}: $1"
}

# Print banner
print_banner() {
  echo ""
  echo -e "${BOLD}ClaraVerse MCP Client Installer${NC}"
  echo ""
}

# Detect OS and architecture
detect_os_arch() {
  OS="$(uname -s)"
  ARCH="$(uname -m)"

  case "$OS" in
    Darwin)  OS="darwin" ;;
    Linux)   OS="linux" ;;
    MINGW*|MSYS*|CYGWIN*)
      echo ""
      echo "For Windows, use PowerShell:"
      echo ""
      echo "  irm https://get.claraverse.com/mcp.ps1 | iex"
      echo ""
      exit 1 ;;
    *) error "Unsupported OS: $OS" ;;
  esac

  case "$ARCH" in
    x86_64)        ARCH="amd64" ;;
    aarch64|arm64) ARCH="arm64" ;;
    *) error "Unsupported architecture: $ARCH" ;;
  esac

  # Validate supported combinations
  if [[ "$OS" == "linux" && "$ARCH" == "arm64" ]]; then
    error "Linux ARM64 not currently supported. Please use x86_64."
  fi

  info "Detected: $OS/$ARCH"
}

# Get latest version from GitHub releases
get_latest_version() {
  local latest
  latest=$(curl -fsSL "https://api.github.com/repos/$GITHUB_REPO/releases/latest" 2>/dev/null |
           grep '"tag_name"' |
           sed -E 's/.*"v?([^"]+)".*/\1/' |
           head -1)

  if [[ -z "$latest" ]]; then
    error "Failed to fetch latest version. Check your internet connection."
  fi

  # Remove 'v' prefix if present
  echo "${latest#v}"
}

# Determine installation directory
get_install_dir() {
  # 1. User-specified via env var
  if [[ -n "$INSTALL_DIR" ]]; then
    mkdir -p "$INSTALL_DIR" 2>/dev/null || true
    echo "$INSTALL_DIR"
    return
  fi

  # 2. /usr/local/bin if writable (no sudo needed on newer macOS)
  if [[ -w "/usr/local/bin" ]]; then
    echo "/usr/local/bin"
    return
  fi

  # 3. ~/.local/bin (XDG standard, always works)
  mkdir -p "$HOME/.local/bin"
  echo "$HOME/.local/bin"
}

# Download binary with checksum verification
download_binary() {
  local version=$1
  local os=$2
  local arch=$3
  local tmpdir="${TMPDIR:-/tmp}"
  local binary_url="https://github.com/$GITHUB_REPO/releases/download/v$version/mcp-client-$os-$arch"
  local checksum_url="https://github.com/$GITHUB_REPO/releases/download/v$version/checksums.txt"

  info "Downloading mcp-client v$version..."

  # Download binary
  if ! curl -fsSL "$binary_url" -o "$tmpdir/mcp-client"; then
    error "Failed to download binary from $binary_url"
  fi

  # Download and verify checksum
  info "Verifying checksum..."
  local expected
  expected=$(curl -fsSL "$checksum_url" 2>/dev/null | grep "mcp-client-$os-$arch" | awk '{print $1}' || true)

  if [[ -n "$expected" ]]; then
    local actual
    if command -v sha256sum &>/dev/null; then
      actual=$(sha256sum "$tmpdir/mcp-client" | awk '{print $1}')
    elif command -v shasum &>/dev/null; then
      actual=$(shasum -a 256 "$tmpdir/mcp-client" | awk '{print $1}')
    else
      warn "Skipping checksum verification (sha256sum not found)"
      return
    fi

    if [[ "$expected" != "$actual" ]]; then
      rm -f "$tmpdir/mcp-client"
      error "Checksum verification failed! Expected: $expected, Got: $actual"
    fi
    success "Checksum verified"
  else
    warn "Skipping checksum verification (checksums.txt not found)"
  fi
}

# Update PATH in shell config
update_path() {
  local install_dir=$1

  # Skip if already in PATH
  if [[ ":$PATH:" == *":$install_dir:"* ]]; then
    return
  fi

  # Detect shell config file
  local shell_rc=""
  case "${SHELL:-}" in
    */zsh)
      if [[ -f "$HOME/.zprofile" ]]; then
        shell_rc="$HOME/.zprofile"
      else
        shell_rc="$HOME/.zshrc"
      fi
      ;;
    */bash)
      if [[ -f "$HOME/.bash_profile" ]]; then
        shell_rc="$HOME/.bash_profile"
      elif [[ -f "$HOME/.bashrc" ]]; then
        shell_rc="$HOME/.bashrc"
      else
        shell_rc="$HOME/.profile"
      fi
      ;;
    */fish)
      shell_rc="$HOME/.config/fish/config.fish"
      mkdir -p "$HOME/.config/fish"
      ;;
    *)
      shell_rc="$HOME/.profile"
      ;;
  esac

  if [[ -n "$shell_rc" ]]; then
    # Check if already added
    if ! grep -q "# mcp-client" "$shell_rc" 2>/dev/null; then
      {
        echo ""
        echo "# mcp-client"
        if [[ "$shell_rc" == *"fish"* ]]; then
          echo "set -gx PATH \"$install_dir\" \$PATH"
        else
          echo "export PATH=\"$install_dir:\$PATH\""
        fi
      } >> "$shell_rc"
      info "Added $install_dir to PATH in $shell_rc"
      echo ""
      echo "  Restart your shell or run:"
      echo "    source $shell_rc"
      echo ""
    fi
  fi
}

# Print next steps after installation
print_next_steps() {
  echo ""
  echo -e "${GREEN}${BOLD}  MCP Client installed successfully!${NC}"
  echo ""
  echo "  Get started:"
  echo "    mcp-client login     # Authenticate with ClaraVerse"
  echo ""
  echo "  This will:"
  echo "    1. Open browser for device authorization"
  echo "    2. Let you select MCP servers"
  echo "    3. Optionally install as background service"
  echo ""
  echo "  Other commands:"
  echo "    mcp-client start     # Start manually"
  echo "    mcp-client status    # Check connection"
  echo "    mcp-client --help    # See all commands"
  echo ""
  echo "  Documentation: https://docs.claraverse.com/mcp-client"
  echo ""
}

# Install action
do_install() {
  print_banner
  detect_os_arch

  if [[ "$VERSION" == "latest" ]]; then
    info "Fetching latest version..."
    VERSION=$(get_latest_version)
  fi

  local install_dir
  install_dir=$(get_install_dir)
  local tmpdir="${TMPDIR:-/tmp}"

  info "Installing mcp-client v$VERSION to $install_dir"

  download_binary "$VERSION" "$OS" "$ARCH"

  # Make executable and move to install dir
  chmod +x "$tmpdir/mcp-client"

  # Try without sudo first, then with sudo if needed
  if [[ -w "$install_dir" ]]; then
    mv "$tmpdir/mcp-client" "$install_dir/mcp-client"
  else
    info "Requesting sudo to install to $install_dir..."
    sudo mv "$tmpdir/mcp-client" "$install_dir/mcp-client"
  fi

  # Update PATH if needed
  update_path "$install_dir"

  success "mcp-client v$VERSION installed to $install_dir/mcp-client"
  print_next_steps
}

# Update action
do_update() {
  print_banner
  local current_path
  current_path=$(command -v mcp-client 2>/dev/null || true)

  if [[ -z "$current_path" ]]; then
    info "mcp-client not found, performing fresh install..."
    do_install
    return
  fi

  # Get current version
  local current_version
  current_version=$("$current_path" --version 2>/dev/null | grep -oE '[0-9]+\.[0-9]+\.[0-9]+' | head -1 || echo "unknown")
  info "Current version: $current_version"

  # Get latest version
  if [[ "$VERSION" == "latest" ]]; then
    info "Fetching latest version..."
    VERSION=$(get_latest_version)
  fi

  if [[ "$current_version" == "$VERSION" ]]; then
    success "Already at version $VERSION"
    return
  fi

  info "Updating to v$VERSION..."
  detect_os_arch

  # Backup current binary
  cp "$current_path" "$current_path.bak"

  # Install to same location
  INSTALL_DIR=$(dirname "$current_path")

  local tmpdir="${TMPDIR:-/tmp}"
  download_binary "$VERSION" "$OS" "$ARCH"

  chmod +x "$tmpdir/mcp-client"

  if [[ -w "$INSTALL_DIR" ]]; then
    mv "$tmpdir/mcp-client" "$current_path"
  else
    sudo mv "$tmpdir/mcp-client" "$current_path"
  fi

  # Remove backup
  rm -f "$current_path.bak"

  success "mcp-client updated to v$VERSION"
  echo ""
  echo "  Run 'mcp-client --version' to verify"
  echo ""
}

# Uninstall action
do_uninstall() {
  print_banner
  local current_path
  current_path=$(command -v mcp-client 2>/dev/null || true)

  if [[ -z "$current_path" ]]; then
    error "mcp-client not found in PATH"
  fi

  info "Uninstalling mcp-client from $current_path..."

  # Stop and uninstall service if running
  "$current_path" service stop 2>/dev/null || true
  "$current_path" service uninstall 2>/dev/null || true

  # Remove binary
  if [[ -w "$(dirname "$current_path")" ]]; then
    rm -f "$current_path"
  else
    sudo rm -f "$current_path"
  fi

  success "Removed $current_path"

  # Optionally remove config
  if [[ "$REMOVE_CONFIG" == "true" ]]; then
    if [[ -d "$HOME/.claraverse" ]]; then
      rm -rf "$HOME/.claraverse"
      info "Removed config directory: ~/.claraverse"
    fi
  else
    echo ""
    echo "  Config directory preserved: ~/.claraverse"
    echo "  To remove: rm -rf ~/.claraverse"
    echo ""
  fi

  success "mcp-client uninstalled"
}

# Show help
show_help() {
  cat << 'EOF'
ClaraVerse MCP Client Installer

Usage:
  curl -fsSL https://get.claraverse.com/mcp | sh
  curl -fsSL https://get.claraverse.com/mcp | sh -s -- [command]

Commands:
  install     Install mcp-client (default)
  update      Update to latest version
  uninstall   Remove mcp-client

Environment Variables:
  MCP_CLIENT_VERSION      Specific version to install (default: latest)
  MCP_CLIENT_INSTALL_DIR  Custom installation directory
  REMOVE_CONFIG           Set to "true" to remove ~/.claraverse on uninstall

Examples:
  # Install latest
  curl -fsSL https://get.claraverse.com/mcp | sh

  # Install specific version
  curl -fsSL https://get.claraverse.com/mcp | MCP_CLIENT_VERSION=1.2.0 sh

  # Install to custom directory
  curl -fsSL https://get.claraverse.com/mcp | MCP_CLIENT_INSTALL_DIR=/opt/bin sh

  # Update
  curl -fsSL https://get.claraverse.com/mcp | sh -s -- update

  # Uninstall with config removal
  curl -fsSL https://get.claraverse.com/mcp | REMOVE_CONFIG=true sh -s -- uninstall

EOF
}

# Main
main() {
  local action="${1:-install}"

  case "$action" in
    install)
      do_install
      ;;
    update)
      do_update
      ;;
    uninstall)
      do_uninstall
      ;;
    -h|--help|help)
      show_help
      ;;
    *)
      error "Unknown action: $action. Use 'install', 'update', or 'uninstall'."
      ;;
  esac
}

main "$@"
