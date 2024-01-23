// Code generated by protoc-gen-go-flipt-sdk. DO NOT EDIT.

package sdk

import (
	context "context"
	os "os"
	sync "sync"
	time "time"

	flipt "go.flipt.io/flipt/rpc/flipt"
	auth "go.flipt.io/flipt/rpc/flipt/auth"
	evaluation "go.flipt.io/flipt/rpc/flipt/evaluation"
	meta "go.flipt.io/flipt/rpc/flipt/meta"
	metadata "google.golang.org/grpc/metadata"
)

var _ *time.Time
var _ *os.File
var _ *sync.Mutex
var _ auth.Method

const (
	defaultServiceAccountTokenPath = "/var/run/secrets/kubernetes.io/serviceaccount/token"
	defaultKubernetesExpiryLeeway  = 10 * time.Second
)

type Transport interface {
	AuthClient() AuthClient
	EvaluationClient() evaluation.EvaluationServiceClient
	FliptClient() flipt.FliptClient
	MetaClient() meta.MetadataServiceClient
}

// ClientTokenProvider is a type which when requested provides a
// client token which can be used to authenticate RPC/API calls
// invoked through the SDK.
// Deprecated: Use ClientAuthenticationProvider instead.
type ClientTokenProvider interface {
	ClientToken() (string, error)
}

// WithClientTokenProviders returns an Option which configures
// any supplied SDK with the provided ClientTokenProvider.
// Deprecated: Use WithAuthenticationProvider instead.
func WithClientTokenProvider(p ClientTokenProvider) Option {
	return func(s *SDK) {
		s.authenticationProvider = authenticationProviderFunc(func(context.Context) (string, error) {
			clientToken, err := p.ClientToken()
			if err != nil {
				return "", err
			}

			return "Bearer " + string(clientToken), nil
		})
	}
}

type authenticationProviderFunc func(context.Context) (string, error)

func (f authenticationProviderFunc) Authentication(ctx context.Context) (string, error) {
	return f(ctx)
}

// StaticClientTokenProvider is a string which is supplied as a static client token
// on each RPC which requires authentication.
// Deprecated: Use StaticTokenAuthenticationProvider instead.
type StaticClientTokenProvider string

// ClientToken returns the underlying string that is the StaticClientTokenProvider.
// Deprecated: Use StaticTokenAuthenticationProvider instead.
func (p StaticClientTokenProvider) ClientToken() (string, error) {
	return string(p), nil
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

// KubernetesClientTokenProvider is an implementation of ClientAuthenticationProvider
// which automatically uses the service account token from the environment and exchanges
// it with Flipt for a client token.
// This provider keeps the client token up to date and refreshes it for a new client
// token before expiry. It re-reads the service account token as Kubernetes can and will refresh
// this token, as it also has its own expiry.
type KubernetesClientTokenProvider struct {
	transport               Transport
	serviceAccountTokenPath string
	leeway                  time.Duration

	mu   sync.RWMutex
	resp *auth.VerifyServiceAccountResponse
}

// KubernetesClientTokenProviderOption is a functional option for configuring KubernetesClientTokenProvider.
type KubernetesClientTokenProviderOption func(*KubernetesClientTokenProvider)

// WithKubernetesServiceAccountTokenPath sets the path on the host to locate the kubernetes service account.
// The KubernetesClientTokenProvider uses the default location set by Kubernetes.
// This option lets you override that if your path happens to differ.
func WithKubernetesServiceAccountTokenPath(p string) KubernetesClientTokenProviderOption {
	return func(kctp *KubernetesClientTokenProvider) {
		kctp.serviceAccountTokenPath = p
	}
}

// WithKubernetesExpiryLeeway configures the duration leeway for deciding when to refresh
// the client token. The default is 10 seconds, which ensures that tokens are automatically refreshed
// when their is less that 10 seconds of lifetime left on the previously fetched client token.
func WithKubernetesExpiryLeeway(d time.Duration) KubernetesClientTokenProviderOption {
	return func(kctp *KubernetesClientTokenProvider) {
		kctp.leeway = d
	}
}

// NewKuberntesClientTokenProvider cosntructs and configures a new KubernetesClientTokenProvider
// using the provided transport.
func NewKuberntesClientTokenProvider(transport Transport) *KubernetesClientTokenProvider {
	return &KubernetesClientTokenProvider{
		transport:               transport,
		serviceAccountTokenPath: defaultServiceAccountTokenPath,
		leeway:                  defaultKubernetesExpiryLeeway,
	}
}

// Authentication returns the authentication header string to be used for a request
// by the client SDK. It is generated via exchanging the local service account token
// with Flipt for a client token. The token is then formatted appropriately for use
// in the Authentication header as a bearer token.
func (k *KubernetesClientTokenProvider) Authentication(ctx context.Context) (string, error) {
	k.mu.RLock()
	resp := k.resp
	k.mu.RUnlock()
	if resp != nil && time.Now().UTC().Add(k.leeway).Before(resp.Authentication.ExpiresAt.AsTime()) {
		return StaticTokenAuthenticationProvider(k.resp.ClientToken).Authentication(ctx)
	}

	k.mu.Lock()
	defer k.mu.Unlock()
	saToken, err := os.ReadFile(k.serviceAccountTokenPath)
	if err != nil {
		return "", err
	}

	resp, err = k.transport.
		AuthClient().
		AuthenticationMethodKubernetesServiceClient().
		VerifyServiceAccount(ctx, &auth.VerifyServiceAccountRequest{
			ServiceAccountToken: string(saToken),
		})
	if err != nil {
		return "", err
	}

	k.resp = resp

	return StaticTokenAuthenticationProvider(k.resp.ClientToken).Authentication(ctx)
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

func (s SDK) Auth() *Auth {
	return &Auth{
		transport:              s.transport.AuthClient(),
		authenticationProvider: s.authenticationProvider,
	}
}

func (s SDK) Evaluation() *Evaluation {
	return &Evaluation{
		transport:              s.transport.EvaluationClient(),
		authenticationProvider: s.authenticationProvider,
	}
}

func (s SDK) Flipt() *Flipt {
	return &Flipt{
		transport:              s.transport.FliptClient(),
		authenticationProvider: s.authenticationProvider,
	}
}

func (s SDK) Meta() *Meta {
	return &Meta{
		transport:              s.transport.MetaClient(),
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
