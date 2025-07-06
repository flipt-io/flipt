package signing

import (
	"bytes"
	"context"
	"fmt"
	"strings"

	"github.com/ProtonMail/go-crypto/openpgp"
	"github.com/ProtonMail/go-crypto/openpgp/armor"
	"github.com/go-git/go-git/v5/plumbing/object"
	"go.flipt.io/flipt/internal/config"
	"go.flipt.io/flipt/internal/secrets"
	"go.flipt.io/flipt/internal/storage/git/signing"
	"go.uber.org/zap"
)

// GPGSigner implements signing.Signer using GPG keys from secrets management.
type GPGSigner struct {
	secretRef     config.SecretReference
	keyID         string
	secretManager secrets.Manager
	logger        *zap.Logger

	// Cached entities to avoid repeated secret fetches
	entity       *openpgp.Entity
	publicKeyPEM string
}

// NewGPGSigner creates a new GPG signer.
func NewGPGSigner(secretRef config.SecretReference, keyID string, manager secrets.Manager, logger *zap.Logger) (*GPGSigner, error) {
	if err := secretRef.Validate(); err != nil {
		return nil, fmt.Errorf("invalid secret reference: %w", err)
	}

	return &GPGSigner{
		secretRef:     secretRef,
		keyID:         keyID,
		secretManager: manager,
		logger:        logger,
	}, nil
}

// SignCommit signs a git commit using GPG.
func (g *GPGSigner) SignCommit(ctx context.Context, commit *object.Commit) (string, error) {
	// Load the signing entity if not cached
	if g.entity == nil {
		if err := g.loadEntity(ctx); err != nil {
			return "", fmt.Errorf("loading signing entity: %w", err)
		}
	}

	// Get the commit data in the format Git expects for signing
	commitData := g.formatCommitForSigning(commit)

	// Create signature
	var signatureBuf bytes.Buffer

	// Sign the commit data
	err := openpgp.DetachSign(&signatureBuf, g.entity, bytes.NewReader(commitData), nil)
	if err != nil {
		return "", fmt.Errorf("signing commit: %w", err)
	}

	// Convert to armored format
	var armoredBuf bytes.Buffer
	armorWriter, err := armor.Encode(&armoredBuf, openpgp.SignatureType, nil)
	if err != nil {
		return "", fmt.Errorf("creating armor encoder: %w", err)
	}

	if _, err := armorWriter.Write(signatureBuf.Bytes()); err != nil {
		return "", fmt.Errorf("writing signature to armor: %w", err)
	}

	if err := armorWriter.Close(); err != nil {
		return "", fmt.Errorf("closing armor writer: %w", err)
	}

	signature := armoredBuf.String()

	g.logger.Debug("signed commit",
		zap.String("key_id", g.keyID),
		zap.String("commit_hash", commit.Hash.String()))

	return signature, nil
}

// formatCommitForSigning formats the commit object in the way Git expects for signing.
func (g *GPGSigner) formatCommitForSigning(commit *object.Commit) []byte {
	var buf bytes.Buffer

	// Write tree
	fmt.Fprintf(&buf, "tree %s\n", commit.TreeHash.String())

	// Write parents
	for _, parent := range commit.ParentHashes {
		fmt.Fprintf(&buf, "parent %s\n", parent.String())
	}

	// Write author
	fmt.Fprintf(&buf, "author %s <%s> %d %s\n",
		commit.Author.Name,
		commit.Author.Email,
		commit.Author.When.Unix(),
		commit.Author.When.Format("-0700"))

	// Write committer
	fmt.Fprintf(&buf, "committer %s <%s> %d %s\n",
		commit.Committer.Name,
		commit.Committer.Email,
		commit.Committer.When.Unix(),
		commit.Committer.When.Format("-0700"))

	// Write message
	fmt.Fprintf(&buf, "\n%s", commit.Message)

	return buf.Bytes()
}

