ARG GO_VERSION=1.20

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

# install nodejs
RUN curl -sSL https://deb.nodesource.com/setup_18.x | bash && \
    apt-get update && \
    apt-get install -y --no-install-recommends \
    nodejs && \
    apt-get clean && \
    rm -rf /var/lib/apt/lists/*

# install mage
RUN git clone https://github.com/magefile/mage && \
    cd mage && \
    go run bootstrap.go

WORKDIR /flipt

COPY go.mod go.mod
COPY go.sum go.sum

COPY . .

EXPOSE 8080
EXPOSE 9000
