package integration_test

import (
	"flag"
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
	"go.flipt.io/flipt/build/integration"
	sdk "go.flipt.io/flipt/sdk/go"
	sdkgrpc "go.flipt.io/flipt/sdk/go/grpc"
	sdkhttp "go.flipt.io/flipt/sdk/go/http"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

var (
	fliptAddr  = flag.String("flipt-addr", "grpc://localhost:9000", "Address for running Flipt instance (gRPC only)")
	fliptToken = flag.String("flipt-token", "", "Authentication token to be used during test suite")
)

func TestFlipt(t *testing.T) {
	var transport sdk.Transport
	protocol, addr, _ := strings.Cut(*fliptAddr, "://")
	switch protocol {
	case "grpc":
		conn, err := grpc.Dial(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
		require.NoError(t, err)

		transport = sdkgrpc.NewTransport(conn)
	case "http", "https":
		transport = sdkhttp.NewTransport(*fliptAddr)
	default:
		t.Fatalf("Unexpected flipt address protocol %q", *fliptAddr)
	}

	var (
		name = fmt.Sprintf("(%s) No Authentication", protocol)
		opts []sdk.Option
	)

	if *fliptToken != "" {
		name = fmt.Sprintf("(%s) With Static Token Authentication", protocol)
		opts = append(opts, sdk.WithClientTokenProvider(
			sdk.StaticClientTokenProvider(*fliptToken),
		))
	}

	t.Run(name, func(t *testing.T) {
		fn := func(t *testing.T) sdk.SDK {
			return sdk.New(transport, opts...)
		}

		integration.Core(t, fn)

		// run extra tests in authenticated context
		if *fliptToken != "" {
			integration.Authenticated(t, fn)
		}
	})
}
