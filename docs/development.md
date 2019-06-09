# Development

The following are instructions for setting up your machine for Flipt development.

## Requirements

Before starting, make sure you have the following installed:

* GCC Compiler
* [SQLite](https://sqlite.org/index.html)
* [Go 1.12+](https://golang.org/doc/install)
* [Protoc Compiler](https://github.com/protocolbuffers/protobuf)

## Setup

1. Clone this repo: `git clone https://github.com/markphelps/flipt`
1. Run `make setup` to download dependencies
1. Run `make test` to execute the test suite
1. Run `make dev` to build and run in development mode
1. Run `make help` to see a full list of possible make commands

## Go Modules

Flipt uses Go 1.12 with [Go Modules](https://github.com/golang/go/wiki/Modules) enabled. To reliably build Flipt, make sure you clone it to a location outside of your `$GOPATH` or set the environment variable `GO111MODULE=on`. For more info see: [https://github.com/golang/go/wiki/Modules#how-to-install-and-activate-module-support](https://github.com/golang/go/wiki/Modules#how-to-install-and-activate-module-support).

## Vagrant

You can also easily get started with a development environment running in a VM using [Vagrant](https://www.vagrantup.com/) and [Virtual Box](https://www.virtualbox.org/wiki/Downloads).

Once you have Vagrant and Virtual Box installed you can `cd` into either the `build/dev/ubuntu` or `build/dev/centos` directories and run `vagrant up`.

This will provision a VM that installs the necessary dev dependencies and runs the Flipt test suite.

Once the provisioning process is complete, run:

```shell
$ vagrant ssh
$ cd ~/app/flipt
$ make dev
```

This will run Flipt in development mode inside your VM.

## Configuration

Configuration for running when developing Flipt can be found at `./config/local.yml`. To run Flipt with this configuration, run:

```shell
make dev
```

## Changes

Changing certain types of files such as the protobuf, ui or documentation files require re-building before they will be picked up in new versions of the binary.

### Updating .proto Files

After changing `flipt.proto`, you'll need to run `make proto`. This will regenerate the following files:

* `flipt.pb.go`
* `flipt.pb.gw.go`

### Updating assets

Running `make assets` will regenerate the embedded assets (ui, api documentation) so that the next time `make dev` is run they will be included.

#### UI components

The UI is built using [Yarn](https://yarnpkg.com/en/) and [webpack](https://webpack.js.org/) and is also statically compiled into the Flipt binary.

The [ui/README.md](https://github.com/markphelps/flipt/tree/master/ui/README.md) has more information on how to build the UI and also how to run it locally during development.
