package authz

import "context"

type Verifier interface {
	// IsAllowed returns whether the user is allowed to access the resource
	IsAllowed(ctx context.Context, input map[string]any) (bool, error)
	// Namespaces returns the list of namespaces the user has access to
	Namespaces(ctx context.Context, input map[string]any) ([]string, error)
	// Shutdown is called when the server is shutting down
	Shutdown(ctx context.Context) error
}

type contextKey string

const NamespacesKey contextKey = "namespaces"
