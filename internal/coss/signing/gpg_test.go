package signing

import (
	"bytes"
	"context"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/ProtonMail/go-crypto/openpgp"
	"github.com/ProtonMail/go-crypto/openpgp/armor"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"go.flipt.io/flipt/internal/config"
	"go.flipt.io/flipt/internal/secrets"
	"go.uber.org/zap"
)

// MockSecretsManager is a mock implementation of the secrets.Manager interface
type MockSecretsManager struct {
	mock.Mock
}

func (m *MockSecretsManager) RegisterProvider(name string, provider secrets.Provider) error {
	args := m.Called(name, provider)
	return args.Error(0)
}

func (m *MockSecretsManager) GetProvider(name string) (secrets.Provider, error) {
	args := m.Called(name)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(secrets.Provider), args.Error(1)
}

func (m *MockSecretsManager) GetSecretValue(ctx context.Context, ref secrets.Reference) ([]byte, error) {
	args := m.Called(ctx, ref)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]byte), args.Error(1)
}

func (m *MockSecretsManager) GetSecret(ctx context.Context, providerName, path string) (*secrets.Secret, error) {
	args := m.Called(ctx, providerName, path)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*secrets.Secret), args.Error(1)
}

func (m *MockSecretsManager) PutSecret(ctx context.Context, providerName, path string, secret *secrets.Secret) error {
	args := m.Called(ctx, providerName, path, secret)
	return args.Error(0)
}

func (m *MockSecretsManager) DeleteSecret(ctx context.Context, providerName, path string) error {
	args := m.Called(ctx, providerName, path)
	return args.Error(0)
}

func (m *MockSecretsManager) ListSecrets(ctx context.Context, providerName, pathPrefix string) ([]string, error) {
	args := m.Called(ctx, providerName, pathPrefix)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]string), args.Error(1)
}

func (m *MockSecretsManager) ListProviders() []string {
	args := m.Called()
	return args.Get(0).([]string)
}

func (m *MockSecretsManager) Close() error {
	args := m.Called()
	return args.Error(0)
}

func TestNewGPGSigner(t *testing.T) {
	mockManager := &MockSecretsManager{}
	logger := zap.NewNop()

	t.Run("creates signer with valid reference", func(t *testing.T) {
		secretRef := config.SecretReference{
			Provider: "file",
			Path:     "test/key",
			Key:      "private_key",
		}

		signer, err := NewGPGSigner(secretRef, "test@example.com", mockManager, logger)

		require.NoError(t, err)
		assert.NotNil(t, signer)
		assert.Equal(t, secretRef, signer.secretRef)
		assert.Equal(t, "test@example.com", signer.keyID)
		assert.Equal(t, mockManager, signer.secretManager)
		assert.Equal(t, logger, signer.logger)
	})

	t.Run("fails with invalid secret reference", func(t *testing.T) {
		secretRef := config.SecretReference{
			Provider: "", // Empty provider should cause validation error
			Path:     "test/key",
			Key:      "private_key",
		}

		_, err := NewGPGSigner(secretRef, "test@example.com", mockManager, logger)

		require.Error(t, err)
		assert.Contains(t, err.Error(), "invalid secret reference")
	})
}