// GetPublicKey returns the public key in ASCII-armored format.
func (g *GPGSigner) GetPublicKey(ctx context.Context) (string, error) {
	// Return cached public key if available
	if g.publicKeyPEM != "" {
		return g.publicKeyPEM, nil
	}

	// Load entity if needed
	if g.entity == nil {
		if err := g.loadEntity(ctx); err != nil {
			return "", fmt.Errorf("loading signing entity: %w", err)
		}
	}

	// Export public key
	var pubKeyBuf bytes.Buffer
	armorWriter, err := armor.Encode(&pubKeyBuf, openpgp.PublicKeyType, nil)
	if err != nil {
		return "", fmt.Errorf("creating armor encoder: %w", err)
	}

	if err := g.entity.Serialize(armorWriter); err != nil {
		return "", fmt.Errorf("serializing public key: %w", err)
	}

	if err := armorWriter.Close(); err != nil {
		return "", fmt.Errorf("closing armor writer: %w", err)
	}

	g.publicKeyPEM = pubKeyBuf.String()
	return g.publicKeyPEM, nil
}

// loadEntity loads the GPG entity from secrets management.
func (g *GPGSigner) loadEntity(ctx context.Context) error {
	// Convert config.SecretReference to secrets.Reference
	ref := secrets.Reference{
		Provider: g.secretRef.Provider,
		Path:     g.secretRef.Path,
		Key:      g.secretRef.Key,
	}

	// Get private key from secrets manager
	privateKeyBytes, err := g.secretManager.GetSecretValue(ctx, ref)
	if err != nil {
		return fmt.Errorf("getting signing key: %w", err)
	}

	// Parse the private key
	entityList, err := openpgp.ReadArmoredKeyRing(bytes.NewReader(privateKeyBytes))
	if err != nil {
		// Try without armor
		entityList, err = openpgp.ReadKeyRing(bytes.NewReader(privateKeyBytes))
		if err != nil {
			return fmt.Errorf("parsing private key: %w", err)
		}
	}

	if len(entityList) == 0 {
		return fmt.Errorf("no keys found in private key data")
	}

	// Find the entity with the matching key ID or email
	var selectedEntity *openpgp.Entity
	for _, entity := range entityList {
		// Check if key ID matches
		if g.keyID != "" {
			for _, identity := range entity.Identities {
				if strings.Contains(identity.Name, g.keyID) ||
					strings.Contains(identity.UserId.Email, g.keyID) {
					selectedEntity = entity
					break
				}
			}

			// Also check key ID directly
			if selectedEntity == nil && entity.PrimaryKey != nil {
				keyIDHex := fmt.Sprintf("%X", entity.PrimaryKey.KeyId)
				if strings.HasSuffix(keyIDHex, strings.ToUpper(g.keyID)) {
					selectedEntity = entity
				}
			}
		}

		if selectedEntity != nil {
			break
		}
	}

	// If no specific key ID requested or found, use the first entity
	if selectedEntity == nil {
		selectedEntity = entityList[0]
		g.logger.Warn("no matching key ID found, using first key",
			zap.String("requested_key_id", g.keyID))
	}

	// Check if the private key is encrypted and decrypt if needed
	if selectedEntity.PrivateKey != nil && selectedEntity.PrivateKey.Encrypted {
		// Try to get passphrase from secrets
		passphraseRef := secrets.Reference{
			Provider: g.secretRef.Provider,
			Path:     g.secretRef.Path,
			Key:      "passphrase",
		}

		passphraseBytes, err := g.secretManager.GetSecretValue(ctx, passphraseRef)
		if err != nil {
			// No passphrase available
			return fmt.Errorf("private key is encrypted but no passphrase found: %w", err)
		}

		if err := selectedEntity.PrivateKey.Decrypt(passphraseBytes); err != nil {
			return fmt.Errorf("decrypting private key: %w", err)
		}
	}

	g.entity = selectedEntity

	g.logger.Info("loaded GPG signing key",
		zap.String("key_id", g.keyID),
		zap.String("fingerprint", fmt.Sprintf("%X", selectedEntity.PrimaryKey.Fingerprint)))

	return nil
}

// Ensure GPGSigner implements signing.Signer
var _ signing.Signer = (*GPGSigner)(nil)
