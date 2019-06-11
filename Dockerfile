FROM golang:1.12.5-alpine AS build

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
RUN go generate -v ./...
RUN go install -v ./cmd/flipt/

FROM alpine:3.9
LABEL maintainer="mark.aaron.phelps@gmail.com"

RUN apk add --no-cache \
    ca-certificates \
    openssl \
    postgresql-client

RUN mkdir -p /etc/flipt && \
    mkdir -p /var/opt/flipt

COPY --from=build /go/bin/flipt /
COPY config /etc/flipt/config

EXPOSE 8080
EXPOSE 9000

CMD ["./flipt"]
