package sdk

import (
	context "context"
	os "os"
	sync "sync"
	time "time"

	auth "go.flipt.io/flipt/rpc/flipt/auth"
)

const (
	defaultServiceAccountTokenPath = "/var/run/secrets/kubernetes.io/serviceaccount/token"
	defaultKubernetesExpiryLeeway  = 10 * time.Second
)

// KubernetesAuthenticationProvider is an implementation of ClientAuthenticationProvider
// which automatically uses the service account token from the environment and verifies it
// with Kubernetes.
// This provider keeps the JWT up to date and refreshes it for a new JWT token before expiry.
// It re-reads the service account token as Kubernetes can and will refresh
// this token, as it also has its own expiry.
type KubernetesAuthenticationProvider struct {
	transport               Transport
	serviceAccountTokenPath string
	leeway                  time.Duration

	mu   sync.RWMutex
	resp *auth.VerifyServiceAccountResponse
}

// KubernetesAuthenticationProviderOption is a functional option for configuring KubernetesAuthenticationProvider.
type KubernetesAuthenticationProviderOption func(*KubernetesAuthenticationProvider)

// WithKubernetesServiceAccountTokenPath sets the path on the host to locate the kubernetes service account.
// The KubernetesAuthenticationProvider uses the default location set by Kubernetes.
// This option lets you override that if your path happens to differ.
func WithKubernetesServiceAccountTokenPath(p string) KubernetesAuthenticationProviderOption {
	return func(kctp *KubernetesAuthenticationProvider) {
		kctp.serviceAccountTokenPath = p
	}
}

// WithKubernetesExpiryLeeway configures the duration leeway for deciding when to refresh
// the JWT. The default is 10 seconds, which ensures that tokens are automatically refreshed
// when their is less that 10 seconds of lifetime left on the previously fetched JWT.
func WithKubernetesExpiryLeeway(d time.Duration) KubernetesAuthenticationProviderOption {
	return func(kctp *KubernetesAuthenticationProvider) {
		kctp.leeway = d
	}
}

// NewKubernetesAuthenticationProvider constructs and configures a new KubernetesAuthenticationProvider
// using the provided transport.
func NewKubernetesAuthenticationProvider(transport Transport, opts ...KubernetesAuthenticationProviderOption) *KubernetesAuthenticationProvider {
	k := &KubernetesAuthenticationProvider{
		transport:               transport,
		serviceAccountTokenPath: defaultServiceAccountTokenPath,
		leeway:                  defaultKubernetesExpiryLeeway,
	}

	for _, opt := range opts {
		opt(k)
	}

	return k
}

// Authentication returns the authentication header string to be used for a request
// by the client SDK. It is generated via verifying the local service account token
// with Kubernetes. The JWT is then formatted appropriately for use
// in the Authentication header as a JWT token.
func (k *KubernetesAuthenticationProvider) Authentication(ctx context.Context) (string, error) {
	k.mu.RLock()
	resp := k.resp
	k.mu.RUnlock()
	if resp != nil && time.Now().UTC().Add(k.leeway).Before(resp.Authentication.ExpiresAt.AsTime()) {
		return JWTAuthenticationProvider(k.resp.ClientToken).Authentication(ctx)
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

	return JWTAuthenticationProvider(k.resp.ClientToken).Authentication(ctx)
}
