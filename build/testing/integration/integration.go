package integration

import (
	"crypto/x509"
	"encoding/pem"
	"flag"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/go-jose/go-jose/v3"
	"github.com/go-jose/go-jose/v3/jwt"
	"github.com/stretchr/testify/require"
	sdk "go.flipt.io/flipt/sdk/go"
	sdkgrpc "go.flipt.io/flipt/sdk/go/grpc"
	sdkhttp "go.flipt.io/flipt/sdk/go/http"
	sdkv2 "go.flipt.io/flipt/sdk/go/v2"
	sdkv2grpc "go.flipt.io/flipt/sdk/go/v2/grpc"
	sdkv2http "go.flipt.io/flipt/sdk/go/v2/http"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

var (
	fliptAddr       = flag.String("flipt-addr", "grpc://localhost:9000", "Address for running Flipt instance (gRPC only)")
	fliptToken      = flag.String("flipt-token", "", "Full-Access authentication token to be used during test suite")
	fliptReferences = flag.Bool("flipt-supports-references", false, "Identifies the backend as supporting references")
)

type Protocol string

const (
	ProtocolHTTP  Protocol = "http"
	ProtocolHTTPS Protocol = "https"
	ProtocolGRPC  Protocol = "grpc"
)

const (
	DefaultNamespace     = "default"
	AlternativeNamespace = "alternative"

	DefaultEnvironment    = "default"
	ProductionEnvironment = "production"
)

type NamespaceExpectation struct {
	Key      string
	Expected string
}

type NamespaceExpectations []NamespaceExpectation

// OtherNamespaceFrom returns any other namespace in the set of expected
// namespaces different from the namespace provided
func (n NamespaceExpectations) OtherNamespaceFrom(from string) string {
	ns := map[string]struct{}{}
	for _, e := range n {
		if e.Expected == from {
			continue
		}

		ns[e.Expected] = struct{}{}
	}

	for to := range ns {
		return to
	}

	panic("we expected to be at-least one alternative")
}

var Namespaces = NamespaceExpectations{
	{Key: "", Expected: DefaultNamespace},
	{Key: DefaultNamespace, Expected: DefaultNamespace},
	{Key: AlternativeNamespace, Expected: AlternativeNamespace},
}

func Harness(t *testing.T, fn func(t *testing.T, opts TestOpts)) {
	u, err := url.Parse(*fliptAddr)
	if err != nil {
		t.Fatal(err)
	}

	fn(t, TestOpts{
		URL:        u,
		References: *fliptReferences,
		Token:      *fliptToken,
	})
}

type TestOpts struct {
	URL        *url.URL
	References bool
	Token      string
}

func (o TestOpts) Protocol() Protocol {
	if o.URL.Scheme == "" {
		return ProtocolHTTP
	}

	return Protocol(strings.TrimSuffix(o.URL.Scheme, ":"))
}

func (o TestOpts) NoAuthClient(t *testing.T) sdk.SDK {
	t.Helper()

	return sdk.New(o.newTransport(t))
}

func (o TestOpts) BootstrapClient(t *testing.T, _ ...ClientOpt) sdk.SDK {
	t.Helper()

	return sdk.New(o.newTransport(t), sdk.WithAuthenticationProvider(
		sdk.StaticTokenAuthenticationProvider(o.Token),
	))
}

type ClientOpts struct {
	Role string
}

type ClientOpt func(*ClientOpts)

func WithRole(role string) ClientOpt {
	return func(co *ClientOpts) {
		co.Role = role
	}
}

type roundTripFunc func(r *http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(r *http.Request) (*http.Response, error) {
	return f(r)
}

func (o TestOpts) HTTPClient(t *testing.T, opts ...ClientOpt) *http.Client {
	t.Helper()

	transport := roundTripFunc(func(r *http.Request) (*http.Response, error) {
		r.Header.Set("Authorization", fmt.Sprintf("Bearer %s", o.Token))
		r.Header.Set("X-Forwarded-Host", o.URL.Host)
		return http.DefaultTransport.RoundTrip(r)
	})

	return &http.Client{Transport: transport}
}

func (o TestOpts) K8sClient(t *testing.T, opts ...ClientOpt) sdk.SDK {
	transport := o.newTransport(t)
	return sdk.New(transport, K8sAuth(t, transport, opts...))
}

func K8sAuth(t *testing.T, transport sdk.Transport, opts ...ClientOpt) sdk.Option {
	t.Helper()

	var copts ClientOpts
	for _, opt := range opts {
		opt(&copts)
	}

	saName := "integration-test"
	if copts.Role != "" {
		saName = copts.Role
	}

	saToken := signWithPrivateKeyClaims(t, "/var/run/secrets/flipt/k8s.pem", map[string]any{
		"exp": time.Now().Add(24 * time.Hour).Unix(),
		"iss": "https://discover.svc",
		"kubernetes.io": map[string]any{
			"namespace": "integration",
			"pod": map[string]any{
				"name": "integration-test-7d26f049-kdurb",
				"uid":  "bd8299f9-c50f-4b76-af33-9d8e3ef2b850",
			},
			"serviceaccount": map[string]any{
				"name": saName,
				"uid":  "4f18914e-f276-44b2-aebd-27db1d8f8def",
			},
		},
	})

	// write out JWT into service account token slot
	require.NoError(t, os.MkdirAll("/var/run/secrets/kubernetes.io/serviceaccount", 0755))

	tokenPath := fmt.Sprintf("/var/run/secrets/kubernetes.io/serviceaccount/%s.token", saName)
	require.NoError(t, os.WriteFile(tokenPath, []byte(saToken), 0644))
	return sdk.WithAuthenticationProvider(
		sdk.NewKubernetesAuthenticationProvider(transport, sdk.WithKubernetesServiceAccountTokenPath(tokenPath)),
	)
}

func (o TestOpts) JWTClient(t *testing.T, opts ...ClientOpt) sdk.SDK {
	return sdk.New(o.newTransport(t), sdk.WithAuthenticationProvider(
		sdk.JWTAuthenticationProvider(jwtClaims(t, opts...)),
	))
}

func jwtClaims(t *testing.T, opts ...ClientOpt) string {
	t.Helper()

	var copts ClientOpts
	for _, opt := range opts {
		opt(&copts)
	}

	claims := map[string]any{
		"iss": "https://flipt.io",
		"iat": time.Now().Unix(),
		"exp": time.Now().Add(3 * time.Minute).Unix(),
	}

	if copts.Role != "" {
		claims["io.flipt.auth.role"] = copts.Role
	}

	return signWithPrivateKeyClaims(t, "/var/run/secrets/flipt/jwt.pem", claims)
}

func (o TestOpts) TokenClientV2(t *testing.T, opts ...ClientOpt) sdkv2.SDK {
	return sdkv2.New(o.newTransportV2(t), sdkv2.WithAuthenticationProvider(
		sdkv2.StaticTokenAuthenticationProvider(o.Token),
	))
}

func (o TestOpts) JWTClientV2(t *testing.T, opts ...ClientOpt) sdkv2.SDK {
	return sdkv2.New(o.newTransportV2(t), sdkv2.WithAuthenticationProvider(
		sdkv2.JWTAuthenticationProvider(jwtClaims(t, opts...)),
	))
}

func (o TestOpts) ClientV2(t *testing.T, opts ...sdkv2.Option) sdkv2.SDK {
	t.Helper()

	return sdkv2.New(o.newTransportV2(t), opts...)
}

func signWithPrivateKeyClaims(t *testing.T, privPath string, claims map[string]any) string {
	t.Helper()

	bytes, err := os.ReadFile(privPath)
	require.NoError(t, err)

	block, _ := pem.Decode(bytes)
	key, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	require.NoError(t, err)

	sig, err := jose.NewSigner(
		jose.SigningKey{Algorithm: jose.RS256, Key: key},
		(&jose.SignerOptions{}).WithType("JWT"),
	)

	raw, err := jwt.Signed(sig).
		Claims(claims).
		CompactSerialize()
	require.NoError(t, err)

	return raw
}

func (o TestOpts) newTransport(t *testing.T) (transport sdk.Transport) {
	switch o.Protocol() {
	case ProtocolGRPC:
		transport = sdkgrpc.NewTransport(o.newGRPCClient(t))
	case ProtocolHTTP, ProtocolHTTPS:
		transport = sdkhttp.NewTransport(o.URL.String())
	default:
		t.Fatalf("Unexpected flipt address protocol %s", o.URL)
	}

	return
}

func (o TestOpts) newTransportV2(t *testing.T) (transport sdkv2.Transport) {
	switch o.Protocol() {
	case ProtocolGRPC:
		transport = sdkv2grpc.NewTransport(o.newGRPCClient(t))
	case ProtocolHTTP, ProtocolHTTPS:
		transport = sdkv2http.NewTransport(o.URL.String())
	default:
		t.Fatalf("Unexpected flipt address protocol %s", o.URL)
	}

	return
}

func (o TestOpts) newGRPCClient(t *testing.T) *grpc.ClientConn {
	conn, err := grpc.NewClient(o.URL.Host,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	require.NoError(t, err)

	return conn
}
