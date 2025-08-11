# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

All of our documentation is at <https://docs.flipt.io>. Weigh the v2 docs heavily when answering questions over the v1 docs.

## Project Overview

Flipt v2 is a Git-native feature management platform with a monorepo structure. This is the v2 version with a Git-native architecture.

- **v1 code**: Located on `main` branch
- **v2 code**: Located on current (`v2`) branch. This is the default branch in the repository.
- **License**: Fair Core License for server, MIT for client code
- **Requirements**: Go 1.24+ and Node.js 18+

## Project Architecture & Structure

### Repository Structure

- `cmd/flipt/` - Main application entry point
- `core/` - Core validation and business logic
- `errors/` - Shared error definitions
- `internal/` - Core application logic (not importable)
  - `cmd/` - Internal command utilities and code generation tools
  - `common/` - Shared types and utilities across internal packages
  - `config/` - Configuration management
  - `containers/` - Container test utilities
  - `coss/` - Commercial open source features (license management, enterprise storage, secrets)
  - `secrets/` - Secret management system with provider interface (OSS file provider, Pro Vault provider)
  - `credentials/` - Credential management for cloud storage and Git authentication
  - `ext/` - Import/export functionality for flags and segments
  - `gateway/` - HTTP gateway implementation
  - `info/` - Version and build information
  - `migrations/` - Database migration files (ClickHouse for analytics)
  - `otel/` - OpenTelemetry integration (logs, metrics, traces)
  - `product/` - Product metadata and feature flags
  - `release/` - Release version checking
  - `server/` - gRPC and HTTP server implementations
    - `analytics/` - Analytics collection server (Prometheus, ClickHouse)
    - `authn/` - Authentication server and middleware
    - `authz/` - Authorization server with Rego engine
    - `environments/` - Environment management server
    - `evaluation/` - Flag evaluation server with OFREP support
    - `metadata/` - Metadata API server
  - `storage/` - Storage abstractions and implementations
    - `analytics/` - Analytics data storage
    - `authn/` - Authentication storage (Redis, memory, static)
    - `environments/` - Environment-aware storage with Git support
    - `fs/` - File system abstraction layer
    - `git/` - Git repository management
  - `telemetry/` - Usage telemetry collection
- `rpc/` - Protocol buffer definitions and generated code
  - `flipt/` - Main Flipt API (v1)
  - `v2/` - V2 APIs (environments, analytics, evaluation)
- `sdk/` - Client SDKs (Go SDK included)
- `ui/` - React/TypeScript frontend

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

##### Key Endpoints

- **Evaluation API**: `/api/v1/evaluate` - Evaluate feature flags
- **Management API**: Full CRUD operations on flags, segments, rules
- **Analytics API**: `/api/v2/analytics` - Flag evaluation metrics
- **Environments API**: `/api/v2/environments` - Environment management
- **OFREP API**: `/ofrep/v1` - OpenFeature Remote Evaluation Protocol

#### UI Architecture

- React 19 with TypeScript
- Redux Toolkit for state management
- Tailwind CSS for styling
- Vite for build tooling
- React Router for navigation

## Development Setup & Commands

### Initial Setup

```bash
# Install required development tools
mage bootstrap
```

### Development Workflow

1. **Setup**: Run `mage bootstrap` to install tools
2. **Backend Development**:
   - Use `mage go:run` for server
   - Server runs on port 8080 (HTTP) and 9000 (gRPC)
3. **Frontend Development**:
   - Run `mage ui:run` for UI dev server on port 5173
   - UI proxies API requests to backend on port 8080
4. **Full Development**: Run both servers simultaneously

### Build & Run Commands

- `mage build` - Builds the project similar to a release build (default target)
- `mage go:build` - Builds Go server for development without bundling assets
- `mage go:run` - Runs Go server in development mode with local config
- `mage dev` - Alias for go:run
- `./bin/flipt server --config config/local.yml` - Run built binary with local config

### UI Development Commands

