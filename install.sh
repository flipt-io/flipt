#!/bin/bash
# flipt installer
#     _________       __ 
#    / ____/ (_)___  / /_
#   / /_  / / / __ \/ __/
#  / __/ / / / /_/ / /_  
# /_/   /_/_/ .___/\__/  
#          /_/           
#
# Usage:
#   curl -fsSL https://github.com/flipt-io/flipt/raw/latest/install.sh | bash
set -e

file_issue_prompt() {
  echo "If you wish us to support your platform, please file an issue"
  echo "https://github.com/flipt-io/flipt/issues/new"
  exit 1
}

get_latest_version() {
  local res=$(curl -fsSL https://api.github.com/repos/flipt-io/flipt/releases/latest | grep tag_name | cut -d '"' -f 4)
  echo "$res"
}

copy() {
  if [[ ":$PATH:" == *":$HOME/.local/bin:"* ]]; then
      if [ ! -d "$HOME/.local/bin" ]; then
        mkdir -p "$HOME/.local/bin"
      fi
      mv /tmp/flipt "$HOME/.local/bin/flipt"
  else
      # Try without sudo first, run with sudo only if mv failed without it.
      mv /tmp/flipt /usr/local/bin/flipt || (
        echo "Cannot write to installation target directory as current user, writing as root."
        sudo mv /tmp/flipt /usr/local/bin/flipt
      )
  fi
}

# This function decides what version will be installed based on the following priority:
# 1. Environment variable `VERSION` is set.
# 2. Command line argument is passed.
# 3. Latest available on GitHub
function get_version() {
  if [[ -z "$VERSION" ]]; then
      if [[ -n "$1" ]]; then
          VERSION="$1"
      else
          VERSION=$(get_latest_version)
      fi
  fi
  # ensure version starts with v
  if [[ "$VERSION" != v* ]]; then
    VERSION="v$VERSION"
  fi
  echo "$VERSION"
}

function install() {
  local version=$(get_version $1);
  echo "Installing version $version"
  if [[ "$OSTYPE" == "linux"* ]]; then
      ARCH=$(uname -m);
      OS="linux";
      if [[ "$ARCH" != "x86_64" && "$ARCH" != "aarch64" ]]; then
          echo "flipt is only available for linux x86_64/aarch64 architecture"
          file_issue_prompt
          exit 1
      fi
      if [[ "$ARCH" == "aarch64" ]]; then
          ARCH="arm64"
      fi
  elif [[ "$OSTYPE" == "darwin"* ]]; then
      ARCH=$(uname -m);
      OS="darwin";
      if [[ "$ARCH" != "arm64" ]]; then
          echo "flipt is only available for mac arm64 architecture"
          file_issue_prompt
          exit 1
      fi
  else
      echo "flipt isn't supported for your platform - $OSTYPE"
      file_issue_prompt
      exit 1
  fi
  curl -o /tmp/flipt.tar.gz -fsSL https://github.com/flipt-io/flipt/releases/download/$version/flipt_$OS\_$ARCH.tar.gz
  tar -xzf /tmp/flipt.tar.gz -C /tmp
  chmod +x /tmp/flipt
  copy
  echo ""
  echo "Flipt installed!"
  echo ""
  echo "For feedback and support, join our Discord server: https://flipt.io/discord, open an issue or discussion on our GitHub: https://github.com/flipt-io/flipt/ or send us an email at dev@flipt.io"
  echo ""
  }

install "$1"