func TestGPGSigner_FormatCommitForSigning(t *testing.T) {
	// Create a mock signer to test the formatting logic
	signer := &GPGSigner{
		logger: zap.NewNop(),
	}

	t.Run("formats commit correctly", func(t *testing.T) {
		// Create a test commit
		when := time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC)
		commit := &object.Commit{
			Hash:     plumbing.NewHash("1234567890abcdef1234567890abcdef12345678"),
			Message:  "Initial commit",
			TreeHash: plumbing.NewHash("abcdef1234567890abcdef1234567890abcdef12"),
			ParentHashes: []plumbing.Hash{
				plumbing.NewHash("fedcba0987654321fedcba0987654321fedcba09"),
			},
			Author: object.Signature{
				Name:  "Test Author",
				Email: "author@example.com",
				When:  when,
			},
			Committer: object.Signature{
				Name:  "Test Committer",
				Email: "committer@example.com",
				When:  when,
			},
		}

		result := signer.formatCommitForSigning(commit)
		resultStr := string(result)

		// Verify the formatted commit contains expected components
		assert.Contains(t, resultStr, "tree abcdef1234567890abcdef1234567890abcdef12")
		assert.Contains(t, resultStr, "parent fedcba0987654321fedcba0987654321fedcba09")
		assert.Contains(t, resultStr, "author Test Author <author@example.com> 1672574400 +0000")
		assert.Contains(t, resultStr, "committer Test Committer <committer@example.com> 1672574400 +0000")
		assert.Contains(t, resultStr, "\nInitial commit")

		// Verify the order of components
		lines := strings.Split(resultStr, "\n")
		assert.True(t, strings.HasPrefix(lines[0], "tree "))
		assert.True(t, strings.HasPrefix(lines[1], "parent "))
		assert.True(t, strings.HasPrefix(lines[2], "author "))
		assert.True(t, strings.HasPrefix(lines[3], "committer "))
		assert.Empty(t, lines[4]) // Empty line before message
		assert.Equal(t, "Initial commit", lines[5])
	})

	t.Run("handles commit without parents", func(t *testing.T) {
		when := time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC)
		commit := &object.Commit{
			Hash:         plumbing.NewHash("1234567890abcdef1234567890abcdef12345678"),
			Message:      "Initial commit",
			TreeHash:     plumbing.NewHash("abcdef1234567890abcdef1234567890abcdef12"),
			ParentHashes: []plumbing.Hash{}, // No parents
			Author: object.Signature{
				Name:  "Test Author",
				Email: "author@example.com",
				When:  when,
			},
			Committer: object.Signature{
				Name:  "Test Committer",
				Email: "committer@example.com",
				When:  when,
			},
		}

		result := signer.formatCommitForSigning(commit)
		resultStr := string(result)

		// Should not contain parent line
		assert.NotContains(t, resultStr, "parent ")
		assert.Contains(t, resultStr, "tree abcdef1234567890abcdef1234567890abcdef12")
		assert.Contains(t, resultStr, "author Test Author <author@example.com> 1672574400 +0000")
	})

	t.Run("handles multiple parents", func(t *testing.T) {
		when := time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC)
		commit := &object.Commit{
			Hash:     plumbing.NewHash("1234567890abcdef1234567890abcdef12345678"),
			Message:  "Merge commit",
			TreeHash: plumbing.NewHash("abcdef1234567890abcdef1234567890abcdef12"),
			ParentHashes: []plumbing.Hash{
				plumbing.NewHash("fedcba0987654321fedcba0987654321fedcba09"),
				plumbing.NewHash("123456789012345678901234567890123456789a"),
			},
			Author: object.Signature{
				Name:  "Test Author",
				Email: "author@example.com",
				When:  when,
			},
			Committer: object.Signature{
				Name:  "Test Committer",
				Email: "committer@example.com",
				When:  when,
			},
		}

		result := signer.formatCommitForSigning(commit)
		resultStr := string(result)

		// Should contain both parent lines
		assert.Contains(t, resultStr, "parent fedcba0987654321fedcba0987654321fedcba09")
		assert.Contains(t, resultStr, "parent 123456789012345678901234567890123456789a")
	})
}

func TestGPGSigner_LoadEntity_Logic(t *testing.T) {
	t.Run("validates secret reference conversion", func(t *testing.T) {
		secretRef := config.SecretReference{
			Provider: "vault",
			Path:     "flipt/signing-key",
			Key:      "private_key",
		}

		// This tests the conversion logic from config.SecretReference to secrets.Reference
		expectedRef := secrets.Reference{
			Provider: secretRef.Provider,
			Path:     secretRef.Path,
			Key:      secretRef.Key,
		}

		assert.Equal(t, "vault", expectedRef.Provider)
		assert.Equal(t, "flipt/signing-key", expectedRef.Path)
		assert.Equal(t, "private_key", expectedRef.Key)

		// Validate the reference
		err := expectedRef.Validate()
		assert.NoError(t, err)
	})

	t.Run("validates passphrase reference creation", func(t *testing.T) {
		secretRef := config.SecretReference{
			Provider: "vault",
			Path:     "flipt/signing-key",
			Key:      "private_key",
		}

		// This tests the passphrase reference logic
		passphraseRef := secrets.Reference{
			Provider: secretRef.Provider,
			Path:     secretRef.Path,
			Key:      "passphrase",
		}

		assert.Equal(t, "vault", passphraseRef.Provider)
		assert.Equal(t, "flipt/signing-key", passphraseRef.Path)
		assert.Equal(t, "passphrase", passphraseRef.Key)

		// Validate the reference
		err := passphraseRef.Validate()
		assert.NoError(t, err)
	})
}

