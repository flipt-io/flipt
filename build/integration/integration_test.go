package integration

import (
	"context"
	"flag"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.flipt.io/flipt/rpc/flipt"
	"go.flipt.io/flipt/sdk"
	sdkgrpc "go.flipt.io/flipt/sdk/grpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

var fliptAddr = flag.String("flipt-addr", "localhost:9000", "Address for running Flipt instance (gRPC only)")

func TestFlipt(t *testing.T) {
	conn, err := grpc.Dial(*fliptAddr,
		grpc.WithTransportCredentials(insecure.NewCredentials()))
	require.NoError(t, err)

	transport := sdkgrpc.NewTransport(conn)
	client := sdk.New(transport)

	ctx := context.Background()

	created, err := client.Flipt().CreateFlag(ctx, &flipt.CreateFlagRequest{
		Key:         "test",
		Name:        "Test",
		Description: "This is a test flag",
	})
	require.NoError(t, err)

	assert.Equal(t, "test", created.Key)
	assert.Equal(t, "Test", created.Name)
	assert.Equal(t, "This is a test flag", created.Description)

	flag, err := client.Flipt().GetFlag(ctx, &flipt.GetFlagRequest{Key: "test"})
	require.NoError(t, err)

	assert.Equal(t, created, flag)
}
