# Changelog

This format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/)
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [2.5.1](https://github.com/flipt-io/flipt/releases/tag/v2.5.1) - 2026-02-05

### Fixed

- **Git ref update after fetch**: Complete ref update after fetch for both bare and normal repositories, ensuring refs/heads are properly created and maintained for shallow fetches (#5144)
- **Annual license key format**: Handle annual license key format correctly in the activation wizard (#5335)
- **GitHub auth error messages**: Improve error message specificity when users fail to meet organization/team security requirements during GitHub authentication (#5331)
- **SSH URLs with non-standard ports**: Support SSH URLs with non-standard ports in Git configuration (#5298)

### Changed

- **gRPC Prometheus metrics**: Migrate from deprecated go-grpc-prometheus to go-grpc-middleware/providers/prometheus (#5272)

### Dependencies

- Updated various dependencies including OPA, ClickHouse client, gRPC ecosystem, and UI libraries

## [2.5.0](https://github.com/flipt-io/flipt/releases/tag/v2.5.0) - 2026-01-09

### Added

- **Custom commit and PR templates**: Configure server-wide default templates for commit messages, PR titles, and PR bodies per storage backend (#5091)
- **GitHub auth fallbacks**: Provide sensible fallback values for name and email when GitHub users don't have public profile information configured (#5154)
- **SSH remote URL normalization**: Automatically normalize HTTPS URLs to SCP-style format when SSH credentials are provided, ensuring Git operations work correctly (#5130)

### Fixed

- **OIDC secret resolution**: Resolve secret references (`${secret:file:...}`) in OIDC provider configuration for `client_id` and `client_secret` fields (#5153)

### Dependencies

- Updated various dependencies including OpenTelemetry, OPA, gRPC ecosystem, Redis client, and UI libraries

## [2.4.0](https://github.com/flipt-io/flipt/releases/tag/v2.4.0) - 2025-11-20

### Added

- **OpenTelemetry modernization**: Migrate from deprecated gRPC interceptors to modern stats handler API for improved tracing instrumentation (#4895)

### Fixed

- **Git remote synchronization**: Automatically sync git remote URLs with Flipt configuration to ensure repository connections stay current (#4975)
- **Git connection error handling**: Properly detect and handle DNS resolution errors and I/O timeouts when checking git repository connections, allowing lenient fetch policy to work correctly when remotes are temporarily unreachable (#4965)
- **Distributed tracing coverage**: Apply OpenTelemetry stats handler to both the main gRPC server and in-process channel for complete tracing coverage (#4971)
- **UI form submission**: Enable namespace form submit button based on form validation state for better user experience (#4991)

### Dependencies

- Updated various dependencies including OpenTelemetry, AWS SDK, gRPC ecosystem, and UI libraries

## [2.3.1](https://github.com/flipt-io/flipt/releases/tag/v2.3.1) - 2025-10-29

### Fixed

- **Evaluation environment lookup**: Extract duplicate environment retrieval logic with context fallback into reusable helper method for consistent environment lookup behavior (#4959)
- **Configuration schema**: Use snake_case for fetch_policy in JSON schema to match expected configuration format (#4957)

## [2.3.0](https://github.com/flipt-io/flipt/releases/tag/v2.3.0) - 2025-10-28

### Added

- **Configurable flag metadata in ListFlags**: New server configuration option to control whether flag metadata is included in ListFlags API responses, allowing operators to expose metadata publicly when needed (#4934, #4848)
- **Git fetch policy configuration**: Add `fetch_policy` configuration option to control behavior when remote git fetch fails during startup, allowing Flipt to continue operating with stale local data (#4920)

### Fixed

- **Install script compatibility**: Prevent install script failure in minimal terminals (TERM=dumb) by adding tput capability check before setting color variables (#4928)
- **Segment deletion validation**: Prevent deletion of segments that are referenced by flag rules or rollouts to ensure data integrity (#4879)
- **SDK default environment compatibility**: Ensure proper compatibility with SDKs using default environment (#4857)
- **OpenTelemetry batch evaluation**: Add missing OpenTelemetry event tracking to batch evaluation endpoint for consistent observability (#4875)
- **UI form submission**: Prevent Enter key from submitting form when adding tags in constraint values, now properly adds tag instead (#4896)
- **Redis configuration schema**: Set default Redis mode to "single" instead of empty string for more explicit configuration (#4926)
- **OpenAPI generation**: Generate OpenAPI specs for v2 API using gnostic (#4847)

### Changed

- **Authentication error handling**: Improve code style and error handling in authentication method utilities (#4775)
- **GitHub API**: Update to go-github/v75 for latest GitHub API features (#4866)
- **Contributor workflow**: Optimize contributor workflow to use check_suite pattern (#4819)
- **OpenTelemetry dependencies**: Group OpenTelemetry dependency updates for better management (#4820)

### Dependencies

- Updated various dependencies including OpenTelemetry group, grpc-ecosystem middleware, ClickHouse client, OPA, and cloud storage libraries

## [2.2.0](https://github.com/flipt-io/flipt/releases/tag/v2.2.0) - 2025-10-02

### Added

- **CLI license management**: New `flipt license` command with interactive TUI wizard for managing Flipt Pro licenses (#4767)
- **Boolean evaluation**: Include segments in boolean flag evaluation response for consistency (#4767)

### Fixed

- **UI environment sync**: Refetch environments when window regains focus or network reconnects to ensure UI stays in sync (#4774)
- **Git branch cleanup**: Remove deleted branches from UI when synced from Git repositories (#4770)
- **Docker build**: Correct ARG declaration placement in Dockerfile to resolve build issues (#4772)
- **OpenTelemetry compatibility**: Resolve schema URL conflicts and update to compatible exporter versions

### Changed

- **OpenTelemetry conventions**: Align event and attribute names with OpenTelemetry semantic conventions specification (#4769)

### Dependencies

- Updated various UI dependencies including @mui/material, @mui/x-charts, and development tools

## [2.1.3](https://github.com/flipt-io/flipt/releases/tag/v2.1.3) - 2025-09-24

### Fixed

- **Analytics batch processing**: Return all batch data for requested flags to prevent incomplete analytics results (#4756)
- **Memory optimization**: Reduce memory usage with OpenTelemetry and ClickHouse integrations (#4755)
- **Authentication middleware**: Pass correct server parameter to skipped() method in EmailMatchingUnaryInterceptor
- **Nightly builds**: Use static v2-nightly version for consistent nightly build versioning

### Changed

- **Documentation**: Added nightly build Docker instructions to README for improved developer experience

## [2.1.2](https://github.com/flipt-io/flipt/releases/tag/v2.1.2) - 2025-09-20

### Fixed

- **Segment validation**: Validate segment references exist when creating flags via v2 API to prevent snapshot creation errors and inconsistent state (#4749)
- **OFREP authentication**: Allow authentication to be excluded for OFREP endpoints when configured (#4753)

### Changed

- **Developer setup**: Improved devenv setup and fixed UI styles loading issue in development mode (#4750)

### Dependencies

- Updated various dependencies including gqlgen, gnostic, and golang.org/x/exp

## [2.1.1](https://github.com/flipt-io/flipt/releases/tag/v2.1.1) - 2025-09-13

### Fixed

- **Authentication middleware**: Replace panic with proper error handling in EmailMatchingUnaryInterceptor when authentication not found in context (#4744)
- **Streaming evaluation**: Get environment key from streaming subscribe request (#4733)
- **HTTP/2 streaming**: Accept HTTP/2 connections for streaming evaluation (#4715)
- **Environment context**: Return error from GetFromContext when environment not found (#4745)

### Changed

- **Go version**: Remove toolchain in go.mod files and set Go directives to 1.25 (#4746)

### Dependencies

- Security update for Vite to v6.3.6 (#4721)
- Updated various dependencies including go-bitbucket, go-jose, go-billy, and bubbletea

## [2.1.0](https://github.com/flipt-io/flipt/releases/tag/v2.1.0) - 2025-09-04

### Added

- **Quickstart TUI**: New terminal user interface for the quickstart command using Bubbletea framework (#4639)
- **Flipt Pro trial banner**: Added promotional banner in UI for OSS users to discover Pro features and start 14-day trial (#4660)
- **Git repository initialization support**: Automatic detection and initialization for existing features files when using local storage backend (#4674)
- **YAML file extension support**: Added support for `.yml` extension for features files in addition to `.yaml` (#4681)

### Fixed

- **Zero value serialization**: Fixed validation errors when using rollout value of 0 in variant flag distributions and percentage value of 0 in boolean flag threshold rollouts (#4678)
- **Namespace deletion**: Handle empty directories correctly during namespace deletion (#4622)

### Changed

- **Default namespace display**: Always show default namespace first in namespace table (#4667)
- **UI component organization**: Consolidated UI components into main components directory for better structure (#4661)
- **Bitbucket SCM support**: Added Bitbucket to supported SCM providers schema (#4666)

### Dependencies

- Bumped various UI dependencies including `@uiw/codemirror-theme-tokyo-night`, `@tanstack/react-table`, and `lucide-react`
- Updated `github.com/clickhouse/clickhouse-go/v2` and `github.com/go-chi/chi/v5`

## [2.0.2](https://github.com/flipt-io/flipt/releases/tag/v2.0.2) - 2025-08-27

### Fixed

- Comprehensive Keygen API rate limit handling improvements including caching, proper error type handling, and intelligent rate limit management (#4602)
- Process Go files in batches for `mage go:fmt` command to avoid "argument list too long" errors (#4570)

### Changed

- Replaced Dependabot with Renovate for dependency management (#4594, #4603)
- Added AI-powered release notes generation for Homebrew releases (#4604)

### Dependencies

- Bumped React and @types/react in UI (#4582)
- Bumped gitlab.com/gitlab-org/api/client-go from 0.129.0 to 0.139.2 (#4575)
- Bumped @uiw/react-codemirror from 4.24.2 to 4.25.1 in UI (#4580)
- Bumped @radix-ui/react-popover from 1.1.14 to 1.1.15 in UI (#4581)
- Bumped @radix-ui/react-select from 2.2.5 to 2.2.6 in UI (#4579)
- Bumped @radix-ui/react-switch from 1.2.5 to 1.2.6 in UI (#4577)
- Bumped github.com/go-chi/cors from 1.2.1 to 1.2.2 (#4573)
- Bumped github.com/hashicorp/go-retryablehttp (#4572)
- Bumped @playwright/test from 1.54.1 to 1.54.2 in UI (#4578)
- Bumped eslint-plugin-prettier from 5.2.3 to 5.5.4 in UI (#4526)
- Bumped actions/checkout from 4 to 5 in GitHub Actions (#4576)

## [2.0.1](https://github.com/flipt-io/flipt/releases/tag/v2.0.1) - 2025-08-16

### Added

- Include `type` and `enabled` flag attributes for v1 ListFlags API call (#4562)

### Fixed

- Add `x-flipt-accept-server-version` to default CORS allowed headers (#4564)
- Correct architecture naming in v2 install script for x86_64 systems (#4555)
- Install script no longer incorrectly installs beta versions (#4553)

### Changed

- Refactored CSRF handling to use Go 1.25 CSRF handler (#4561)
- Upgraded Go from 1.24 to 1.25 (#4558)

### Dependencies

- Bumped TypeScript from 5.8.3 to 5.9.2 in UI (#4549)
- Bumped @uiw/react-codemirror from 4.23.8 to 4.24.2 in UI (#4548)
- Bumped lucide-react from 0.534.0 to 0.539.0 in UI (#4547)

## [2.0.0](https://github.com/flipt-io/flipt/releases/tag/v2.0.0) - 2025-08-12

Flipt v2 is a complete re-architecture of Flipt, focused on Git-native operations and multi-environment support. This major release represents a fundamental shift in how Flipt manages feature flags, bringing enterprise-grade GitOps workflows to feature management.

### Added

#### Core Features

- **Git-Native Architecture**: Feature flags are now stored directly in Git repositories with full read/write capabilities through the API and UI
- **Multi-Environment Support**: Manage feature flags across multiple environments using:
  - Different Git repositories per environment
  - Different directories within the same repository
  - Different branches within the same repository
- **Environment Branching**: Create isolated branches of any environment for testing changes without affecting production
- **Git SCM Integration** (Pro): Native integration with GitHub, GitLab, BitBucket, Azure DevOps, and Gitea for merge proposals and GPG-signed commits
- **Offline Mode**: Continue serving flags even when the source Git repository is temporarily unavailable

#### Security & Secrets Management

- **Secret References in Configuration**: Support for referencing secrets from environment variables and files using `{{ secret "key" }}` syntax
- **Integrated Secrets Management**: Built-in secure storage for GPG keys with HashiCorp Vault integration (Pro) and file system support (OSS)
- **GPG Commit Signing** (Pro): Sign all Git commits for maximum security and auditability
- **JWT Claims Mapping**: Configurable mapping of JWT claims to well-known authentication metadata

#### UI & Developer Experience

- **Redesigned UI**: Complete UI overhaul for improved user experience and intuitive navigation
- **Environment Switcher**: Seamless switching between environments in the UI
- **Merge Proposals**: Create pull/merge requests directly from the UI (Pro)

### Changed

- **Storage Backend**: Default storage is now Git-based (memory or disk) rather than database-backed
- **Configuration Format**: New configuration structure to support environments and Git storage options
- **API Versioning**: New v2 APIs for environments, analytics, and evaluation while maintaining v1 compatibility
- **License Model**: Introduction of Pro tier with enterprise features while maintaining OSS core
