ARG GO_VERSION=1.16

FROM golang:${GO_VERSION}-alpine AS build

RUN apk add --no-cache \
    bash \
    curl \
    gcc \
    git \
    make \
    musl-dev \
    nodejs \
    openssl \
    postgresql-client \
    yarn

WORKDIR /flipt

COPY go.mod go.mod
COPY go.sum go.sum
RUN go mod download

COPY . .

RUN make bootstrap

EXPOSE 8080
EXPOSE 9000

CMD ["make", "dev"]
