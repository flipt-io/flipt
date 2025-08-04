# CLAUDE.md - Build & Integration Testing

This file provides guidance to Claude Code for working with Flipt's build system and integration testing infrastructure.

## Build System Overview

Flipt uses a multi-layered build system combining:

- **Mage**: Build automation tool for Go projects (like Make but in Go)
- **Dagger**: Container-based CI/CD engine for reproducible builds
- **npm**: UI build system for React frontend
- **GitHub Actions**: CI/CD workflows

## Directory Structure

```
build/
├── README.md                 # Dagger build documentation
├── main.go                   # Dagger module entry point
├── dagger.gen.go            # Dagger generated code
├── internal/                # Build system internals
│   ├── dagger/              # Dagger client code
│   ├── flipt.go            # Main Flipt build functions
│   ├── ui.go               # UI build functions
│   └── cmd/                # Build utilities
│       ├── discover/       # OIDC test server for k8s auth
│       └── gitea/          # Gitea setup utilities
└── testing/                # Integration test framework
    ├── integration.go      # Main integration test orchestration
    ├── helpers.go         # Test helper functions
    ├── test.go            # Test container setup
    └── integration/       # Individual test suites
        ├── authn/         # Authentication tests
        ├── authz/         # Authorization tests
        ├── environments/  # Environment API tests
        ├── ofrep/         # OpenFeature tests
        ├── signing/       # Git signing tests
        └── snapshot/      # Snapshot API tests
```

## Build Commands

### Dagger Commands (Container-based)

#### Build Commands

```bash
dagger call build                    # Build production container
dagger call test --source=. unit    # Run unit tests
dagger call test --source=. ui      # Run UI tests

# Registry cache optimization (for CI)
dagger call build-with-cache --source . --registry-cache ghcr.io/owner/repo-cache --cache-tag cache-key
dagger call test-with-cache --source . --cached-image ghcr.io/owner/repo-cache:cache-key integration --cases="authn"
dagger call check-cache-exists --image-ref ghcr.io/owner/repo-cache:cache-key
```

#### Integration Testing

```bash
# Run all integration tests
dagger call test --source=. integration

# Run specific test cases
dagger call test --source=. integration --cases="authn"
dagger call test --source=. integration --cases="authn authz"

# Run with coverage output
dagger call test-coverage --source=. integration-coverage --cases="snapshot"
```

## Integration Testing Framework

### Test Architecture

The integration testing system uses a sophisticated container orchestration approach:

1. **Base Container**: Contains Go toolchain and test binaries
2. **Flipt Container**: Runs Flipt server with various configurations
3. **Service Containers**: External dependencies (Gitea, Vault, OIDC providers)
4. **Test Execution**: Runs in isolated containers with service bindings

### Available Test Cases

Test cases are defined in `build/testing/integration.go:AllCases`:

### Test Infrastructure

#### Service Dependencies

- **Gitea**: Git repository hosting for storage backends
- **Vault**: HashiCorp Vault for secret management
- **OIDC Provider**: Custom OIDC server for Kubernetes auth testing

#### Test Data

- `build/testing/integration/testdata/`: Test configuration files
- `build/testing/integration/*/`: Individual test suite implementations

### Running Integration Tests

#### Local Development

```bash
# Install Dagger CLI
brew install dagger/tap/dagger

# Run all tests (parallel execution)
dagger call test --source=. integration

# Run specific authentication tests
dagger call test --source=. integration --cases="authn/token authn/jwt"

# Run with coverage collection
dagger call test-coverage --source=. integration-coverage --cases="snapshot" export --path=/tmp/coverage.out
```

#### CI/CD Pipeline

GitHub Actions workflow (`.github/workflows/integration-test.yml`):

1. **Test Discovery**: Extract test cases from `integration.go`
2. **Cache Build**: Build Flipt once and push to GHCR for sharing across jobs
3. **Matrix Execution**: Run each test case in parallel using cached images
4. **Coverage Collection**: Aggregate coverage from all test runs
5. **Codecov Upload**: Submit coverage reports

#### Registry Cache Optimization

The CI pipeline uses Docker BuildKit registry cache to share container images across GitHub Actions jobs:

**Cache Strategy:**
- **Cache Backend**: GitHub Container Registry (GHCR) - `ghcr.io/owner/repo-cache`
- **Cache Key Format**: `flipt-{git-commit}-{source-hash}`
- **Cache Scope**: Per repository, shared across all PR and push workflows

**Performance Benefits:**
- **Before**: Each integration test job rebuilds Flipt (~5-10 minutes each)
- **After**: One build job (~10 minutes) + fast test jobs using cache (~2-3 minutes each)
- **Total reduction**: ~60% in pipeline execution time
- **Resource savings**: Significant reduction in GitHub Actions minutes

**Cache Workflow:**
```yaml
build-cache:
  - Generate cache key from git commit + source file hashes
  - Check if cached image exists in GHCR
  - Build and push only if cache miss
  - Output image reference for test jobs

integration-tests:
  - Pull cached image from GHCR
  - Run tests using pre-built Flipt container
  - Maintain full test isolation and coverage
```

**Cache Management:**
- **Invalidation**: Automatic via content-based cache keys
- **Cleanup**: Images cleaned up after successful pipeline runs
- **Fallback**: Graceful fallback to fresh builds if cache fails
- **Authentication**: Uses `GITHUB_TOKEN` with `packages: write` permission

### Test Configuration Patterns

#### Authentication Setup

Tests configure Flipt with various auth methods:

