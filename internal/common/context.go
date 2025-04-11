package common

import (
	"context"

	"go.flipt.io/flipt/rpc/flipt"
)

type fliptEnvironmentContextKey struct{}

// WithFliptEnvironment sets the flipt environment in the context.
func WithFliptEnvironment(ctx context.Context, environment string) context.Context {
	return context.WithValue(ctx, fliptEnvironmentContextKey{}, environment)
}

func FliptEnvironmentFromContext(ctx context.Context) (string, bool) {
	v, ok := ctx.Value(fliptEnvironmentContextKey{}).(string)
	if !ok {
		return flipt.DefaultEnvironment, false
	}
	return v, true
}

type fliptNamespaceContextKey struct{}

// WithFliptNamespace sets the flipt namespace in the context.
func WithFliptNamespace(ctx context.Context, namespace string) context.Context {
	return context.WithValue(ctx, fliptNamespaceContextKey{}, namespace)
}

func FliptNamespaceFromContext(ctx context.Context) (string, bool) {
	v, ok := ctx.Value(fliptNamespaceContextKey{}).(string)
	if !ok {
		return flipt.DefaultNamespace, false
	}
	return v, true
}
