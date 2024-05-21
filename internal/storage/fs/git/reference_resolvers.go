package git

import (
	"errors"

	"github.com/Masterminds/semver/v3"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
)

// referenceResolver is a function type used to describe reference resolver functions.
type referenceResolver func(repo *git.Repository, ref string) (plumbing.Hash, error)

// staticResolver is a resolver which just resolve static references.
func staticResolver() referenceResolver {
	return func(repo *git.Repository, ref string) (plumbing.Hash, error) {
		if plumbing.IsHash(ref) {
			return plumbing.NewHash(ref), nil
		}

		reference, err := repo.Reference(plumbing.NewBranchReferenceName(ref), true)
		if err != nil {
			return plumbing.ZeroHash, err
		}

		return reference.Hash(), nil
	}
}

// semverResolver is a resolver which resolver semantic versioning references for tags.
func semverResolver() referenceResolver {
	return func(repo *git.Repository, ref string) (plumbing.Hash, error) {
		constraint, err := semver.NewConstraint(ref)
		if err != nil {
			return plumbing.ZeroHash, err
		}

		tags, err := repo.Tags()
		if err != nil {
			return plumbing.ZeroHash, err
		}

		maxVersion := semver.New(0, 0, 0, "", "")
		maxVersionHash := plumbing.ZeroHash
		err = tags.ForEach(func(reference *plumbing.Reference) error {
			version, err := semver.NewVersion(reference.Name().Short())
			if err != nil {
				// We are bypassing the error as the repository can have tags noncompliant with semver
				return nil //nolint:nilerr
			}

			if constraint.Check(version) && version.GreaterThan(maxVersion) {
				maxVersion = version
				maxVersionHash = reference.Hash()
			}

			return nil
		})
		if err != nil {
			return plumbing.ZeroHash, err
		}

		if maxVersionHash == plumbing.ZeroHash {
			return plumbing.ZeroHash, errors.New("could not find the specified tag reference")
		}

		return maxVersionHash, nil
	}
}
