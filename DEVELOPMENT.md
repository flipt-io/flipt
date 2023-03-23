# Development

The following are instructions for setting up your local machine for Flipt development. For info on using VSCode Remote Containers / GitHub Codespaces, see [#cdes](#cdes) below.

## Requirements

Before starting, make sure you have the following installed:

- GCC Compiler
- [SQLite](https://sqlite.org/index.html)
- [Go 1.20+](https://golang.org/doc/install)
- [NodeJS >= 18](https://nodejs.org/en/)
- [Mage](https://magefile.org/)
- [Docker](https://docs.docker.com/install/) (for running tests)

## Setup

1. Clone this repo: `git clone https://github.com/flipt-io/flipt`
1. Run `mage bootstrap` to install required development tools. See [#bootstrap](#bootstrap) below.
1. Run `mage test` to execute the test suite
1. Run `mage build` to build the binary with embedded assets.
1. Run `mage -l` to see a full list of possible commands

## Go

Flipt is built with Go 1.20+.

## Bootstrap

The `bootstrap` task will install all of the necessary tools used for development and testing. It does this using a seperate tools modules as described here: [https://marcofranssen.nl/manage-go-tools-via-go-modules](https://marcofranssen.nl/manage-go-tools-via-go-modules)

## Configuration

Configuration for running when developing Flipt can be found at `./config/local.yml`. To run Flipt with this configuration, run:

`./bin/flipt --config ./config/local.yml`

## Changes

Changing certain types of files such as the proto or ui files require re-building before they will be picked up in new versions of the binary.

### Updating .proto Files

After changing `flipt.proto`, you'll need to run `mage proto`. This will regenerate the necessary files in the `rpc` directory.

## UI

The UI is built using [NPM](https://nodejs.org/en/) and [Vite](https://vitejs.dev/) and is also statically compiled into the Flipt binary using [go:embed](https://golang.org/pkg/embed/).

The UI code has been migrated to it's own repository: [flipt-ui](https://github.com/flipt-io/flipt-ui), clone it somewhere on your machine and follow the instructions in the README to get started.

To develop the project with the UI also in development mode (with hot reloading):

1. Run `npm run dev` where you have the `flipt-ui` repository checked out. This will start a development server on port `5173` and proxy API requests to the Flipt API on port `8080`.
2. Run `mage dev` from the this repository. This will create a dev build that will serve the UI from the development server, while still making it accessible on port `8080`.
3. Run the binary with the local config: `./bin/flipt --config ./config/local.yml`. This will enable CORS with the correct configuration and start the server on port `8080`.
4. Visit `http://localhost:8080` to see the UI.
5. Any changes made in the `flipt-ui` repository will be picked up by the development server and the UI will be reloaded.

### Ports

In development, the two ports that Flipt users are:

- `8080`: The port for the Flipt REST API
- `9000`: The port for the Flipt GRPC Server

These ports will be forwarded to your local machine automatically if you are developing Flipt in a VSCode Remote Container or GitHub Codespace.

## CDEs

Flipt now supports Containerized Development Environments (CDE) [VSCode Remote Containers](https://github.com/Microsoft/vscode-dev-containers)/[GitHub Codespaces](https://github.com/features/codespaces).

These technologies allow you to quickly get setup with a Flipt development environment either locally or 'in the cloud'.

For VSCode Remote Containers (devcontainers), make sure you have [Docker](https://www.docker.com/get-started) and the [ms-vscode-remote.remote-containers](https://marketplace.visualstudio.com/items?itemName=ms-vscode-remote.remote-containers) extension installed. Then simply clone this repo, open it in VSCode and run the [`Remote-Containers: Open Folder in Container`](https://code.visualstudio.com/docs/remote/containers#_quick-start-open-an-existing-folder-in-a-container) command in VSCode.

If you have access to [GitHub Codespaces](https://github.com/features/codespaces), simply open Flipt in a codespaces from the `Code` tab in the repo on GitHub or click the button below.

[![Open in Codespaces](https://github.com/codespaces/badge.svg)](https://github.com/codespaces/new/?repo=flipt-io/flipt)
