package github

import (
	"context"
	"net"
	"net/http"
	"net/url"
	"testing"
	"time"

	grpc_middleware "github.com/grpc-ecosystem/go-grpc-middleware"
	"github.com/h2non/gock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.flipt.io/flipt/internal/config"
	middleware "go.flipt.io/flipt/internal/server/middleware/grpc"
	"go.flipt.io/flipt/internal/storage/auth/memory"
	"go.flipt.io/flipt/rpc/flipt/auth"
	"go.uber.org/zap/zaptest"
	"golang.org/x/oauth2"
	"google.golang.org/grpc"
	"google.golang.org/grpc/test/bufconn"
)

type OAuth2Mock struct{}

func (o *OAuth2Mock) AuthCodeURL(state string, opts ...oauth2.AuthCodeOption) string {
	u := url.URL{
		Scheme: "http",
		Host:   "github.com",
		Path:   "login/oauth/authorize",
	}

	v := url.Values{}
	v.Set("state", state)

	u.RawQuery = v.Encode()
	return u.String()
}

func (o *OAuth2Mock) Exchange(ctx context.Context, code string, opts ...oauth2.AuthCodeOption) (*oauth2.Token, error) {
	return &oauth2.Token{
		AccessToken: "AccessToken",
		Expiry:      time.Now().Add(1 * time.Hour),
	}, nil
}

func (o *OAuth2Mock) Client(ctx context.Context, t *oauth2.Token) *http.Client {
	return &http.Client{}
}

func TestServer_Github(t *testing.T) {
	var (
		logger   = zaptest.NewLogger(t)
		store    = memory.NewStore()
		listener = bufconn.Listen(1024 * 1024)
		server   = grpc.NewServer(
			grpc_middleware.WithUnaryServerChain(
				middleware.ErrorUnaryInterceptor,
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

	s := &Server{
		logger: logger,
		store:  store,
		config: config.AuthenticationConfig{
			Methods: config.AuthenticationMethods{
				Github: config.AuthenticationMethod[config.AuthenticationMethodGithubConfig]{
					Enabled: true,
					Method: config.AuthenticationMethodGithubConfig{
						ClientSecret:    "topsecret",
						ClientId:        "githubid",
						RedirectAddress: "test.flipt.io",
						Scopes:          []string{"user", "email"},
					},
				},
			},
		},
		oauth2Config: &OAuth2Mock{},
	}

	auth.RegisterAuthenticationMethodGithubServiceServer(server, s)

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

	client := auth.NewAuthenticationMethodGithubServiceClient(conn)

	a, err := client.AuthorizeURL(ctx, &auth.AuthorizeURLRequest{
		Provider: "github",
		State:    "random-state",
	})
	require.NoError(t, err)

	assert.Equal(t, "http://github.com/login/oauth/authorize?state=random-state", a.AuthorizeUrl)

	defer gock.Off()

	gock.New("https://api.github.com").
		MatchHeader("Authorization", "Bearer AccessToken").
		MatchHeader("Accept", "application/vnd.github+json").
		Get("/user").
		Reply(200).
		JSON(map[string]string{"name": "fliptuser", "email": "user@flipt.io"})

	c, err := client.Callback(ctx, &auth.CallbackRequest{Code: "github_code"})
	require.NoError(t, err)

	assert.NotEmpty(t, c.ClientToken)
	assert.Equal(t, auth.Method_METHOD_GITHUB, c.Authentication.Method)
	assert.Equal(t, map[string]string{
		storageMetadataGithubEmail: "user@flipt.io",
		storageMetadataGithubName:  "fliptuser",
	}, c.Authentication.Metadata)
}
