# Changelog

This format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/)
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [v1.30.0](https://github.com/flipt-io/flipt/releases/tag/v1.30.0) - 2023-10-31

### Added

- add flag key to variant and boolean evaluation responses (#2318)
- add subject to authentication metadata for auditing (#2299)
- `git`: support SSH authentication (#2272)

### Changed

- `ui`: redux all flag and variant state management (#2301)
- dependency updates

### Fixed

- readOnly config option should be read_only (#2298)
- Audit log path will be created if doesn't exist (#2295)

## [v1.29.1](https://github.com/flipt-io/flipt/releases/tag/v1.29.1) - 2023-10-26

### Fixed

- error about invalid offset with FS backend (#2286)

### Changed

- security: GRPC dependency updates

## [v1.29.0](https://github.com/flipt-io/flipt/releases/tag/v1.29.0) - 2023-10-23

### Added

- impl grpc health check (#2257)
- `distributions`: support updating distribution variant (#2235)
- allow disabling profiling endpoints (#2195)
- LibSQL/Turso support

### Changed

- Add default namespace in the sync store layer for FS implementations

## [v1.28.2](https://github.com/flipt-io/flipt/releases/tag/v1.28.2) - 2023-10-13

### Fixed

- `service/middleware`: use separate cache keys for old and new evaluations (#2237)

## [v1.28.1](https://github.com/flipt-io/flipt/releases/tag/v1.28.1) - 2023-10-04

### Fixed

- csp header for monaco editor (#2200)
- specify region to fix s3 feature (#2193)

## [v1.28.0](https://github.com/flipt-io/flipt/releases/tag/v1.28.0) - 2023-10-02

### Added

- `ui`: Rules tab (#2176)
- Add requestID to batch response (#2177)
- `cli`: Gen docs (#2173)
- `storage/fs`: Allow FS backends to understand multi-document YAML files
- `cli`: config edit command (#2169)
- support export otel telemetry directly over http(s) (#2155)
- `cli`: cmd shell completions (#2168)
- `import/export`: add support for exporting multiple namespaces and all namespaces
- `cli`: config init (#2147)
- `import/export`: add functionality for importing YAML stream
- `tracing`: add OTLP headers config (#2148)
- `audit`: Add webhook templating for generic request to webhooks

### Changed

- Dependency Updates

### Fixed

- export over https (#2152)

## [v1.27.2](https://github.com/flipt-io/flipt/releases/tag/v1.27.2) - 2023-09-21

### Fixed

- `security`: dont marshal secrets to JSON

## [v1.27.1](https://github.com/flipt-io/flipt/releases/tag/v1.27.1) - 2023-09-18

### Added

- Add token type and actions to audit event filter (#2121)

### Fixed

- `webhook`: Rewind HTTP request body upon request failure (#2143)
- Default config allow env overrides (#2144)

## [v1.27.0](https://github.com/flipt-io/flipt/releases/tag/v1.27.0) - 2023-09-13

### Added

- Audit info to telemetry (#2116)
- Webhook configuration for sending event payloads (#2087)
- `events`: Add feature for scoping the emission of events
- Datadog as an exporter for OTLP traces (#2083)
- Default config (#2067)
- Support page instead of link to docs (#2063)
- Ability to set default theme for UI (#2062)

### Changed

- `flipt/validate`: build entire snapshot on validate (#2093)
- Dependency updates

### Fixed

- `build/generate`: add exposed port to flipt (#2060)

## [v1.26.1](https://github.com/flipt-io/flipt/releases/tag/v1.26.1) - 2023-09-09

### Fixed

- `cmd/import`: check for correct not found error on create namespace (#2082)
- `ext`: attempt to import when namespace exists on create namespace (#2089)
- `schema`: make rollout description optional in cue schema (#2091)

## [v1.26.0](https://github.com/flipt-io/flipt/releases/tag/v1.26.0) - 2023-08-28

### Added

- exp: Homebrew tap (#2044)
- Switch to monaco editor for console (#2040)
- default to ~/Library/Application\ Support/flipt for db sqlite file on mac (#2039)
- `auth`: Add parameters for pkce flow
- `auth`: OAuth for GitHub (#2016)

### Changed

- Dependency updates

### Fixed

- hover icon visibility for copy key to clipboard in lightmode (#2056)

## [v1.25.2](https://github.com/flipt-io/flipt/releases/tag/v1.25.2) - 2023-08-23

### Fixed

- `internal/ext`: add 1.1 to versions and validate explicit 1.2 (#2042)

## [v1.25.1](https://github.com/flipt-io/flipt/releases/tag/v1.25.1) - 2023-08-21

### Fixed

- cache shadow var when setting up middleware (#2017)

## [v1.25.0](https://github.com/flipt-io/flipt/releases/tag/v1.25.0) - 2023-08-16

### Added

- `ui`: copy flag/segment key to clipboard button in ui (#1987)
- `storage/cache`: add redis connection optional params (#1983)
- `ui`: Force readonly (#1973)
- `fs/s3`: Add support for prefix filtering (#1938)
- `config/flipt.schema.*`: add oidc provider scopes property (#1946)
- `segment-anding`: the ability to and or or multiple segments in rules and rollouts (#1941)

### Changed

- `fs`: Graduate filesystem storage (#1976)

### Fixed

- `ui`: make flags/segments inputs disabled in readonly mode (#1986)

## [v1.24.2](https://github.com/flipt-io/flipt/releases/tag/v1.24.2) - 2023-08-07

### Fixed

- `ui`: rule form inputs in darkmode (#1968)
- store namespace in localstorage (#1950)
- dont clear all localstorage on 401 (#1951)

## [v1.24.1](https://github.com/flipt-io/flipt/releases/tag/v1.24.1) - 2023-08-01

### Fixed

- fix(auth/middleware): only enforce email matches on OIDC method tokens (#1934)

## [v1.24.0](https://github.com/flipt-io/flipt/releases/tag/v1.24.0) - 2023-07-31

### Added

- boolean flag type, rollouts, and new evaluation routes :tada: (#1824)
- `fs/s3`: Initial support for s3 filesystem backend (#1900)
- `email`: Add email matching logic for whitelisting OIDC
- `build`: add mage dagger:run generate:screenshots (#1868)
- add cache to get evaluation rules storage method (#1910)

### Changed

- unhide validate command (#1909)
- add rudineirk as a contributor for code (#1913)
- add ahobson as a contributor for code (#1905)
- Dependency updates

### Fixed

- renable pprof (#1912)
- dagger cli test (#1899)
- `ui`: memoize selectors for namespaces (#1837)

## [v1.23.3](https://github.com/flipt-io/flipt/releases/tag/v1.23.3) - 2023-07-06

### Fixed

- potential panic in `flipt validate` (#1831)
- support variant desc in cue import schema (#1828)

### Changed

- Dependency updates

## [v1.23.2](https://github.com/flipt-io/flipt/releases/tag/v1.23.2) - 2023-07-03

### Added

- 'Share Feedback' link in app footer (#1812)

### Changed

- trim trailing slash off of http request if exists
- `config`: ensure JSON schema validates the default configuration (#1788)
- `config`: ensure default config passes CUE validation (#1758)
- return 400 for invalid page token"
- `integration`: ensure all list limits are enforced correctly
- Dependency updates

### Fixed

- Fix `flipt validate` command (#1811)
- context cancelled/deadline exceeded errors coming back internal (#1803)
- prevent overwriting URL struct (#1784) for trailing slash change

## [v1.23.1](https://github.com/flipt-io/flipt/releases/tag/v1.23.1) - 2023-06-15

### Added

- Ability to configure use of prepared statements (for enabling PGBouncer) (#1750)
- `telemetry`: track storage type and experimental features (#1745)

### Changed

- Dependency updates

### Fixed

- `storage/sql`: set binary_parameters=yes for lib/pq when prepared statements disabled (#1753)
- UI: input bg / text in darkmode (#1752)
- count rules was not taking flagKey into account (#1738)
- tag prefix for build info/release links (#1757)

## [v1.23.0](https://github.com/flipt-io/flipt/releases/tag/v1.23.0) - 2023-06-12

### Added

- UI: Darkmode (#1692)
- UI: Copy to namespace (#1731)
- UI: Support for read-only mode (#1709)
- check source file from different places for configuration if they exist (#1650)
- Experimental: support for filesystem/git backends (#1714)
- Flipt Validate (#1642)
- add namespace to export, check for match on import (#1645)
- add storage configuration to support multiple backend types (#1644)

### Changed

- `integration`: ensure public and auth self routes are reachable (#1729)
- `cli`: ensure user configuration path supported (#1727)
- `storage/sql`: handle null description on constraint during GetSegment (#1654)
- UI: switch namespaces to use Redux (#1680)
- `build`: define CLI tests using dagger pipelines (#1679)
- Dependency updates

### Fixed

- UI: disable flag toggle in read-only mode (#1719)
- cross namespaced constraints (#1721)
- `mage`: install package-changed as dev dependency (#1716)
- UI: catch errs on call to namespaces API in namespace form (#1687)
- Get the openfeature example working (#1665)
- panic if no import file name provided (#1658)

## [v1.22.1](https://github.com/flipt-io/flipt/releases/tag/v1.22.1) - 2023-05-24

### Fixed

- `storage/sql`: handle null description on constraint during GetSegment

## [v1.22.0](https://github.com/flipt-io/flipt/releases/tag/v1.22.0) - 2023-05-23

### Added

- Datetime constraint type (#1602)
- Add optional description to constraints (#1581)

### Changed

- Dependency updates

## [v1.21.1](https://github.com/flipt-io/flipt/releases/tag/v1.21.1) - 2023-05-05

### Fixed

- `storage/sql`: paginated walk for resources using joins (#1584)

## [v1.21.0](https://github.com/flipt-io/flipt/releases/tag/v1.21.0) - 2023-05-02

### Added

- OTEL implementation for audit sinks (#1458)
- Error highlighting in console (#1528)
- Additional telemetry data captured re: database, cache and authentication (#1527)

### Fixed

- `cmd/flipt`: restore console logger defaults on fatal (#1550)
- `grpc/middleware`: set timestamp on each batch evaluate response (#1545)

### Changed

- Dependency updates

## [v1.20.2](https://github.com/flipt-io/flipt/releases/tag/v1.20.2) - 2023-05-05

### Fixed

- `storage/sql`: paginated walk for resources using joins (#1584)

## [v1.20.1](https://github.com/flipt-io/flipt/releases/tag/v1.20.1) - 2023-04-27

### Fixed

- `grpc/middleware`: set timestamp on each batch evaluate response (#1545)

## [v1.20.0](https://github.com/flipt-io/flipt/releases/tag/v1.20.0) - 2023-04-11

### Added

- Support for 'namespacing' / multi-environments. All types can now belong to a namespace allowing you to seperate your flags/segments/etc.

### Changed

- `sdk/go`: add details regarding the HTTP transport (#1435)
- All existing objects have been moved to the 'default' namespace to be fully backward compatible.
- Import/Export have been updated to be 'namespace-aware'
- Dependency updates

### Fixed

- `cmd/import`: re-open migration after dropping on import
- protojson to use DiscardUnknown option for backwards compatibility (#1453)
- `rpc/flipt`: move all openapi annotations into yaml file (#1437)

## [v1.19.5](https://github.com/flipt-io/flipt/releases/tag/v1.19.5) - 2023-05-05

### Fixed

- `storage/sql`: paginated walk for resources using joins (#1584)

## [v1.19.4](https://github.com/flipt-io/flipt/releases/tag/v1.19.4) - 2023-04-27

### Fixed

- `grpc/middleware`: set timestamp on each batch evaluate response (#1545)

## [v1.19.3](https://github.com/flipt-io/flipt/releases/tag/v1.19.3) - 2023-03-22

### Changed

- Upgraded to Go 1.20
- Dependency updates

## [v1.19.2](https://github.com/flipt-io/flipt/releases/tag/v1.19.2) - 2023-03-15

### Changed

- Dependency updates
- Return better error messages for grpc-gateway errors [#1397](https://github.com/flipt-io/flipt/pull/1397)

### Fixed

- Import/Export for CockroachDB [#1399](https://github.com/flipt-io/flipt/pull/1399)

## [v1.19.1](https://github.com/flipt-io/flipt/releases/tag/v1.19.1) - 2023-02-27

### Changed

- Dependency updates
- UI: combobox shows 'No results found' if no input matches

### Fixed

- UI: issue where Edit Rule submit button was disabled [ui #108](https://github.com/flipt-io/flipt-ui/pull/108)

## [v1.19.0](https://github.com/flipt-io/flipt/releases/tag/v1.19.0) - 2023-02-22 :lock:

### Added

- UI: Settings/API Tokens management
- Enable ability to specify bootstrap token [#1350](https://github.com/flipt-io/flipt/1350)
- Kubernetes Authentication Method [#1344](https://github.com/flipt-io/flipt/1344)

### Changed

- Dependency updates
- Switch to official redis client and redis-cache/v9 [#1345](https://github.com/flipt-io/flipt/1345)

## [v1.18.2](https://github.com/flipt-io/flipt/releases/tag/v1.18.2) - 2023-02-14 :heart:

### Added

- OpenTelemetry: support for Zipkin exporter [#1323](https://github.com/flipt-io/flipt/pull/1323)
- OpenTelemetry: support for OTLP exporter [#1324](https://github.com/flipt-io/flipt/pull/1324)

### Changed

- Deprecated `tracing.jaeger.enabled` [#1316](https://github.com/flipt-io/flipt/pull/1316)
- Added `frame-ancestors` directive to `Content-Security-Policy` header [#1317](https://github.com/flipt-io/flipt/pull/1317)

### Fixed

- UI: Clear session if 401 is returned from server [ui #88](https://github.com/flipt-io/flipt-ui/pull/88)
- Ensure all authentication methods are being cleaned up [#1337](https://github.com/flipt-io/flipt/pull/1337)
- Ensure failed cookie auth attempt clears cookies [#1336](https://github.com/flipt-io/flipt/pull/1336)
- 500 error when providing invalid base64 encoded page token [#1314](https://github.com/flipt-io/flipt/pull/1314)

## [v1.18.1](https://github.com/flipt-io/flipt/releases/tag/v1.18.1) - 2023-02-02

### Added

- Set Content-Security-Policy and `X-Content-Type-Options` headers [#1293](https://github.com/flipt-io/flipt/pull/1293)
- Ability to customize log keys/format [#1295](https://github.com/flipt-io/flipt/pull/1295)
- Additional OpenTelemetry annotations for `GetFlag` and `Evaluate` calls [#1306](https://github.com/flipt-io/flipt/pull/1306)
- UI: Visual indicator on submitting + success message [ui #52](https://github.com/flipt-io/flipt-ui/pull/52)

### Changed

- Dependency updates
- UI: Clear session on logout, change session storage format [ui #64](https://github.com/flipt-io/flipt-ui/pull/64)

### Fixed

- ListX calls could potentially lock the DB if running in SQLite. Also a nice performance boost [#1297](https://github.com/flipt-io/flipt/pull/1297)
- UI: Bug where editing constraint values did not pre-populate with existing values [ui #67](https://github.com/flipt-io/flipt-ui/pull/67)
- UI: Regression where we weren't showing help text when creating rule with no variants [ui #58](https://github.com/flipt-io/flipt-ui/pull/58)

## [v1.18.0](https://github.com/flipt-io/flipt/releases/tag/v1.18.0) - 2023-01-23

### Added

- UI: Login via OIDC [ui #41](https://github.com/flipt-io/flipt-ui/pull/41)
- UI: Validate distribution percentages [ui #37](https://github.com/flipt-io/flipt-ui/pull/37)
- New `auth.ExpireAuthenticationSelf` endpoint to expire a user's own authentication [#1279](https://github.com/flipt-io/flipt/pull/1279)

### Changed

- Dev: Switched from Task to Mage [#1273](https://github.com/flipt-io/flipt/pull/1273)
- Authentication metadata structure changed to JSON object [#1275](https://github.com/flipt-io/flipt/pull/1275)
- Dev: Make developing the UI easier by proxying `:8080` to vite [#1278](https://github.com/flipt-io/flipt/pull/1278)
- Dependency updates

### Fixed

- Setting Authentication cookies on localhost [#1274](https://github.com/flipt-io/flipt/pull/1274)
- Panic in `/meta` endpoints when Authentication was required but not present [#1277](https://github.com/flipt-io/flipt/pull/1277)
- Don't set empty CSRF cookie [#1280](https://github.com/flipt-io/flipt/pull/1280)
- Bootstrapping command in mage [#1281](https://github.com/flipt-io/flipt/pull/1281)
- Removed duplicate shutdown log [#1282](https://github.com/flipt-io/flipt/pull/1282)

## [v1.17.1](https://github.com/flipt-io/flipt/releases/tag/v1.17.0) - 2023-01-13

### Fixed

- UI: Fix useEffect warnings from eslint [ui #21](https://github.com/flipt-io/flipt-ui/pull/21)
- UI: Show info about creating rule with no variants [ui #22](https://github.com/flipt-io/flipt-ui/pull/22)
- UI: Fix issue where key was changing when updating flag/segment name [ui #25](https://github.com/flipt-io/flipt-ui/pull/25)
- UI: Fixed issue where segment form would reset when changing matchType [ui #26](https://github.com/flipt-io/flipt-ui/pull/26)
- UI: Fixed auto key generation from name weirdness [ui #27](https://github.com/flipt-io/flipt-ui/pull/27)

## [v1.17.0](https://github.com/flipt-io/flipt/releases/tag/v1.17.0) - 2023-01-12

### Added

- Brand new UI / UX :tada:
- JSON/CUE schema [#1196](https://github.com/flipt-io/flipt/pull/1196)
- Support config file versioning [#1225](https://github.com/flipt-io/flipt/pull/1225)
- OIDC Auth Methods [#1197](https://github.com/flipt-io/flipt/pull/1197)
- List API for Auth Methods [#1240](https://github.com/flipt-io/flipt/pull/1240)

### Changed

- Move `/meta` API endpoints behind authentication [#1250](https://github.com/flipt-io/flipt/pull/1250)
- Dependency updates

### Deprecated

- Deprecates `ui.enabled` in favor of always enabling the UI

### Fixed

- Don't print to stdout with color when using JSON log format [#1188](https://github.com/flipt-io/flipt/pull/1188) 

### Removed

- Embedded swagger/openapiv2 API docs in favor of hosted / openapiv3 API docs [#1241](https://github.com/flipt-io/flipt/pull/1241)

## [v1.16.0](https://github.com/flipt-io/flipt/releases/tag/v1.16.0) - 2022-11-30

### Added

- Automatic authentication background cleanup process [#1161](https://github.com/flipt-io/flipt/pull/1161).

### Fixed

- Fix configuration unmarshalling from `string` to `[]string` to delimit on `" "` vs `","` [#1179](https://github.com/flipt-io/flipt/pull/1179)
- Dont log warnings when telemetry cannot report [#1156](https://github.com/flipt-io/flipt/pull/1156)

### Changed

- Switched to use otel abstractions for recording metrics [#1147](https://github.com/flipt-io/flipt/pull/1147).
- Dependency updates

## [v1.15.1](https://github.com/flipt-io/flipt/releases/tag/v1.15.1) - 2022-11-28

### Fixed

- Authentication API read operations have been appropriately mounted and no longer return 404.

## [v1.15.0](https://github.com/flipt-io/flipt/releases/tag/v1.15.0) - 2022-11-17

### Added

- Token-based authentication [#1097](https://github.com/flipt-io/flipt/pull/1097)

### Changed

- Linting for markdown
- Merge automation for dependabot updates
- Dependency updates

## [v1.14.0](https://github.com/flipt-io/flipt/releases/tag/v1.14.0) - 2022-11-02

### Added

- `reason` field in `EvaluationResponse` payload detailing why the request evaluated to the given result [#1099](https://github.com/flipt-io/flipt/pull/1099)

### Deprecated

- Deprecated both `db.migrations.path` and `db.migrations_path` [#1096](https://github.com/flipt-io/flipt/pull/1096)

### Fixed

- Propogating OpenTelemetry spans through Flipt [#1112](https://github.com/flipt-io/flipt/pull/1112)

## [v1.13.0](https://github.com/flipt-io/flipt/releases/tag/v1.13.0) - 2022-10-17

### Added

- Page token based pagination for `list` methods for forward compatibility with
  future versions of the API [#936](https://github.com/flipt-io/flipt/issues/936)
- Support for CockroachDB :tada: [#1064](https://github.com/flipt-io/flipt/pull/1064)

### Changed

- Validation for `list` methods now requires a `limit` if requesting with an `offset` or `page_token`
- Replaced OpenTracing with OpenTelemetry [#576](https://github.com/flipt-io/flipt/issues/576)
- Updated favicon to new logo
- Documentation link in app no longer uses redirects
- Dependency updates

### Deprecated

- Deprecated `offset` in `list` methods in favor of `page_token` [#936](https://github.com/flipt-io/flipt/issues/936)

### Fixed

- Correctly initialize shutdown context after interrupt [#1057](https://github.com/flipt-io/flipt/pull/1057)

### Removed

- Removed stacktrace from error logs [#1066](https://github.com/flipt-io/flipt/pull/1066)

## [v1.12.1](https://github.com/flipt-io/flipt/releases/tag/v1.12.1) - 2022-09-30

### Fixed

- Issue where parsing value with incorrect type would return 500 from the evaluation API [#1047](https://github.com/flipt-io/flipt/pull/1047)

### Changed

- Dependency updates
- Use testcontainers for MySQL/Postgres tests to run locally [#1045](https://github.com/flipt-io/flipt/pull/1045)

## [v1.12.0](https://github.com/flipt-io/flipt/releases/tag/v1.12.0) - 2022-09-22

### Added

- Ability to log as structure JSON [#1027](https://github.com/flipt-io/flipt/pull/1027)
- Ability to configure GRPC log level seperately from main service log level [#1029](https://github.com/flipt-io/flipt/pull/1029)

### Changed

- Configuration 'enums' are encoding correctly as JSON [#1030](https://github.com/flipt-io/flipt/pull/1030)
- Dependency updates

### Fixed

- Rendering of clone icon for variants/constraints [#1038](https://github.com/flipt-io/flipt/pull/1038)

## [v1.11.0](https://github.com/flipt-io/flipt/releases/tag/v1.11.0) - 2022-09-12

### Added

- Redis example [#968](https://github.com/flipt-io/flipt/pull/968)
- Support for arm64 builds [#1005](https://github.com/flipt-io/flipt/pull/1005)

### Changed

- Updated to Go 1.18 [#1016](https://github.com/flipt-io/flipt/pull/1016)
- Replaces Logrus with Zap logger [#1020](https://github.com/flipt-io/flipt/pull/1020)
- Updated Buf googleapis version [#1011](https://github.com/flipt-io/flipt/pull/1011)
- Dependency updates

## [v1.10.1](https://github.com/flipt-io/flipt/releases/tag/v1.10.1) - 2022-09-01

### Fixed

- (Ported from v1.9.1) Issue when not using `netgo` build tag during build, resulting in native dns resolution not working for some cases in k8s. See  [https://github.com/flipt-io/flipt/issues/993](https://github.com/flipt-io/flipt/issues/993).

## [v1.10.0](https://github.com/flipt-io/flipt/releases/tag/v1.10.0) - 2022-07-27

### Added

- Redis cache support :tada: [https://github.com/flipt-io/flipt/issues/633](https://github.com/flipt-io/flipt/issues/633)
- Support for pretty printing JSON responses from API (via ?pretty=true or setting `Accept: application/json+pretty` header)
- Configuration warnings/deprecations are displayed in console at startup

### Changed

- Ping database on startup to check if it's alive
- Default cache TTL is 1m. Previously there was no TTL for the in memory cache.
- Dependency updates

### Deprecated

- `cache.memory.enabled` config value is deprecated. See [Deprecations](DEPRECATIONS.md) for more info
- `cache.memory.expiration` config value is deprecated. See [Deprecations](DEPRECATIONS.md) for more info

### Fixed

- Build date was incorrect and always showed current date/time
- Button spacing issue on targeting page
- Docker compose examples run again after switch to non-root user 

## [v1.9.1](https://github.com/flipt-io/flipt/releases/tag/v1.9.1) - 2022-09-01

### Fixed

- Issue when not using `netgo` build tag during build, resulting in native dns resolution not working for some cases in k8s. See  [https://github.com/flipt-io/flipt/issues/993](https://github.com/flipt-io/flipt/issues/993).

## [v1.9.0](https://github.com/flipt-io/flipt/releases/tag/v1.9.0) - 2022-07-06

### Changed

- Module name changed to `go.flipt.io/flipt` [https://github.com/flipt-io/flipt/pull/898](https://github.com/flipt-io/flipt/pull/898)
- Upgraded NodeJS to v18 [https://github.com/flipt-io/flipt/pull/911](https://github.com/flipt-io/flipt/pull/911)
- Removed Yarn in favor of NPM [https://github.com/flipt-io/flipt/pull/916](https://github.com/flipt-io/flipt/pull/916)
- Switched to ViteJS for UI build instead of Webpack [https://github.com/flipt-io/flipt/pull/924](https://github.com/flipt-io/flipt/pull/924)
- All UI dependencies are now bundled as well, instead of pulling from external sources (e.g. FontAwesome) [https://github.com/flipt-io/flipt/pull/924](https://github.com/flipt-io/flipt/pull/924)
- Telemetry no longer outputs log messages in case of errors or in-ability to connect [https://github.com/flipt-io/flipt/pull/926](https://github.com/flipt-io/flipt/pull/926)
- Telemetry will not run in dev or snapshot mode [https://github.com/flipt-io/flipt/pull/926](https://github.com/flipt-io/flipt/pull/926)
- Dependency updates

## [v1.8.3](https://github.com/flipt-io/flipt/releases/tag/v1.8.3) - 2022-06-08

### Changed

- Re-added ability to create rules for flags without any variants [https://github.com/flipt-io/flipt/issues/874](https://github.com/flipt-io/flipt/issues/874)

## [v1.8.2](https://github.com/flipt-io/flipt/releases/tag/v1.8.2) - 2022-04-27

### Fixed

- Broken rules reordering resulting from GRPC update [https://github.com/flipt-io/flipt/pull/836](https://github.com/flipt-io/flipt/pull/836)

## [v1.8.1](https://github.com/flipt-io/flipt/releases/tag/v1.8.1) - 2022-04-18

### Changed

- Updated telemetry to not run if `CI` is set
- Updated telemtry to flush on each batch
- Dependency updates

### Fixed

- Update CORS middleware to handle Vary header properly [https://github.com/flipt-io/flipt/pull/803](https://github.com/flipt-io/flipt/pull/803)

## [v1.8.0](https://github.com/flipt-io/flipt/releases/tag/v1.8.0) - 2022-04-06

### Added

- Helm Chart [https://github.com/flipt-io/flipt/pull/777](https://github.com/flipt-io/flipt/pull/777)
- Basic anonymous telemetry [https://github.com/flipt-io/flipt/pull/790](https://github.com/flipt-io/flipt/pull/790). Can be disabled via config.

### Changed

- Updated protoc/protobuf to v1.28 [https://github.com/flipt-io/flipt/pull/768](https://github.com/flipt-io/flipt/pull/768)
- Updated CODE_OF_CONDUCT.md with my new email address
- Updated README.md with link to [Flipt Sandbox](https://try.flipt.io)
- Updated README.md with link to Discord

## [v1.7.0](https://github.com/flipt-io/flipt/releases/tag/v1.7.0) - 2022-03-22

### Added

- Ability to quickly copy constraints/variants for easier creation [https://github.com/flipt-io/flipt/pull/754](https://github.com/flipt-io/flipt/pull/754)

### Changed

- Disallow empty rule creation in UI [https://github.com/flipt-io/flipt/issues/758](https://github.com/flipt-io/flipt/issues/758)
- `trimpath` added as [`go build` flag](https://pkg.go.dev/cmd/go#hdr-Compile_packages_and_dependencies) for builds [https://github.com/flipt-io/flipt/pull/722](https://github.com/flipt-io/flipt/pull/722)
- Base alpine image updated to 3.15.1 [https://github.com/flipt-io/flipt/pull/757](https://github.com/flipt-io/flipt/pull/757)
- Dependency updates

## [v1.6.3](https://github.com/flipt-io/flipt/releases/tag/v1.6.3) - 2022-02-21

### Added

- Test for failure to increment expected migration versions [https://github.com/flipt-io/flipt/issues/706](https://github.com/flipt-io/flipt/issues/706)
- All dependency licenses [https://github.com/flipt-io/flipt/pull/714](https://github.com/flipt-io/flipt/pull/714)

### Changed

- Dependency updates

### Fixed

- Potential null pointer bugs in importer found by fuzzing [https://github.com/flipt-io/flipt/pull/713](https://github.com/flipt-io/flipt/pull/713)

## [v1.6.2](https://github.com/flipt-io/flipt/releases/tag/v1.6.2) - 2022-02-19

### Fixed

- Issue with missing Segment.MatchType in export [https://github.com/flipt-io/flipt/issues/710](https://github.com/flipt-io/flipt/issues/710)
- Issue with version not showing in UI (again)

## [v1.6.1](https://github.com/flipt-io/flipt/releases/tag/v1.6.1) - 2022-02-13

### Fixed

- Issue where migrations were not checked against latest version
- Issue where version was not showing in UI

## [v1.6.0](https://github.com/flipt-io/flipt/releases/tag/v1.6.0) - 2022-02-13

### Added

- Flipt now shows if there is an update available in the UI [https://github.com/flipt-io/flipt/pull/650](https://github.com/flipt-io/flipt/pull/650). Can be disabled via config.
- Variants now support JSON attachments :tada: ! [https://github.com/flipt-io/flipt/issues/188](https://github.com/flipt-io/flipt/issues/188)
- Import/Export of variant attachment JSON marshal as YAML for human readability [https://github.com/flipt-io/flipt/issues/697](https://github.com/flipt-io/flipt/issues/697)

### Changed

- Dependency updates
- Update JS to ES6 syntax
- Flipt now runs without root user in Docker [https://github.com/flipt-io/flipt/pull/659](https://github.com/flipt-io/flipt/pull/659)
- Changed development task runner to [Task](https://taskfile.dev/#/) from `make`
- Re-configured how Flipt is built in a [devcontainer](https://code.visualstudio.com/docs/remote/devcontainer-cli#_building-a-dev-container-image)

## [v1.5.1](https://github.com/flipt-io/flipt/releases/tag/v1.5.1) - 2022-01-26

### Fixed

- Backwards compatability issue with using `null` as a string field in the `context` map for `Evaluate`. [https://github.com/flipt-io/flipt/issues/664](https://github.com/flipt-io/flipt/issues/664)

## [v1.5.0](https://github.com/flipt-io/flipt/releases/tag/v1.5.0) - 2022-01-11

### Changed

- Dependency updates
- Upgrade UI packages and fix linter errors [https://github.com/flipt-io/flipt/pull/625](https://github.com/flipt-io/flipt/pull/625)
- Upgraded required nodeJS version to 16

## [v1.4.0](https://github.com/flipt-io/flipt/releases/tag/v1.4.0) - 2021-08-07

### Added

- Ability to exclude `NotFound` errors in `BatchEvaluate` calls [https://github.com/flipt-io/flipt/pull/518](https://github.com/flipt-io/flipt/pull/518)

### Changed

- Dependency updates
- Add commit hash to dev build [https://github.com/flipt-io/flipt/pull/517](https://github.com/flipt-io/flipt/pull/517)
- Remove `packr` in favor of using Go 1.16 `go embed` for embedding assets [https://github.com/flipt-io/flipt/pull/492](https://github.com/flipt-io/flipt/pull/492)
- Updated process to create new releases [https://github.com/flipt-io/flipt/pull/482](https://github.com/flipt-io/flipt/pull/482)

### Fixed

- Bug when trying to add two variants of a flag to the same rule in the UI [https://github.com/flipt-io/flipt/issues/515](https://github.com/flipt-io/flipt/issues/515)

## [v1.3.0](https://github.com/flipt-io/flipt/releases/tag/v1.3.0) - 2021-06-14

### Changed

- Bunch of dependencies updated.
- [Development](DEVELOPMENT.md) updated to work with VS Remote Containers / GH Codespaces.
- Development docker image changed to be Debian based.

### Fixed

- Segment search autocompletion now works correctly when defining a rule [https://github.com/flipt-io/flipt/issues/462](https://github.com/flipt-io/flipt/issues/462)

## [v1.2.1](https://github.com/flipt-io/flipt/releases/tag/v1.2.1) - 2021-03-09

### Fixed

- Downgrade Buefy dependency to fix [https://github.com/flipt-io/flipt/issues/394](https://github.com/flipt-io/flipt/issues/394) and [https://github.com/flipt-io/flipt/issues/391](https://github.com/flipt-io/flipt/issues/391)

## [v1.2.0](https://github.com/flipt-io/flipt/releases/tag/v1.2.0) - 2021-02-14 :heart:

### Changed

- Don't error on evaluating flags that are disabled, return no match instead [https://github.com/flipt-io/flipt/pull/382](https://github.com/flipt-io/flipt/pull/382)
- Dependency updates

## [v1.1.0](https://github.com/flipt-io/flipt/releases/tag/v1.1.0) - 2021-01-15

### Changed

- Bumped dependencies
- Ignore disabled flags on batch evaluate instead of erroring [https://github.com/flipt-io/flipt/pull/376](https://github.com/flipt-io/flipt/pull/376)

## [v1.0.0](https://github.com/flipt-io/flipt/releases/tag/v1.0.0) - 2020-10-31

Happy Halloween! Flipt goes 1.0.0 today! :jack_o_lantern:

### Changed

- Bumped dependencies
- Upgrade to Go 1.15

## [v0.18.1](https://github.com/flipt-io/flipt/releases/tag/v0.18.1) - 2020-09-30

### Added

- Reflection to grpc server for usage with generic clients [https://github.com/flipt-io/flipt/pull/345](https://github.com/flipt-io/flipt/pull/345)

### Changed

- Bumped dependencies
- Added colorful output for new version available instead of normal log message
- Publishing Docker images to [ghcr.io](https://docs.github.com/en/free-pro-team@latest/packages/getting-started-with-github-container-registry/about-github-container-registry)

## [v0.18.0](https://github.com/flipt-io/flipt/releases/tag/v0.18.0) - 2020-08-02

### Added

- Ability to configure database without using URL [https://github.com/flipt-io/flipt/issues/316](https://github.com/flipt-io/flipt/issues/316)

## [v0.17.1](https://github.com/flipt-io/flipt/releases/tag/v0.17.1) - 2020-07-16

### Fixed

- Don't log database url/credentials on startup [https://github.com/flipt-io/flipt/issues/319](https://github.com/flipt-io/flipt/issues/319)

## [v0.17.0](https://github.com/flipt-io/flipt/releases/tag/v0.17.0) - 2020-07-10

### Added

- Check for newer versions of Flipt on startup. Can be disabled by setting `meta.check_for_updates=false` in config. [https://github.com/flipt-io/flipt/pull/311](https://github.com/flipt-io/flipt/pull/311)
- Ability to configure database connections [https://github.com/flipt-io/flipt/pull/313](https://github.com/flipt-io/flipt/pull/313)
- Prometheus metrics around database connections [https://github.com/flipt-io/flipt/pull/314](https://github.com/flipt-io/flipt/pull/314)
- OpenTracing/Jaeger support [https://github.com/flipt-io/flipt/pull/315](https://github.com/flipt-io/flipt/pull/315)

### Changed

- Update FQDN for cache metrics

## [v0.16.1](https://github.com/flipt-io/flipt/releases/tag/v0.16.1) [Backport] - 2020-07-16

### Fixed

- Don't log database url/credentials on startup [https://github.com/flipt-io/flipt/issues/319](https://github.com/flipt-io/flipt/issues/319)

## [v0.16.0](https://github.com/flipt-io/flipt/releases/tag/v0.16.0) - 2020-06-29

### Added

- MySQL support: [https://github.com/flipt-io/flipt/issues/224](https://github.com/flipt-io/flipt/issues/224)

## [v0.15.0](https://github.com/flipt-io/flipt/releases/tag/v0.15.0) - 2020-06-03

### Added

- Batch Evaluation [https://github.com/flipt-io/flipt/issues/61](https://github.com/flipt-io/flipt/issues/61)

## [v0.14.1](https://github.com/flipt-io/flipt/releases/tag/v0.14.1) - 2020-05-27

### Changed

- Colons are no longer allowed in flag or segment keys [https://github.com/flipt-io/flipt/issues/262](https://github.com/flipt-io/flipt/issues/262)

## [v0.14.0](https://github.com/flipt-io/flipt/releases/tag/v0.14.0) - 2020-05-16

### Added

- Ability to import from STDIN [https://github.com/flipt-io/flipt/issues/278](https://github.com/flipt-io/flipt/issues/278)

### Changed

- Updated several dependencies

## [v0.13.1](https://github.com/flipt-io/flipt/releases/tag/v0.13.1) - 2020-04-05

### Added

- End to end UI tests [https://github.com/flipt-io/flipt/issues/216](https://github.com/flipt-io/flipt/issues/216)

### Changed

- Updated several dependencies

### Fixed

- Variant/Constraint columns always showing 'yes' in UI [https://github.com/flipt-io/flipt/issues/258](https://github.com/flipt-io/flipt/issues/258)

## [v0.13.0](https://github.com/flipt-io/flipt/releases/tag/v0.13.0) - 2020-03-01

### Added

- `export` and `import` commands to export and import data [https://github.com/flipt-io/flipt/issues/225](https://github.com/flipt-io/flipt/issues/225)
- Enable response time histogram metrics for Prometheus [https://github.com/flipt-io/flipt/pull/234](https://github.com/flipt-io/flipt/pull/234)

### Changed

- List calls are no longer cached because of pagination

### Fixed

- Issue where `GetRule` would not return proper error if rule did not exist

## [v0.12.1](https://github.com/flipt-io/flipt/releases/tag/v0.12.1) - 2020-02-18

### Fixed

- Issue where distributions did not always maintain order during evaluation when using Postgres [https://github.com/flipt-io/flipt/issues/229](https://github.com/flipt-io/flipt/issues/229)

## [v0.12.0](https://github.com/flipt-io/flipt/releases/tag/v0.12.0) - 2020-02-01

### Added

- Caching support for segments, rules AND evaluation! [https://github.com/flipt-io/flipt/issues/100](https://github.com/flipt-io/flipt/issues/100)
- `cache.memory.expiration` configuration option
- `cache.memory.eviction_interval` configuration option

### Changed

- Fixed documentation link in app
- Underlying caching library from golang-lru to go-cache

### Removed

- `cache.memory.items` configuration option

## [v0.11.1](https://github.com/flipt-io/flipt/releases/tag/v0.11.1) - 2020-01-28

### Changed

- Moved evaluation logic to be independent of datastore
- Updated several dependencies
- Moved documentation to [https://github.com/flipt-io/flipt.io](https://github.com/flipt-io/flipt.io)
- Updated documentation link in UI

### Fixed

- Potential index out of range issue with 0% distributions: [https://github.com/flipt-io/flipt/pull/213](https://github.com/flipt-io/flipt/pull/213)

## [v0.11.0](https://github.com/flipt-io/flipt/releases/tag/v0.11.0) - 2019-12-01

### Added

- Ability to match ANY constraints for a segment: [https://github.com/flipt-io/flipt/issues/180](https://github.com/flipt-io/flipt/issues/180)
- Pagination for Flag/Segment grids
- Required fields in API Swagger Documentation

### Changed

- All JSON fields in responses are returned, even empty ones
- Various UI Tweaks: [https://github.com/flipt-io/flipt/issues/190](https://github.com/flipt-io/flipt/issues/190)

## [v0.10.6](https://github.com/flipt-io/flipt/releases/tag/v0.10.6) - 2019-11-28

### Changed

- Require value for constraints unless operator is one of `[empty, not empty, present, not present]`

### Fixed

- Fix issue where != would evaluate to true if constraint value was not present: [https://github.com/flipt-io/flipt/pull/193](https://github.com/flipt-io/flipt/pull/193)

## [v0.10.5](https://github.com/flipt-io/flipt/releases/tag/v0.10.5) - 2019-11-25

### Added

- Evaluation benchmarks: [https://github.com/flipt-io/flipt/pull/185](https://github.com/flipt-io/flipt/pull/185)

### Changed

- Update UI dependencies: [https://github.com/flipt-io/flipt/pull/183](https://github.com/flipt-io/flipt/pull/183)
- Update go-sqlite3 version

### Fixed

- Calculate distribution percentages so they always add up to 100% in UI: [https://github.com/flipt-io/flipt/pull/189](https://github.com/flipt-io/flipt/pull/189)

## [v0.10.4](https://github.com/flipt-io/flipt/releases/tag/v0.10.4) - 2019-11-19

### Added

- Example using Prometheus for capturing metrics: [https://github.com/flipt-io/flipt/pull/178](https://github.com/flipt-io/flipt/pull/178)

### Changed

- Update Go versions to 1.13.4
- Update Makefile to only build assets on change
- Update go-sqlite3 version

### Fixed

- Remove extra dashes when auto-generating flag/segment key in UI: [https://github.com/flipt-io/flipt/pull/177](https://github.com/flipt-io/flipt/pull/177)

## [v0.10.3](https://github.com/flipt-io/flipt/releases/tag/v0.10.3) - 2019-11-15

### Changed

- Update swagger docs to be built completely from protobufs: [https://github.com/flipt-io/flipt/pull/175](https://github.com/flipt-io/flipt/pull/175)

### Fixed

- Handle flags/segments not found in UI: [https://github.com/flipt-io/flipt/pull/175](https://github.com/flipt-io/flipt/pull/175)

## [v0.10.2](https://github.com/flipt-io/flipt/releases/tag/v0.10.2) - 2019-11-11

### Changed

- Updated grpc and protobuf versions
- Updated spf13/viper version

### Fixed

- Update chi compress middleware to fix large number of memory allocations

## [v0.10.1](https://github.com/flipt-io/flipt/releases/tag/v0.10.1) - 2019-11-09

### Changed

- Use go 1.13 style errors
- Updated outdated JS dependencies
- Updated prometheus client version

### Fixed

- Inconsistent matching of rules: [https://github.com/flipt-io/flipt/issues/166](https://github.com/flipt-io/flipt/issues/166)

## [v0.10.0](https://github.com/flipt-io/flipt/releases/tag/v0.10.0) - 2019-10-20

### Added

- Ability to write logs to file instead of STDOUT: [https://github.com/flipt-io/flipt/issues/141](https://github.com/flipt-io/flipt/issues/141)

### Changed

- Automatically populate flag/segment key based on name: [https://github.com/flipt-io/flipt/pull/155](https://github.com/flipt-io/flipt/pull/155)

## [v0.9.0](https://github.com/flipt-io/flipt/releases/tag/v0.9.0) - 2019-10-02

### Added

- Support evaluating flags without variants: [https://github.com/flipt-io/flipt/issues/138](https://github.com/flipt-io/flipt/issues/138)

### Changed

- Dropped support for Go 1.12

### Fixed

- Segments not matching on all constraints: [https://github.com/flipt-io/flipt/issues/140](https://github.com/flipt-io/flipt/issues/140)
- Modal content streching to fit entire screen

## [v0.8.0](https://github.com/flipt-io/flipt/releases/tag/v0.8.0) - 2019-09-15

### Added

- HTTPS support

## [v0.7.1](https://github.com/flipt-io/flipt/releases/tag/v0.7.1) - 2019-07-25

### Added

- Exposed errors metrics via Prometheus

### Changed

- Updated JS dev dependencies
- Updated grpc and protobuf versions
- Updated pq version
- Updated go-sqlite3 version

## [v0.7.0](https://github.com/flipt-io/flipt/releases/tag/v0.7.0) - 2019-07-07

### Added

- CORS support with `cors` config options
- [Prometheus](https://prometheus.io/) metrics exposed at `/metrics`

## [v0.6.1](https://github.com/flipt-io/flipt/releases/tag/v0.6.1) - 2019-06-14

### Fixed

- Missing migrations folders in release archive: [https://github.com/flipt-io/flipt/issues/97](https://github.com/flipt-io/flipt/issues/97)

## [v0.6.0](https://github.com/flipt-io/flipt/releases/tag/v0.6.0) - 2019-06-10

### Added

- `migrate` subcommand to run database migrations
- 'Has Prefix' and 'Has Suffix' constraint operators

### Changed

- Variant keys are now only required to be unique per flag, not globally: [https://github.com/flipt-io/flipt/issues/87](https://github.com/flipt-io/flipt/issues/87)

### Removed

- `db.migrations.auto` in config. DB migrations must now be run explicitly with the `flipt migrate` command

## [v0.5.0](https://github.com/flipt-io/flipt/releases/tag/v0.5.0) - 2019-05-27

### Added

- Beta support for Postgres! :tada:
- `/meta/info` endpoint for version/build info
- `/meta/config` endpoint for running configuration info

### Changed

- `cache.enabled` config becomes `cache.memory.enabled`
- `cache.size` config becomes `cache.memory.items`
- `db.path` config becomes `db.url`

### Removed

- `db.name` in config

## [v0.4.2](https://github.com/flipt-io/flipt/releases/tag/v0.4.2) - 2019-05-12

### Fixed

- Segments with no constraints now match all requests by default: [https://github.com/flipt-io/flipt/issues/60](https://github.com/flipt-io/flipt/issues/60)
- Clear Debug Console response on error

## [v0.4.1](https://github.com/flipt-io/flipt/releases/tag/v0.4.1) - 2019-05-11

### Added

- `/debug/pprof` [pprof](https://golang.org/pkg/net/http/pprof/) endpoint for profiling

### Fixed

- Issue in evaluation: [https://github.com/flipt-io/flipt/issues/63](https://github.com/flipt-io/flipt/issues/63)

## [v0.4.0](https://github.com/flipt-io/flipt/releases/tag/v0.4.0) - 2019-04-06

### Added

- `ui` config section to allow disabling the ui:

  ```yaml
  ui:
    enabled: true
  ```

- `/health` HTTP healthcheck endpoint

### Fixed

- Issue where updating a Constraint or Variant via the UI would not show the update values until a refresh: [https://github.com/flipt-io/flipt/issues/43](https://github.com/flipt-io/flipt/issues/43)
- Potential IndexOutOfRange error if distribution percentage didn't add up to 100: [https://github.com/flipt-io/flipt/issues/42](https://github.com/flipt-io/flipt/issues/42)

## [v0.3.0](https://github.com/flipt-io/flipt/releases/tag/v0.3.0) - 2019-03-03

### Changed

- Renamed generated proto package to `flipt` for use with external GRPC clients
- Updated docs and example to reference [GRPC go client](https://github.com/flipt-io/flipt-grpc-go)

### Fixed

- Don't return error on graceful shutdown of HTTP server

## [v0.2.0](https://github.com/flipt-io/flipt/releases/tag/v0.2.0) - 2019-02-24

### Added

- `server` config section to consolidate and rename `host`, `api.port` and `backend.port`:

  ```yaml
  server:
    host: 127.0.0.1
    http_port: 8080
    grpc_port: 9000
  ```

- Implemented flag caching! Preliminary testing shows about a 10x speedup for retrieving flags with caching enabled. See the docs for more info.

  ```yaml
  cache:
    enabled: true
  ```

### Deprecated

- `host`, `api.port` and `backend.port`. These values have been moved and renamed under the `server` section and will be removed in the 1.0 release.

## [v0.1.0](https://github.com/flipt-io/flipt/releases/tag/v0.1.0) - 2019-02-19

### Added

- Moved proto/client code to proto directory and added MIT License

## [v0.0.0](https://github.com/flipt-io/flipt/releases/tag/v0.0.0) - 2019-02-16

Initial Release!
