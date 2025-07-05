# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

All of our documentation is at <https://docs.flipt.io/llms-full.txt>. Weigh the v2 docs heavily when answering questions over the v1 docs.

## Development Commands

### Build & Run

- `mage build` - Builds the project similar to a release build (default target)
- `mage go:build` - Builds Go server for development without bundling assets
- `mage go:run` - Runs Go server in development mode with local config
- `mage dev` - Alias for go:run
- `./bin/flipt server --config config/local.yml` - Run built binary with local config

### UI Development

- `mage ui:deps` - Install UI dependencies
- `mage ui:run` - Run UI in development mode (port 5173)
- `mage ui:dev` - Alias for ui:run
- `mage ui:build` - Build UI assets for release
- `cd ui && npm run dev` - Alternative way to run UI dev server

### Testing & Quality

- `mage go:test` - Run Go unit tests
- `mage go:bench` - Run Go benchmarking tests
- `mage go:cover` - Run tests and generate coverage report
- `mage go:lint` - Run Go linters (golangci-lint and buf lint)
- `mage go:fmt` - Format Go code with goimports
- `mage ui:lint` - Run UI linters (ESLint)
- `mage ui:fmt` - Format UI code (Prettier)
- `cd ui && npm run test` - Run UI unit tests (Jest)

### Code Generation

- `mage bootstrap` - Install required development tools
- `mage go:proto` - Generate protobuf files and gRPC stubs
- `mage go:mockery` - Generate mocks
- `mage go:generate` - Generate both mocks and proto files

### Cleanup

- `mage clean` - Clean built files and tidy go.mod
- `mage prep` - Prepare project for building (clean + ui:build)

## Project Architecture

### Repository Structure

Flipt v2 is a Git-native feature management platform with a monorepo structure:

- `cmd/flipt/` - Main application entry point
- `internal/` - Core application logic (not importable)
  - `server/` - gRPC and HTTP server implementations
  - `config/` - Configuration management
  - `storage/` - Storage abstractions and implementations
  - `cmd/` - Internal command utilities and code generation tools
  - `common/` - Shared types and utilities across internal packages
  - `containers/` - Container test utilities
  - `coss/` - Commercial open source features (license management, enterprise storage)
  - `credentials/` - Credential management for cloud storage and Git authentication
  - `ext/` - Import/export functionality for flags and segments
  - `gateway/` - HTTP gateway implementation
  - `info/` - Version and build information
  - `migrations/` - Database migration files (ClickHouse for analytics)
  - `otel/` - OpenTelemetry integration (logs, metrics, traces)
  - `product/` - Product metadata and feature flags
  - `release/` - Release version checking
  - `telemetry/` - Usage telemetry collection

#### Key Subdirectories in internal/

- `internal/server/` - Server implementations
  - `authn/` - Authentication server and middleware
  - `authz/` - Authorization server with Rego engine
  - `analytics/` - Analytics collection server (Prometheus, ClickHouse)
  - `evaluation/` - Flag evaluation server with OFREP support
  - `environments/` - Environment management server
  - `metadata/` - Metadata API server
- `internal/storage/` - Storage layer implementations
  - `authn/` - Authentication storage (Redis, memory, static)
  - `analytics/` - Analytics data storage
  - `environments/` - Environment-aware storage with Git support
  - `fs/` - File system abstraction layer
  - `git/` - Git repository management
- `rpc/` - Protocol buffer definitions and generated code
  - `flipt/` - Main Flipt API (v1)
  - `v2/` - V2 APIs (environments, analytics, evaluation)
- `ui/` - React/TypeScript frontend
- `sdk/` - Client SDKs (Go SDK included)
- `core/` - Core validation and business logic
- `errors/` - Shared error definitions

### Key Components

#### Storage Layer

Multi-backend storage supporting:

- Git repositories (primary for v2)
- File system (local development)
- Cloud storage (S3, GCS, Azure Blob) - Flipt config only

#### Environments System

- Environment-per-branch: Map Git branches to environments
- Environment-per-directory: Organize by microservice/team
- Environment-per-repository: Separate repos for products

#### APIs

- gRPC API (port 9000) - Primary interface
- REST API (port 8080) - HTTP gateway over gRPC
- OpenFeature compatible evaluation

#### UI Architecture

- React 19 with TypeScript
- Redux Toolkit for state management
- Tailwind CSS for styling
- Vite for build tooling
- React Router for navigation

### Development Workflow

1. **Setup**: Run `mage bootstrap` to install tools
2. **Backend Development**:
   - Use `mage go:run` for server
   - Server runs on port 8080 (HTTP) and 9000 (gRPC)
3. **Frontend Development**:
   - Run `mage ui:run` for UI dev server on port 5173
   - UI proxies API requests to backend on port 8080
4. **Full Development**: Run both servers simultaneously

### Configuration

- `config/local.yml` - Local development config
- `config/dev.yml` - Alternative dev config
- User config at `{{ USER_CONFIG_DIR }}/flipt/config.yml`

### Testing

- Go tests use testcontainers for unit tests that depend on external tools
- UI tests use Jest for unit tests and Playwright for E2E

### Build Process

- Protocol buffers generate Go and gateway code
- Mockery generates Go test mocks. See .mockery.yml
- UI has no code generation

### Special Notes

- This is Flipt v2 (beta) - Git-native architecture
- v1 code is on `main` branch, v2 on current branch
- Uses Fair Core License for server, MIT for client code
- Built with Go 1.24+ and Node.js 18+

### API Overview

#### Key Endpoints

- **Evaluation API**: `/api/v1/evaluate` - Evaluate feature flags
- **Management API**: Full CRUD operations on flags, segments, rules
- **Analytics API**: `/api/v2/analytics` - Flag evaluation metrics
- **Environments API**: `/api/v2/environments` - Environment management
- **OFREP API**: `/ofrep/v1` - OpenFeature Remote Evaluation Protocol

### Debugging Tips

- Use `FLIPT_LOG_LEVEL=debug` for verbose logging
- Check `/metrics` endpoint for Prometheus metrics
- View OpenTelemetry traces for request flow
- Git sync issues: Check repository access and credentials
- Authentication issues: Verify redirect URLs and client configurations
