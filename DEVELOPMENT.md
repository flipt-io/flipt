# Development

The following are instructions for setting up your local machine for Flipt development. For info on using VSCode Remote Containers / GitHub Codespaces, see [#cdes](#cdes) below.

> [!TIP]
> Try our new [devenv](#devenv) solution to quickly get setup developing Flipt!

Also check out our [Contributing](CONTRIBUTING.md) guide for more information on how to get changes merged into the project.

## Requirements

Before starting, make sure you have the following installed:

- [GCC Compiler](https://gcc.gnu.org/install/binaries.html)
- [SQLite](https://sqlite.org/index.html)
- [Go 1.20+](https://golang.org/doc/install)
- [NodeJS >= 18](https://nodejs.org/en/ )
- [Mage](https://magefile.org/)
- [Docker](https://docs.docker.com/install/) (for running tests)

## CGO

Flipt uses [CGO](https://golang.org/cmd/cgo/) to compile SQLite.

If you run into errors such as:

```console
undefined: sqlite3.Error
```

Then you need to enable CGO. 

### Windows

- Make sure you have the [GCC Compiler](https://gcc.gnu.org/install/binaries.html) installed and in your PATH.
- `set CGO_ENABLED=1`

### Linux/Mac

- Make sure you have the [GCC Compiler](https://gcc.gnu.org/install/binaries.html) installed and in your PATH.
- `export CGO_ENABLED=1`

## Setup

1. Clone this repo: `git clone https://github.com/flipt-io/flipt`.
1. Run `mage bootstrap` to install required development tools. See [#bootstrap](#bootstrap) below.
1. Run `mage go:test` to execute the Go test suite. For more information on tests, see also [here](build/README.md)
1. Run `mage` to build the binary with embedded assets.
1. Run `mage -l` to see a full list of possible commands.

## Conventional Commits

Flipt uses [Conventional Commits](https://www.conventionalcommits.org/en/v1.0.0/) for commit messages. This allows us to automatically generate changelogs and releases. To help with this, we use [pre-commit](https://pre-commit.com/) to automatically lint commit messages. To install pre-commit, run:

`pip install pre-commit` or `brew install pre-commit` (if you're on a Mac)

Then run `pre-commit install` to install the git hook.

## Go

Flipt is built with Go 1.20+.

## Bootstrap

The `bootstrap` task will install all of the necessary tools used for development and testing. It does this using a seperate tools modules as described here: [https://marcofranssen.nl/manage-go-tools-via-go-modules](https://marcofranssen.nl/manage-go-tools-via-go-modules)

## Configuration

A sample configuration for running and developing against Flipt can be found at `./config/local.yml`. To run Flipt with this configuration, run:

`./bin/flipt [--config ./config/local.yml`]

To prevent providing the config via config flag every time, you have the option of writing configuration at the location:

```shell
{{ USER_CONFIG_DIR }}/flipt/config.yml
```

The flipt binary will check in that location if a `--config` override is not provided, so you can just invoke the binary:

`./bin/flipt`

The `USER_CONFIG_DIR` is different based on your architecture, and specified [here](https://pkg.go.dev/os#UserConfigDir).

## Changes

Changing certain types of files such as the proto or ui files require re-building before they will be picked up in new versions of the binary.

### Updating .proto Files

After changing `flipt.proto`, you'll need to run `mage proto`. This will regenerate the necessary files in the `rpc` directory.

## UI

The UI is built using [NPM](https://nodejs.org/en/) and [Vite](https://vitejs.dev/) and is also statically compiled into the Flipt binary using [go:embed](https://golang.org/pkg/embed/).

To develop the project with the UI also in development mode (with hot reloading):

1. Run `mage ui:dev` from the root of this repository. This will start a development server on port `5173` and proxy API requests to the Flipt API on port `8080`.
2. In another terminal, run `mage dev` (or `mage go:run`) from the root of this repository. This will run the backend server making it accessible on port `8080`.
3. Visit `http://localhost:8080` to see the UI.
4. Any changes made in the `ui` directory will be picked up by the development server and the UI will be reloaded.

### Ports

In development, the two ports that Flipt uses are:

- `8080`: The port for the Flipt REST API
- `9000`: The port for the Flipt GRPC Server

These ports will be forwarded to your local machine automatically if you are developing Flipt in a VSCode Remote Container or GitHub Codespace.

## Docker Compose

If you want to develop Flipt using Docker Compose, you can use the `docker-compose.yml` file in the root of this repository. 

This will start two Docker containers:

- `server` will run the Flipt server, mounting the contents of this repository as a bind mount. This means that the database (SQLite by default) will be persisted between runs. Currently the server does not support hot reloading, so you'll need to restart the container to pick up any changes.
- `ui` will run the UI development server, mounting the `ui` directory as a bind mount. This means that any changes made to the UI will be picked up by the development server and the UI will be reloaded (thanks to Vite).

To start the containers, run `docker-compose up` from the root of this repository. After the containers are started, you can visit `http://localhost:8080` to see the UI.

## CDEs

Flipt also supports Containerized Development Environments (CDE) [VSCode Remote Containers](https://github.com/Microsoft/vscode-dev-containers)/[GitHub Codespaces](https://github.com/features/codespaces)/[Gitpod](https://www.gitpod.io/).

These technologies allow you to quickly get setup with a Flipt development environment either locally or 'in the cloud'.

### VSCode Remote Containers

For VSCode Remote Containers (devcontainers), make sure you have [Docker](https://www.docker.com/get-started) and the [ms-vscode-remote.remote-containers](https://marketplace.visualstudio.com/items?itemName=ms-vscode-remote.remote-containers) extension installed. Then simply clone this repo, open it in VSCode and run the [`Remote-Containers: Open Folder in Container`](https://code.visualstudio.com/docs/remote/containers#_quick-start-open-an-existing-folder-in-a-container) command in VSCode.

### GitHub Codespaces

Simply open Flipt in a codespaces from the `Code` tab in the repo on GitHub or click the button below:

[![Open in Codespaces](https://github.com/codespaces/badge.svg)](https://github.com/codespaces/new/?repo=flipt-io/flipt)

### Gitpod

To use [Gitpod](https://www.gitpod.io/), simply open Flipt in Gitpod by clicking the button below:

[![Open in Gitpod](https://gitpod.io/button/open-in-gitpod.svg)](https://gitpod.io/#https://github.com/flipt-io/flipt)

## devenv

[devenv](https://devenv.sh) is a solution that creates fast, declarative, reproducible, and composable developer environments using Nix.

To use it for developing Flipt, you'll first need to install it. See the devenv [getting started](https://devenv.sh/getting-started/) guide for more information.

Once you have devenv installed, you can run `devenv up` from the root of this repository to start a development environment.

This will start a Docker container with the Flipt server running on port `8080` and the UI development server running on port `5173`.
