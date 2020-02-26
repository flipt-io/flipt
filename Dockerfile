ARG GO_VERSION=1.13.7

FROM golang:$GO_VERSION-alpine AS build

RUN apk add --no-cache \
    gcc \
    git \
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
RUN cd ./ui && yarn install && yarn run build
RUN go get -v github.com/gobuffalo/packr/packr
RUN packr -v -i cmd/flipt
RUN go install -v ./cmd/flipt/

FROM alpine:3.10
LABEL maintainer="mark.aaron.phelps@gmail.com"

RUN apk add --no-cache \
    ca-certificates \
    openssl \
    postgresql-client

RUN mkdir -p /etc/flipt && \
    mkdir -p /var/opt/flipt

COPY --from=build /go/bin/flipt /
COPY config/migrations/ /etc/flipt/config/migrations/
COPY config/*.yml /etc/flipt/config/

EXPOSE 8080
EXPOSE 9000

CMD ["./flipt"]
