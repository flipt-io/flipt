package common

import "context"

// ContextKey is a type for context keys used across Flipt packages.
type ContextKey string

const (
	HeaderFliptEnvironment            = "X-Flipt-Environment"
	HeaderFliptNamespace              = "X-Flipt-Namespace"
	HeaderFliptOFREPStream            = "x-flipt-ofrep-stream"
	ContextKeyOFREPStream  ContextKey = "ofrep-stream"
)

// WithOFREPStream sets the OFREP stream marker on the context.
func WithOFREPStream(ctx context.Context) context.Context {
	return context.WithValue(ctx, ContextKeyOFREPStream, true)
}

// IsOFREPStream reports whether the context has the OFREP stream marker set.
func IsOFREPStream(ctx context.Context) bool {
	v := ctx.Value(ContextKeyOFREPStream)
	ok, _ := v.(bool)
	return ok
}
