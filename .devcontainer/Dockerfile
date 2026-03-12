FROM mcr.microsoft.com/devcontainers/base:ubuntu-24.04

ENV DEBIAN_FRONTEND=noninteractive

RUN apt-get update && apt-get install -y --no-install-recommends \
    bash \
    build-essential \
    ca-certificates \
    curl \
    git \
    pkg-config \
  && rm -rf /var/lib/apt/lists/*

USER vscode

ENV PATH="/home/vscode/.local/bin:${PATH}"

# Install mise
RUN curl https://mise.run | sh && /home/vscode/.local/bin/mise --version

# Disable GPG verification for Node.js (gpg/keyboxd not available in containers)
RUN /home/vscode/.local/bin/mise settings set node.gpg_verify false

# Pre-create volume mount points so Docker initializes them with correct ownership
RUN mkdir -p /home/vscode/.local/share/mise \
    /home/vscode/go/pkg \
    /home/vscode/.cache/go-build
