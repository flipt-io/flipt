# Flipt Build

This directory contains a Go module dedicated to building and testing Flipt using [Dagger](dagger.io).
It is currently under active development. We're experimenting with it as our engine for CI/CD.

## Dependencies

- Go 1.20 (required for Dagger 0.4+)
- Mage

## Tips

All the commands in this directory can be invoked from the root of the Flipt repo using `mage dagger:run` as a prefixed.
For example:

```sh
cd ..

mage dagger:run test:ui
```

This version of the command runs using the `dagger` cli (`brew install dagger/tap/dagger`).
It comes with a nice TUI.

## Build

The `build` namespace within the Mage targets can be used to build Flipt into a target Docker container.

```sh
  build:base           builds Flipts base image via Dagger and buildkit.
  build:flipt          builds a development version of Flipt as a Docker image and loads it into a local Docker instance.
```

There exist two sub-targets in this namespace `base` and `flipt`.

`mage build:flipt` will build Flipt's service into a target Docker image and load it into your local docker.

The result will be a docker image with a name such as `flipt:dev-<sha>`.
Where `<sha>` is the local head SHA for this project.

## Test

The test section of the Mage targets handles running Flipts various unit tests with different configurations.

```sh
  test:all             runs Flipt's unit test suite against all the databases Flipt supports.
  test:cli             runs a suite of test cases which exercise the `flipt` binary CLI.
  test:integration     runs the entire integration test suite (one of ["*", "list", "<test-case>"] use "list" to see available cases).
  test:ui              runs the entire integration test suite for the UI.
  test:unit            runs the base suite of tests for all of Flipt.
```

### Unit Tests

`mage test:all` runs the entire suite for each database concurrently.
`mage test:unit` runs the entire test suite with `SQLite` as the backing database.

These tests run [Flipts Go suite](testing/test.go) of unit tests.

### Integration Tests (End to End)

`mage test:integration` runs the [integration test suite](./testing/integration.go) against an instance of Flipt.

These tests exercise the Flipt Go SDK against a matrix of Flipt configurations.

### UI Tests

`mage test:ui` runs the UIs [playwright test suite](../ui/tests) against a configure instance of Flipt.

### CLI Tests

`mage test:cli` runs a suite of [CLI tests](./testing/cli.go) invoking the `flipt` binary and its subcommands.

## Hack

The `hack` namespace within the Mage targets can be used to run various tasks that are under active development.

```sh
  hack:loadTest       runs a load test against a running instance of Flipt using Pyroscope and vegeta.
```

### Load Test

`mage hack:loadTest` runs a load test against a running instance of Flipt using [Pyroscope](https://pyroscope.io) and vegeta.

The test will import data from [testdata/main/default.yaml](testing/integration/readonly/testdata/main/default.yaml) into Flipt and then run a series of evaluation requests against a running instance of Flipt for a period of 60 seconds.

After running this command, the results will be available in the `./build/hack/out/profiles` directory.

You can view the results using the Pyroscope UI by running `pyroscope server --adhoc-data-path="$(pwd)/hack/out/profiles"`. 

This will start the Pyroscope server on `http://localhost:4040`. 

**Note:** You will need to have Pyroscope installed locally to run this command (See [Pyroscope Quick Start](https://pyroscope.io/docs/server-install-macos/)). (TODO: run this in a container)
