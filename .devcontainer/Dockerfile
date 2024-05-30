ARG GO_VERSION=1.22

FROM golang:${GO_VERSION}

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

RUN git config --global --add safe.directory '*'

# install mage
RUN git clone https://github.com/magefile/mage && \
    cd mage && \
    go run bootstrap.go

WORKDIR /flipt

COPY go.mod go.mod
COPY go.sum go.sum
COPY ./core ./core
COPY ./errors ./errors
COPY ./rpc/flipt ./rpc/flipt
COPY ./sdk ./sdk

RUN go mod download -x

COPY . .

ENV GOPATH=/workspace/go \
    PATH=/workspace/go/bin:$PATH

# set the sqlite DB path to the user's home because of potential permission issues
ENV FLIPT_DB_URL=file:$HOME/flipt/flipt.db

EXPOSE 8080
EXPOSE 9000
