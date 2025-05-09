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
	fliptAddr = flag.String("flipt-addr", "grpc://localhost:9000", "Address for running Flipt instance (gRPC only)")
	// tokens are used to authenticate requests to the Flipt instance
	// they are used to create a new token for each role
	// and then use that token to authenticate requests to the Flipt instance
	tokens = map[string]string{
		"bootstrap":          "s3cr3t",
		"admin":              "admin123",
		"editor":             "editor456",
		"viewer":             "viewer789",
		"default_viewer":     "default_viewer1111",
		"alternative_viewer": "alternative_viewer2222",
	}
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

type KeyedExpectation struct {
	Key      string
	Expected string
}

type KeyedExpectations []KeyedExpectation

// OtherKeyFrom returns any other key in the set that has a different Expected value
// than the Expected value of the provided key
func (n KeyedExpectations) OtherKeyFrom(from string) string {
	// First find the Expected value for our input key
	var fromExpected string
	for _, e := range n {
		if e.Key == from {
			fromExpected = e.Expected
			break
		}
	}

	// Now find a key with a different Expected value
	for _, e := range n {
		if e.Expected != fromExpected {
			return e.Key
		}
	}

	panic("we expected to be at-least one alternative with different Expected value")
}

var Environments = KeyedExpectations{
	{Key: "", Expected: DefaultEnvironment},
	{Key: DefaultEnvironment, Expected: DefaultEnvironment},
	{Key: ProductionEnvironment, Expected: ProductionEnvironment},
}

var Namespaces = KeyedExpectations{
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
		URL:    u,
		Tokens: tokens,
	})
}

type TestOpts struct {
	URL    *url.URL
	Tokens map[string]string
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

type ClientOpts struct {
	Role      string
	Transport http.RoundTripper
}

type ClientOpt func(*ClientOpts)

func WithTransport(transport http.RoundTripper) ClientOpt {
	return func(co *ClientOpts) {
		co.Transport = transport
	}
}

func WithRole(role string) ClientOpt {
	return func(co *ClientOpts) {
		co.Role = role
	}
}

type roundTripFunc func(r *http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(r *http.Request) (*http.Response, error) {
	return f(r)
}

func (o TestOpts) BootstrapClient(t *testing.T, _ ...ClientOpt) sdk.SDK {
	t.Helper()

	return sdk.New(o.newTransport(t), sdk.WithAuthenticationProvider(
		sdk.StaticTokenAuthenticationProvider(o.Tokens["bootstrap"]),
	))
}

func (o TestOpts) HTTPClient(t *testing.T, opts ...ClientOpt) *http.Client {
	t.Helper()

	var options ClientOpts
	for _, opt := range opts {
		opt(&options)
	}

	// Default to bootstrap token if no specific token name provided
	token := o.Tokens["bootstrap"]
	if options.Role != "" {
		if tok, ok := o.Tokens[options.Role]; ok {
			token = tok
		} else {
			t.Fatalf("token with name %q not found", options.Role)
		}
	}

	baseTransport := options.Transport
	if baseTransport == nil {
		baseTransport = http.DefaultTransport
	}

	transport := roundTripFunc(func(r *http.Request) (*http.Response, error) {
		r.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))
		r.Header.Set("X-Forwarded-Host", o.URL.Host)
		return baseTransport.RoundTrip(r)
	})

	return &http.Client{Transport: transport}
}

func (o TestOpts) JWTClient(t *testing.T, opts ...ClientOpt) sdk.SDK {
	return sdk.New(o.newTransport(t), sdk.WithAuthenticationProvider(
		sdk.JWTAuthenticationProvider(jwtClaims(t, opts...)),
	))
}

func (o TestOpts) K8sClient(t *testing.T, opts ...ClientOpt) sdk.SDK {
	t.Helper()

	transport := o.newTransport(t)
	return sdk.New(transport, sdk.WithAuthenticationProvider(
		sdk.NewKubernetesAuthenticationProvider(transport, sdk.WithKubernetesServiceAccountTokenPath(k8sServiceAccountToken(t, opts...))),
	))
}

func k8sServiceAccountToken(t *testing.T, opts ...ClientOpt) string {
	t.Helper()

	var copts ClientOpts
	for _, opt := range opts {
		opt(&copts)
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
				"name": "flipt",
				"uid":  "4f18914e-f276-44b2-aebd-27db1d8f8def",
			},
		},
	})

	// write out JWT into service account token slot
	require.NoError(t, os.MkdirAll("/var/run/secrets/kubernetes.io/serviceaccount", 0755))

	tokenPath := "/var/run/secrets/kubernetes.io/serviceaccount/token"
	require.NoError(t, os.WriteFile(tokenPath, []byte(saToken), 0644))

	return tokenPath
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

func (o TestOpts) NoAuthClientV2(t *testing.T, opts ...sdkv2.Option) sdkv2.SDK {
	t.Helper()

	return sdkv2.New(o.newTransportV2(t), opts...)
}

func (o TestOpts) TokenClientV2(t *testing.T, opts ...ClientOpt) sdkv2.SDK {
	t.Helper()

	var options ClientOpts
	for _, opt := range opts {
		opt(&options)
	}

	// Default to bootstrap token if no specific token name provided
	token := o.Tokens["bootstrap"]
	if options.Role != "" {
		if tok, ok := o.Tokens[options.Role]; ok {
			token = tok
		} else {
			t.Fatalf("token with name %q not found", options.Role)
		}
	}

	return sdkv2.New(o.newTransportV2(t), sdkv2.WithAuthenticationProvider(
		sdkv2.StaticTokenAuthenticationProvider(token),
	))
}

func (o TestOpts) JWTClientV2(t *testing.T, opts ...ClientOpt) sdkv2.SDK {
	t.Helper()

	return sdkv2.New(o.newTransportV2(t), sdkv2.WithAuthenticationProvider(
		sdkv2.JWTAuthenticationProvider(jwtClaims(t, opts...)),
	))
}

func (o TestOpts) K8sClientV2(t *testing.T, opts ...ClientOpt) sdkv2.SDK {
	t.Helper()

	return sdkv2.New(o.newTransportV2(t), sdkv2.WithAuthenticationProvider(
		sdk.NewKubernetesAuthenticationProvider(o.newTransport(t), sdk.WithKubernetesServiceAccountTokenPath(k8sServiceAccountToken(t, opts...))),
	))
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
