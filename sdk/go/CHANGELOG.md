# Changelog

This format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/)
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [v0.2.0](https://github.com/flipt-io/flipt/releases/tag/sdk/go/v0.2.0) - 2023-04-11

### Added

- `sdk/go`: regenerate clients to include new namespaces features

### Changed

- `sdk/go`: add details regarding the HTTP transport (#1435)

### Fixed

- protojson to use DiscardUnknown option for backwards compatibility (#1453)

## v0.1.1

### Changed

- Upgraded `rpc/flipt` to `v1.19.3`
- Upgraded `errors` to `v1.19.3`

## v0.1.0

### Added

- A new Go SDK for Flipt.
