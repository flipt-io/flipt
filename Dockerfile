FROM golang:1.20-alpine3.16

WORKDIR /server

RUN apk add --update --no-cache git bash gcc build-base binutils-gold \
    openssl \
    ca-certificates \
    postgresql-client

RUN git clone https://github.com/magefile/mage && \
    cd mage && \
    go run bootstrap.go

COPY go.mod .
COPY go.sum .

RUN go mod download -x

COPY ./errors ./errors
COPY ./rpc/flipt ./rpc/flipt
COPY ./sdk ./sdk

RUN mkdir -p /etc/flipt && \
    mkdir -p /var/opt/flipt

EXPOSE 8080
EXPOSE 9000