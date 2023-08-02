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
	fliptAddr      = flag.String("flipt-addr", "grpc://localhost:9000", "Address for running Flipt instance (gRPC only)")
	fliptToken     = flag.String("flipt-token", "", "Authentication token to be used during test suite")
	fliptNamespace = flag.String("flipt-namespace", "", "Namespace used to scope API calls.")
)

func Harness(t *testing.T, fn func(t *testing.T, sdk sdk.SDK, ns string, authenticated bool)) {
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
	)

	if authentication = *fliptToken != ""; authentication {
		opts = append(opts, sdk.WithClientTokenProvider(
			sdk.StaticClientTokenProvider(*fliptToken),
		))
	}

	namespace := "default"
	if *fliptNamespace != "" {
		namespace = *fliptNamespace
	}

	name := fmt.Sprintf("[Protocol %q; Namespace %q; Authentication %v]", protocol, namespace, authentication)
	t.Run(name, func(t *testing.T) {
		fn(t, sdk.New(transport, opts...), namespace, authentication)
	})
}
