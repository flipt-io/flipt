# Changelog

This format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/)
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [v0.14.0](https://github.com/markphelps/flipt/releases/tag/v0.14.0) - 2020-05-16

### Added

* Ability to import from STDIN [https://github.com/markphelps/flipt/issues/278](https://github.com/markphelps/flipt/issues/278)

### Changed

* `import` no longer requires a file argument
* Updated several dependencies

## [v0.13.1](https://github.com/markphelps/flipt/releases/tag/v0.13.1) - 2020-04-05

### Added

* End to end UI tests [https://github.com/markphelps/flipt/issues/216](https://github.com/markphelps/flipt/issues/216)

### Changed

* Updated several dependencies

### Fixed

* Variant/Constraint columns always showing 'yes' in UI [https://github.com/markphelps/flipt/issues/258](https://github.com/markphelps/flipt/issues/258)

## [v0.13.0](https://github.com/markphelps/flipt/releases/tag/v0.13.0) - 2020-03-01

### Added

* `export` and `import` commands to export and import data [https://github.com/markphelps/flipt/issues/225](https://github.com/markphelps/flipt/issues/225)
* Enable response time histogram metrics for Prometheus [https://github.com/markphelps/flipt/pull/234](https://github.com/markphelps/flipt/pull/234)

### Changed

* List calls are no longer cached because of pagination

### Fixed

* Issue where `GetRule` would not return proper error if rule did not exist

## [v0.12.1](https://github.com/markphelps/flipt/releases/tag/v0.12.1) - 2020-02-18

### Fixed

* Issue where distributions did not always maintain order during evaluation when using Postgres [https://github.com/markphelps/flipt/issues/229](https://github.com/markphelps/flipt/issues/229)

## [v0.12.0](https://github.com/markphelps/flipt/releases/tag/v0.12.0) - 2020-02-01

### Added

* Caching support for segments, rules AND evaluation! [https://github.com/markphelps/flipt/issues/100](https://github.com/markphelps/flipt/issues/100)
* `cache.memory.expiration` configuration option
* `cache.memory.eviction_interval` configuration option

### Changed

* Fixed documentation link in app
* Underlying caching library from golang-lru to go-cache

### Removed

* `cache.memory.items` configuration option

## [v0.11.1](https://github.com/markphelps/flipt/releases/tag/v0.11.1) - 2020-01-28

### Changed

* Moved evaluation logic to be independent of datastore
* Updated several dependencies
* Moved documentation to https://github.com/markphelps/flipt.io
* Updated documentation link in UI

### Fixed

* Potential index out of range issue with 0% distributions: [https://github.com/markphelps/flipt/pull/213](https://github.com/markphelps/flipt/pull/213)

## [v0.11.0](https://github.com/markphelps/flipt/releases/tag/v0.11.0) - 2019-12-01

### Added

* Ability to match ANY constraints for a segment: [https://github.com/markphelps/flipt/issues/180](https://github.com/markphelps/flipt/issues/180)
* Pagination for Flag/Segment grids
* Required fields in API Swagger Documentation

### Changed

* All JSON fields in responses are returned, even empty ones
* Various UI Tweaks: [https://github.com/markphelps/flipt/issues/190](https://github.com/markphelps/flipt/issues/190)

## [v0.10.6](https://github.com/markphelps/flipt/releases/tag/v0.10.6) - 2019-11-28

### Changed

* Require value for constraints unless operator is one of `[empty, not empty, present, not present]`

### Fixed

* Fix issue where != would evaluate to true if constraint value was not present: [https://github.com/markphelps/flipt/pull/193](https://github.com/markphelps/flipt/pull/193)

## [v0.10.5](https://github.com/markphelps/flipt/releases/tag/v0.10.5) - 2019-11-25

### Added

* Evaluation benchmarks: [https://github.com/markphelps/flipt/pull/185](https://github.com/markphelps/flipt/pull/185)

### Changed

* Update UI dependencies: [https://github.com/markphelps/flipt/pull/183](https://github.com/markphelps/flipt/pull/183)
* Update go-sqlite3 version

### Fixed

* Calculate distribution percentages so they always add up to 100% in UI: [https://github.com/markphelps/flipt/pull/189](https://github.com/markphelps/flipt/pull/189)

## [v0.10.4](https://github.com/markphelps/flipt/releases/tag/v0.10.4) - 2019-11-19

### Added

* Example using Prometheus for capturing metrics: [https://github.com/markphelps/flipt/pull/178](https://github.com/markphelps/flipt/pull/178)

### Changed

* Update Go versions to 1.13.4
* Update Makefile to only build assets on change
* Update go-sqlite3 version

### Fixed

* Remove extra dashes when auto-generating flag/segment key in UI: [https://github.com/markphelps/flipt/pull/177](https://github.com/markphelps/flipt/pull/177)

## [v0.10.3](https://github.com/markphelps/flipt/releases/tag/v0.10.3) - 2019-11-15

### Changed

* Update swagger docs to be built completely from protobufs: [https://github.com/markphelps/flipt/pull/175](https://github.com/markphelps/flipt/pull/175)

### Fixed

* Handle flags/segments not found in UI: [https://github.com/markphelps/flipt/pull/175](https://github.com/markphelps/flipt/pull/175)

## [v0.10.2](https://github.com/markphelps/flipt/releases/tag/v0.10.2) - 2019-11-11

### Changed

* Updated grpc and protobuf versions
* Updated spf13/viper version

### Fixed

* Update chi compress middleware to fix large number of memory allocations

## [v0.10.1](https://github.com/markphelps/flipt/releases/tag/v0.10.1) - 2019-11-09

### Changed

* Use go 1.13 style errors
* Updated outdated JS dependencies
* Updated prometheus client version

### Fixed

* Inconsistent matching of rules: [https://github.com/markphelps/flipt/issues/166](https://github.com/markphelps/flipt/issues/166)

## [v0.10.0](https://github.com/markphelps/flipt/releases/tag/v0.10.0) - 2019-10-20

### Added

* Ability to write logs to file instead of STDOUT: [https://github.com/markphelps/flipt/issues/141](https://github.com/markphelps/flipt/issues/141)

### Changed

* Automatically populate flag/segment key based on name: [https://github.com/markphelps/flipt/pull/155](https://github.com/markphelps/flipt/pull/155)

## [v0.9.0](https://github.com/markphelps/flipt/releases/tag/v0.9.0) - 2019-10-02

### Added

* Support evaluating flags without variants: [https://github.com/markphelps/flipt/issues/138](https://github.com/markphelps/flipt/issues/138)

### Changed

* Dropped support for Go 1.12

### Fixed

* Segments not matching on all constraints: [https://github.com/markphelps/flipt/issues/140](https://github.com/markphelps/flipt/issues/140)
* Modal content streching to fit entire screen

## [v0.8.0](https://github.com/markphelps/flipt/releases/tag/v0.8.0) - 2019-09-15

### Added

* HTTPS support

## [v0.7.1](https://github.com/markphelps/flipt/releases/tag/v0.7.1) - 2019-07-25

### Added

* Exposed errors metrics via Prometheus

### Changed

* Updated JS dev dependencies
* Updated grpc and protobuf versions
* Updated pq version
* Updated go-sqlite3 version

## [v0.7.0](https://github.com/markphelps/flipt/releases/tag/v0.7.0) - 2019-07-07

### Added

* CORS support with `cors` config options
* [Prometheus](https://prometheus.io/) metrics exposed at `/metrics`

## [v0.6.1](https://github.com/markphelps/flipt/releases/tag/v0.6.1) - 2019-06-14

### Fixed

* Missing migrations folders in release archive: [https://github.com/markphelps/flipt/issues/97](https://github.com/markphelps/flipt/issues/97)

## [v0.6.0](https://github.com/markphelps/flipt/releases/tag/v0.6.0) - 2019-06-10

### Added

* `migrate` subcommand to run database migrations
* 'Has Prefix' and 'Has Suffix' constraint operators

### Changed

* Variant keys are now only required to be unique per flag, not globally: [https://github.com/markphelps/flipt/issues/87](https://github.com/markphelps/flipt/issues/87)

### Removed

* `db.migrations.auto` in config. DB migrations must now be run explicitly with the `flipt migrate` command

## [v0.5.0](https://github.com/markphelps/flipt/releases/tag/v0.5.0) - 2019-05-27

### Added

* Beta support for Postgres! :tada:
* `/meta/info` endpoint for version/build info
* `/meta/config` endpoint for running configuration info

### Changed

* `cache.enabled` config becomes `cache.memory.enabled`
* `cache.size` config becomes `cache.memory.items`
* `db.path` config becomes `db.url`

### Removed

* `db.name` in config

## [v0.4.2](https://github.com/markphelps/flipt/releases/tag/v0.4.2) - 2019-05-12

### Fixed

* Segments with no constraints now match all requests by default: [https://github.com/markphelps/flipt/issues/60](https://github.com/markphelps/flipt/issues/60)
* Clear Debug Console response on error

## [v0.4.1](https://github.com/markphelps/flipt/releases/tag/v0.4.1) - 2019-05-11

### Added

* `/debug/pprof` [pprof](https://golang.org/pkg/net/http/pprof/) endpoint for profiling

### Fixed

* Issue in evaluation: [https://github.com/markphelps/flipt/issues/63](https://github.com/markphelps/flipt/issues/63)

## [v0.4.0](https://github.com/markphelps/flipt/releases/tag/v0.4.0) - 2019-04-06

### Added

* `ui` config section to allow disabling the ui:

    ```yaml
    ui:
      enabled: true
    ```

* `/health` HTTP healthcheck endpoint

### Fixed

* Issue where updating a Constraint or Variant via the UI would not show the update values until a refresh: [https://github.com/markphelps/flipt/issues/43](https://github.com/markphelps/flipt/issues/43)
* Potential IndexOutOfRange error if distribution percentage didn't add up to 100: [https://github.com/markphelps/flipt/issues/42](https://github.com/markphelps/flipt/issues/42)

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
