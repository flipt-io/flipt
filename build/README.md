# Flipt Build

This directory contains a Go module dedicated to building and testing Flipt using [Dagger](dagger.io).

## Dependencies

- Dagger CLI (e.g. for MacOS `brew install dagger/tap/dagger`)

## Tips

All the commands in this directory can be invoked from the root of the Flipt repo using `dagger call`.

```console
USAGE
  dagger call [options] [arguments] <function>

EXAMPLES
  dagger call test
  dagger call build -o ./bin/myapp
  dagger call lint stdout

FUNCTIONS
  base             Returns a container with all the assets compiled and ready for testing and distribution
  base-container
  build            Return container with Flipt binaries in a thinner alpine distribution
  generate         Execute generate function with subcommand
  source
  test             Execute test specific by subcommand
```

This version of the command runs using the `dagger` cli (`brew install dagger/tap/dagger`).
It comes with a nice TUI.

## Build

```console
dagger call build
```

This builds a distribution appropriate version of Flipt.
It results in a container image based on Alpine, with the Flipt binary and an appropriate fileystem structure.

## Test

The test section of the dagger command tree contains a bunch of functions for running tests suites with Flipt.

```console
➜  dagger call test --help
Execute test specific by subcommand
see all available subcommands with dagger call test --help

USAGE
  dagger call test [arguments] <function>

FUNCTIONS
  base-container
  cli               Run all cli tests
  flipt-container
  integration       Run all integration tests
  load              Run all load tests
  migration         Run all migration tests
  source
  ui                Run all ui tests
  uicontainer
  unit              Run all unit tests

ARGUMENTS
      --source Directory   [required]

Use "dagger call test [command] --help" for more information about a command.
```

- `dagger call test --source=. cli`: Runs a suite of tests against `flipt` CLI commands
- `dagger call test --source=. unit`: Runs the entire suite of unit style tests
- `dagger call test --source=. ui`: Runs the entire suite of UI integration style tests (playwright against running Flipt)
- `dagger call test --source=. migration`: Ensures backwards compatibility after running database migrations
- `dagger call test --source=. load`: Run Flipt and measure performance running evaluations under load

### Integration Tests

The `integration` section of the `test` functions runs our matrix of integration tests against Flipt under a number of different configuration scenarios.

```console
➜  dagger call test integration --help
Run all integration tests

USAGE
  dagger call test integration [arguments]

ARGUMENTS
      --cases string    (default "*")
```

By default, it runs all cases concurrently against the target dagger runtime.
However, cases can be constrained to the individual top-level suites via the `--cases` flag.
This flag expects a space delimited list of the case names.
The case names are maintained near the top of [integration.go](./testing/integration.go).
