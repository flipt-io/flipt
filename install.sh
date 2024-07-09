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
set -e

file_issue_prompt() {
  echo "If you wish us to support your platform, please file an issue"
  echo "https://github.com/flipt-io/flipt/issues/new"
  exit 1
}

get_latest_version() {
  res=$(curl -fsSL https://api.github.com/repos/flipt-io/flipt/releases/latest | grep tag_name | cut -d '"' -f 4)
  echo "$res"
}

copy() {
  case ":PATH:" in 
    *":HOME/.local/bin:"*) 
      [ ! -d "$HOME/.local/bin" ] && mkdir -p "$HOME/.local/bin"
      mv /tmp/flipt "$HOME/.local/bin/flipt"
      ;;
    *)
      mv /tmp/flipt /usr/local/bin/flipt || (
      echo "Cannot write to installation target directory as current user, writing as root."
      sudo mv /tmp/flipt /usr/local/bin/flipt
    )
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
    v*) : ;;
    *) VERSION="v$VERSION" ;;
  esac
  echo "$VERSION"
}

install() {
  version=$(get_version "$1");
  echo "Installing version $version"

  OSCHECK=$(uname | tr '[:upper:]' '[:lower:]')
  
  case "$OSCHECK" in 
    linux*)
      ARCH=$(uname -m)
      OS="linux"
      case "$ARCH" in 
        x86_64|aarch64) : ;; 
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
        x86_64|arm64) : ;;
        *) 
          echo "flipt is only available for mac x86_64/arm64 architecture"
          file_issue_prompt
          ;;
      esac 
      [ "$ARCH" = "aarch64" ] && ARCH="arm64"   
      ;;
    *)
      echo "flipt isn't supported for your platform - $OSTYPE"
      file_issue_prompt
      ;;
  esac

  curl -o /tmp/flipt.tar.gz -fsSL https://download.flipt.io/flipt/$version/$OS\_$ARCH.tar.gz
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
