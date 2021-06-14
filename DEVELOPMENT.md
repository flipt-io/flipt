# Development

The following are instructions for setting up your local machine for Flipt development. For info on using VSCode Remote Containers / GitHub Codespaces, see [#remote containers](#remote-containers) below.

## Requirements

Before starting, make sure you have the following installed:

- GCC Compiler
- [SQLite](https://sqlite.org/index.html)
- [Go 1.16+](https://golang.org/doc/install)
- [Protoc Compiler](https://github.com/protocolbuffers/protobuf)

## Setup

1. Clone this repo: `git clone https://github.com/markphelps/flipt`
1. Run `make bootstrap` to install required development tools
1. Run `make test` to execute the test suite
1. Run `make server` to build and run the server locally
1. Run `cd ui && yarn run dev` to build and run the ui dev server
1. Run `make help` to see a full list of possible make commands

## Go

Flipt is built with Go 1.16+. To reliably build Flipt, make sure you clone it to a location outside of your `$GOPATH`.

## Configuration

Configuration for running when developing Flipt can be found at `./config/local.yml`. To run Flipt with this configuration, run:

```shell
make dev
```

## Changes

Changing certain types of files such as the protobuf, ui or documentation files require re-building before they will be picked up in new versions of the binary.

### Updating .proto Files

After changing `flipt.proto`, you'll need to run `make proto`. This will regenerate the following files:

- `rpc/flipt.pb.go`
- `rpc/flipt.pb.gw.go`

### Updating assets

Running `make assets` will regenerate the embedded assets (ui, api documentation).

#### UI components

The UI is built using [Yarn](https://yarnpkg.com/en/) and [webpack](https://webpack.js.org/) and is also statically compiled into the Flipt binary.

The [ui/README.md](https://github.com/markphelps/flipt/tree/master/ui/README.md) has more information on how to build the UI and also how to run it locally during development.

## Remote Containers

Flipt now supports [VSCode Remote Containers](https://github.com/Microsoft/vscode-dev-containers)/[GitHub Codespaces](https://github.com/features/codespaces) as of #464/#465 respectively.

These technologies allow you to quickly get setup with a Flipt development environment either locally or 'in the cloud'.

For VSCode Remote Containers (devcontainers), make sure you have [Docker](https://www.docker.com/get-started) and the [ms-vscode-remote.remote-containers](https://marketplace.visualstudio.com/items?itemName=ms-vscode-remote.remote-containers) extension installed. Then simply clone this repo, open it in VSCode and run the [`Remote-Containers: Open Folder in Container`](https://code.visualstudio.com/docs/remote/containers#_quick-start-open-an-existing-folder-in-a-container) command in VSCode.

If you have access to [GitHub Codespaces](https://github.com/features/codespaces), simply open Flipt in a codespaces from the `Code` tab in the repo on GitHub.

### Building/Running

Regardless of wether you are using Remote Containers or GitHub Codespaces, you'll need to run a couple tasks to successfully build/run Flipt in these environments.

Then run the [Server/Client](.vscode/launch.json) run command to start the server and client.

This should publish three ports `8080`, `8081`, `9000`. `8081` is the UI dev port that runs the `yarn dev server` that you can open in your browser.
