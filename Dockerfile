ARG GO_VERSION=1.16

FROM golang:${GO_VERSION}-alpine AS build

ENV NVM_VERSION=v0.38.0
ENV NODE_VERSION 12

RUN apk add --no-cache \
    bash \
    curl \
    gcc \
    git \
    make \
    musl-dev \
    openssl \
    postgresql-client


RUN curl -o- "https://raw.githubusercontent.com/nvm-sh/nvm/${NVM_VERSION}/install.sh" | /bin/bash && \
    . ~/.nvm/nvm.sh && \
    nvm install $NODE_VERSION

# WORKDIR /flipt

# COPY go.mod go.mod
# COPY go.sum go.sum
# RUN go mod download

# COPY . .

# RUN make bootstrap

# EXPOSE 8080
# EXPOSE 9000

# CMD ["make", "dev"]
