package integration_test

import (
	"flag"
	"testing"

	"github.com/stretchr/testify/require"
	"go.flipt.io/flipt/build/integration"
	"go.flipt.io/flipt/sdk"
	sdkgrpc "go.flipt.io/flipt/sdk/grpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

var (
	fliptAddr  = flag.String("flipt-addr", "localhost:9000", "Address for running Flipt instance (gRPC only)")
	fliptToken = flag.String("flipt-token", "", "Authentication token to be used during test suite")
)

func TestFlipt(t *testing.T) {
	var (
		name = "No Authentication"
		opts []sdk.Option
	)

	if *fliptToken != "" {
		name = "With Static Token Authentication"
		opts = append(opts, sdk.WithClientTokenProvider(
			sdk.StaticClientTokenProvider(*fliptToken),
		))
	}

	t.Run(name, func(t *testing.T) {
		integration.Harness(t, func(t *testing.T) sdk.SDK {
			conn, err := grpc.Dial(*fliptAddr,
				grpc.WithTransportCredentials(insecure.NewCredentials()))
			require.NoError(t, err)

			return sdk.New(sdkgrpc.NewTransport(conn), opts...)
		})
	})
}
