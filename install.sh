#!/bin/sh
#
#     _________       __ 
#    / ____/ (_)___  / /_
#   / /_  / / / __ \/ __/
#  / __/ / / / /_/ / /_  
# /_/   /_/_/ .___/\__/  
#          /_/           
#
# Flipt installer
#
# Usage:
#   curl -fsSL https://github.com/flipt-io/flipt/raw/main/install.sh | sh
#   curl -fsSL https://github.com/flipt-io/flipt/raw/main/install.sh | sh -s -- v1.45.0

set -e

# Configuration
REPO="flipt-io/flipt"
RELEASE_URL_BASE="https://api.github.com/repos/${REPO}/releases"
DOWNLOAD_URL_BASE="https://download.flipt.io/flipt"

# Color and formatting
if command -v tput >/dev/null 2>&1 && [ -t 1 ]; then
    RED=$(tput setaf 1)
    GREEN=$(tput setaf 2)
    YELLOW=$(tput setaf 3)
    RESET=$(tput sgr0)
else
    RED=""
    GREEN=""
    YELLOW=""
    RESET=""
fi

# Logging functions
log_info() {
    printf "%s[INFO]%s %s\n" "${GREEN}" "${RESET}" "$1" >&2
}

log_warn() {
    printf "%s[WARN]%s %s\n" "${YELLOW}" "${RESET}" "$1" >&2
}

log_error() {
    printf "%s[ERROR]%s %s\n" "${RED}" "${RESET}" "$1" >&2
}

# Detect OS
uname_os() {
    os=$(uname -s | tr '[:upper:]' '[:lower:]')
    case "$os" in
        msys_nt*) os="windows" ;;
        mingw*) os="windows" ;;
        cygwin*) os="windows" ;;
        win*) os="windows" ;;
    esac
    echo "$os"
}

# Detect architecture
uname_arch() {
    arch=$(uname -m)
    case $arch in
        x86_64) arch="x86_64" ;;
        x64) arch="x86_64" ;;
        i686) arch="386" ;;
        i386) arch="386" ;;
        aarch64) arch="arm64" ;;
        armv5*) arch="armv5" ;;
        armv6*) arch="armv6" ;;
        armv7*) arch="armv7" ;;
    esac
    echo "${arch}"
}

# Check if OS is supported
uname_os_check() {
    os=$(uname_os)
    case "$os" in
        darwin) return 0 ;;
        linux) return 0 ;;
        windows) return 0 ;;
    esac
    log_error "OS $os is not supported"
    log_error "Please open an issue: https://github.com/flipt-io/flipt/issues/new"
    return 1
}

# Check if architecture is supported
uname_arch_check() {
    arch=$(uname_arch)
    case "$arch" in
        x86_64) return 0 ;;
        arm64) return 0 ;;
    esac
    log_error "Architecture $arch is not supported"
    log_error "Please open an issue: https://github.com/flipt-io/flipt/issues/new"
    return 1
}

# Get latest v1 version from GitHub
get_latest_version() {
    tmpfile=$(mktemp)
    
    log_info "Fetching latest v1 release information..."
    
    # Use /releases endpoint to get all releases and filter for v1
    url="${RELEASE_URL_BASE}"
    
    if [ -n "$GITHUB_TOKEN" ]; then
        if ! curl -fsSL -H "Authorization: Bearer $GITHUB_TOKEN" "$url" > "$tmpfile" 2>/dev/null; then
            rm -f "$tmpfile"
            log_error "Failed to fetch releases from GitHub API"
            log_error "You can specify a version manually: curl -fsSL https://github.com/flipt-io/flipt/raw/main/install.sh | sh -s -- v1.45.0"
            return 1
        fi
    else
        if ! curl -fsSL "$url" > "$tmpfile" 2>/dev/null; then
            rm -f "$tmpfile"
            log_error "Failed to fetch releases from GitHub API"
            log_error "You can specify a version manually: curl -fsSL https://github.com/flipt-io/flipt/raw/main/install.sh | sh -s -- v1.45.0"
            return 1
        fi
    fi
    
    # Extract v1.x.x stable releases (excluding prereleases and drafts)
    # Process each release object and filter for stable v1 releases
    version=$(awk '
    BEGIN { 
        in_release = 0
        tag_name = ""
        is_prerelease = "true"
        is_draft = "true"
    }
    /{/ { in_release = 1 }
    /"tag_name":/ { 
        gsub(/.*"tag_name": *"/, "", $0)
        gsub(/".*/, "", $0)
        tag_name = $0
    }
    /"prerelease":/ {
        gsub(/.*"prerelease": */, "", $0)
        gsub(/,.*/, "", $0)
        is_prerelease = $0
    }
    /"draft":/ {
        gsub(/.*"draft": */, "", $0)
        gsub(/,.*/, "", $0)
        is_draft = $0
    }
    /}/ { 
        if (in_release && tag_name ~ /^v1\./ && is_prerelease == "false" && is_draft == "false") {
            print tag_name
            exit
        }
        in_release = 0
        tag_name = ""
        is_prerelease = "true"
        is_draft = "true"
    }
    ' "$tmpfile")
    
    rm -f "$tmpfile"
    
    # Validate that we found a version
    if [ -z "$version" ]; then
        log_error "No stable v1.x.x releases found"
        log_error "You can specify a version manually: curl -fsSL https://github.com/flipt-io/flipt/raw/main/install.sh | sh -s -- v1.45.0"
        return 1
    fi
    
    log_info "Found latest v1 version: $version"
    echo "$version"
    return 0
}

# Normalize version (ensure it starts with v)
normalize_version() {
    version="$1"
    case "$version" in
        v*) echo "$version" ;;
        *) echo "v$version" ;;
    esac
}

# Validate that version is a v1.x.x release
validate_v1_version() {
    version="$1"
    case "$version" in
        v1.*)
            return 0
            ;;
        *)
            log_error "Version $version is not a v1.x.x release"
            log_error "This installer only supports v1.x.x releases"
            return 1
            ;;
    esac
}

