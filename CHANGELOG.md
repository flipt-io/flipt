# Changelog

This format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/)
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

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