```go
// Token authentication
flipt = flipt.
    WithEnvVariable("FLIPT_AUTHENTICATION_REQUIRED", "true").
    WithEnvVariable("FLIPT_AUTHENTICATION_METHODS_TOKEN_ENABLED", "true").
    WithEnvVariable("FLIPT_AUTHENTICATION_METHODS_TOKEN_STORAGE_TOKENS_BOOTSTRAP_CREDENTIAL", "s3cr3t")

// JWT authentication  
flipt = flipt.
    WithEnvVariable("FLIPT_AUTHENTICATION_METHODS_JWT_ENABLED", "true").
    WithEnvVariable("FLIPT_AUTHENTICATION_METHODS_JWT_PUBLIC_KEY_FILE", "/etc/flipt/jwt.pem")
```

#### Storage Configuration

```go
// Git-based storage
flipt = flipt.
    WithEnvVariable("FLIPT_STORAGE_DEFAULT_REMOTE", "http://gitea:3000/root/features.git").
    WithEnvVariable("FLIPT_STORAGE_DEFAULT_BRANCH", "main").
    WithEnvVariable("FLIPT_CREDENTIALS_DEFAULT_TYPE", "basic")
```

#### Secrets Management

```go
// File-based secrets
flipt = flipt.
    WithEnvVariable("FLIPT_SECRETS_PROVIDERS_FILE_ENABLED", "true").
    WithEnvVariable("FLIPT_SECRETS_PROVIDERS_FILE_BASE_PATH", "/home/flipt/secrets")

// Vault secrets
flipt = flipt.
    WithEnvVariable("FLIPT_SECRETS_PROVIDERS_VAULT_ENABLED", "true").
    WithEnvVariable("FLIPT_SECRETS_PROVIDERS_VAULT_ADDRESS", "http://vault:8200")
```

## Coverage Collection

### Go Coverage

- Unit tests: Standard Go coverage with `go test -coverprofile`
- Integration tests: Uses `GOCOVERDIR` environment variable for runtime coverage
- Combined reporting via Codecov

### UI Coverage

- Jest for unit tests
- Playwright for E2E tests
- Coverage reports integrated with Codecov

## Development Workflow

### Pre-commit Checks

Always run before committing:

```bash
mage go:fmt        # Format Go code
mage go:lint       # Lint Go code  
mage go:modernize  # Update to modern Go style
mage ui:fmt        # Format UI code
mage ui:lint       # Lint UI code
```

### Adding New Integration Tests

1. **Add test case** to `AllCases` map in `build/testing/integration.go`
2. **Create test function** following existing patterns
3. **Add test suite** in `build/testing/integration/{testname}/`
4. **Update CI discovery** (automatic via script)

### Debugging Integration Tests

#### Local Container Inspection

```bash
# Run with debug logging
dagger call test --source=. integration --cases="snapshot" 

# Export coverage for analysis
dagger call test-coverage --source=. integration-coverage --cases="snapshot" export --path=/tmp/debug.out
```

#### Service Debugging

- Gitea UI: Available during test execution at `http://gitea:3000`
- Vault UI: Available at `http://vault:8200`
- OIDC endpoints: Custom discovery at `https://discover.svc`

## Performance Considerations

### Concurrency Control

- Integration tests use semaphore (`sema := make(chan struct{}, 6)`) to limit concurrent test execution
- Maximum 6 concurrent test containers to prevent resource exhaustion

### Container Optimization

- **Base containers**: Cached between test runs using Dagger volumes
- **Registry cache**: Shared Flipt images across CI jobs via GHCR
- **Coverage volumes**: Persisted across test executions
- **Service containers**: Reused within test suites (Gitea, Vault, OIDC)

### CI Resource Management

- Tests timeout after 20 minutes
- Matrix execution distributes load across GitHub runners
- Coverage aggregation minimizes storage overhead

## Troubleshooting

### Common Issues

#### Build Failures

1. **Go version mismatch**: Ensure Go 1.24+
2. **Missing tools**: Run `mage bootstrap`
3. **Dirty modules**: Run `mage clean`

#### Integration Test Failures

1. **Service startup**: Check container logs in Dagger output
2. **Authentication**: Verify token/credential configuration
3. **Git access**: Confirm Gitea repository setup
4. **Coverage issues**: Check `GOCOVERDIR` permissions

#### Performance Issues

1. **Slow tests**: Check semaphore limits and container resources
2. **Timeouts**: Increase test timeout in CI configuration
3. **Memory**: Monitor container memory usage in complex test suites
4. **Cache issues**: Check GHCR authentication and registry cache permissions

#### Registry Cache Issues

1. **Cache miss**: Verify cache key generation and source file hashing
2. **Authentication**: Ensure `GITHUB_TOKEN` has `packages: write` permission
3. **Registry quota**: Monitor GHCR storage usage and cleanup old images
4. **Fallback behavior**: Check that tests work when cache is unavailable

### Debug Commands

```bash
# Verbose test output
dagger call test --source=. integration --cases="authn" --verbose

# Container inspection
dagger call test --source=. base-container terminal

# Coverage debugging  
dagger call test-coverage --source=. integration-coverage --cases="snapshot" export --path=/tmp/debug-coverage.out

# Registry cache debugging
dagger call check-cache-exists --image-ref ghcr.io/owner/repo-cache:cache-key
docker manifest inspect ghcr.io/owner/repo-cache:cache-key
dagger call build-with-cache --source . --registry-cache ghcr.io/owner/repo-cache --cache-tag debug-test
```

This comprehensive build and testing infrastructure ensures Flipt maintains high quality through automated testing across multiple scenarios, authentication methods, and deployment configurations.

