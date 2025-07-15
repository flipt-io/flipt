# Commit Signing Integration Tests

This directory contains integration tests for Flipt's commit signing functionality using HashiCorp Vault as the secrets provider.

## Overview

These tests verify that:
1. Flipt can successfully connect to Vault and retrieve GPG private keys
2. Git commits are properly signed when commit signing is enabled
3. The signing configuration works end-to-end in a containerized environment
4. Error conditions are handled gracefully

## Test Architecture

### Components

- **Vault Container**: HashiCorp Vault running in dev mode with test secrets
- **Flipt Container**: Flipt server configured with commit signing enabled
- **Gitea Container**: Git repository server for storing feature flag configurations
- **Test GPG Keys**: Generated test key pairs for signing operations

### Test Flow

```
1. Start Vault container with dev token
2. Store test GPG private key in Vault KV store
3. Start Flipt with Vault secrets provider configured
4. Start Gitea with test repositories
5. Perform flag operations that trigger commits
6. Verify commits are signed and operations succeed
```

## Running the Tests

### Prerequisites

- Go 1.24+
- Docker or compatible container runtime
- Dagger CLI (for local development)

### Run All Integration Tests

```bash
# From project root
go run ./build/main.go test integration
```

### Run Only Signing Tests

```bash
# From project root  
go run ./build/main.go test integration --cases signing
```

### Run Tests with Verbose Output

```bash
# From project root
go run ./build/main.go test integration --cases signing --verbose
```

## Test Configuration

### Storage Configuration

The tests use **local storage** instead of remote git repositories for easier signature verification:

- **Backend Type**: `local`
- **Repository Path**: `/tmp/flipt-repo` (inside container)
- **Branch**: `main`
- **Mounted Volume**: Dagger cache volume for persistence

### Vault Configuration

The tests use the following Vault setup:

- **Address**: `http://vault:8200`
- **Auth Method**: Token-based authentication
- **Root Token**: `test-root-token`
- **KV Mount**: `secret/` (KV v2 engine)
- **Secret Path**: `secret/flipt/signing-key`

### Flipt Configuration

Environment variables set for commit signing:

```bash
# Vault secrets configuration
FLIPT_SECRETS_PROVIDERS_VAULT_ENABLED=true
FLIPT_SECRETS_PROVIDERS_VAULT_ADDRESS=http://vault:8200
FLIPT_SECRETS_PROVIDERS_VAULT_AUTH_METHOD=token
FLIPT_SECRETS_PROVIDERS_VAULT_TOKEN=test-root-token
FLIPT_SECRETS_PROVIDERS_VAULT_MOUNT=secret

# Local storage configuration
FLIPT_STORAGE_DEFAULT_BACKEND_TYPE=local
FLIPT_STORAGE_DEFAULT_BACKEND_PATH=/tmp/flipt-repo
FLIPT_STORAGE_DEFAULT_BRANCH=main

# Commit signing configuration
FLIPT_STORAGE_DEFAULT_SIGNATURE_ENABLED=true
FLIPT_STORAGE_DEFAULT_SIGNATURE_TYPE=gpg
FLIPT_STORAGE_DEFAULT_SIGNATURE_KEY_REF_PROVIDER=vault
FLIPT_STORAGE_DEFAULT_SIGNATURE_KEY_REF_PATH=flipt/signing-key
FLIPT_STORAGE_DEFAULT_SIGNATURE_KEY_REF_KEY=private_key
FLIPT_STORAGE_DEFAULT_SIGNATURE_NAME=Flipt Test Bot
FLIPT_STORAGE_DEFAULT_SIGNATURE_EMAIL=test-bot@flipt.io
FLIPT_STORAGE_DEFAULT_SIGNATURE_KEY_ID=test-bot@flipt.io
```

## Test Data

### GPG Test Key

The tests use a pre-generated GPG key pair for consistent testing:

- **Name**: Flipt Test Bot
- **Email**: test-bot@flipt.io
- **Key Type**: RSA 2048-bit
- **Purpose**: Signing only (test key)

The private key is stored in Vault and used by Flipt for commit signing.

## File Structure

```
signing/
├── README.md                    # This file
├── signing_test.go             # Main integration tests
├── verification.go             # Git signature verification utilities  
├── helpers.go                  # Test helper functions
├── gpg_test_key.go            # GPG key generation utilities
└── testdata/
    └── test-gpg-key.asc       # Test GPG key data
```

## Test Cases

### TestCommitSigning

**Purpose**: Verify that commits are properly signed when signing is enabled

**Steps**:
1. Setup Vault with test GPG private key
2. Configure Flipt with local storage and commit signing
3. Create a test feature flag to trigger a Git commit
4. Verify the server operates without signing-related errors
5. Check that the signing configuration was loaded successfully

**Current Verification**: 
- Server health check confirms signing configuration loaded
- Operations complete without errors
- No signature verification failures in logs

### TestCommitSigningWithVerification

**Purpose**: Enhanced test with actual Git signature verification

**Status**: Planned implementation using Dagger container exec

**Planned Steps**:
1. All steps from TestCommitSigning
2. Execute git commands inside the container to verify signatures
3. Check for PGP signature blocks in commit objects
4. Validate signature status using `git show --show-signature`

### TestCommitSigningDisabled

**Purpose**: Verify behavior when signing is disabled (placeholder)

**Note**: Currently skipped as it requires a separate Flipt instance configuration

### TestGPGKeyGeneration

**Purpose**: Test the GPG key generation utility functions

**Steps**:
1. Generate a test GPG key pair
2. Verify the key format and structure
3. Ensure keys can be parsed by the OpenPGP library

## Debugging

### Common Issues

1. **Vault Connection Failed**
   - Check container networking
   - Verify Vault is running and accessible
   - Confirm token authentication

2. **GPG Key Not Found**
   - Verify key was stored in Vault correctly
   - Check secret path and key name configuration
   - Ensure Vault KV v2 engine is enabled

3. **Signing Failed**
   - Check GPG key format and validity
   - Verify key permissions and passphrase (if any)
   - Review Flipt logs for signing errors

### Log Analysis

Enable debug logging for detailed information:

```bash
FLIPT_LOG_LEVEL=DEBUG
```

Look for log entries related to:
- Vault connection and authentication
- Secrets manager initialization
- GPG signer creation and key loading
- Git commit operations and signing

## Contributing

When adding new signing integration tests:

1. Follow the existing test patterns and naming conventions
2. Use the provided helper functions for Vault and GPG setup
3. Include proper error handling and cleanup
4. Add documentation for new test scenarios
5. Ensure tests are deterministic and don't rely on external state

## Security Notes

⚠️ **Test Keys Only**: The GPG keys used in these tests are for testing purposes only and should never be used in production environments.

The test keys are:
- Generated specifically for testing
- Not associated with real identities
- Safe to include in version control
- Should be rotated periodically for security best practices

## Performance Considerations

- Tests use lightweight containers and minimal configurations
- GPG operations may add latency to commit operations
- Vault connections are reused where possible
- Test keys use smaller key sizes (2048-bit) for faster generation

## Future Enhancements

Potential improvements for the integration tests:

1. **Multi-Key Scenarios**: Test with multiple GPG keys and key rotation
2. **Different Auth Methods**: Test Vault with Kubernetes and AppRole auth
3. **Error Injection**: Test various failure modes and recovery
4. **Performance Testing**: Measure impact of signing on commit latency
5. **Real Git Verification**: Extract and verify actual GPG signatures from commits
6. **Cross-Platform Testing**: Ensure compatibility across different container platforms