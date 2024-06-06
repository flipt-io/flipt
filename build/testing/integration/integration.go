package integration

import (
	"flag"
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
	sdk "go.flipt.io/flipt/sdk/go"
	sdkgrpc "go.flipt.io/flipt/sdk/go/grpc"
	sdkhttp "go.flipt.io/flipt/sdk/go/http"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

var (
	fliptAddr       = flag.String("flipt-addr", "grpc://localhost:9000", "Address for running Flipt instance (gRPC only)")
	fliptTokenType  = flag.String("flipt-token-type", "static", "Type of token to be used during test suite (static, jwt, k8s)")
	fliptToken      = flag.String("flipt-token", "", "Authentication token to be used during test suite")
	fliptReferences = flag.Bool("flipt-supports-references", false, "Identifies the backend as supporting references")
)

type AuthConfig int

const (
	NoAuth AuthConfig = iota
	StaticTokenAuth
	JWTAuth
	K8sAuth
)

func (a AuthConfig) String() string {
	switch a {
	case NoAuth:
		return "NoAuth"
	case StaticTokenAuth:
		return "StaticTokenAuth"
	case JWTAuth:
		return "JWTAuth"
	case K8sAuth:
		return "K8sAuth"
	default:
		return "Unknown"
	}
}

func (a AuthConfig) StaticToken() bool {
	return a == StaticTokenAuth
}

func (a AuthConfig) Required() bool {
	return a != NoAuth
}

type Protocol string

const (
	ProtocolHTTP  Protocol = "http"
	ProtocolHTTPS Protocol = "https"
	ProtocolGRPC  Protocol = "grpc"
)

type TestOpts struct {
	Addr       string
	Protocol   Protocol
	AuthConfig AuthConfig
	References bool
}

const (
	DefaultNamespace    = "default"
	ProductionNamespace = "production"
)

var Namespaces = []struct {
	Key      string
	Expected string
}{
	{Key: "", Expected: DefaultNamespace},
	{Key: DefaultNamespace, Expected: DefaultNamespace},
	{Key: ProductionNamespace, Expected: ProductionNamespace},
}

func Harness(t *testing.T, fn func(t *testing.T, sdk sdk.SDK, opts TestOpts)) {
	var transport sdk.Transport

	p, host, _ := strings.Cut(*fliptAddr, "://")
	protocol := Protocol(p)

	switch protocol {
	case ProtocolGRPC:
		conn, err := grpc.NewClient(host, grpc.WithTransportCredentials(insecure.NewCredentials()))
		require.NoError(t, err)

		transport = sdkgrpc.NewTransport(conn)
	case ProtocolHTTP, ProtocolHTTPS:
		transport = sdkhttp.NewTransport(fmt.Sprintf("%s://%s", protocol, host))
	default:
		t.Fatalf("Unexpected flipt address protocol %s://%s", protocol, host)
	}

	var (
		opts       []sdk.Option
		authConfig AuthConfig
		client     sdk.SDK
	)

	switch *fliptTokenType {
	case "static":
		if *fliptToken != "" {
			authConfig = StaticTokenAuth

			opts = append(opts, sdk.WithAuthenticationProvider(
				sdk.StaticTokenAuthenticationProvider(*fliptToken),
			))

			client = sdk.New(transport, opts...)
		}
	case "jwt":
		if authentication := *fliptToken != ""; authentication {
			authConfig = JWTAuth
			opts = append(opts, sdk.WithAuthenticationProvider(
				sdk.JWTAuthenticationProvider(*fliptToken),
			))
		}
	case "k8s":
		authConfig = K8sAuth
		opts = append(opts, sdk.WithAuthenticationProvider(
			sdk.NewKubernetesAuthenticationProvider(transport),
		))
	}

	client = sdk.New(transport, opts...)

	name := fmt.Sprintf("[Protocol %q; Authentication %s]", protocol, authConfig)
	t.Run(name, func(t *testing.T) {
		fn(t, client, TestOpts{
			Protocol:   protocol,
			Addr:       *fliptAddr,
			AuthConfig: authConfig,
			References: *fliptReferences,
		})
	})
}
