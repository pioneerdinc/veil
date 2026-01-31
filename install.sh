#!/bin/bash

# Veil installer script
# Usage: curl -fsSL https://raw.githubusercontent.com/ossydotpy/veil/master/install.sh | bash

set -e

REPO="ossydotpy/veil"
BINARY="veil"
INSTALL_DIR="${INSTALL_DIR:-/usr/local/bin}"

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Detect OS and architecture
detect_platform() {
    local os=""
    local arch=""

    case "$(uname -s)" in
        Linux*)     os="linux";;
        Darwin*)    os="darwin";;
        CYGWIN*|MINGW*|MSYS*) os="windows";;
        *)          os="unknown";;
    esac

    case "$(uname -m)" in
        x86_64|amd64)   arch="amd64";;
        arm64|aarch64)  arch="arm64";;
        i386|i686)      arch="386";;
        armv7l)         arch="arm";;
        *)              arch="unknown";;
    esac

    if [ "$os" = "unknown" ] || [ "$arch" = "unknown" ]; then
        echo -e "${RED}Error: Unsupported platform: $(uname -s) $(uname -m)${NC}" >&2
        exit 1
    fi

    echo "${os}_${arch}"
}

# Get latest release version
get_latest_version() {
    curl -s "https://api.github.com/repos/${REPO}/releases/latest" | grep '"tag_name":' | sed -E 's/.*"([^"]+)".*/\1/'
}

# Download binary
download_binary() {
    local version="$1"
    local platform="$2"
    local tmp_dir="$3"

    local download_url="https://github.com/${REPO}/releases/download/${version}/${BINARY}_${platform}"
    local output_path="${tmp_dir}/${BINARY}"

    echo -e "${YELLOW}Downloading ${BINARY} ${version} for ${platform}...${NC}" >&2

    if ! curl -fsSL "$download_url" -o "$output_path"; then
        echo -e "${RED}Error: Failed to download from ${download_url}${NC}" >&2
        echo -e "${YELLOW}Trying alternative URL...${NC}" >&2
        
        # Try with .tar.gz extension
        download_url="https://github.com/${REPO}/releases/download/${version}/${BINARY}_${version#v}_${platform}.tar.gz"
        if ! curl -fsSL "$download_url" -o "${tmp_dir}/veil.tar.gz"; then
            echo -e "${RED}Error: Failed to download binary${NC}" >&2
            exit 1
        fi
        
        tar -xzf "${tmp_dir}/veil.tar.gz" -C "$tmp_dir"
    fi

    chmod +x "$output_path"
    echo "$output_path"
}

# Install binary
install_binary() {
    local binary_path="$1"
    local install_dir="$2"

    # Check if we need sudo
    if [ -w "$install_dir" ]; then
        mv "$binary_path" "${install_dir}/${BINARY}"
    else
        echo -e "${YELLOW}Requires sudo to install to ${install_dir}${NC}" >&2
        sudo mv "$binary_path" "${install_dir}/${BINARY}"
    fi

    echo -e "${GREEN}✓ Installed ${BINARY} to ${install_dir}/${BINARY}${NC}" >&2
}

# Check if binary is in PATH
check_path() {
    if ! command -v "$BINARY" &> /dev/null; then
        echo -e "${YELLOW}Warning: ${INSTALL_DIR} is not in your PATH${NC}" >&2
        echo -e "${YELLOW}Add this to your shell profile:${NC}" >&2
        echo "export PATH=\"\$PATH:${INSTALL_DIR}\"" >&2
    else
        echo -e "${GREEN}✓ ${BINARY} is now available in your PATH${NC}" >&2
    fi
}

# Main installation
main() {
    echo -e "${GREEN}Installing Veil...${NC}" >&2

    # Detect platform
    platform=$(detect_platform)
    echo "Detected platform: $platform" >&2

    # Get version
    version=$(get_latest_version)
    if [ -z "$version" ]; then
        echo -e "${RED}Error: Could not determine latest version${NC}" >&2
        exit 1
    fi
    echo "Latest version: $version" >&2

    # Create temp directory
    tmp_dir=$(mktemp -d)
    trap "rm -rf $tmp_dir" EXIT

    # Download
    binary_path=$(download_binary "$version" "$platform" "$tmp_dir")

    # Install
    install_binary "$binary_path" "$INSTALL_DIR"

    # Verify
    check_path

    echo "" >&2
    echo -e "${GREEN}Installation complete!${NC}" >&2
    echo "Run 'veil --help' to get started"
    echo ""
    echo "Next steps:"
    echo "  1. veil init          # Generate your master key"
    echo "  2. veil --help        # See all commands"
}

# Allow custom install directory
while [[ $# -gt 0 ]]; do
    case $1 in
        --dir)
            INSTALL_DIR="$2"
            shift 2
            ;;
        --version)
            VERSION="$2"
            shift 2
            ;;
        -h|--help)
            echo "Usage: install.sh [OPTIONS]"
            echo ""
            echo "Options:"
            echo "  --dir DIR       Install to custom directory (default: /usr/local/bin)"
            echo "  --version VER   Install specific version (default: latest)"
            echo "  -h, --help      Show this help message"
            exit 0
            ;;
        *)
            echo "Unknown option: $1"
            exit 1
            ;;
    esac
done

main
