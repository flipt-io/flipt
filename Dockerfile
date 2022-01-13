ARG GO_VERSION=1.16

FROM golang:${GO_VERSION}

RUN apt-get update && \
    export DEBIAN_FRONTEND=noninteractive \
    apt-get -y --install --no-install-recommends \
    curl \
    gnupg \ 
    sudo \
    openssh-server \
    postgresql-client

RUN curl -sS https://dl.yarnpkg.com/debian/pubkey.gpg | apt-key add - && \
    echo "deb https://dl.yarnpkg.com/debian/ stable main" | tee /etc/apt/sources.list.d/yarn.list

RUN curl -sL https://deb.nodesource.com/setup_16.x | bash

RUN apt-get update && \
    apt-get install -y nodejs yarn && \
    apt-get clean -y

WORKDIR /flipt

COPY go.mod go.mod
COPY go.sum go.sum
RUN go mod download

COPY . .

RUN make bootstrap

EXPOSE 8080
EXPOSE 8081
EXPOSE 9000
