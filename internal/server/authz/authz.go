package authz

import "context"

type Verifier interface {
	IsAllowed(ctx context.Context, input map[string]any) (bool, error)
	Shutdown(ctx context.Context) error
}
