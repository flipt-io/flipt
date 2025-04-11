package common

import (
	"context"

	"github.com/blang/semver/v4"
	"go.flipt.io/flipt/rpc/flipt"
)

type fliptAcceptServerVersionContextKey struct{}

// WithFliptAcceptServerVersion sets the flipt version in the context.
func WithFliptAcceptServerVersion(ctx context.Context, version semver.Version) context.Context {
	return context.WithValue(ctx, fliptAcceptServerVersionContextKey{}, version)
}

// The last version that does not support the x-flipt-accept-server-version header.
var preFliptAcceptServerVersion = semver.MustParse("1.37.1")

// FliptAcceptServerVersionFromContext returns the flipt-accept-server-version from the context if it exists or the default version.
func FliptAcceptServerVersionFromContext(ctx context.Context) semver.Version {
	v, ok := ctx.Value(fliptAcceptServerVersionContextKey{}).(semver.Version)
	if !ok {
		return preFliptAcceptServerVersion
	}
	return v
}

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
