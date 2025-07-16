# Commit Signing Integration Tests

This directory contains integration tests for Flipt's commit signing functionality using HashiCorp Vault as the secrets provider.

## Overview

These tests verify that:
1. Flipt can successfully connect to Vault and retrieve GPG private keys
2. Git commits are properly signed when commit signing is enabled
3. The signing configuration works end-to-end in a containerized environment

## Test Architecture

### Components

- **Vault Container**: HashiCorp Vault running in dev mode with test secrets
- **Flipt Container**: Flipt server configured with local storage and commit signing enabled
- **Test GPG Keys**: Pre-generated test key pairs for signing operations

### Test Flow

```
1. Start Vault container with dev token
2. Store test GPG private key in Vault KV store
3. Start Flipt with Vault secrets provider and local storage configured
4. Perform flag operations that trigger commits
5. Verify operations succeed without signing errors
6. Execute git commands inside container to verify commit signatures
7. Verify signatures are from expected GPG key and contain PGP blocks
```

## Running the Tests

### Prerequisites

- Go 1.24+
- Docker or compatible container runtime
- Dagger CLI (for local development)

### Run All Integration Tests

```bash
# From project root
dagger call test integration
```

### Run Only Signing Tests

```bash
# From project root  
dagger call test integration --cases signing
```

## Test Configuration

### Storage Configuration

The tests use **local storage** for easier signature verification:

- **Backend Type**: `local`
- **Repository Path**: `/tmp/flipt-repo` (inside container)
- **Branch**: `main`

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

The tests dynamically generate a GPG key pair during setup:

- **Name**: Flipt Test Bot
- **Email**: test-bot@flipt.io
- **Key Type**: RSA 2048-bit
- **Purpose**: Signing only (test key)
- **Generation**: Created fresh for each test run using GPG batch mode

The private key is generated inside the test container and stored in Vault for Flipt to use for commit signing.

## File Structure

```
signing/
├── README.md                    # This file
└── signing_test.go             # Main integration test
```

## Test Cases

### TestCommitSigning

**Purpose**: Verify that commits are properly signed when signing is enabled

**Steps**:
1. Setup Vault container as a background service
2. Generate test GPG key pair inside container using GPG batch mode
3. Store GPG private key in Vault KV store
4. Configure Flipt with local storage and commit signing
5. Create test feature flags to trigger Git commits
6. Verify flag operations complete without signing-related errors
7. Execute git commands inside container to verify commit signatures
8. Verify signatures are from the expected GPG key (test-bot@flipt.io)
9. Check that commits contain PGP signature blocks

**Verification**: 
- Server starts successfully with signing configuration
- Flag operations complete without errors
- Commits are verified to contain valid GPG signatures
- Signatures are verified to be from the expected test key
- Repository is verified inside container using git commands

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

4. **Signature Verification Failed**
   - Verify git commands can access the repository in the container
   - Check that commits contain expected signature format
   - Ensure test GPG key is properly configured in Vault

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
- Signature verification results from git commands

## Security Notes

⚠️ **Test Keys Only**: The GPG keys used in these tests are for testing purposes only and should never be used in production environments.

The test keys are:
- Generated specifically for testing
- Not associated with real identities
- Safe to include in version control
- Should be rotated periodically for security best practices

## Implementation Details

The integration test uses Dagger containers to orchestrate the test environment:

- **Vault Setup**: The `withVault` function creates a Vault container and stores the test GPG key
- **Flipt Configuration**: Environment variables configure Flipt to use Vault for secrets and local storage for commits
- **Test Execution**: The `signingTestSuite` function runs the test and then verifies signatures using git commands inside the Flipt container
- **Verification**: Git commands (`git log --show-signature` and `git cat-file commit HEAD`) verify that commits are properly signed

The test validates both that Flipt can successfully sign commits and that the signatures are cryptographically valid.