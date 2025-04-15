package authz

import "context"

type Verifier interface {
	// IsAllowed returns whether the user is allowed to access the resource
	IsAllowed(ctx context.Context, input map[string]any) (bool, error)
	// ViewableEnvironments returns the list of environments the user has access to
	ViewableEnvironments(ctx context.Context, input map[string]any) ([]string, error)
	// ViewableNamespaces returns the list of namespaces the user has access to in a specific environment
	ViewableNamespaces(ctx context.Context, env string, input map[string]any) ([]string, error)
	// Shutdown is called when the server is shutting down
	Shutdown(ctx context.Context) error
}

type contextKey string

const (
	// EnvironmentsKey is the context key for storing viewable environments
	EnvironmentsKey contextKey = "environments"
	// NamespacesKey is the context key for storing viewable namespaces
	NamespacesKey contextKey = "namespaces"
)
