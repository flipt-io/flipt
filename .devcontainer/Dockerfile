FROM mcr.microsoft.com/devcontainers/base:ubuntu
ARG USERNAME=vscode
ENV HOME=/home/${USERNAME}

RUN apt-get update && \
  apt-get -y install --no-install-recommends \
  curl \
  gnupg \
  sudo \
  openssh-server \
  postgresql-client && \
  apt-get clean && \
  rm -rf /var/lib/apt/lists/*

RUN git config --global --add safe.directory '*'

EXPOSE 5173
EXPOSE 8080
EXPOSE 9000
