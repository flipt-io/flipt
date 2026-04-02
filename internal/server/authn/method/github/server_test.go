package github

import (
	"context"
	"encoding/json"
	"net"
	"net/http"
	"net/url"
	"sync"
	"testing"
	"time"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"

	"github.com/h2non/gock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.flipt.io/flipt/errors"
	"go.flipt.io/flipt/internal/config"
	middleware "go.flipt.io/flipt/internal/server/middleware/grpc"
	storageauth "go.flipt.io/flipt/internal/storage/authn"
	"go.flipt.io/flipt/internal/storage/authn/memory"
	"go.flipt.io/flipt/rpc/flipt/auth"
	"go.uber.org/zap/zaptest"
	"golang.org/x/oauth2"
	"google.golang.org/grpc"
	"google.golang.org/grpc/test/bufconn"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type OAuth2Mock struct{}

func (o *OAuth2Mock) AuthCodeURL(state string, opts ...oauth2.AuthCodeOption) string {
	return (&oauth2.Config{
		Endpoint: oauth2.Endpoint{
			AuthURL: "http://github.com/login/oauth/authorize",
		},
	}).AuthCodeURL(state, opts...)
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

type testGithubClient struct {
	auth.AuthenticationMethodGithubServiceClient
}

func (c *testGithubClient) Callback(ctx context.Context, req *auth.CallbackRequest, opts ...grpc.CallOption) (*auth.CallbackResponse, error) {
	if req.State == "" {
		_, err := c.AuthorizeURL(ctx, &auth.AuthorizeURLRequest{
			Provider: "github",
			State:    "random-state",
		})
		if err != nil {
			return nil, err
		}

		ctx = metadata.NewOutgoingContext(ctx, metadata.Pairs("flipt_client_state", "random-state"))
		req.State = "random-state"
	}

	return c.AuthenticationMethodGithubServiceClient.Callback(ctx, req, opts...)
}

func Test_Server(t *testing.T) {
	t.Run("should return the authorize url correctly", func(t *testing.T) {
		ctx := t.Context()
		client := newTestServer(t, config.AuthenticationMethod[config.AuthenticationMethodGithubConfig]{
			Enabled: true,
			Method: config.AuthenticationMethodGithubConfig{
				ClientSecret:    "topsecret",
				ClientId:        "githubid",
				RedirectAddress: "test.flipt.io",
				Scopes:          []string{"user", "email"},
				UsePKCE:         true,
			},
		})

		resp, err := client.AuthorizeURL(ctx, &auth.AuthorizeURLRequest{
			Provider: "github",
			State:    "random-state",
		})

		require.NoError(t, err)

		u, err := url.Parse(resp.AuthorizeUrl)
		require.NoError(t, err)
		q := u.Query()
		assert.Equal(t, "random-state", q.Get("state"))
		assert.Equal(t, "S256", q.Get("code_challenge_method"))
		assert.NotEmpty(t, q.Get("code_challenge"))
	})

	t.Run("should return an error when authorize state is missing", func(t *testing.T) {
		ctx := t.Context()
		client := newTestServer(t, config.AuthenticationMethod[config.AuthenticationMethodGithubConfig]{
			Enabled: true,
			Method: config.AuthenticationMethodGithubConfig{
				ClientSecret:    "topsecret",
				ClientId:        "githubid",
				RedirectAddress: "test.flipt.io",
				Scopes:          []string{"user", "email"},
				UsePKCE:         true,
			},
		})

		_, err := client.AuthorizeURL(ctx, &auth.AuthorizeURLRequest{
			Provider: "github",
		})
		require.Error(t, err)
		assert.ErrorContains(t, err, "missing state parameter")
	})

	t.Run("should round trip the verifier through the oauth challenge store", func(t *testing.T) {
		ctx := t.Context()
		store := newChallengeCaptureStore(memory.NewStore(zaptest.NewLogger(t)))
		client := newTestServer(t, config.AuthenticationMethod[config.AuthenticationMethodGithubConfig]{
			Enabled: true,
			Method: config.AuthenticationMethodGithubConfig{
				ClientSecret:    "topsecret",
				ClientId:        "githubid",
				RedirectAddress: "test.flipt.io",
				Scopes:          []string{"user", "email"},
				UsePKCE:         true,
			},
		}, store)

		resp, err := client.AuthorizeURL(ctx, &auth.AuthorizeURLRequest{
			Provider: "github",
			State:    "random-state",
		})
		require.NoError(t, err)

		u, err := url.Parse(resp.AuthorizeUrl)
		require.NoError(t, err)
		q := u.Query()
		verifier := store.putChallenge("random-state")
		require.NotEmpty(t, verifier)
		assert.Equal(t, oauth2.S256ChallengeFromVerifier(verifier), q.Get("code_challenge"))
		assert.Equal(t, "S256", q.Get("code_challenge_method"))

		defer gock.Off()
		gock.New("https://api.github.com").
			MatchHeader("Authorization", "Bearer AccessToken").
			MatchHeader("Accept", "application/vnd.github+json").
			Get("/user").
			Reply(200).
			JSON(map[string]any{"name": "fliptuser", "email": "user@flipt.io", "avatar_url": "https://thispicture.com", "id": 1234567890})

		callbackCtx := metadata.NewOutgoingContext(ctx, metadata.Pairs("flipt_client_state", "random-state"))
		callback, err := client.Callback(callbackCtx, &auth.CallbackRequest{
			Code:  "github_code",
			State: "random-state",
		})
		require.NoError(t, err)

		require.Equal(t, verifier, store.popChallenge("random-state"))
		assert.NotEmpty(t, callback.ClientToken)
	})

	t.Run("should return unauthenticated when callback state is missing", func(t *testing.T) {
		ctx := t.Context()
		client := newRawTestServer(t, config.AuthenticationMethod[config.AuthenticationMethodGithubConfig]{
			Enabled: true,
			Method: config.AuthenticationMethodGithubConfig{
				ClientSecret:    "topsecret",
				ClientId:        "githubid",
				RedirectAddress: "test.flipt.io",
				Scopes:          []string{"user", "email"},
				UsePKCE:         true,
			},
		})

		_, err := client.Callback(ctx, &auth.CallbackRequest{Code: "github_code"})
		require.Error(t, err)
		assert.ErrorContains(t, err, "missing state parameter")
	})

	t.Run("should return unauthenticated when callback challenge is missing", func(t *testing.T) {
		ctx := metadata.NewOutgoingContext(t.Context(), metadata.Pairs("flipt_client_state", "random-state"))
		client := newRawTestServer(t, config.AuthenticationMethod[config.AuthenticationMethodGithubConfig]{
			Enabled: true,
			Method: config.AuthenticationMethodGithubConfig{
				ClientSecret:    "topsecret",
				ClientId:        "githubid",
				RedirectAddress: "test.flipt.io",
				Scopes:          []string{"user", "email"},
				UsePKCE:         true,
			},
		})

		_, err := client.Callback(ctx, &auth.CallbackRequest{
			Code:  "github_code",
			State: "random-state",
		})
		require.Error(t, err)
		assert.ErrorContains(t, err, "missing or expired oauth challenge")
	})

	t.Run("should return internal error when callback challenge lookup fails unexpectedly", func(t *testing.T) {
		ctx := metadata.NewOutgoingContext(t.Context(), metadata.Pairs("flipt_client_state", "random-state"))
		client := newRawTestServer(t, config.AuthenticationMethod[config.AuthenticationMethodGithubConfig]{
			Enabled: true,
			Method: config.AuthenticationMethodGithubConfig{
				ClientSecret:    "topsecret",
				ClientId:        "githubid",
				RedirectAddress: "test.flipt.io",
				Scopes:          []string{"user", "email"},
				UsePKCE:         true,
			},
		}, &popErrorStore{
			Store: memory.NewStore(zaptest.NewLogger(t)),
			err:   errors.New("boom"),
		})

		_, err := client.Callback(ctx, &auth.CallbackRequest{
			Code:  "github_code",
			State: "random-state",
		})
		require.Error(t, err)
		assert.NotContains(t, err.Error(), "missing or expired oauth challenge")
		assert.ErrorContains(t, err, "boom")
	})

	t.Run("should authorize the user correctly", func(t *testing.T) {
		ctx := t.Context()
		client := newTestServer(t, config.AuthenticationMethod[config.AuthenticationMethodGithubConfig]{
			Enabled: true,
			Method: config.AuthenticationMethodGithubConfig{
				ClientSecret:    "topsecret",
				ClientId:        "githubid",
				RedirectAddress: "test.flipt.io",
				Scopes:          []string{"user", "email"},
			},
		})

		defer gock.Off()
		gock.New("https://api.github.com").
			MatchHeader("Authorization", "Bearer AccessToken").
			MatchHeader("Accept", "application/vnd.github+json").
			Get("/user").
			Reply(200).
			JSON(map[string]any{"name": "fliptuser", "email": "user@flipt.io", "avatar_url": "https://thispicture.com", "id": 1234567890})

		callback, err := client.Callback(ctx, &auth.CallbackRequest{Code: "github_code"})
		require.NoError(t, err)

		assert.NotEmpty(t, callback.ClientToken)
		assert.Equal(t, auth.Method_METHOD_GITHUB, callback.Authentication.Method)
		assert.Equal(t, map[string]string{
			"io.flipt.auth.github.email":   "user@flipt.io",
			"io.flipt.auth.email":          "user@flipt.io",
			"io.flipt.auth.github.name":    "fliptuser",
			"io.flipt.auth.name":           "fliptuser",
			"io.flipt.auth.github.picture": "https://thispicture.com",
			"io.flipt.auth.picture":        "https://thispicture.com",
			"io.flipt.auth.github.sub":     "1234567890",
		}, callback.Authentication.Metadata)
	})

	t.Run("should authorize the user and generate name and email", func(t *testing.T) {
		ctx := t.Context()
		client := newTestServer(t, config.AuthenticationMethod[config.AuthenticationMethodGithubConfig]{
			Enabled: true,
			Method: config.AuthenticationMethodGithubConfig{
				ClientSecret:    "topsecret",
				ClientId:        "githubid",
				RedirectAddress: "test.flipt.io",
			},
		})

		defer gock.Off()
		gock.New("https://api.github.com").
			MatchHeader("Authorization", "Bearer AccessToken").
			MatchHeader("Accept", "application/vnd.github+json").
			Get("/user").
			Reply(200).
			JSON(map[string]any{"name": nil, "email": nil, "avatar_url": "https://thispicture.com", "id": 1234567890, "login": "someuser"})

		callback, err := client.Callback(ctx, &auth.CallbackRequest{Code: "github_code"})
		require.NoError(t, err)

		assert.NotEmpty(t, callback.ClientToken)
		assert.Equal(t, auth.Method_METHOD_GITHUB, callback.Authentication.Method)
		assert.Subset(t, callback.Authentication.Metadata, map[string]string{
			"io.flipt.auth.github.email": "1234567890+someuser@users.noreply.github.com",
			"io.flipt.auth.email":        "1234567890+someuser@users.noreply.github.com",
			"io.flipt.auth.github.name":  "someuser",
			"io.flipt.auth.name":         "someuser",
		})
	})

	t.Run("store all orgs and teams to metadata", func(t *testing.T) {
		ctx := t.Context()
		client := newTestServer(t, config.AuthenticationMethod[config.AuthenticationMethodGithubConfig]{
			Enabled: true,
			Method: config.AuthenticationMethodGithubConfig{
				ClientSecret:    "topsecret",
				ClientId:        "githubid",
				RedirectAddress: "test.flipt.io",
				Scopes:          []string{"read:org"},
			},
		})

		defer gock.Off()
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

		gock.New("https://api.github.com").
			MatchHeader("Authorization", "Bearer AccessToken").
			MatchHeader("Accept", "application/vnd.github+json").
			Get("/user/teams").
			Reply(200).
			JSON([]githubSimpleTeam{
				{
					Slug: "flipt-team",
					Organization: githubSimpleOrganization{
						Login: "flipt-io",
					},
				},
			})

		callback, err := client.Callback(ctx, &auth.CallbackRequest{Code: "github_code"})
		require.NoError(t, err)

		assert.NotEmpty(t, callback.ClientToken)
		assert.Equal(t, auth.Method_METHOD_GITHUB, callback.Authentication.Method)
		assert.Equal(t, map[string]string{
			"io.flipt.auth.github.email":         "user@flipt.io",
			"io.flipt.auth.email":                "user@flipt.io",
			"io.flipt.auth.github.name":          "fliptuser",
			"io.flipt.auth.name":                 "fliptuser",
			"io.flipt.auth.github.picture":       "https://thispicture.com",
			"io.flipt.auth.picture":              "https://thispicture.com",
			"io.flipt.auth.github.sub":           "1234567890",
			"io.flipt.auth.github.organizations": "[\"flipt-io\"]",
			"io.flipt.auth.github.teams":         "{\"flipt-io\":[\"flipt-team\"]}",
		}, callback.Authentication.Metadata)
	})

	t.Run("store all allowed orgs that matches orgs of the user to metadata", func(t *testing.T) {
		ctx := t.Context()
		client := newTestServer(t, config.AuthenticationMethod[config.AuthenticationMethodGithubConfig]{
			Enabled: true,
			Method: config.AuthenticationMethodGithubConfig{
				ClientSecret:         "topsecret",
				ClientId:             "githubid",
				RedirectAddress:      "test.flipt.io",
				Scopes:               []string{"read:org"},
				AllowedOrganizations: []string{"flipt-io", "flipt-io-2"},
			},
		})

		defer gock.Off()
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
			JSON([]githubSimpleOrganization{{Login: "flipt-io"}, {Login: "test-flipt-io"}})

		gock.New("https://api.github.com").
			MatchHeader("Authorization", "Bearer AccessToken").
			MatchHeader("Accept", "application/vnd.github+json").
			Get("/user/teams").
			Reply(200).
			JSON([]githubSimpleTeam{})

		callback, err := client.Callback(ctx, &auth.CallbackRequest{Code: "github_code"})
		require.NoError(t, err)

		assert.NotEmpty(t, callback.ClientToken)
		assert.Equal(t, auth.Method_METHOD_GITHUB, callback.Authentication.Method)
		assert.Equal(t, map[string]string{
			"io.flipt.auth.github.email":         "user@flipt.io",
			"io.flipt.auth.email":                "user@flipt.io",
			"io.flipt.auth.github.name":          "fliptuser",
			"io.flipt.auth.name":                 "fliptuser",
			"io.flipt.auth.github.picture":       "https://thispicture.com",
			"io.flipt.auth.picture":              "https://thispicture.com",
			"io.flipt.auth.github.sub":           "1234567890",
			"io.flipt.auth.github.organizations": "[\"flipt-io\"]",
			"io.flipt.auth.github.teams":         "{}",
		}, callback.Authentication.Metadata)
	})

	t.Run("store all allowed teams that matches teams of the user to metadata", func(t *testing.T) {
		ctx := t.Context()
		client := newTestServer(t, config.AuthenticationMethod[config.AuthenticationMethodGithubConfig]{
			Enabled: true,
			Method: config.AuthenticationMethodGithubConfig{
				ClientSecret:         "topsecret",
				ClientId:             "githubid",
				RedirectAddress:      "test.flipt.io",
				Scopes:               []string{"read:org"},
				AllowedOrganizations: []string{"flipt-io"},
				AllowedTeams: map[string][]string{
					"flipt-io": {"flipt-team", "flipt-team-2"},
				},
			},
		})

		defer gock.Off()
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
			JSON([]githubSimpleOrganization{{Login: "flipt-io"}, {Login: "test-flipt-io"}})

		gock.New("https://api.github.com").
			MatchHeader("Authorization", "Bearer AccessToken").
			MatchHeader("Accept", "application/vnd.github+json").
			Get("/user/teams").
			Reply(200).
			JSON([]githubSimpleTeam{
				{
					Slug: "flipt-team",
					Organization: githubSimpleOrganization{
						Login: "flipt-io",
					},
				},
				{
					Slug: "test-flipt-team",
					Organization: githubSimpleOrganization{
						Login: "flipt-io",
					},
				},
			})

		callback, err := client.Callback(ctx, &auth.CallbackRequest{Code: "github_code"})
		require.NoError(t, err)

		assert.NotEmpty(t, callback.ClientToken)
		assert.Equal(t, auth.Method_METHOD_GITHUB, callback.Authentication.Method)
		assert.Equal(t, map[string]string{
			"io.flipt.auth.github.email":         "user@flipt.io",
			"io.flipt.auth.email":                "user@flipt.io",
			"io.flipt.auth.github.name":          "fliptuser",
			"io.flipt.auth.name":                 "fliptuser",
			"io.flipt.auth.github.picture":       "https://thispicture.com",
			"io.flipt.auth.picture":              "https://thispicture.com",
			"io.flipt.auth.github.sub":           "1234567890",
			"io.flipt.auth.github.organizations": "[\"flipt-io\"]",
			"io.flipt.auth.github.teams":         "{\"flipt-io\":[\"flipt-team\"]}",
		}, callback.Authentication.Metadata)
	})

	t.Run("store all teams of allowed orgs to metadata", func(t *testing.T) {
		ctx := t.Context()
		client := newTestServer(t, config.AuthenticationMethod[config.AuthenticationMethodGithubConfig]{
			Enabled: true,
			Method: config.AuthenticationMethodGithubConfig{
				ClientSecret:         "topsecret",
				ClientId:             "githubid",
				RedirectAddress:      "test.flipt.io",
				Scopes:               []string{"read:org"},
				AllowedOrganizations: []string{"flipt-io"},
			},
		})

		defer gock.Off()
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
			JSON([]githubSimpleOrganization{{Login: "flipt-io"}, {Login: "test-flipt-io"}})

		gock.New("https://api.github.com").
			MatchHeader("Authorization", "Bearer AccessToken").
			MatchHeader("Accept", "application/vnd.github+json").
			Get("/user/teams").
			Reply(200).
			JSON([]githubSimpleTeam{
				{
					Slug: "flipt-team",
					Organization: githubSimpleOrganization{
						Login: "flipt-io",
					},
				},
			})

		callback, err := client.Callback(ctx, &auth.CallbackRequest{Code: "github_code"})
		require.NoError(t, err)

		assert.NotEmpty(t, callback.ClientToken)
		assert.Equal(t, auth.Method_METHOD_GITHUB, callback.Authentication.Method)
		assert.Equal(t, map[string]string{
			"io.flipt.auth.github.email":         "user@flipt.io",
			"io.flipt.auth.email":                "user@flipt.io",
			"io.flipt.auth.github.name":          "fliptuser",
			"io.flipt.auth.name":                 "fliptuser",
			"io.flipt.auth.github.picture":       "https://thispicture.com",
			"io.flipt.auth.picture":              "https://thispicture.com",
			"io.flipt.auth.github.sub":           "1234567890",
			"io.flipt.auth.github.organizations": "[\"flipt-io\"]",
			"io.flipt.auth.github.teams":         "{\"flipt-io\":[\"flipt-team\"]}",
		}, callback.Authentication.Metadata)
	})

	t.Run("should authorize successfully when the user is a member of one of the allowed organizations", func(t *testing.T) {
		ctx := t.Context()
		client := newTestServer(t, config.AuthenticationMethod[config.AuthenticationMethodGithubConfig]{
			Enabled: true,
			Method: config.AuthenticationMethodGithubConfig{
				ClientSecret:         "topsecret",
				ClientId:             "githubid",
				RedirectAddress:      "test.flipt.io",
				Scopes:               []string{"read:org"},
				AllowedOrganizations: []string{"flipt-io"},
			},
		})

		defer gock.Off()
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

		gock.New("https://api.github.com").
			MatchHeader("Authorization", "Bearer AccessToken").
			MatchHeader("Accept", "application/vnd.github+json").
			Get("/user/teams").
			Reply(200).
			JSON([]githubSimpleTeam{
				{
					Slug: "flipt-team",
					Organization: githubSimpleOrganization{
						Login: "flipt-io",
					},
				},
			})

		callback, err := client.Callback(ctx, &auth.CallbackRequest{Code: "github_code"})
		require.NoError(t, err)
		assert.NotEmpty(t, callback.ClientToken)
	})

	t.Run("should not authorize when the user is not a member of one of the allowed organizations", func(t *testing.T) {
		ctx := t.Context()
		client := newTestServer(t, config.AuthenticationMethod[config.AuthenticationMethodGithubConfig]{
			Enabled: true,
			Method: config.AuthenticationMethodGithubConfig{
				ClientSecret:         "topsecret",
				ClientId:             "githubid",
				RedirectAddress:      "test.flipt.io",
				Scopes:               []string{"read:org"},
				AllowedOrganizations: []string{"io-flipt"},
			},
		})

		defer gock.Off()
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

		gock.New("https://api.github.com").
			MatchHeader("Authorization", "Bearer AccessToken").
			MatchHeader("Accept", "application/vnd.github+json").
			Get("/user/teams").
			Reply(200).
			JSON([]githubSimpleTeam{
				{
					Slug: "flipt-team",
					Organization: githubSimpleOrganization{
						Login: "flipt-io",
					},
				},
			})

		_, err := client.Callback(ctx, &auth.CallbackRequest{Code: "github_code"})
		assert.ErrorContains(t, err, "account does not satisfy the organization requirements")
	})

	t.Run("should authorize successfully when the user is a member of one of the allowed organizations and it's inside the allowed team", func(t *testing.T) {
		ctx := t.Context()
		client := newTestServer(t, config.AuthenticationMethod[config.AuthenticationMethodGithubConfig]{
			Enabled: true,
			Method: config.AuthenticationMethodGithubConfig{
				ClientSecret:         "topsecret",
				ClientId:             "githubid",
				RedirectAddress:      "test.flipt.io",
				Scopes:               []string{"read:org"},
				AllowedOrganizations: []string{"flipt-io"},
				AllowedTeams: map[string][]string{
					"flipt-io": {"flipt-team"},
				},
			},
		})

		defer gock.Off()
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

		gock.New("https://api.github.com").
			MatchHeader("Authorization", "Bearer AccessToken").
			MatchHeader("Accept", "application/vnd.github+json").
			Get("/user/teams").
			Reply(200).
			JSON([]githubSimpleTeam{
				{
					Slug: "flipt-team",
					Organization: githubSimpleOrganization{
						Login: "flipt-io",
					},
				},
			})

		callback, err := client.Callback(ctx, &auth.CallbackRequest{Code: "github_code"})
		require.NoError(t, err)
		assert.NotEmpty(t, callback.ClientToken)
	})

	t.Run("should not authorize when the user is a member of one of the allowed organizations and but it's not inside the allowed team", func(t *testing.T) {
		ctx := t.Context()
		client := newTestServer(t, config.AuthenticationMethod[config.AuthenticationMethodGithubConfig]{
			Enabled: true,
			Method: config.AuthenticationMethodGithubConfig{
				ClientSecret:         "topsecret",
				ClientId:             "githubid",
				RedirectAddress:      "test.flipt.io",
				Scopes:               []string{"read:org"},
				AllowedOrganizations: []string{"flipt-io"},
				AllowedTeams: map[string][]string{
					"flipt-io": {"flipt-team"},
				},
			},
		})

		defer gock.Off()
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

		gock.New("https://api.github.com").
			MatchHeader("Authorization", "Bearer AccessToken").
			MatchHeader("Accept", "application/vnd.github+json").
			Get("/user/teams").
			Reply(200).
			JSON([]githubSimpleTeam{
				{
					Slug: "simple-team",
					Organization: githubSimpleOrganization{
						Login: "flipt-io",
					},
				},
			})

		_, err := client.Callback(ctx, &auth.CallbackRequest{Code: "github_code"})
		require.ErrorIs(t, err, status.Error(codes.Unauthenticated, "account does not satisfy the team requirements"))
	})

	t.Run("should return an internal error when github user route returns a code different from 200 (OK)", func(t *testing.T) {
		ctx := t.Context()
		client := newTestServer(t, config.AuthenticationMethod[config.AuthenticationMethodGithubConfig]{
			Enabled: true,
			Method: config.AuthenticationMethodGithubConfig{
				ClientSecret:    "topsecret",
				ClientId:        "githubid",
				RedirectAddress: "test.flipt.io",
				Scopes:          []string{"user", "email"},
			},
		})

		defer gock.Off()
		gock.New("https://api.github.com").
			MatchHeader("Authorization", "Bearer AccessToken").
			MatchHeader("Accept", "application/vnd.github+json").
			Get("/user").
			Reply(400)

		_, err := client.Callback(ctx, &auth.CallbackRequest{Code: "github_code"})
		assert.ErrorIs(t, err, status.Error(codes.Internal, `github /user info response status: "400 Bad Request"`))
	})

	t.Run("should return an internal error when github user orgs route returns a code different from 200 (OK)", func(t *testing.T) {
		ctx := t.Context()
		client := newTestServer(t, config.AuthenticationMethod[config.AuthenticationMethodGithubConfig]{
			Enabled: true,
			Method: config.AuthenticationMethodGithubConfig{
				ClientSecret:         "topsecret",
				ClientId:             "githubid",
				RedirectAddress:      "test.flipt.io",
				Scopes:               []string{"read:org"},
				AllowedOrganizations: []string{"flipt-io"},
			},
		})

		defer gock.Off()
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
			Reply(429).
			BodyString("too many requests")

		_, err := client.Callback(ctx, &auth.CallbackRequest{Code: "github_code"})
		assert.ErrorIs(t, err, status.Error(codes.Internal, `github /user/orgs info response status: "429 Too Many Requests"`))
	})

	t.Run("should return an internal error when github user teams route returns a code different from 200 (OK)", func(t *testing.T) {
		ctx := t.Context()
		client := newTestServer(t, config.AuthenticationMethod[config.AuthenticationMethodGithubConfig]{
			Enabled: true,
			Method: config.AuthenticationMethodGithubConfig{
				ClientSecret:         "topsecret",
				ClientId:             "githubid",
				RedirectAddress:      "test.flipt.io",
				Scopes:               []string{"read:org"},
				AllowedOrganizations: []string{"flipt-io"},
				AllowedTeams: map[string][]string{
					"flipt-io": {"flipt-team"},
				},
			},
		})

		defer gock.Off()
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

		gock.New("https://api.github.com").
			MatchHeader("Authorization", "Bearer AccessToken").
			MatchHeader("Accept", "application/vnd.github+json").
			Get("/user/teams").
			Reply(400).
			BodyString("bad request")

		_, err := client.Callback(ctx, &auth.CallbackRequest{Code: "github_code"})
		assert.ErrorIs(t, err, status.Error(codes.Internal, `github /user/teams info response status: "400 Bad Request"`))
	})
}

func Test_Server_SkipsAuthentication(t *testing.T) {
	server := &Server{}
	assert.True(t, server.SkipsAuthentication(t.Context()))
}

func TestCallbackURL(t *testing.T) {
	callback := callbackURL("https://flipt.io")
	assert.Equal(t, "https://flipt.io/auth/v1/method/github/callback", callback)

	callback = callbackURL("https://flipt.io/")
	assert.Equal(t, "https://flipt.io/auth/v1/method/github/callback", callback)
}

type challengeCaptureStore struct {
	storageauth.Store

	mu     sync.Mutex
	puts   map[string]string
	pops   map[string]string
	popErr error
}

func newChallengeCaptureStore(store storageauth.Store) *challengeCaptureStore {
	return &challengeCaptureStore{
		Store: store,
		puts:  map[string]string{},
		pops:  map[string]string{},
	}
}

func (s *challengeCaptureStore) PutOAuthChallenge(ctx context.Context, state, value string, expiresAt *timestamppb.Timestamp) error {
	s.mu.Lock()
	s.puts[state] = value
	s.mu.Unlock()

	return s.Store.PutOAuthChallenge(ctx, state, value, expiresAt)
}

func (s *challengeCaptureStore) PopOAuthChallenge(ctx context.Context, state string) (string, error) {
	if s.popErr != nil {
		return "", s.popErr
	}

	value, err := s.Store.PopOAuthChallenge(ctx, state)
	if err != nil {
		return "", err
	}

	s.mu.Lock()
	s.pops[state] = value
	s.mu.Unlock()

	return value, nil
}

func (s *challengeCaptureStore) putChallenge(state string) string {
	s.mu.Lock()
	defer s.mu.Unlock()

	return s.puts[state]
}

func (s *challengeCaptureStore) popChallenge(state string) string {
	s.mu.Lock()
	defer s.mu.Unlock()

	return s.pops[state]
}

type popErrorStore struct {
	storageauth.Store
	err error
}

func (s *popErrorStore) PopOAuthChallenge(context.Context, string) (string, error) {
	return "", s.err
}

func (s *popErrorStore) Shutdown(ctx context.Context) error {
	if s.Store == nil {
		return nil
	}

	return s.Store.Shutdown(ctx)
}

func TestGithubSimpleOrganizationDecode(t *testing.T) {
	body := `[{
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
	assert.Len(t, githubUserOrgsResponse, 1)
	assert.Equal(t, "github", githubUserOrgsResponse[0].Login)
}

func newTestServer(t *testing.T, cfg config.AuthenticationMethod[config.AuthenticationMethodGithubConfig], store ...storageauth.Store) auth.AuthenticationMethodGithubServiceClient {
	return newTestServerClient(t, cfg, true, store...)
}

func newRawTestServer(t *testing.T, cfg config.AuthenticationMethod[config.AuthenticationMethodGithubConfig], store ...storageauth.Store) auth.AuthenticationMethodGithubServiceClient {
	return newTestServerClient(t, cfg, false, store...)
}

func newTestServerClient(t *testing.T, cfg config.AuthenticationMethod[config.AuthenticationMethodGithubConfig], autoChallenge bool, store ...storageauth.Store) auth.AuthenticationMethodGithubServiceClient {
	t.Helper()

	listener := bufconn.Listen(1024 * 1024)

	authStore := storageauth.Store(memory.NewStore(zaptest.NewLogger(t)))
	if len(store) > 0 && store[0] != nil {
		authStore = store[0]
	}

	// Setup gRPC Server
	server := grpc.NewServer(
		grpc.ChainUnaryInterceptor(
			middleware.ErrorUnaryInterceptor,
		),
	)

	auth.RegisterAuthenticationMethodGithubServiceServer(server, &Server{
		logger: zaptest.NewLogger(t),
		store:  authStore,
		config: config.AuthenticationConfig{
			Methods: config.AuthenticationMethodsConfig{
				Github: cfg,
			},
		},
		oauth2Config: &OAuth2Mock{},
	})

	go func() {
		if err := server.Serve(listener); err != nil {
			t.Errorf("error closing grpc server: %s", err)
		}
	}()
	t.Cleanup(server.Stop)
	t.Cleanup(func() {
		require.NoError(t, authStore.Shutdown(context.Background())) // nolint: usetesting
	})

	// Setup gRPC Client
	conn, err := grpc.DialContext(
		t.Context(),
		"",
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithContextDialer(func(context.Context, string) (net.Conn, error) {
			return listener.Dial()
		}),
	)
	require.NoError(t, err)
	t.Cleanup(func() { conn.Close() })

	client := auth.NewAuthenticationMethodGithubServiceClient(conn)
	if autoChallenge {
		return &testGithubClient{AuthenticationMethodGithubServiceClient: client}
	}
	return client
}
