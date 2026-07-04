package authn_test

import (
	"context"
	"encoding/json"
	"net"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.flipt.io/flipt/errors"
	"go.flipt.io/flipt/internal/config"
	"go.flipt.io/flipt/internal/server/authn"
	authoidc "go.flipt.io/flipt/internal/server/authn/method/oidc"
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

	ctx := metadata.NewIncomingContext(t.Context(), map[string][]string{"x-forwarded-for": {"127.0.0.1"}})
	ctx = authmiddlewaregrpc.ContextWithAuthentication(ctx, &rpcauth.Authentication{Method: rpcauth.Method_METHOD_TOKEN})

	actor := authn.ActorFromContext(ctx)

	assert.Equal(t, ipAddress, actor.Ip)
	assert.Equal(t, authentication, actor.Authentication)
}

func TestServer(t *testing.T) {
	var (
		logger   = zaptest.NewLogger(t)
		store    = memory.NewStore(logger)
		listener = bufconn.Listen(1024 * 1024)
		server   = grpc.NewServer(
			grpc.ChainUnaryInterceptor(
				authmiddlewaregrpc.ClientTokenAuthenticationUnaryInterceptor(logger, store),
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

	rpcauth.RegisterAuthenticationServiceServer(server, authn.NewServer(logger, store, config.AuthenticationConfig{}, nil))

	go func() {
		errC <- server.Serve(listener)
	}()

	var (
		ctx    = t.Context()
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

	t.Run("RevokeAuthenticationSelf", func(t *testing.T) {
		t.Run("unauthenticated", func(t *testing.T) {
			_, err := client.RevokeAuthenticationSelf(ctx, &rpcauth.RevokeAuthenticationSelfRequest{})
			require.ErrorContains(t, err, "request was not authenticated")
		})

		t.Run("token method (no next_uri)", func(t *testing.T) {
			req := &storageauth.CreateAuthenticationRequest{
				Method:    rpcauth.Method_METHOD_TOKEN,
				ExpiresAt: timestamppb.New(time.Now().Add(time.Hour).UTC()),
			}

			clientToken, _, err := store.CreateAuthentication(ctx, req)
			require.NoError(t, err)

			authCtx := metadata.AppendToOutgoingContext(
				ctx,
				"authorization",
				"Bearer "+clientToken,
			)

			resp, err := client.RevokeAuthenticationSelf(authCtx, &rpcauth.RevokeAuthenticationSelfRequest{})
			require.NoError(t, err)
			require.NotNil(t, resp)
			assert.Nil(t, resp.NextUri)
		})

		t.Run("OIDC method with IDToken (next_uri present)", func(t *testing.T) {
			// Start an OIDC-like test server for the end_session_endpoint
			var oidcSrv *httptest.Server
			oidcSrv = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				host := strings.TrimPrefix(oidcSrv.URL, "http://")
				w.Header().Set("Content-Type", "application/json")
				_ = json.NewEncoder(w).Encode(map[string]string{
					"issuer":                 "http://" + host,
					"authorization_endpoint": "http://" + host + "/auth",
					"token_endpoint":         "http://" + host + "/token",
					"jwks_uri":               "http://" + host + "/certs",
					"end_session_endpoint":   "http://" + host + "/logout",
				})
			}))
			t.Cleanup(oidcSrv.Close)

			// Create a new server and client with OIDC auth config
			oidcStore := memory.NewStore(logger)
			oidcServer := grpc.NewServer(
				grpc.ChainUnaryInterceptor(
					authmiddlewaregrpc.ClientTokenAuthenticationUnaryInterceptor(logger, oidcStore),
					middlewaregrpc.ErrorUnaryInterceptor,
				),
			)
			oidcListener := bufconn.Listen(1024 * 1024)

			oidcAuthConfig := config.AuthenticationConfig{
				Methods: config.AuthenticationMethodsConfig{
					OIDC: config.AuthenticationMethod[config.AuthenticationMethodOIDCConfig]{
						Enabled: true,
						Method: config.AuthenticationMethodOIDCConfig{
							Providers: map[string]config.AuthenticationMethodOIDCProvider{
								"google": {
									IssuerURL:             oidcSrv.URL,
									ClientID:              "test-client",
									ClientSecret:          "test-secret",
									RedirectAddress:       "http://localhost:8080",
									UseEndSessionEndpoint: true,
								},
							},
						},
					},
				},
			}

			rpcauth.RegisterAuthenticationServiceServer(oidcServer, authn.NewServer(logger, oidcStore, oidcAuthConfig, authoidc.NewRegistry(oidcAuthConfig)))

			oidcErrC := make(chan error)
			go func() {
				oidcErrC <- oidcServer.Serve(oidcListener)
			}()
			t.Cleanup(func() {
				oidcServer.Stop()
				<-oidcErrC
			})

			oidcDialer := func(context.Context, string) (net.Conn, error) {
				return oidcListener.Dial()
			}

			oidcConn, err := grpc.DialContext(ctx, "", grpc.WithInsecure(), grpc.WithContextDialer(oidcDialer))
			require.NoError(t, err)
			t.Cleanup(func() { oidcConn.Close() })

			oidcClient := rpcauth.NewAuthenticationServiceClient(oidcConn)

			// Create OIDC authentication with IDToken
			oidcToken, _, err := oidcStore.CreateAuthentication(ctx, &storageauth.CreateAuthenticationRequest{
				Method:    rpcauth.Method_METHOD_OIDC,
				ExpiresAt: timestamppb.New(time.Now().Add(time.Hour).UTC()),
				Metadata: map[string]string{
					"io.flipt.auth.oidc.provider": "google",
				},
				IDToken: "test-id-token",
			})
			require.NoError(t, err)

			authCtx := metadata.AppendToOutgoingContext(
				ctx,
				"authorization",
				"Bearer "+oidcToken,
			)

			resp, err := oidcClient.RevokeAuthenticationSelf(authCtx, &rpcauth.RevokeAuthenticationSelfRequest{})
			require.NoError(t, err)
			require.NotNil(t, resp)

			// Should have next_uri pointing to the end_session_endpoint
			require.NotNil(t, resp.NextUri)
			assert.Contains(t, *resp.NextUri, oidcSrv.URL+"/logout")
			assert.Contains(t, *resp.NextUri, "id_token_hint=test-id-token")
			assert.Contains(t, *resp.NextUri, "post_logout_redirect_uri=http%3A%2F%2Flocalhost%3A8080")
		})
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
