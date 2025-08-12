#!/bin/sh
#
#     _________       __ 
#    / ____/ (_)___  / /_
#   / /_  / / / __ \/ __/
#  / __/ / / / /_/ / /_  
# /_/   /_/_/ .___/\__/  
#          /_/           
# (v2)
#
# Flipt installer
#
# Usage:
#   curl -fsSL https://github.com/flipt-io/flipt/raw/v2/install.sh | sh
set -e

file_issue_prompt() {
    echo "If you wish us to support your platform, please file an issue"
    echo "https://github.com/flipt-io/flipt/issues/new?labels=v2"
    exit 1
}

get_latest_version() {
    # Get the latest stable v2.x.x release using the simplest reliable approach
    # Use /releases/latest endpoint to avoid rate limiting issues
    
    result=""
    
    # Method 1: Check if latest release is a v2.x version (simplest and most reliable)
    if command -v jq >/dev/null 2>&1; then
        latest=$(curl -fsSL "https://api.github.com/repos/flipt-io/flipt/releases/latest" 2>/dev/null | \
                 jq -r '.tag_name' 2>/dev/null || echo "")
        case "$latest" in
            v2.*)
                result="$latest"
                ;;
        esac
    fi
    
    # Method 2: If jq failed, try grep approach on latest release
    if [ -z "$result" ]; then
        latest=$(curl -fsSL "https://api.github.com/repos/flipt-io/flipt/releases/latest" 2>/dev/null | \
                 grep '"tag_name"' | cut -d '"' -f 4 | grep '^v2\.' || echo "")
        if [ -n "$latest" ]; then
            result="$latest"
        fi
    fi
    
    # Method 3: Fallback - search recent releases for v2.x (if latest isn't v2.x)
    if [ -z "$result" ] && command -v jq >/dev/null 2>&1; then
        result=$(curl -fsSL "https://api.github.com/repos/flipt-io/flipt/releases?per_page=10" 2>/dev/null | \
                 jq -r '.[] | select(.prerelease == false) | select(.tag_name | startswith("v2.")) | .tag_name' 2>/dev/null | \
                 head -1 || echo "")
    fi
    
    # Validate result
    if [ -z "$result" ]; then
        echo "Error: Unable to fetch v2.x.x stable release" >&2
        echo "Please check your internet connection or try again later" >&2
        exit 1
    fi
    
    echo "$result"
}

copy() {
    case ":$PATH:" in
        *":$HOME/.local/bin:"*)
            [ ! -d "$HOME/.local/bin" ] && mkdir -p "$HOME/.local/bin"
            mv /tmp/flipt "$HOME/.local/bin/flipt"
            ;;
        *)
            if ! mv /tmp/flipt /usr/local/bin/flipt; then
                echo "Cannot write to installation target directory as current user, writing as root."
                sudo mv /tmp/flipt /usr/local/bin/flipt
            fi
            ;;
    esac 
}

# This function decides what version will be installed based on the following priority:
# 1. Environment variable `VERSION` is set.
# 2. Latest available on GitHub
get_version() {
    if [ -z "$VERSION" ]; then
        if [ -n "$1" ]; then
            VERSION="$1"
        else
            VERSION=$(get_latest_version)
        fi
    fi
  
    # ensure version starts with v
    case "$VERSION" in
        v*) 
            : ;;
        *) 
            VERSION="v$VERSION" ;;
    esac
    echo "$VERSION"
}

install() {
    version=$(get_version "$1")
    echo "Installing version $version"

    OSCHECK=$(uname | tr '[:upper:]' '[:lower:]')
  
    case "$OSCHECK" in 
        linux*)
            ARCH=$(uname -m)
            OS="linux"
            case "$ARCH" in 
                x86_64|aarch64)
                    : ;; 
                *)
                    echo "flipt is only available for linux x86_64/arm64 architecture"
                    file_issue_prompt
                    ;; 
            esac 
            [ "$ARCH" = "aarch64" ] && ARCH="arm64"
            ;;
        darwin*) 
            ARCH=$(uname -m)
            OS="darwin"
            case "$ARCH" in 
                x86_64|arm64)
                    : ;;
                *) 
                    echo "flipt is only available for mac x86_64/arm64 architecture"
                    file_issue_prompt
                    ;;
            esac 
            [ "$ARCH" = "aarch64" ] && ARCH="arm64"   
            ;;
        *)
            echo "flipt isn't supported for your platform - $OSCHECK"
            file_issue_prompt
            ;;
    esac

    curl -o /tmp/flipt.tar.gz -fsSL "https://download.flipt.io/flipt/$version/${OS}_${ARCH}.tar.gz"
    tar -xzf /tmp/flipt.tar.gz -C /tmp
    chmod +x /tmp/flipt
    copy
    echo ""
    echo "Flipt installed!"
    echo ""
    echo "For feedback and support, join our Discord server: https://flipt.io/discord,"
    echo "open an issue or discussion on GitHub: https://github.com/flipt-io/flipt/"
    echo "or send us an email at dev@flipt.io"
    echo ""
}

install "$1"