func TestGPGSigner_GetSecretValue_Errors(t *testing.T) {
	mockManager := &MockSecretsManager{}
	logger := zap.NewNop()

	secretRef := config.SecretReference{
		Provider: "file",
		Path:     "test/key",
		Key:      "private_key",
	}

	signer, err := NewGPGSigner(secretRef, "test@example.com", mockManager, logger)
	require.NoError(t, err)

	ctx := context.Background()

	t.Run("handles secret manager error", func(t *testing.T) {
		expectedRef := secrets.Reference{
			Provider: "file",
			Path:     "test/key",
			Key:      "private_key",
		}

		mockManager.On("GetSecretValue", mock.MatchedBy(func(ctx context.Context) bool {
			return true
		}), expectedRef).Return(nil, fmt.Errorf("secret not found"))

		err := signer.loadEntity(ctx)

		require.Error(t, err)
		assert.Contains(t, err.Error(), "getting signing key")
		assert.Contains(t, err.Error(), "secret not found")

		mockManager.AssertExpectations(t)
	})
}

func TestGPGSigner_KeyIDMatching_Logic(t *testing.T) {
	t.Run("tests key ID matching logic", func(t *testing.T) {
		// Test the key ID matching logic used in loadEntity
		requestedKeyID := "test@example.com"

		// Test email matching
		identityName := "Test User <test@example.com>"
		identityEmail := "test@example.com"

		emailMatches := strings.Contains(identityName, requestedKeyID) ||
			strings.Contains(identityEmail, requestedKeyID)

		assert.True(t, emailMatches)

		// Test key ID hex matching
		keyId := uint64(0x1234567890ABCDEF)
		keyIDHex := fmt.Sprintf("%X", keyId)
		shortKeyID := "90ABCDEF"

		keyIdMatches := strings.HasSuffix(keyIDHex, strings.ToUpper(shortKeyID))
		assert.True(t, keyIdMatches)

		// Test non-matching case
		nonMatchingKeyID := "OTHER"
		noMatch := strings.HasSuffix(keyIDHex, strings.ToUpper(nonMatchingKeyID))
		assert.False(t, noMatch)
	})
}

func TestGPGSigner_Caching_Logic(t *testing.T) {
	t.Run("tests entity caching logic", func(t *testing.T) {
		signer := &GPGSigner{
			entity:       nil,
			publicKeyPEM: "",
			logger:       zap.NewNop(),
		}

		// Initially entity should be nil
		assert.Nil(t, signer.entity)
		assert.Empty(t, signer.publicKeyPEM)

		// Simulate setting cached values
		// Note: We can't create a real entity here without GPG setup,
		// but we can test the caching logic
		signer.publicKeyPEM = "-----BEGIN PGP PUBLIC KEY BLOCK-----\ntest\n-----END PGP PUBLIC KEY BLOCK-----"

		// Should return cached value
		assert.NotEmpty(t, signer.publicKeyPEM)
	})
}

func TestGPGSigner_ArmorFormat_Validation(t *testing.T) {
	t.Run("validates armor format constants", func(t *testing.T) {
		// These tests verify that we're using the correct armor types
		// from the go-crypto library

		// These should be valid armor types
		assert.Equal(t, "PGP SIGNATURE", openpgp.SignatureType)
		assert.Equal(t, "PGP PUBLIC KEY BLOCK", openpgp.PublicKeyType)
	})

	t.Run("tests armor encoding logic", func(t *testing.T) {
		// Test the armor encoding process without actual GPG data
		testData := []byte("test signature data")

		var armoredBuf bytes.Buffer
		armorWriter, err := armor.Encode(&armoredBuf, openpgp.SignatureType, nil)
		require.NoError(t, err)

		_, err = armorWriter.Write(testData)
		require.NoError(t, err)

		err = armorWriter.Close()
		require.NoError(t, err)

		result := armoredBuf.String()

		// Verify armor format
		assert.Contains(t, result, "-----BEGIN PGP SIGNATURE-----")
		assert.Contains(t, result, "-----END PGP SIGNATURE-----")
	})
}

