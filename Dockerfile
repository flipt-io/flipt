FROM golang:1.22-alpine3.19 AS build

WORKDIR /home/flipt

RUN apk add --update --no-cache npm git bash gcc build-base binutils-gold

RUN git clone https://github.com/magefile/mage && \
    cd mage && \
    go run bootstrap.go

COPY go.mod .
COPY go.sum .
COPY ./errors ./errors
COPY ./rpc/flipt ./rpc/flipt
COPY ./sdk ./sdk
COPY ./core ./core

RUN go mod download

COPY . /home/flipt

ENV CGO_ENABLED=1

RUN mage bootstrap && \
    mage build

FROM alpine:3.19

LABEL maintainer="dev@flipt.io"
LABEL org.opencontainers.image.name="flipt"
LABEL org.opencontainers.image.source="https://github.com/flipt-io/flipt"

RUN apk add --update --no-cache postgresql-client \
    openssl \
    ca-certificates

RUN mkdir -p /etc/flipt && \
    mkdir -p /var/opt/flipt && \
    mkdir -p /var/log/flipt

COPY --from=build /home/flipt/bin/flipt /
COPY config/*.yml /etc/flipt/config/

RUN addgroup flipt && \
    adduser -S -D -g '' -G flipt -s /bin/sh flipt && \
    chown -R flipt:flipt /etc/flipt /var/opt/flipt /var/log/flipt

EXPOSE 8080
EXPOSE 9000

USER flipt

CMD ["./flipt"]
