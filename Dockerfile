ARG GO_VERSION=1.16

FROM golang:$GO_VERSION-alpine AS build

RUN apk add --no-cache \
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

RUN make bootstrap clean assets

EXPOSE 8080
EXPOSE 9000

# Docker exposes container ports to the IP address 0.0.0.0
ENV FLIPT_SERVER_HOST 0.0.0.0

CMD ["go", "run", "./cmd/flipt/.", "--config", "./config/local.yml", "--force-migrate"]
