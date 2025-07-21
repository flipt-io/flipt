package signing

import (
	"context"

	"github.com/go-git/go-git/v6/plumbing/object"
)

// Signer defines the interface for commit signing implementations.
type Signer interface {
	// SignCommit creates a signature for the given commit.
	SignCommit(ctx context.Context, commit *object.Commit) (string, error)

	// GetPublicKey returns the public key used for verification.
	GetPublicKey(ctx context.Context) (string, error)
}

// Type represents the type of signing method.
type Type string

const (
	// TypeGPG represents GPG signing.
	TypeGPG Type = "gpg"

	// TypeSSH represents SSH signing (future).
	TypeSSH Type = "ssh"
)
