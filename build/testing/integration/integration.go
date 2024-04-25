package integration

import (
	"context"
	"flag"
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
	"go.flipt.io/flipt/rpc/flipt/auth"
	sdk "go.flipt.io/flipt/sdk/go"
	sdkgrpc "go.flipt.io/flipt/sdk/go/grpc"
	sdkhttp "go.flipt.io/flipt/sdk/go/http"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

var (
	fliptAddr                  = flag.String("flipt-addr", "grpc://localhost:9000", "Address for running Flipt instance (gRPC only)")
	fliptTokenType             = flag.String("flipt-token-type", "static", "Type of token to be used during test suite (static, jwt, k8s)")
	fliptToken                 = flag.String("flipt-token", "", "Authentication token to be used during test suite")
	fliptCreateNamespacedToken = flag.Bool("flipt-create-namespaced-token", false, "Create a namespaced token for the test suite")
	fliptNamespace             = flag.String("flipt-namespace", "", "Namespace used to scope API calls.")
	fliptReferences            = flag.Bool("flipt-supports-references", false, "Identifies the backend as supporting references")
)

type AuthConfig int

const (
	NoAuth AuthConfig = iota
	StaticTokenAuth
	StaticTokenAuthNamespaced
	JWTAuth
	K8sAuth
)

func (a AuthConfig) String() string {
	switch a {
	case NoAuth:
		return "NoAuth"
	case StaticTokenAuth:
		return "StaticTokenAuth"
	case StaticTokenAuthNamespaced:
		return "StaticTokenAuthNamespaced"
	case JWTAuth:
		return "JWTAuth"
	case K8sAuth:
		return "K8sAuth"
	default:
		return "Unknown"
	}
}

func (a AuthConfig) StaticToken() bool {
	return a == StaticTokenAuth || a == StaticTokenAuthNamespaced
}

func (a AuthConfig) Required() bool {
	return a != NoAuth
}

func (a AuthConfig) NamespaceScoped() bool {
	return a == StaticTokenAuthNamespaced
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
	Namespace  string
	AuthConfig AuthConfig
	References bool
}

func Harness(t *testing.T, fn func(t *testing.T, sdk sdk.SDK, opts TestOpts)) {
	var transport sdk.Transport

	p, host, _ := strings.Cut(*fliptAddr, "://")
	protocol := Protocol(p)

	switch protocol {
	case ProtocolGRPC:
		conn, err := grpc.Dial(host, grpc.WithTransportCredentials(insecure.NewCredentials()))
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

	namespace := "default"
	if *fliptNamespace != "" {
		namespace = *fliptNamespace
	}

	switch *fliptTokenType {
	case "static":
		if *fliptToken != "" {
			authConfig = StaticTokenAuth

			opts = append(opts, sdk.WithAuthenticationProvider(
				sdk.StaticTokenAuthenticationProvider(*fliptToken),
			))

			client = sdk.New(transport, opts...)

			if *fliptCreateNamespacedToken {
				authConfig = StaticTokenAuthNamespaced

				t.Log("Creating namespaced token for test suite")

				authn, err := client.Auth().AuthenticationMethodTokenService().CreateToken(
					context.Background(),
					&auth.CreateTokenRequest{
						Name:         "Integration Test Token",
						NamespaceKey: namespace,
					},
				)

				t.Log("Created token", authn.ClientToken, authn.Authentication.Metadata)

				require.NoError(t, err)

				opts = append(opts, sdk.WithAuthenticationProvider(
					sdk.StaticTokenAuthenticationProvider(authn.ClientToken),
				))
			}
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

	name := fmt.Sprintf("[Protocol %q; Namespace %q; Authentication %s]", protocol, namespace, authConfig)
	t.Run(name, func(t *testing.T) {
		fn(t, client, TestOpts{
			Protocol:   protocol,
			Addr:       *fliptAddr,
			Namespace:  namespace,
			AuthConfig: authConfig,
			References: *fliptReferences,
		})
	})
}
