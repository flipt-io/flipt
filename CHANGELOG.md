# Changelog

This format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/)
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [v0.4.1](https://github.com/markphelps/flipt/releases/tag/v0.4.1) - 2019-05-11

### Added

* `/debug/pprof` [pprof](https://golang.org/pkg/net/http/pprof/) endpoint for profiling

### Fixed

* Issue in evaluation: [https://github.com/markphelps/flipt/issues/63](https://github.com/markphelps/flipt/issues/63)

## [v0.4.0](https://github.com/markphelps/flipt/releases/tag/v0.4.0) - 2019-04-06

### Fixed

* Issue where updating a Constraint or Variant via the UI would not show the update values until a refresh: [https://github.com/markphelps/flipt/issues/43](https://github.com/markphelps/flipt/issues/43)
* Potential IndexOutOfRange error if distribution percentage didn't add up to 100: [https://github.com/markphelps/flipt/issues/42](https://github.com/markphelps/flipt/issues/42)

### Added

* `ui` config section to allow disabling the ui:

    ```yaml
    ui:
      enabled: true
    ```

* `/health` HTTP healthcheck endpoint

## [v0.3.0](https://github.com/markphelps/flipt/releases/tag/v0.3.0) - 2019-03-03

### Changed

* Renamed generated proto package to `flipt` for use with external GRPC clients
* Updated docs and example to reference [GRPC go client](https://github.com/markphelps/flipt-grpc-go)

### Fixed

* Don't return error on graceful shutdown of HTTP server

## [v0.2.0](https://github.com/markphelps/flipt/releases/tag/v0.2.0) - 2019-02-24

### Added

* `server` config section to consolidate and rename `host`, `api.port` and `backend.port`:

    ```yaml
    server:
      host: 0.0.0.0
      http_port: 8080
      grpc_port: 9000
    ```

* Implemented flag caching! Preliminary testing shows about a 10x speedup for retrieving flags with caching enabled. See the docs for more info.

    ```yaml
    cache:
      enabled: true
    ```

### Deprecated

* `host`, `api.port` and `backend.port`. These values have been moved and renamed under the `server` section and will be removed in the 1.0 release.

## [v0.1.0](https://github.com/markphelps/flipt/releases/tag/v0.1.0) - 2019-02-19

### Added

* Moved proto/client code to proto directory and added MIT License

## [v0.0.0](https://github.com/markphelps/flipt/releases/tag/v0.0.0) - 2019-02-16

Initial Release!
