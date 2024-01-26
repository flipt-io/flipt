# Changelog

This format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/)
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [v0.10.0](https://github.com/flipt-io/flipt/releases/tag/sdk/go/v0.10.0) - 2024-01-26

### Added

- Add Kubernetes authentication provider (#2703)

## [v0.9.0](https://github.com/flipt-io/flipt/releases/tag/sdk/go/v0.9.0) - 2024-01-09

### Added

- Add JWT authentication provider (#2620)
- `rpc/flipt`: add reference to read request protocols (#2570)

## [v0.8.0](https://github.com/flipt-io/flipt/releases/tag/sdk/go/v0.8.0) - 2023-11-15

- `sdk/go`: update rpc/flipt to 1.31.0

## [v0.5.0](https://github.com/flipt-io/flipt/releases/tag/sdk/go/v0.5.0) - 2023-08-08

- `sdk/go`: update rpc/flipt to 1.25.0

## [v0.4.0](https://github.com/flipt-io/flipt/releases/tag/sdk/go/v0.4.0) - 2023-08-08

### Added

- `sdk/go`: update rpc/flipt to 1.24.0
- new evaluation routes (#1824)

## [v0.3.0](https://github.com/flipt-io/flipt/releases/tag/sdk/go/v0.3.0) - 2023-05-23

### Added

- `sdk/go`: Datetime constraint type (#1602)
- `sdk/go`: Add optional description to constraints (#1581)

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
