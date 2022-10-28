package token

import (
	"context"
	"net"
	"testing"

	"github.com/google/go-cmp/cmp"
	grpc_middleware "github.com/grpc-ecosystem/go-grpc-middleware"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.flipt.io/flipt/internal/server"
	"go.flipt.io/flipt/internal/storage/auth/memory"
	"go.flipt.io/flipt/rpc/flipt/auth"
	"go.uber.org/zap/zaptest"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/grpc/test/bufconn"
	"google.golang.org/protobuf/testing/protocmp"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func TestServer(t *testing.T) {
	var (
		logger   = zaptest.NewLogger(t)
		store    = memory.NewStore()
		listener = bufconn.Listen(1024 * 1024)
		server   = grpc.NewServer(
			grpc_middleware.WithUnaryServerChain(
				server.ErrorUnaryInterceptor,
			),
		)
		errC     = make(chan error)
		shutdown = func(t *testing.T) {
			t.Helper()

			server.Stop()
			if err := <-errC; err != nil {
				t.Fatal(err)
			}
		}
	)

	defer shutdown(t)

	auth.RegisterAuthenticationMethodTokenServiceServer(server, NewServer(logger, store))

	go func() {
		errC <- server.Serve(listener)
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
	retrieved, err := store.GetAuthenticationByClientToken(ctx, resp.ClientToken)
	require.NoError(t, err)

	// switch to go-cmp here to do the comparisons since assert trips up
	// on the unexported sizeCache values.
	if diff := cmp.Diff(retrieved, resp.Authentication, protocmp.Transform()); err != nil {
		t.Errorf("-exp/+got:\n%s", diff)
	}

	// attempt to create token with invalid expires at
	_, err = client.CreateToken(ctx, &auth.CreateTokenRequest{
		Name:        "access_all_areas",
		Description: "Super secret skeleton key",
		// invalid expires at, nanos must be positive
		ExpiresAt: &timestamppb.Timestamp{Nanos: -1},
	})
	require.ErrorIs(t, err, status.Error(codes.InvalidArgument, "attempting to create token: invalid expiry time: nanos:-1"))
}