- `mage ui:deps` - Install UI dependencies
- `mage ui:run` - Run UI in development mode (port 5173)
- `mage ui:dev` - Alias for ui:run
- `mage ui:build` - Build UI assets for release
- `cd ui && npm run dev` - Alternative way to run UI dev server

### Testing Commands

- `mage go:test` - Run all Go unit tests
- `go test -v {path} -run {test}` - Run a specific Go test
- `mage go:bench` - Run Go benchmarking tests
- `mage go:cover` - Run tests and generate coverage report
- `cd ui && npm run test` - Run UI unit tests (Jest)

### Code Quality Commands

- `mage go:lint` - Run Go linters (golangci-lint and buf lint)
- `mage go:fmt` - Format Go code with goimports
- `mage ui:lint` - Run UI linters (ESLint)
- `mage ui:fmt` - Format UI code (Prettier)

### Code Generation Commands

- `mage go:proto` - Generate protobuf files and gRPC stubs
- `mage go:mockery` - Generate mocks
- `mage go:generate` - Generate both mocks and proto files

### Cleanup Commands

- `mage clean` - Clean built files and tidy go.mod
- `mage prep` - Prepare project for building (clean + ui:build)

### Build Process Notes

- Protocol buffers generate Go and gateway code
- Mockery generates Go test mocks. See .mockery.yml
- UI has no code generation

## Code Style Guidelines

### Go Code Style

#### Foundation: Google Go Style Guide

