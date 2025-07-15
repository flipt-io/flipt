package signing

import (
	"context"
	"fmt"
	"strings"

	"go.flipt.io/build/internal/dagger"
)

// VerifyCommitSignatures uses Dagger to verify commit signatures in the container
func VerifyCommitSignatures(ctx context.Context, flipt *dagger.Container, repoPath string) (*CommitVerificationResult, error) {
	// First check if the repository exists and has commits
	repoCheck := flipt.WithExec([]string{"sh", "-c", fmt.Sprintf("test -d %s/.git && echo 'repo_exists' || echo 'repo_missing'", repoPath)})
	
	output, err := repoCheck.Stdout(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to check repository: %w", err)
	}
	
	if strings.TrimSpace(output) != "repo_exists" {
		return &CommitVerificationResult{
			RepositoryExists: false,
			Error:           "git repository not found",
		}, nil
	}

	// Get the latest commit hash
	commitHashCmd := flipt.WithExec([]string{"git", "-C", repoPath, "rev-parse", "HEAD"})
	commitHash, err := commitHashCmd.Stdout(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get commit hash: %w", err)
	}
	commitHash = strings.TrimSpace(commitHash)

	// Check the commit signature status
	sigCheckCmd := flipt.WithExec([]string{"git", "-C", repoPath, "show", "--show-signature", "--format=%G?", "-s", commitHash})
	sigStatus, err := sigCheckCmd.Stdout(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to check signature: %w", err)
	}
	sigStatus = strings.TrimSpace(sigStatus)

	// Get commit details with signature information
	detailsCmd := flipt.WithExec([]string{"git", "-C", repoPath, "log", "--show-signature", "-1", "--pretty=fuller", commitHash})
	details, err := detailsCmd.Stdout(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get commit details: %w", err)
	}

	// Check for PGP signature markers in the commit object
	rawCommitCmd := flipt.WithExec([]string{"git", "-C", repoPath, "cat-file", "commit", commitHash})
	rawCommit, err := rawCommitCmd.Stdout(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get raw commit: %w", err)
	}

	result := &CommitVerificationResult{
		RepositoryExists:   true,
		CommitHash:         commitHash,
		SignatureStatus:    sigStatus,
		CommitDetails:      details,
		RawCommit:          rawCommit,
		HasSignature:       sigStatus != "N" && sigStatus != "",
		HasPGPSignature:    strings.Contains(rawCommit, "-----BEGIN PGP SIGNATURE-----"),
	}

	return result, nil
}

// CommitVerificationResult contains the results of commit signature verification
type CommitVerificationResult struct {
	RepositoryExists bool
	CommitHash       string
	SignatureStatus  string // Git signature status (%G?)
	CommitDetails    string // Full commit log with signature info
	RawCommit        string // Raw commit object
	HasSignature     bool   // Whether git detected any signature
	HasPGPSignature  bool   // Whether PGP signature block is present
	Error            string // Any error encountered
}

// IsValid returns true if the commit has a valid signature
func (r *CommitVerificationResult) IsValid() bool {
	return r.RepositoryExists && r.HasSignature && r.HasPGPSignature
}

// Summary returns a human-readable summary of the verification
func (r *CommitVerificationResult) Summary() string {
	if !r.RepositoryExists {
		return fmt.Sprintf("Repository not found: %s", r.Error)
	}

	status := "unsigned"
	if r.HasPGPSignature {
		status = "signed (PGP)"
	} else if r.HasSignature {
		status = "signed (other)"
	}

	return fmt.Sprintf("Commit %s is %s (status: %s)", 
		r.CommitHash[:8], status, r.SignatureStatus)
}