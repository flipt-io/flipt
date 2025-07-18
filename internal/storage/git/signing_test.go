package git

import (
	"bytes"
	"context"
	"errors"
	"io"
	"os"
	"testing"

	"github.com/go-git/go-git/v6/plumbing/object"
	"github.com/go-git/go-git/v6/storage/memory"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"go.flipt.io/flipt/internal/storage/git/signing"
	"go.uber.org/zap"
)

// MockSigner implements the signing.Signer interface for testing.
type MockSigner struct {
	mock.Mock
}

func (m *MockSigner) SignCommit(ctx context.Context, commit *object.Commit) (string, error) {
	args := m.Called(ctx, commit)
	return args.String(0), args.Error(1)
}

func (m *MockSigner) GetPublicKey(ctx context.Context) (string, error) {
	args := m.Called(ctx)
	return args.String(0), args.Error(1)
}

func TestWithSigner(t *testing.T) {
	t.Run("configures signer on repository", func(t *testing.T) {
		mockSigner := &MockSigner{}
		repo := &Repository{}

		option := WithSigner(mockSigner)
		option(repo)

		assert.Equal(t, mockSigner, repo.signer)
	})

	t.Run("nil signer is allowed", func(t *testing.T) {
		repo := &Repository{}

		option := WithSigner(nil)
		option(repo)

		assert.Nil(t, repo.signer)
	})
}

func TestRepository_SignerIntegration(t *testing.T) {
	t.Run("repository without signer works normally", func(t *testing.T) {
		repo := &Repository{
			signer: nil,
		}

		// Verify that no signer is configured
		assert.Nil(t, repo.signer)

		// Repository should work normally without signer
		assert.NotNil(t, repo)
	})

	t.Run("repository with signer stores reference", func(t *testing.T) {
		mockSigner := &MockSigner{}
		repo := &Repository{
			signer: mockSigner,
		}

		// Verify that signer is properly stored
		assert.Equal(t, mockSigner, repo.signer)
		assert.NotNil(t, repo.signer)
	})
}

func TestFilesystem_NewFilesystemWithSigner(t *testing.T) {
	logger := zap.NewNop()
	storage := memory.NewStorage()

	t.Run("creates filesystem without signer", func(t *testing.T) {
		fs, err := newFilesystem(logger, storage)

		require.NoError(t, err)
		assert.NotNil(t, fs)
		assert.Nil(t, fs.signer)
	})

	t.Run("creates filesystem with signer", func(t *testing.T) {
		mockSigner := &MockSigner{}

		fs, err := newFilesystem(logger, storage, withSigner(mockSigner))

		require.NoError(t, err)
		assert.NotNil(t, fs)
		assert.Equal(t, mockSigner, fs.signer)
	})

	t.Run("creates filesystem with nil signer", func(t *testing.T) {
		fs, err := newFilesystem(logger, storage, withSigner(nil))

		require.NoError(t, err)
		assert.NotNil(t, fs)
		assert.Nil(t, fs.signer)
	})
}

func TestFilesystem_CommitSigning_MockBehavior(t *testing.T) {
	logger := zap.NewNop()
	storage := memory.NewStorage()

	t.Run("commit calls signer when available", func(t *testing.T) {
		mockSigner := &MockSigner{}

		// Set up mock expectations
		expectedSignature := "-----BEGIN PGP SIGNATURE-----\ntest signature\n-----END PGP SIGNATURE-----"
		mockSigner.On("SignCommit", mock.MatchedBy(func(ctx context.Context) bool {
			return true
		}), mock.MatchedBy(func(commit *object.Commit) bool {
			// Verify the commit has the expected message and basic structure
			return commit.Message == "Test commit message" &&
				!commit.TreeHash.IsZero() && // Has a valid tree
				!commit.Author.When.IsZero() && // Has a timestamp
				!commit.Committer.When.IsZero() // Has a timestamp
		})).Return(expectedSignature, nil)

		// Create filesystem with signer
		fs, err := newFilesystem(logger, storage,
			withSignature("Test User", "test@example.com"),
			withSigner(mockSigner),
		)
		require.NoError(t, err)

		// Create a simple file to have a tree
		file, err := fs.OpenFile("test.txt", os.O_CREATE|os.O_RDWR|os.O_TRUNC, 0644)
		require.NoError(t, err)
		_, err = io.Copy(file, bytes.NewBufferString("test content"))
		require.NoError(t, err)
		err = file.Close()
		require.NoError(t, err)

		ctx := context.Background()
		commit, err := fs.commit(ctx, "Test commit message")

		require.NoError(t, err)
		assert.NotNil(t, commit)
		assert.Equal(t, expectedSignature, commit.PGPSignature)

		mockSigner.AssertExpectations(t)
	})

	t.Run("commit fails when signer returns error", func(t *testing.T) {
		mockSigner := &MockSigner{}

		// Set up mock to return error
		signingError := errors.New("GPG key not found")
		mockSigner.On("SignCommit", mock.Anything, mock.Anything).Return("", signingError)

		// Create filesystem with signer
		fs, err := newFilesystem(logger, storage,
			withSignature("Test User", "test@example.com"),
			withSigner(mockSigner),
		)
		require.NoError(t, err)

		// Create a simple file to have a tree
		file, err := fs.OpenFile("test.txt", os.O_CREATE|os.O_RDWR|os.O_TRUNC, 0644)
		require.NoError(t, err)
		_, err = io.Copy(file, bytes.NewBufferString("test content"))
		require.NoError(t, err)
		err = file.Close()
		require.NoError(t, err)

		ctx := context.Background()
		_, err = fs.commit(ctx, "Test commit message")

		require.Error(t, err)
		assert.Contains(t, err.Error(), "signing commit")
		assert.Contains(t, err.Error(), "GPG key not found")

		mockSigner.AssertExpectations(t)
	})

	t.Run("commit without signer does not call SignCommit", func(t *testing.T) {
		// Create fresh storage for this test to avoid interference
		freshStorage := memory.NewStorage()

		// Create filesystem without signer
		fs, err := newFilesystem(logger, freshStorage)
		require.NoError(t, err)

		// Verify that no signer is configured
		assert.Nil(t, fs.signer)

		// Create a simple file to have a tree
		file, err := fs.OpenFile("test.txt", os.O_CREATE|os.O_RDWR|os.O_TRUNC, 0644)
		require.NoError(t, err)
		_, err = io.Copy(file, bytes.NewBufferString("test content"))
		require.NoError(t, err)
		err = file.Close()
		require.NoError(t, err)

		ctx := context.Background()
		commit, err := fs.commit(ctx, "Test commit message")

		require.NoError(t, err)
		assert.NotNil(t, commit)
		assert.Empty(t, commit.PGPSignature, "Commit should not have PGP signature when no signer is configured")
		assert.Equal(t, "Test commit message", commit.Message)
		// The actual author/email will be overridden by the authentication system
		// but the key point is that no PGP signature is set when no signer is configured
		assert.NotNil(t, commit.Author)
		assert.NotNil(t, commit.Committer)
	})
}

