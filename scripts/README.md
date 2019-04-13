# Scripts

The scripts contained in this directory are to aid in running CI/CD.

The following is a list of scripts and their primary responsibilities:

* `build` - builds a Docker image `markphelps/flipt:latest` with the current code
* `integration` - runs integration tests against `0.0.0.0:8080`
* `release` - runs goreleaser to release the built binary and Docker images
