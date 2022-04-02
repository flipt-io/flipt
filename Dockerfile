ARG GO_VERSION=1.17

FROM golang:${GO_VERSION}

SHELL ["/bin/bash", "-o", "pipefail", "-c"]

RUN apt-get update && \
    apt-get -y install --no-install-recommends \
    curl \
    gnupg \ 
    sudo \
    openssh-server \
    postgresql-client && \
    apt-get clean && \
    rm -rf /var/lib/apt/lists/*

# install yarn & nodejs
RUN curl -sSL https://dl.yarnpkg.com/debian/pubkey.gpg | apt-key add - && \
    echo "deb https://dl.yarnpkg.com/debian/ stable main" | tee /etc/apt/sources.list.d/yarn.list

RUN curl -sSL https://deb.nodesource.com/setup_16.x | bash && \
    apt-get update && \
    apt-get install -y --no-install-recommends \
    nodejs \
    yarn && \
    apt-get clean && \
    rm -rf /var/lib/apt/lists/*

# install helm
RUN curl -sSL https://baltocdn.com/helm/signing.asc | sudo apt-key add - && \
    sudo apt-get install -y --no-install-recommends apt-transport-https && \
    echo "deb https://baltocdn.com/helm/stable/debian/ all main" | sudo tee /etc/apt/sources.list.d/helm-stable-debian.list && \
    apt-get update && \
    apt-get install -y --no-install-recommends helm

# install task
RUN sh -c "$(curl --location https://taskfile.dev/install.sh)" -- -d -b /usr/local/bin

WORKDIR /flipt

COPY go.mod go.mod
COPY go.sum go.sum
RUN go mod download

COPY . .

EXPOSE 8080
EXPOSE 8081
EXPOSE 9000
