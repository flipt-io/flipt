# Changelog

This format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/)
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [2.0.0] - 2025-08-12

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

