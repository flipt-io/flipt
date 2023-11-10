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
	fliptToken                 = flag.String("flipt-token", "", "Authentication token to be used during test suite")
	fliptCreateNamespacedToken = flag.Bool("flipt-create-namespaced-token", false, "Create a namespaced token for the test suite")
	fliptNamespace             = flag.String("flipt-namespace", "", "Namespace used to scope API calls.")
)

type TestOpts struct {
	Addr          string
	Protocol      string
	Namespace     string
	Authenticated bool
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
		opts           []sdk.Option
		authentication bool
		client         sdk.SDK
	)

	if authentication = *fliptToken != ""; authentication {
		opts = append(opts, sdk.WithClientTokenProvider(
			sdk.StaticClientTokenProvider(*fliptToken),
		))

		client = sdk.New(transport, opts...)

		if *fliptCreateNamespacedToken {
			t.Log("Creating namespaced token for test suite")

			authn, err := sdk.New(transport).Auth().AuthenticationMethodTokenService().CreateToken(
				context.Background(),
				&auth.CreateTokenRequest{
					Name:         "Integration Test Token",
					NamespaceKey: *fliptNamespace,
				},
			)

			require.NoError(t, err)

			opts = append(opts, sdk.WithClientTokenProvider(
				sdk.StaticClientTokenProvider(authn.ClientToken),
			))

			// re-initialize client with new token
			client = sdk.New(transport, opts...)
		}
	}

	namespace := "default"
	if *fliptNamespace != "" {
		namespace = *fliptNamespace
	}

	name := fmt.Sprintf("[Protocol %q; Namespace %q; Authentication %v]", protocol, namespace, authentication)
	t.Run(name, func(t *testing.T) {
		fn(t, client, TestOpts{
			Protocol:      protocol,
			Addr:          *fliptAddr,
			Namespace:     namespace,
			Authenticated: authentication},
		)
	})
}