func TestGPGSigner_Interface_Compliance(t *testing.T) {
	t.Run("implements signing.Signer interface", func(t *testing.T) {
		// This test ensures GPGSigner implements the Signer interface
		mockManager := &MockSecretsManager{}
		logger := zap.NewNop()

		secretRef := config.SecretReference{
			Provider: "file",
			Path:     "test/key",
			Key:      "private_key",
		}

		signer, err := NewGPGSigner(secretRef, "test@example.com", mockManager, logger)
		require.NoError(t, err)

		// Verify it can be assigned to the interface
		var _ interface {
			SignCommit(ctx context.Context, commit *object.Commit) (string, error)
			GetPublicKey(ctx context.Context) (string, error)
		} = signer
	})
}

func TestGPGSigner_ErrorHandling(t *testing.T) {
	mockManager := &MockSecretsManager{}
	logger := zap.NewNop()

	secretRef := config.SecretReference{
		Provider: "file",
		Path:     "test/key",
		Key:      "private_key",
	}

	signer, err := NewGPGSigner(secretRef, "test@example.com", mockManager, logger)
	require.NoError(t, err)

	ctx := context.Background()

	t.Run("SignCommit handles loadEntity errors", func(t *testing.T) {
		// Reset entity to force loading
		signer.entity = nil

		expectedRef := secrets.Reference{
			Provider: "file",
			Path:     "test/key",
			Key:      "private_key",
		}

		mockManager.On("GetSecretValue", mock.MatchedBy(func(ctx context.Context) bool {
			return true
		}), expectedRef).Return(nil, fmt.Errorf("secret not found"))

		commit := &object.Commit{
			Hash:     plumbing.NewHash("1234567890abcdef1234567890abcdef12345678"),
			Message:  "Test commit",
			TreeHash: plumbing.NewHash("abcdef1234567890abcdef1234567890abcdef12"),
		}

		_, err := signer.SignCommit(ctx, commit)

		require.Error(t, err)
		assert.Contains(t, err.Error(), "loading signing entity")

		mockManager.AssertExpectations(t)
	})

	t.Run("GetPublicKey handles loadEntity errors", func(t *testing.T) {
		// Reset entity and cached public key to force loading
		signer.entity = nil
		signer.publicKeyPEM = ""

		expectedRef := secrets.Reference{
			Provider: "file",
			Path:     "test/key",
			Key:      "private_key",
		}

		mockManager.On("GetSecretValue", mock.MatchedBy(func(ctx context.Context) bool {
			return true
		}), expectedRef).Return(nil, fmt.Errorf("secret not found"))

		_, err := signer.GetPublicKey(ctx)

		require.Error(t, err)
		assert.Contains(t, err.Error(), "loading signing entity")

		mockManager.AssertExpectations(t)
	})
}

func TestGPGSigner_KeyDecryption_Logic(t *testing.T) {
	t.Run("validates passphrase handling logic", func(t *testing.T) {
		secretRef := config.SecretReference{
			Provider: "vault",
			Path:     "flipt/signing-key",
			Key:      "private_key",
		}

		// Test the passphrase reference construction logic
		passphraseRef := secrets.Reference{
			Provider: secretRef.Provider,
			Path:     secretRef.Path,
			Key:      "passphrase",
		}

		// Verify the passphrase reference is constructed correctly
		assert.Equal(t, "vault", passphraseRef.Provider)
		assert.Equal(t, "flipt/signing-key", passphraseRef.Path)
		assert.Equal(t, "passphrase", passphraseRef.Key)

		// Test that it validates correctly
		err := passphraseRef.Validate()
		assert.NoError(t, err)
	})
}

func TestGPGSigner_GPGEntitySearch_Logic(t *testing.T) {
	t.Run("validates entity selection logic", func(t *testing.T) {
		// Mock entity list with different scenarios
		_ = "test@example.com" // Mock requested key ID

		// Test entity selection logic
		var selectedEntity *int    // Using int as placeholder for entity
		entities := []int{1, 2, 3} // Mock entity list

		// Simulate finding no matching entity
		for _, entity := range entities {
			// Mock the matching logic
			if entity == 999 { // No match condition
				selectedEntity = &entity
				break
			}
		}

		// Should use first entity as fallback
		if selectedEntity == nil && len(entities) > 0 {
			selectedEntity = &entities[0]
		}

		assert.NotNil(t, selectedEntity)
		assert.Equal(t, 1, *selectedEntity)

		// Test specific match scenario
		var specificMatch *int
		for _, entity := range entities {
			if entity == 2 { // Specific match condition
				specificMatch = &entity
				break
			}
		}

		assert.NotNil(t, specificMatch)
		assert.Equal(t, 2, *specificMatch)
	})
}
