# Flipt Build

This directory contains a Go module dedicated to building and testing Flipt using [Dagger](dagger.io).
It is currently under active development. We're experimenting with it as our engine for CI/CD.

## Dependencies

- Go 1.20 (required for Dagger 0.4+)
- Mage

## Build

The `build` namespace within the Mage targets can be used to build Flipt into a target Docker container.

There exist two sub-targets in this namespace `base` and `flipt`.

`mage build:flipt` will build Flipt's service into a target Docker image and load it into your local docker.

The result will be a docker image with a name such as `flipt:dev-<sha>`.
Where `<sha>` is the local head SHA for this project.

## Test

The test section of the Mage targets handles running Flipts various unit tests with different configurations.

`mage test:unit` runs the entire test suite with `SQLite` as the backing database.

`mage test:database <db>` runs the entire test suite with the desired database (`sqlite`, `postgres`, `mysql` and `cockroach` available).

`mage test:all` runs the entire suite for each database concurrently.

`mage test:integration` run the [integration test suite](./build/integration) against an instance of Flipt.
