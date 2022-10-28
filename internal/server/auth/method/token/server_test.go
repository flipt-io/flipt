package auth

import (
	"context"
	"net"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.flipt.io/flipt/internal/storage/auth/memory"
	"go.flipt.io/flipt/rpc/flipt/auth"
	"go.uber.org/zap/zaptest"
	"google.golang.org/grpc"
	"google.golang.org/grpc/test/bufconn"
	"google.golang.org/protobuf/testing/protocmp"
)

func TestServer(t *testing.T) {
	var (
		logger   = zaptest.NewLogger(t)
		store    = memory.NewStore()
		listener = bufconn.Listen(1024 * 1024)
		server   = grpc.NewServer()
	)

	auth.RegisterAuthenticationMethodTokenServiceServer(server, NewServer(logger, store))

	go func() {
		if err := server.Serve(listener); err != nil {
			t.Fatal(err)
		}
	}()

	var (
		ctx    = context.Background()
		dialer = func(context.Context, string) (net.Conn, error) {
			return listener.Dial()
		}
	)

	conn, err := grpc.DialContext(ctx, "", grpc.WithInsecure(), grpc.WithContextDialer(dialer))
	require.NoError(t, err)
	defer conn.Close()

	client := auth.NewAuthenticationMethodTokenServiceClient(conn)

	// attempt to create token
	resp, err := client.CreateToken(ctx, &auth.CreateTokenRequest{
		Name:        "access_all_areas",
		Description: "Super secret skeleton key",
	})
	require.NoError(t, err)

	// assert auth is as expected
	metadata := resp.Authentication.Metadata
	assert.Equal(t, "access_all_areas", metadata["io.flipt.auth.token.name"])
	assert.Equal(t, "Super secret skeleton key", metadata["io.flipt.auth.token.description"])

	// ensure client token can be used on store to fetch authentication
	// and that the authentication returned matches the one received
	// by the client
	auth, err := store.GetAuthenticationByClientToken(ctx, resp.ClientToken)
	require.NoError(t, err)

	// switch to go-cmp here to do the comparisons since assert trips up
	// on the unexported sizeCache values.
	if diff := cmp.Diff(auth, resp.Authentication, protocmp.Transform()); err != nil {
		t.Errorf("-exp/+got:\n%s", diff)
	}
}
