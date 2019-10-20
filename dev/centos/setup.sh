#!/bin/bash
set -euxo pipefail

NODE_VERSION='12.10.0'
YARN_VERSION='1.17.3'
GO_VERSION='1.13.3'
PROTOC_VERSION='3.9.2'

echo '=== Installing necessary system libraries to build'

sudo yum update -y
sudo yum install -y \
    epel-release
    gcc gcc-c++ git make \
    postgresql \
    npm \
    unzip \
    wget

# install node and yarn
sudo npm cache clean -f
sudo npm install -g n
sudo n "$NODE_VERSION"

curl -o- -L https://yarnpkg.com/install.sh | bash -s -- --version "$YARN_VERSION"
export PATH="$HOME/.yarn/bin:$PATH"

# install go and setup GOPATH if not installed
if [ ! -d /usr/local/go ]; then
    echo '=== Installing Go'
    wget -q "https://dl.google.com/go/go$GO_VERSION.linux-amd64.tar.gz"
    sudo tar -C /usr/local -xzf "go$GO_VERSION.linux-amd64.tar.gz"
    rm -f "go$GO_VERSION.linux-amd64.tar.gz"
    echo 'export PATH=/usr/local/go/bin:$HOME/go/bin:$PATH' >> ~/.bash_profile
    echo 'export GOPATH=$HOME/go' >> ~/.bash_profile
    export PATH=/usr/local/go/bin:$HOME/go/bin:$PATH
    export GOPATH=$HOME/go
    go version
fi

# install protoc and necessary libraries if not installed
if [ ! -f /usr/local/bin/protoc ]; then
    echo '=== Installing protoc'
    wget -q "https://github.com/protocolbuffers/protobuf/releases/download/v$PROTOC_VERSION/protoc-$PROTOC_VERSION-linux-x86_64.zip"
    sudo unzip "protoc-$PROTOC_VERSION-linux-x86_64.zip" -d /usr/local
    rm -f "protoc-$PROTOC_VERSION-linux-x86_64.zip"
    protoc --version
fi

# make /usr/local readable, required for protoc
sudo chmod -R 755 /usr/local/

mkdir -p ~/app
cd ~/app && ln -s /flipt flipt && cd flipt

echo '=== Installing dev tools'
make setup

echo '=== Running test suite'
make test

echo '=== Generating assets'
make assets

echo "=== Done. To run Flipt, run 'make dev' from $(pwd)"
