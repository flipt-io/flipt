// Code generated by protoc-gen-go-flipt-sdk. DO NOT EDIT.

package sdk

import (
	context "context"
	environments "go.flipt.io/flipt/rpc/v2/environments"
	metadata "google.golang.org/grpc/metadata"
)

type Transport interface {
	EnvironmentsClient() environments.EnvironmentsServiceClient
}

// ClientAuthenticationProvider is a type which when requested provides a
// client authentication which can be used to authenticate RPC/API calls
// invoked through the SDK.
type ClientAuthenticationProvider interface {
	Authentication(context.Context) (string, error)
}

// SDK is the definition of Flipt's Go SDK.
// It depends on a pluggable transport implementation and exposes
// a consistent API surface area across both transport implementations.
// It also provides consistent client-side instrumentation and authentication
// lifecycle support.
type SDK struct {
	transport              Transport
	authenticationProvider ClientAuthenticationProvider
}

// Option is a functional option which configures the Flipt SDK.
type Option func(*SDK)

// WithAuthenticationProviders returns an Option which configures
// any supplied SDK with the provided ClientAuthenticationProvider.
func WithAuthenticationProvider(p ClientAuthenticationProvider) Option {
	return func(s *SDK) {
		s.authenticationProvider = p
	}
}

// StaticTokenAuthenticationProvider is a string which is supplied as a static client authentication
// on each RPC which requires authentication.
type StaticTokenAuthenticationProvider string

// Authentication returns the underlying string that is the StaticTokenAuthenticationProvider.
func (p StaticTokenAuthenticationProvider) Authentication(context.Context) (string, error) {
	return "Bearer " + string(p), nil
}

// JWTAuthenticationProvider is a string which is supplied as a JWT client authentication
// on each RPC which requires authentication.
type JWTAuthenticationProvider string

// Authentication returns the underlying string that is the JWTAuthenticationProvider.
func (p JWTAuthenticationProvider) Authentication(context.Context) (string, error) {
	return "JWT " + string(p), nil
}

// New constructs and configures a Flipt SDK instance from
// the provided Transport implementation and options.
func New(t Transport, opts ...Option) SDK {
	sdk := SDK{transport: t}

	for _, opt := range opts {
		opt(&sdk)
	}

	return sdk
}

func (s SDK) Environments() *Environments {
	return &Environments{
		transport:              s.transport.EnvironmentsClient(),
		authenticationProvider: s.authenticationProvider,
	}
}

func authenticate(ctx context.Context, p ClientAuthenticationProvider) (context.Context, error) {
	if p != nil {
		authentication, err := p.Authentication(ctx)
		if err != nil {
			return ctx, err
		}

		ctx = metadata.AppendToOutgoingContext(ctx, "authorization", authentication)
	}

	return ctx, nil
}
