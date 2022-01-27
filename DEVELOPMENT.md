# Development

The following are instructions for setting up your local machine for Flipt development. For info on using VSCode Remote Containers / GitHub Codespaces, see [#remote containers](#remote-containers) below.

## Requirements

Before starting, make sure you have the following installed:

- GCC Compiler
- [SQLite](https://sqlite.org/index.html)
- [Go 1.16+](https://golang.org/doc/install)
- [Buf](https://docs.buf.build/introduction)
- [Protoc Compiler](https://github.com/protocolbuffers/protobuf)
- [NodeJS >= 16](https://nodejs.org/en/)
- [Yarn](https://yarnpkg.com/en/)

## Setup

1. Clone this repo: `git clone https://github.com/markphelps/flipt`
1. Run `make bootstrap` to install required development tools
1. Run `make test` to execute the test suite
1. Run `modd` to run the server and ui in development mode.
1. Run `make help` to see a full list of possible make commands

## Go

Flipt is built with Go 1.16+.

## Configuration

Configuration for running when developing Flipt can be found at `./config/local.yml`. To run Flipt with this configuration, run:

```shell
make server
```

## Changes

Changing certain types of files such as the protobuf, ui or documentation files require re-building before they will be picked up in new versions of the binary.

### Updating .proto Files

After changing `flipt.proto`, you'll need to run `make generate`. This will regenerate the following files:

- `rpc/flipt/flipt.pb.go`
- `rpc/flipt/flipt_grpc.pb.go`
- `rpc/flipt/flipt.pb.gw.go`

### Updating assets

Running `make assets` will regenerate the embedded assets (ui, api documentation).

#### UI components

The UI is built using [Yarn](https://yarnpkg.com/en/) and [webpack](https://webpack.js.org/) and is also statically compiled into the Flipt binary.

The [ui/README.md](https://github.com/markphelps/flipt/tree/master/ui/README.md) has more information on how to build the UI and also how to run it locally during development.

## Remote Containers/GitHub Codespaces

Flipt now supports [VSCode Remote Containers](https://github.com/Microsoft/vscode-dev-containers)/[GitHub Codespaces](https://github.com/features/codespaces).

These technologies allow you to quickly get setup with a Flipt development environment either locally or 'in the cloud'.

For VSCode Remote Containers (devcontainers), make sure you have [Docker](https://www.docker.com/get-started) and the [ms-vscode-remote.remote-containers](https://marketplace.visualstudio.com/items?itemName=ms-vscode-remote.remote-containers) extension installed. Then simply clone this repo, open it in VSCode and run the [`Remote-Containers: Open Folder in Container`](https://code.visualstudio.com/docs/remote/containers#_quick-start-open-an-existing-folder-in-a-container) command in VSCode.

If you have access to [GitHub Codespaces](https://github.com/features/codespaces), simply open Flipt in a codespaces from the `Code` tab in the repo on GitHub.

### Building/Running

Flipt uses [modd](https://github.com/cortesi/modd) for managing processes during development.

Run `modd` from the project root. This will intelligently rebuild/restart the backend server if any `*.go` files change which is helpful while developing. See [modd.conf](modd.conf) for configuration.

The `webpack-dev-server` that is used when running the UI in development mode will also rebuild the UI assets when applicable files in the `ui` folder change. See [ui/README.md](https://github.com/markphelps/flipt/tree/master/ui/README.md) for more info.

### Ports

The three ports `8080`, `8081`, `9000` will be forwarded to your local machine automatically.

`8081` is the UI dev port that runs the `yarn dev server` that you can open in your browser.