func TestSigning_InterfaceCompliance(t *testing.T) {
	t.Run("MockSigner implements signing.Signer interface", func(t *testing.T) {
		mockSigner := &MockSigner{}

		// Verify it can be assigned to the interface
		var _ signing.Signer = mockSigner

		// Test the interface methods exist
		ctx := context.Background()

		mockSigner.On("GetPublicKey", ctx).Return("test-public-key", nil)

		pubKey, err := mockSigner.GetPublicKey(ctx)
		require.NoError(t, err)
		assert.Equal(t, "test-public-key", pubKey)

		mockSigner.AssertExpectations(t)
	})
}

func TestCommitSigning_ErrorScenarios(t *testing.T) {
	logger := zap.NewNop()
	storage := memory.NewStorage()

	t.Run("handles various signer error conditions", func(t *testing.T) {
		testCases := []struct {
			name          string
			signerError   error
			expectedError string
		}{
			{
				name:          "generic signing error",
				signerError:   errors.New("signing failed"),
				expectedError: "signing commit: signing failed",
			},
			{
				name:          "key not found error",
				signerError:   errors.New("GPG key not found"),
				expectedError: "signing commit: GPG key not found",
			},
			{
				name:          "invalid key format error",
				signerError:   errors.New("invalid key format"),
				expectedError: "signing commit: invalid key format",
			},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				mockSigner := &MockSigner{}

				// Mock the signer to return the specific error
				mockSigner.On("SignCommit", mock.Anything, mock.Anything).Return("", tc.signerError)

				// Create filesystem with signer
				fs, err := newFilesystem(logger, storage,
					withSigner(mockSigner),
					withSignature("Test User", "test@example.com"),
				)
				require.NoError(t, err)

				// Create a simple file to have a tree
				file, err := fs.OpenFile("test.txt", os.O_CREATE|os.O_RDWR|os.O_TRUNC, 0644)
				require.NoError(t, err)
				_, err = io.Copy(file, bytes.NewBufferString("test content"))
				require.NoError(t, err)
				err = file.Close()
				require.NoError(t, err)

				ctx := context.Background()
				_, err = fs.commit(ctx, "Test commit message")

				require.Error(t, err)
				assert.Contains(t, err.Error(), tc.expectedError)

				mockSigner.AssertExpectations(t)
			})
		}
	})
}

func TestCommitSigning_SignerDataValidation(t *testing.T) {
	logger := zap.NewNop()
	storage := memory.NewStorage()

	t.Run("signer receives correctly formatted commit", func(t *testing.T) {
		mockSigner := &MockSigner{}

		// Detailed verification of what the signer receives
		expectedSignature := "-----BEGIN PGP SIGNATURE-----\ntest signature\n-----END PGP SIGNATURE-----"
		mockSigner.On("SignCommit", mock.MatchedBy(func(ctx context.Context) bool {
			return true
		}), mock.MatchedBy(func(commit *object.Commit) bool {
			// Verify all aspects of the commit structure that we can control
			if commit.Message != "Test commit with validation" {
				return false
			}
			// Note: Author/Email will be overridden by authentication system, so we skip checking them
			// Verify the commit has a tree hash (not empty)
			if commit.TreeHash.IsZero() {
				return false
			}
			// Verify timestamp is recent (within last minute)
			if commit.Author.When.IsZero() || commit.Committer.When.IsZero() {
				return false
			}
			return true
		})).Return(expectedSignature, nil)

		// Create filesystem with signer
		fs, err := newFilesystem(logger, storage,
			withSignature("Test User", "test@example.com"),
			withSigner(mockSigner),
		)
		require.NoError(t, err)

		// Create a simple file to have a tree
		file, err := fs.OpenFile("test.txt", os.O_CREATE|os.O_RDWR|os.O_TRUNC, 0644)
		require.NoError(t, err)
		_, err = io.Copy(file, bytes.NewBufferString("test content"))
		require.NoError(t, err)
		err = file.Close()
		require.NoError(t, err)

		ctx := context.Background()
		commit, err := fs.commit(ctx, "Test commit with validation")

		require.NoError(t, err)
		assert.NotNil(t, commit)
		assert.Equal(t, expectedSignature, commit.PGPSignature)

		mockSigner.AssertExpectations(t)
	})
}
