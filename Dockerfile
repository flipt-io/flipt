# https://goreleaser.com/docker/

FROM golang:1.18-alpine3.16 AS build

WORKDIR /home/flipt

RUN apk add npm git bash gcc build-base binutils-gold

RUN npm install -g @go-task/cli

COPY go.mod .
COPY go.sum .

RUN go mod download

COPY . /home/flipt

ENV CGO_ENABLED=1

RUN task

FROM alpine:3.16.2

LABEL maintainer="dev@flipt.io"
LABEL org.opencontainers.image.name="flipt"
LABEL org.opencontainers.image.source="https://github.com/flipt-io/flipt"

RUN apk add --no-cache postgresql-client \
    openssl \
    ca-certificates

RUN mkdir -p /etc/flipt && \
    mkdir -p /var/opt/flipt

COPY --from=build /home/flipt/bin/flipt /
COPY config/migrations/ /etc/flipt/config/migrations/
COPY config/*.yml /etc/flipt/config/

RUN addgroup flipt && \
    adduser -S -D -g '' -G flipt -s /bin/sh flipt && \
    chown -R flipt:flipt /etc/flipt /var/opt/flipt

EXPOSE 8080
EXPOSE 9000

USER flipt

CMD ["./flipt"]
