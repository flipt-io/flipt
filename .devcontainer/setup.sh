#!/usr/bin/env bash
set -euo pipefail

# Install build dependencies
sudo apt-get update
sudo apt-get install -y --no-install-recommends build-essential pkg-config
sudo rm -rf /var/lib/apt/lists/*

# Fix ownership on volume mounts and mise state directory
sudo mkdir -p /home/vscode/.local/state/mise
sudo chown -R vscode:vscode \
  /home/vscode/.local/share/mise \
  /home/vscode/.local/state/mise \
  /home/vscode/go/pkg \
  /home/vscode/.cache/

# Install tools and dependencies via mise
mise settings set node.gpg_verify false
mise run bootstrap
mise run ui:deps
