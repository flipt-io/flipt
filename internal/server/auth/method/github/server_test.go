package github

import (
	"context"
	"encoding/json"
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
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
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

func Test_Server(t *testing.T) {
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

	gock.New("https://api.github.com").
		MatchHeader("Authorization", "Bearer AccessToken").
		MatchHeader("Accept", "application/vnd.github+json").
		Get("/user").
		Reply(200).
		JSON(map[string]any{"name": "fliptuser", "email": "user@flipt.io", "avatar_url": "https://thispicture.com", "id": 1234567890})

	c, err := client.Callback(ctx, &auth.CallbackRequest{Code: "github_code"})
	require.NoError(t, err)

	assert.NotEmpty(t, c.ClientToken)
	assert.Equal(t, auth.Method_METHOD_GITHUB, c.Authentication.Method)
	assert.Equal(t, map[string]string{
		storageMetadataGithubEmail:   "user@flipt.io",
		storageMetadataGithubName:    "fliptuser",
		storageMetadataGithubPicture: "https://thispicture.com",
		storageMetadataGithubSub:     "1234567890",
	}, c.Authentication.Metadata)

	gock.Off()

	gock.New("https://api.github.com").
		MatchHeader("Authorization", "Bearer AccessToken").
		MatchHeader("Accept", "application/vnd.github+json").
		Get("/user").
		Reply(400)

	_, err = client.Callback(ctx, &auth.CallbackRequest{Code: "github_code"})
	assert.EqualError(t, err, "rpc error: code = Internal desc = github /user info response status: \"400 Bad Request\"")

	gock.Off()

	// check allowed organizations successfully
	s.config.Methods.Github.Method.AllowedOrganizations = []string{"flipt-io"}
	gock.New("https://api.github.com").
		MatchHeader("Authorization", "Bearer AccessToken").
		MatchHeader("Accept", "application/vnd.github+json").
		Get("/user").
		Reply(200).
		JSON(map[string]any{"name": "fliptuser", "email": "user@flipt.io", "avatar_url": "https://thispicture.com", "id": 1234567890})

	gock.New("https://api.github.com").
		MatchHeader("Authorization", "Bearer AccessToken").
		MatchHeader("Accept", "application/vnd.github+json").
		Get("/user/orgs").
		Reply(200).
		JSON([]githubSimpleOrganization{{Login: "flipt-io"}})

	c, err = client.Callback(ctx, &auth.CallbackRequest{Code: "github_code"})
	require.NoError(t, err)
	assert.NotEmpty(t, c.ClientToken)
	gock.Off()

	// check allowed organizations  unsuccessfully
	s.config.Methods.Github.Method.AllowedOrganizations = []string{"flipt-io"}
	gock.New("https://api.github.com").
		MatchHeader("Authorization", "Bearer AccessToken").
		MatchHeader("Accept", "application/vnd.github+json").
		Get("/user").
		Reply(200).
		JSON(map[string]any{"name": "fliptuser", "email": "user@flipt.io", "avatar_url": "https://thispicture.com", "id": 1234567890})

	gock.New("https://api.github.com").
		MatchHeader("Authorization", "Bearer AccessToken").
		MatchHeader("Accept", "application/vnd.github+json").
		Get("/user/orgs").
		Reply(200).
		JSON([]githubSimpleOrganization{{Login: "github"}})

	_, err = client.Callback(ctx, &auth.CallbackRequest{Code: "github_code"})
	require.ErrorIs(t, err, status.Error(codes.Unauthenticated, "request was not authenticated"))
	gock.Off()
}

func Test_Server_SkipsAuthentication(t *testing.T) {
	server := &Server{}
	assert.True(t, server.SkipsAuthentication(context.Background()))
}

func TestCallbackURL(t *testing.T) {
	callback := callbackURL("https://flipt.io")
	assert.Equal(t, callback, "https://flipt.io/auth/v1/method/github/callback")

	callback = callbackURL("https://flipt.io/")
	assert.Equal(t, callback, "https://flipt.io/auth/v1/method/github/callback")
}

func TestGithubSimpleOrganizationDecode(t *testing.T) {
	var body = `[{
	"login": "github",
	"id": 1,
	"node_id": "MDEyOk9yZ2FuaXphdGlvbjE=",
	"url": "https://api.github.com/orgs/github",
	"repos_url": "https://api.github.com/orgs/github/repos",
	"events_url": "https://api.github.com/orgs/github/events",
	"hooks_url": "https://api.github.com/orgs/github/hooks",
	"issues_url": "https://api.github.com/orgs/github/issues",
	"members_url": "https://api.github.com/orgs/github/members{/member}",
	"public_members_url": "https://api.github.com/orgs/github/public_members{/member}",
	"avatar_url": "https://github.com/images/error/octocat_happy.gif",
	"description": "A great organization"
	}]`

	var githubUserOrgsResponse []githubSimpleOrganization
	err := json.Unmarshal([]byte(body), &githubUserOrgsResponse)
	require.NoError(t, err)
	require.Len(t, githubUserOrgsResponse, 1)
	require.Equal(t, "github", githubUserOrgsResponse[0].Login)
}
