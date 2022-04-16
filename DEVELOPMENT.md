# Development

The following are instructions for setting up your local machine for Flipt development. For info on using VSCode Remote Containers / GitHub Codespaces, see [#remote containers](#remote-containers) below.

## Requirements

Before starting, make sure you have the following installed:

- GCC Compiler
- [SQLite](https://sqlite.org/index.html)
- [Go 1.17+](https://golang.org/doc/install)
- [NodeJS >= 16](https://nodejs.org/en/)
- [Yarn](https://yarnpkg.com/en/)
- [Task](https://taskfile.dev/#/)

## Setup

1. Clone this repo: `git clone https://github.com/markphelps/flipt`
1. Run `task bootstrap` to install required development tools. See [#bootstrap](#bootstrap) below.
1. Run `task test` to execute the test suite
1. Run `task dev` to run the server and ui in development mode.
1. Run `task build` to build the binary with embedded assets.
1. Run `task --list-all` to see a full list of possible commands

## Go

Flipt is built with Go 1.17+.

## Bootstrap

The `bootstrap` task will install all of the necessary tools used for development and testing. It does this using a seperate tools modules as described here: https://marcofranssen.nl/manage-go-tools-via-go-modules

## Configuration

Configuration for running when developing Flipt can be found at `./config/local.yml`. To run Flipt with this configuration, run:

```shell
task dev
```

## Changes

Changing certain types of files such as the proto, ui or documentation files require re-building before they will be picked up in new versions of the binary.

### Updating .proto Files

After changing `flipt.proto`, you'll need to run `task build:proto`. This will regenerate the following files:

- `rpc/flipt/flipt.pb.go`
- `rpc/flipt/flipt_grpc.pb.go`
- `rpc/flipt/flipt.pb.gw.go`

### Updating assets

Running `task assets` will regenerate the embedded assets (ui, api documentation).

#### UI components

The UI is built using [Yarn](https://yarnpkg.com/en/) and [webpack](https://webpack.js.org/) and is also statically compiled into the Flipt binary.

The [ui/README.md](https://github.com/markphelps/flipt/tree/main/ui/README.md) has more information on how to build the UI and also how to run it locally during development.

## Remote Containers/GitHub Codespaces

Flipt now supports [VSCode Remote Containers](https://github.com/Microsoft/vscode-dev-containers)/[GitHub Codespaces](https://github.com/features/codespaces).

These technologies allow you to quickly get setup with a Flipt development environment either locally or 'in the cloud'.

For VSCode Remote Containers (devcontainers), make sure you have [Docker](https://www.docker.com/get-started) and the [ms-vscode-remote.remote-containers](https://marketplace.visualstudio.com/items?itemName=ms-vscode-remote.remote-containers) extension installed. Then simply clone this repo, open it in VSCode and run the [`Remote-Containers: Open Folder in Container`](https://code.visualstudio.com/docs/remote/containers#_quick-start-open-an-existing-folder-in-a-container) command in VSCode.

If you have access to [GitHub Codespaces](https://github.com/features/codespaces), simply open Flipt in a codespaces from the `Code` tab in the repo on GitHub.

### Building/Running

Flipt uses [modd](https://github.com/cortesi/modd) for managing processes during development.

Run `task dev` from the project root. See [modd.conf](modd.conf) for configuration.

The `webpack-dev-server` that is used when running the UI in development mode will rebuild the UI assets when applicable files in the `ui` folder change. See [ui/README.md](https://github.com/markphelps/flipt/tree/main/ui/README.md) for more info.

### Ports

The three ports `8080`, `8081`, `9000` will be forwarded to your local machine automatically.

`8081` is the UI dev port that runs the `yarn dev server` that you can open in your browser.
