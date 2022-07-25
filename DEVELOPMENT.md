# Development

The following are instructions for setting up your local machine for Flipt development. For info on using VSCode Remote Containers / GitHub Codespaces, see [#remote containers](#remote-containers) below.

## Requirements

Before starting, make sure you have the following installed:

- GCC Compiler
- [SQLite](https://sqlite.org/index.html)
- [Go 1.17+](https://golang.org/doc/install)
- [NodeJS >= 18](https://nodejs.org/en/)
- [Task](https://taskfile.dev/#/)
- [Docker](https://docs.docker.com/install/) (for running tests)

## Setup

1. Clone this repo: `git clone https://github.com/flipt-io/flipt`
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

`task server` or `task dev`

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

The UI is built using [NPM](https://nodejs.org/en/) and [Vite](https://vitejs.dev/) and is also statically compiled into the Flipt binary.

The [ui/README.md](https://github.com/flipt-io/flipt/tree/main/ui/README.md) has more information on how to build the UI and also how to run it locally during development.

## Building/Running

**Run `task dev` from the project root.**

Vite will rebuild the UI assets when applicable files in the `ui` folder change. See [ui/README.md](https://github.com/flipt-io/flipt/tree/main/ui/README.md) for more info.

You'll need to stop and re-run for any changes in the server (Go) code :exclamation:

### Ports

In development, the three ports that Flipt users are:

- `8080`: The port for the Flipt REST API
- `8081`: The port for the Flipt UI (via `npm run dev`)
- `9000`: The port for the Flipt GRPC Server

These three ports will be forwarded to your local machine automatically if you are developing Flipt in a VSCode Remote Container or GitHub Codespace.

## Remote Containers/GitHub Codespaces

Flipt now supports [VSCode Remote Containers](https://github.com/Microsoft/vscode-dev-containers)/[GitHub Codespaces](https://github.com/features/codespaces).

These technologies allow you to quickly get setup with a Flipt development environment either locally or 'in the cloud'.

For VSCode Remote Containers (devcontainers), make sure you have [Docker](https://www.docker.com/get-started) and the [ms-vscode-remote.remote-containers](https://marketplace.visualstudio.com/items?itemName=ms-vscode-remote.remote-containers) extension installed. Then simply clone this repo, open it in VSCode and run the [`Remote-Containers: Open Folder in Container`](https://code.visualstudio.com/docs/remote/containers#_quick-start-open-an-existing-folder-in-a-container) command in VSCode.

If you have access to [GitHub Codespaces](https://github.com/features/codespaces), simply open Flipt in a codespaces from the `Code` tab in the repo on GitHub or click the button below.

[![Open in Codespaces](https://github.com/codespaces/badge.svg)](https://github.com/codespaces/new/?repo=flipt-io/flipt)
