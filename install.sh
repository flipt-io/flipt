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
    # Get the latest stable v2.x.x release from GitHub API
    # We need semantic version sorting to handle backports correctly
    
    # Debug: Show environment info
    echo "DEBUG: GITHUB_TOKEN is $([ -n "$GITHUB_TOKEN" ] && echo "set (${#GITHUB_TOKEN} chars)" || echo "not set")" >&2
    echo "DEBUG: jq available: $(command -v jq >/dev/null 2>&1 && echo "yes" || echo "no")" >&2
    
    # Try using jq if available for better JSON parsing
    if command -v jq >/dev/null 2>&1; then
        # Get all stable v2.x releases and sort them semantically
        if [ -n "$GITHUB_TOKEN" ]; then
            echo "DEBUG: Using authenticated API call" >&2
            api_response=$(curl -fsSL -H "Authorization: Bearer $GITHUB_TOKEN" "https://api.github.com/repos/flipt-io/flipt/releases?per_page=100" 2>&1)
            curl_exit_code=$?
            echo "DEBUG: curl exit code: $curl_exit_code" >&2
            echo "DEBUG: API response length: ${#api_response}" >&2
            echo "DEBUG: First 500 chars of response: ${api_response:0:500}" >&2
            releases=$(echo "$api_response" | jq -r '.[] | select(.prerelease == false) | select(.tag_name | startswith("v2.")) | .tag_name' 2>/dev/null)
        else
            echo "DEBUG: Using unauthenticated API call" >&2
            api_response=$(curl -fsSL "https://api.github.com/repos/flipt-io/flipt/releases?per_page=100" 2>&1)
            curl_exit_code=$?
            echo "DEBUG: curl exit code: $curl_exit_code" >&2
            echo "DEBUG: API response length: ${#api_response}" >&2
            echo "DEBUG: First 500 chars of response: ${api_response:0:500}" >&2
            releases=$(echo "$api_response" | jq -r '.[] | select(.prerelease == false) | select(.tag_name | startswith("v2.")) | .tag_name' 2>/dev/null)
        fi
    else
        echo "DEBUG: Using fallback method (no jq)" >&2
        # Fallback: Get all v2.x.x releases and filter out pre-releases
        if [ -n "$GITHUB_TOKEN" ]; then
            echo "DEBUG: Using authenticated fallback API call" >&2
            releases=$(curl -fsSL -H "Authorization: Bearer $GITHUB_TOKEN" "https://api.github.com/repos/flipt-io/flipt/releases?per_page=100" 2>/dev/null | \
                       grep '"tag_name"' | \
                       cut -d '"' -f 4 | \
                       grep '^v2\.' | \
                       grep -v -E 'alpha|beta|rc|pre|dev')
        else
            echo "DEBUG: Using unauthenticated fallback API call" >&2
            releases=$(curl -fsSL "https://api.github.com/repos/flipt-io/flipt/releases?per_page=100" 2>/dev/null | \
                       grep '"tag_name"' | \
                       cut -d '"' -f 4 | \
                       grep '^v2\.' | \
                       grep -v -E 'alpha|beta|rc|pre|dev')
        fi
    fi
    
    echo "DEBUG: Found releases:" >&2
    echo "$releases" >&2
    echo "DEBUG: Number of releases found: $(echo "$releases" | wc -l)" >&2
    
    # Check if curl failed or no releases were found
    if [ -z "$releases" ]; then
        echo "Error: Unable to fetch releases from GitHub or no v2.x.x stable release found" >&2
        echo "Please check your internet connection or try again later" >&2
        exit 1
    fi
    
    # Sort versions semantically
    # Remove 'v' prefix, sort by major.minor.patch numerically, add 'v' back
    res=$(echo "$releases" | sed 's/^v//' | sort -t. -k1,1n -k2,2n -k3,3n | tail -1 | sed 's/^/v/')
    
    echo "$res"
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
