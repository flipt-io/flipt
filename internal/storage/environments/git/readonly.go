package git

import (
	"context"
	"fmt"

	"go.flipt.io/flipt/errors"
	serverenvironments "go.flipt.io/flipt/internal/server/environments"
	rpcenvironments "go.flipt.io/flipt/rpc/v2/environments"
)

var (
	_ serverenvironments.Environment = ReadOnlyEnvironment{}

	// ErrReadOnly is returned when a mutating operation is attempted on a read-only environment
	ErrReadOnly = errors.ErrNotImplemented("cannot perform operation on read-only environment")
)

type ReadOnlyEnvironment struct {
	*Environment
}

func (r ReadOnlyEnvironment) PutNamespace(_ context.Context, _ string, ns *rpcenvironments.Namespace) (string, error) {
	return "", fmt.Errorf("put namespace %q: %w", ns.Key, ErrReadOnly)
}

func (r ReadOnlyEnvironment) DeleteNamespace(_ context.Context, _ string, key string) (string, error) {
	return "", fmt.Errorf("delete namespace %q: %w", key, ErrReadOnly)
}

func (r ReadOnlyEnvironment) Update(_ context.Context, _ string, typ serverenvironments.ResourceType, _ serverenvironments.UpdateFunc) (string, error) {
	return "", fmt.Errorf("update resource type %q: %w", typ, ErrReadOnly)
}