# Get version to install
get_version() {
    version=""
    
    # Priority order: 1. Environment variable, 2. Command line argument, 3. Latest
    if [ -n "$VERSION" ]; then
        version="$VERSION"
        log_info "Using version from environment: $version"
    elif [ -n "$1" ]; then
        version="$1"
        log_info "Using version from argument: $version"
    else
        version=$(get_latest_version) || return 1
    fi
    
    version=$(normalize_version "$version")
    validate_v1_version "$version" || return 1
    
    echo "$version"
}

# Install binary to appropriate location
install_binary() {
    bin_dir=""
    flipt_bin="/tmp/flipt"
    
    # Determine installation directory
    if [ -n "$BIN_DIR" ]; then
        bin_dir="$BIN_DIR"
    elif echo ":$PATH:" | grep -q ":$HOME/.local/bin:"; then
        bin_dir="$HOME/.local/bin"
    else
        bin_dir="/usr/local/bin"
    fi
    
    log_info "Installing flipt to $bin_dir"
    
    # Create directory if it doesn't exist
    if [ ! -d "$bin_dir" ]; then
        if ! mkdir -p "$bin_dir" 2>/dev/null; then
            log_warn "Cannot create $bin_dir as current user, trying with sudo"
            sudo mkdir -p "$bin_dir"
        fi
    fi
    
    # Install binary
    if cp "$flipt_bin" "$bin_dir/flipt" 2>/dev/null && chmod +x "$bin_dir/flipt" 2>/dev/null; then
        : # Success
    else
        log_warn "Cannot install to $bin_dir as current user, trying with sudo"
        if ! sudo install "$flipt_bin" "$bin_dir/flipt"; then
            log_error "Failed to install binary"
            return 1
        fi
    fi
    
    # Clean up
    rm -f "$flipt_bin"
    
    log_info "Successfully installed flipt to $bin_dir/flipt"
}

# Download and install
install_flipt() {
    version=""
    os=""
    arch=""
    archive_url=""
    archive_name=""
    
    # Get version
    version=$(get_version "$1") || return 1
    
    # Check platform support
    uname_os_check || return 1
    uname_arch_check || return 1
    
    # Get platform details
    os=$(uname_os)
    arch=$(uname_arch)
    
    log_info "Installing Flipt $version for $os/$arch"
    
    # Construct download URL
    archive_name="${os}_${arch}.tar.gz"
    archive_url="${DOWNLOAD_URL_BASE}/${version}/${archive_name}"
    
    log_info "Downloading from: $archive_url"
    
    # Download archive
    if ! curl -fsSL -o "/tmp/flipt.tar.gz" "$archive_url"; then
        log_error "Failed to download $archive_url"
        return 1
    fi
    
    # Extract archive
    log_info "Extracting archive..."
    if ! tar -xzf "/tmp/flipt.tar.gz" -C "/tmp"; then
        log_error "Failed to extract archive"
        return 1
    fi
    
    # Clean up archive
    rm -f "/tmp/flipt.tar.gz"
    
    # Make binary executable
    chmod +x "/tmp/flipt"
    
    # Install binary
    install_binary
    
    # Success message
    echo ""
    log_info "${GREEN}Flipt $version installed successfully!${RESET}"
    echo ""
    echo "Get started with: flipt --help"
    echo ""
    echo "For feedback and support:"
    echo "  Discord: https://flipt.io/discord"
    echo "  GitHub: https://github.com/flipt-io/flipt"
    echo "  Email: dev@flipt.io"
    echo ""
}

# Main execution
install_flipt "$1"