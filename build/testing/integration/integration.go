package integration

import (
	"context"
	"crypto/x509"
	"encoding/pem"
	"flag"
	"net/url"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/go-jose/go-jose/v3"
	"github.com/go-jose/go-jose/v3/jwt"
	"github.com/stretchr/testify/require"
	"go.flipt.io/flipt/rpc/flipt/auth"
	sdk "go.flipt.io/flipt/sdk/go"
	sdkgrpc "go.flipt.io/flipt/sdk/go/grpc"
	sdkhttp "go.flipt.io/flipt/sdk/go/http"
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
	DefaultNamespace    = "default"
	ProductionNamespace = "production"
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
	{Key: ProductionNamespace, Expected: ProductionNamespace},
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

func (o TestOpts) bootstrapClient(t *testing.T) sdk.SDK {
	t.Helper()

	return sdk.New(o.newTransport(t), sdk.WithAuthenticationProvider(
		sdk.StaticTokenAuthenticationProvider(o.Token),
	))
}

type ClientOpts struct {
	Namespace string
	Metadata  map[string]string
}

type ClientOpt func(*ClientOpts)

func WithNamespace(ns string) ClientOpt {
	return func(co *ClientOpts) {
		co.Namespace = ns
	}
}

func WithRole(role string) ClientOpt {
	return func(co *ClientOpts) {
		if co.Metadata == nil {
			co.Metadata = map[string]string{}
		}

		co.Metadata["io.flipt.auth.role"] = role
	}
}

func (o TestOpts) TokenClient(t *testing.T, opts ...ClientOpt) sdk.SDK {
	t.Helper()

	var copts ClientOpts
	for _, opt := range opts {
		opt(&copts)
	}

	resp, err := o.bootstrapClient(t).Auth().AuthenticationMethodTokenService().CreateToken(context.Background(), &auth.CreateTokenRequest{
		Name:         t.Name(),
		NamespaceKey: copts.Namespace,
		Metadata:     copts.Metadata,
	})
	require.NoError(t, err)

	return sdk.New(o.newTransport(t), sdk.WithAuthenticationProvider(
		sdk.StaticTokenAuthenticationProvider(resp.ClientToken),
	))
}

func (o TestOpts) K8sClient(t *testing.T) sdk.SDK {
	t.Helper()

	transport := o.newTransport(t)
	return sdk.New(transport, sdk.WithAuthenticationProvider(
		sdk.NewKubernetesAuthenticationProvider(transport),
	))
}

func (o TestOpts) JWTClient(t *testing.T) sdk.SDK {
	t.Helper()

	bytes, err := os.ReadFile("/var/run/secrets/flipt/private.pem")
	if err != nil {
		t.Fatal(err)
	}

	block, _ := pem.Decode(bytes)
	key, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		t.Fatal(err)
	}

	sig, err := jose.NewSigner(
		jose.SigningKey{Algorithm: jose.RS256, Key: key},
		(&jose.SignerOptions{}).WithType("JWT"),
	)

	raw, err := jwt.Signed(sig).
		Claims(map[string]any{
			"iss": "https://flipt.io",
			"iat": time.Now().Unix(),
			"exp": time.Now().Add(3 * time.Minute).Unix(),
		}).
		CompactSerialize()
	if err != nil {
		panic(err)
	}

	return sdk.New(o.newTransport(t), sdk.WithAuthenticationProvider(
		sdk.JWTAuthenticationProvider(raw),
	))
}

func (o TestOpts) newTransport(t *testing.T) (transport sdk.Transport) {
	switch o.Protocol() {
	case ProtocolGRPC:
		conn, err := grpc.NewClient(o.URL.Host, grpc.WithTransportCredentials(insecure.NewCredentials()))
		require.NoError(t, err)

		transport = sdkgrpc.NewTransport(conn)
	case ProtocolHTTP, ProtocolHTTPS:
		transport = sdkhttp.NewTransport(o.URL.String())
	default:
		t.Fatalf("Unexpected flipt address protocol %s", o.URL)
	}

	return
}
