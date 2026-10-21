package environments

import (
	"regexp"
	"strings"

	"go.flipt.io/flipt/errors"
)

// newNamespaceKeyPattern is the format required of new namespace keys. It
// must match the namespace key in the features schema (core/validation/flipt.cue),
// which the UI and v1 also enforce.
var newNamespaceKeyPattern = regexp.MustCompile(`^[-_,A-Za-z0-9]+$`)

func (r *GetNamespaceRequest) Validate() error {
	return ValidateNamespaceKey(r.GetKey())
}

func (r *UpdateNamespaceRequest) Validate() error {
	return ValidateNamespaceKey(r.GetKey())
}

func (r *DeleteNamespaceRequest) Validate() error {
	return ValidateNamespaceKey(r.GetKey())
}

func (r *GetResourceRequest) Validate() error {
	if err := ValidateNamespaceKey(r.GetNamespaceKey()); err != nil {
		return err
	}

	return requireKey(r)
}

func (r *ListResourcesRequest) Validate() error {
	return ValidateNamespaceKey(r.GetNamespaceKey())
}

func (r *UpdateResourceRequest) Validate() error {
	if err := ValidateNamespaceKey(r.GetNamespaceKey()); err != nil {
		return err
	}

	return requireKey(r)
}

func (r *DeleteResourceRequest) Validate() error {
	if err := ValidateNamespaceKey(r.GetNamespaceKey()); err != nil {
		return err
	}

	return requireKey(r)
}

// ValidateNamespaceKey rejects namespace keys that are empty or unsafe to use
// as a directory in an environment: keys containing path separators or NUL,
// and keys starting with a dot (".", ".." and ".git"). It is deliberately
// looser than ValidateNewNamespaceKey so namespaces created before keys were
// validated can still be read, updated and deleted to clean them up.
func ValidateNamespaceKey(key string) error {
	switch {
	case key == "":
		return errors.ErrInvalid("namespace key must not be empty")
	case strings.ContainsAny(key, "/\\\x00"), strings.HasPrefix(key, "."):
		return errors.ErrInvalidf("namespace key %q is not a valid directory name", key)
	}

	return nil
}

// ValidateNewNamespaceKey requires new namespace keys to contain only letters,
// numbers, hyphens, underscores and commas.
func ValidateNewNamespaceKey(key string) error {
	if !newNamespaceKeyPattern.MatchString(key) {
		return errors.ErrInvalidf("namespace key %q must contain only letters, numbers, hyphens, underscores and commas", key)
	}

	return nil
}

func requireKey(k keyed) error {
	if k.GetKey() == "" {
		return errors.ErrInvalid("key must not be empty")
	}

	return nil
}

type keyed interface {
	GetKey() string
}