**All Go code should follow the [Google Go Style Guide](https://google.github.io/styleguide/go/guide) as the base style guide.**

The guidelines below are Flipt-specific conventions that build upon the Google style guide. When in doubt, refer to the Google guide first, then apply these project-specific patterns.

Key areas from the Google guide to pay special attention to:

- **Formatting**: Use `gofmt` (we use `goimports` which includes `gofmt`)
- **Naming**: Follow Go naming conventions for packages, types, functions, and variables
- **Package organization**: Keep packages focused and avoid circular dependencies
- **Error handling**: Return errors as the last return value, handle errors explicitly
- **Documentation**: Write clear godoc comments for exported types and functions

#### Naming Conventions

- **Public functions**: Use PascalCase (`NewServer`, `ListFlags`, `RegisterGRPC`)
- **Private functions**: Use camelCase (`getStore`, `buildConfig`, `handleError`)
- **Factory functions**: Use `New` prefix for constructors (`NewRepository`, `NewEnvironmentFactory`)
- **Variables**: Use camelCase (`environmentKey`, `namespaceKey`, `startTime`)
- **Constants**: Use PascalCase for exported, camelCase for unexported
- **Interfaces**: Use descriptive names (`Store`, `EnvironmentStore`, `Validator`)

#### Variable Declarations

```go
// Prefer := for single variable declarations
result := SomeType{}

// Use var blocks for multiple related variables
var (
    environmentKey = r.EnvironmentKey
    namespaceKey   = r.NamespaceKey
    startTime      = time.Now().UTC()
)
```

#### Constants

Use constants when appropriate for values that are unlikely to change. Even in tests.

**❌ Bad Examples**

```go
var (
    maxRetries = 3
    secretProvider = "vault"
)
```

```go
  maxRetries := 3
  secretProvider := "vault"
```

**✅ Good Example**

```go
const (
    maxRetries = 3
    secretProvider = "vault"
)
```

#### Error Handling

##### Error Types

Use the custom error types when possible:

```go
// Preferred - use custom error types
return nil, errs.ErrNotFoundf("flag %q not found", key)
return nil, errs.ErrInvalidf("invalid configuration: %s", reason)

// Fallback - standard error handling
return nil, fmt.Errorf("failed to process: %w", err)
```

##### Inline Errors

Prefer inline error declarations and checking when it makes sense.

**❌ Bad Example**

```go
err := doSomething()
if err != nil {
    // ...
}
```

**✅ Good Example**

```go
if err := doSomething(); err != nil {
    // ...
}
```

#### Imports

Follow the three-group import pattern (extends Google guide):

```go
import (
    // Standard library
    "context"
    "fmt"
    "time"

    // Third-party packages
    "github.com/go-git/go-git/v6"
    "go.uber.org/zap"

    // Local packages
    "go.flipt.io/flipt/errors"
    "go.flipt.io/flipt/internal/config"
)
```

#### Logging

Use structured logging with zap:

```go
// Preferred - structured logging with context
s.logger.Debug("processing request", 
    zap.String("environment", envKey),
    zap.String("namespace", nsKey),
    zap.Int("count", len(items)))

// Use appropriate log levels
s.logger.Debug("debug info for development")  // Most common
s.logger.Info("important application events")
s.logger.Warn("recoverable error conditions")
s.logger.Error("error conditions that need attention")

// Avoid overusing log statements - be selective about what to log
```

#### Comments and Documentation

```go
// Public functions need godoc comments starting with function name
// ListFlags lists all flags in the specified environment and namespace
func (s *Server) ListFlags(ctx context.Context, r *flipt.ListFlagRequest) (*flipt.FlagList, error) {
    // Inline comments explain complex logic
    // Check for X-Environment header for backward compatibility
    environmentKey := r.EnvironmentKey
    if headerEnv, ok := common.FliptEnvironmentFromContext(ctx); ok && headerEnv != "" {
        environmentKey = headerEnv
    }
}
```

#### Project Organization

- Use `internal/` for private packages that shouldn't be imported externally
- Organize by domain (`server/`, `storage/`, `config/`, `analytics/`)
- Keep interfaces in domain root, implementations in subdirectories
- Co-locate tests with implementation (`*_test.go`)
- Use `testdata/` directories for test fixtures

#### Pre-commit Quality Checks (Go)

**Always run these commands when adding or editing Go code:**

```bash
# Format Go code
mage go:fmt

# Lint Go code
mage go:lint

# Run modernize to update code to 1.24+ style
mage go:modernize
```

### UI/React/TypeScript Code Style

#### File Naming and Organization

- **Components**: Use PascalCase for `.tsx` files (`FlagForm.tsx`, `FlagTable.tsx`)
- **API files**: Use camelCase with suffix (`flagsApi.ts`, `authApi.ts`)
- **Types**: Use PascalCase for `.ts` files (`Flag.ts`, `Analytics.ts`)
- **Utilities**: Use descriptive lowercase names (`helpers.ts`)

#### Component Structure

- Use functional components with hooks
- Organize components by domain in directories
- Keep reusable UI components in `ui/` directory
- Co-locate component tests with implementation

#### State Management

- Use Redux Toolkit for application state
- Create API slices for domain-specific data (`flagsApi.ts`)
- Use TypeScript for type safety throughout

#### Pre-commit Quality Checks (UI)

**Always run these commands when adding or editing UI code:**

```bash
# Format UI code (TypeScript/React)
mage ui:fmt

# Lint UI code (ESLint)
mage ui:lint
```

### General Pre-commit Checks

```bash
# Format specific markdown files (replace with actual filename)
npx prettier -w CLAUDE.md
```

## Configuration

### Configuration Files

- `config/local.yml` - Local development config
- `config/dev.yml` - Alternative dev config
- User config at `{{ USER_CONFIG_DIR }}/flipt/config.yml`

### Debugging Tips

- Use `FLIPT_LOG_LEVEL=debug` for verbose logging
- Check `/metrics` endpoint for Prometheus metrics
- View OpenTelemetry traces for request flow
- Git sync issues: Check repository access and credentials
- Authentication issues: Verify redirect URLs and client configurations

## Testing Guidelines

### Test Structure

- Go tests use testcontainers for unit tests that depend on external tools
- UI tests use Jest for unit tests and Playwright for E2E

### Writing Effective Tests

When writing tests, ensure they actually verify the functionality being tested:

**❌ Bad Example - Test doesn't verify actual behavior:**

```go
// This test only verifies the method runs without error, not that the logic works
func TestFeature(t *testing.T) {
    result, err := service.DoSomething(input)
    require.NoError(t, err)
    assert.NotNil(t, result)
}
```

**✅ Good Example - Test verifies specific behavior:**

```go
// This test verifies the actual logic and behavior
func TestFeature(t *testing.T) {
    // Use mock.MatchedBy to verify the correct data is passed through
    store.On("Method", mock.MatchedBy(func(ctx context.Context) bool {
        // Verify the context contains the expected values
        value, ok := somePackage.ValueFromContext(ctx)
        return ok && value == expectedValue
    }), expectedArg).Return(expectedResult, nil)
    
    result, err := service.DoSomething(input)
    require.NoError(t, err)
    
    // Verify the actual business logic worked correctly
    assert.Equal(t, expectedOutput, result.ImportantField)
    store.AssertExpectations(t) // Ensures mocks were called as expected
}
```

**Key Testing Principles:**

- **Verify behavior, not just success**: Don't just check that methods return without error
- **Test the logic**: Use `mock.MatchedBy` to verify correct data flows through the system
- **Test edge cases**: Include tests for when conditions are met vs. not met
- **Use meaningful assertions**: Assert on the specific values that matter to the business logic
- **Mock verification**: Use `mock.AssertExpectations(t)` to ensure mocks were called with expected parameters

## Git Workflow & Commits

### Creating Effective Commits

#### Commit Message Format

Follow conventional commit format:

```
type(scope): brief description

Longer description if needed explaining the why, not the what.

- List specific changes if helpful
- Include any breaking changes
- Reference issues (e.g., "Fixes #123")
```

#### Commit Types

- `feat:` - New features
- `fix:` - Bug fixes
- `docs:` - Documentation changes
- `style:` - Code style changes (formatting, etc.)
- `refactor:` - Code refactoring without functional changes
- `test:` - Adding or modifying tests
- `chore:` - Build process or auxiliary tool changes

#### Examples

```bash
# Good commit messages
git commit -s -m "feat: add X-Environment header support to ListFlags endpoint

The v1 ListFlags endpoint now checks for X-Environment header 
for backward compatibility. When present, the header value 
takes precedence over the request environment parameter.

- Modified ListFlags method to check context for environment header
- Added ForwardFliptEnvironment middleware to v1 API gateway
- Added comprehensive test coverage

Fixes #4411"

git commit -s -m "fix: resolve race condition in flag evaluation cache"

git commit -s -m "test: improve X-Environment header test coverage

- Use mock.MatchedBy to verify context contains correct environment
- Add test for when header is not present
- Ensure tests actually validate the header precedence logic"
```

### Pull Request Guidelines

#### PR Title Format

Use the same conventional commit format for PR titles:

```
type(scope): brief description
```

#### PR Description Template

```markdown
## Summary

Brief description of what this PR does and why.

## Changes

- **Modified `file1.go`**: Specific change description
- **Added `file2_test.go`**: Test coverage description
- **Updated `config.yml`**: Configuration change description

## Backward Compatibility

Describe any backward compatibility or breaking change considerations.

## Additional Notes

Any other context, screenshots, or related issues.
```

#### PR Best Practices

- **Keep PRs focused**: One logical change per PR
- **Write descriptive titles**: Use conventional commit format
- **Provide context**: Explain the "why" behind changes
- **Include tests**: Always add tests for new functionality
- **Run quality checks**: Format and lint before creating PR
- **Reference issues**: Use "Fixes #123" to auto-close issues
- **Review your own PR**: Look through the diff before requesting review
- **Base PRs on the correct branch**: Base PRs on the correct branch (e.g. `v2` for v2 changes, `main` for v1 changes)
- **Label PRs correctly**: When creating a feature based off of the v2 branch, label the PR with the `v2` label. Similarly for a v1 (main) feature, use the `v1` label.

#### Quality Checklist

Before creating a PR, ensure:

- [ ] Code follows style guidelines
- [ ] Tests are added and passing
- [ ] Linting passes (`mage go:lint` / `mage ui:lint`)
- [ ] Code is formatted (`mage go:fmt` / `mage ui:fmt`)
- [ ] Commit messages are clear and descriptive
- [ ] PR description explains the change and its purpose
