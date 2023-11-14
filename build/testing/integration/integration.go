package integration

import (
	"context"
	"flag"
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
	"go.flipt.io/flipt/rpc/flipt"
	"go.flipt.io/flipt/rpc/flipt/auth"
	sdk "go.flipt.io/flipt/sdk/go"
	sdkgrpc "go.flipt.io/flipt/sdk/go/grpc"
	sdkhttp "go.flipt.io/flipt/sdk/go/http"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

var (
	fliptAddr                  = flag.String("flipt-addr", "grpc://localhost:9000", "Address for running Flipt instance (gRPC only)")
	fliptToken                 = flag.String("flipt-token", "", "Authentication token to be used during test suite")
	fliptCreateNamespacedToken = flag.Bool("flipt-create-namespaced-token", false, "Create a namespaced token for the test suite")
	fliptNamespace             = flag.String("flipt-namespace", "", "Namespace used to scope API calls.")
)

type AuthConfig int

const (
	NoAuth AuthConfig = iota
	AuthNoNamespace
	AuthNamespaced
)

func (a AuthConfig) String() string {
	switch a {
	case NoAuth:
		return "NoAuth"
	case AuthNoNamespace:
		return "AuthNoNamespace"
	case AuthNamespaced:
		return "AuthNamespaced"
	default:
		return "Unknown"
	}
}

func (a AuthConfig) Required() bool {
	return a != NoAuth
}

type TestOpts struct {
	Addr       string
	Protocol   string
	Namespace  string
	AuthConfig AuthConfig
}

func Harness(t *testing.T, fn func(t *testing.T, sdk sdk.SDK, opts TestOpts)) {
	var transport sdk.Transport

	protocol, host, _ := strings.Cut(*fliptAddr, "://")
	switch protocol {
	case "grpc":
		conn, err := grpc.Dial(host, grpc.WithTransportCredentials(insecure.NewCredentials()))
		require.NoError(t, err)

		transport = sdkgrpc.NewTransport(conn)
	case "http", "https":
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

	if authentication := *fliptToken != ""; authentication {
		authConfig = AuthNoNamespace

		opts = append(opts, sdk.WithClientTokenProvider(
			sdk.StaticClientTokenProvider(*fliptToken),
		))

		client = sdk.New(transport, opts...)

		if *fliptCreateNamespacedToken {
			authConfig = AuthNamespaced

			t.Log("Creating namespaced token for test suite")

			ctx := context.Background()
			authn, err := client.Auth().AuthenticationMethodTokenService().CreateToken(
				ctx,
				&auth.CreateTokenRequest{
					Name:         "Integration Test Token",
					NamespaceKey: namespace,
				},
			)

			t.Log("Created token", authn.ClientToken, authn.Authentication.Metadata)

			require.NoError(t, err)

			t.Log("Creating namespace for test suite")

			if *fliptNamespace != "" {
				_, err := client.Flipt().CreateNamespace(ctx, &flipt.CreateNamespaceRequest{
					Key:  namespace,
					Name: namespace,
				})
				require.NoError(t, err)
			}

			require.NoError(t, err)

			opts = append(opts, sdk.WithClientTokenProvider(
				sdk.StaticClientTokenProvider(authn.ClientToken),
			))
		}
	}

	client = sdk.New(transport, opts...)

	name := fmt.Sprintf("[Protocol %q; Namespace %q; Authentication %s]", protocol, namespace, authConfig)
	t.Run(name, func(t *testing.T) {
		fn(t, client, TestOpts{
			Protocol:   protocol,
			Addr:       *fliptAddr,
			Namespace:  namespace,
			AuthConfig: authConfig,
		})
	})
}
