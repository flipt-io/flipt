package authn_test

import (
	"context"
	"net"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.flipt.io/flipt/errors"
	"go.flipt.io/flipt/internal/server/authn"
	authmiddlewaregrpc "go.flipt.io/flipt/internal/server/authn/middleware/grpc"
	middlewaregrpc "go.flipt.io/flipt/internal/server/middleware/grpc"
	storageauth "go.flipt.io/flipt/internal/storage/authn"
	"go.flipt.io/flipt/internal/storage/authn/memory"
	rpcauth "go.flipt.io/flipt/rpc/flipt/auth"
	"go.uber.org/zap/zaptest"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/test/bufconn"
	"google.golang.org/protobuf/testing/protocmp"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func TestActorFromContext(t *testing.T) {
	const (
		ipAddress      = "127.0.0.1"
		authentication = "token"
	)

	ctx := metadata.NewIncomingContext(context.Background(), map[string][]string{"x-forwarded-for": {"127.0.0.1"}})
	ctx = authmiddlewaregrpc.ContextWithAuthentication(ctx, &rpcauth.Authentication{Method: rpcauth.Method_METHOD_TOKEN})

	actor := authn.ActorFromContext(ctx)

	assert.Equal(t, ipAddress, actor.IP)
	assert.Equal(t, authentication, actor.Authentication)
}

func TestServer(t *testing.T) {
	var (
		logger   = zaptest.NewLogger(t)
		store    = memory.NewStore()
		listener = bufconn.Listen(1024 * 1024)
		server   = grpc.NewServer(
			grpc.ChainUnaryInterceptor(
				authmiddlewaregrpc.ClientTokenAuthenticationInterceptor(logger, store),
				middlewaregrpc.ErrorUnaryInterceptor,
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

	rpcauth.RegisterAuthenticationServiceServer(server, authn.NewServer(logger, store, authn.WithAuditLoggingEnabled(true)))

	go func() {
		errC <- server.Serve(listener)
	}()

	var (
		ctx    = context.Background()
		dialer = func(context.Context, string) (net.Conn, error) {
			return listener.Dial()
		}
	)

	req := &storageauth.CreateAuthenticationRequest{
		Method:    rpcauth.Method_METHOD_TOKEN,
		ExpiresAt: timestamppb.New(time.Now().Add(time.Hour).UTC()),
	}

	clientToken, authentication, err := store.CreateAuthentication(ctx, req)
	require.NoError(t, err)

	conn, err := grpc.DialContext(ctx, "", grpc.WithInsecure(), grpc.WithContextDialer(dialer))
	require.NoError(t, err)
	defer conn.Close()

	client := rpcauth.NewAuthenticationServiceClient(conn)

	authorize := func(context.Context) context.Context {
		return metadata.AppendToOutgoingContext(
			ctx,
			"authorization",
			"Bearer "+clientToken,
		)
	}

	t.Run("GetAuthenticationSelf", func(t *testing.T) {
		_, err := client.GetAuthenticationSelf(ctx, &emptypb.Empty{})
		require.ErrorContains(t, err, "request was not authenticated")

		retrievedAuth, err := client.GetAuthenticationSelf(authorize(ctx), &emptypb.Empty{})
		require.NoError(t, err)

		if diff := cmp.Diff(retrievedAuth, authentication, protocmp.Transform()); err != nil {
			t.Errorf("-exp/+got:\n%s", diff)
		}
	})

	t.Run("GetAuthentication", func(t *testing.T) {
		retrievedAuth, err := client.GetAuthentication(authorize(ctx), &rpcauth.GetAuthenticationRequest{
			Id: authentication.Id,
		})
		require.NoError(t, err)

		if diff := cmp.Diff(retrievedAuth, authentication, protocmp.Transform()); err != nil {
			t.Errorf("-exp/+got:\n%s", diff)
		}
	})

	t.Run("ListAuthentications", func(t *testing.T) {
		expected := &rpcauth.ListAuthenticationsResponse{
			Authentications: []*rpcauth.Authentication{
				authentication,
			},
		}

		response, err := client.ListAuthentications(authorize(ctx), &rpcauth.ListAuthenticationsRequest{})
		require.NoError(t, err)

		if diff := cmp.Diff(response, expected, protocmp.Transform()); err != nil {
			t.Errorf("-exp/+got:\n%s", diff)
		}

		// by method token
		response, err = client.ListAuthentications(authorize(ctx), &rpcauth.ListAuthenticationsRequest{
			Method: rpcauth.Method_METHOD_TOKEN,
		})
		require.NoError(t, err)

		if diff := cmp.Diff(response, expected, protocmp.Transform()); err != nil {
			t.Errorf("-exp/+got:\n%s", diff)
		}
	})

	t.Run("DeleteAuthentication", func(t *testing.T) {
		ctx := authorize(ctx)
		// delete self
		_, err := client.DeleteAuthentication(ctx, &rpcauth.DeleteAuthenticationRequest{
			Id: authentication.Id,
		})
		require.NoError(t, err)

		// get self with authenticated context now unauthorized
		_, err = client.GetAuthenticationSelf(ctx, &emptypb.Empty{})

		require.ErrorContains(t, err, "request was not authenticated")

		// no longer can be retrieved from store by client ID
		_, err = store.GetAuthenticationByClientToken(ctx, clientToken)
		var notFound errors.ErrNotFound
		require.ErrorAs(t, err, &notFound)
	})

	t.Run("ExpireAuthenticationSelf", func(t *testing.T) {
		// create new authentication
		req := &storageauth.CreateAuthenticationRequest{
			Method:    rpcauth.Method_METHOD_TOKEN,
			ExpiresAt: timestamppb.New(time.Now().Add(time.Hour).UTC()),
		}

		ctx := context.TODO()

		clientToken, _, err := store.CreateAuthentication(ctx, req)
		require.NoError(t, err)

		ctx = metadata.AppendToOutgoingContext(
			ctx,
			"authorization",
			"Bearer "+clientToken,
		)

		// get self with authenticated context not unauthorized
		_, err = client.GetAuthenticationSelf(ctx, &emptypb.Empty{})
		require.NoError(t, err)

		// expire self
		_, err = client.ExpireAuthenticationSelf(ctx, &rpcauth.ExpireAuthenticationSelfRequest{})
		require.NoError(t, err)

		// get self with authenticated context now unauthorized
		_, err = client.GetAuthenticationSelf(ctx, &emptypb.Empty{})
		require.ErrorContains(t, err, "request was not authenticated")
	})
}

func Test_Server_DisallowsNamespaceScopedAuthentication(t *testing.T) {
	var (
		logger = zaptest.NewLogger(t)
		store  = memory.NewStore()
		server = authn.NewServer(logger, store)
	)

	assert.False(t, server.AllowsNamespaceScopedAuthentication(context.Background()))
}
