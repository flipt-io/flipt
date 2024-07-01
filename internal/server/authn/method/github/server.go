package github

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"slices"
	"strings"
	"time"

	"go.flipt.io/flipt/errors"
	"go.flipt.io/flipt/internal/config"
	"go.flipt.io/flipt/internal/server/authn/method"
	storageauth "go.flipt.io/flipt/internal/storage/authn"
	"go.flipt.io/flipt/rpc/flipt/auth"
	"go.uber.org/zap"
	"golang.org/x/oauth2"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type endpoint string

const (
	githubServer                     = "https://github.com"
	githubAPI                        = "https://api.github.com"
	githubUser              endpoint = "/user"
	githubUserOrganizations endpoint = "/user/orgs"
	githubUserTeams         endpoint = "/user/teams?per_page=100"
)

// OAuth2Client is our abstraction of communication with an OAuth2 Provider.
type OAuth2Client interface {
	AuthCodeURL(state string, opts ...oauth2.AuthCodeOption) string
	Exchange(ctx context.Context, code string, opts ...oauth2.AuthCodeOption) (*oauth2.Token, error)
	Client(ctx context.Context, t *oauth2.Token) *http.Client
}

const (
	storageMetadataGithubEmail             = "io.flipt.auth.github.email"
	storageMetadataGithubName              = "io.flipt.auth.github.name"
	storageMetadataGithubPicture           = "io.flipt.auth.github.picture"
	storageMetadataGithubSub               = "io.flipt.auth.github.sub"
	storageMetadataGitHubPreferredUsername = "io.flipt.auth.github.preferred_username"
)

// Server is an Github server side handler.
type Server struct {
	logger       *zap.Logger
	store        storageauth.Store
	config       config.AuthenticationConfig
	oauth2Config OAuth2Client

	auth.UnimplementedAuthenticationMethodGithubServiceServer
}

// NewServer constructs a Server.
func NewServer(
	logger *zap.Logger,
	store storageauth.Store,
	config config.AuthenticationConfig,
) *Server {
	serverURL := githubServer
	if config.Methods.Github.Method.ServerURL != "" {
		serverURL = config.Methods.Github.Method.ServerURL
	}

	return &Server{
		logger: logger,
		store:  store,
		config: config,
		oauth2Config: &oauth2.Config{
			ClientID:     config.Methods.Github.Method.ClientId,
			ClientSecret: config.Methods.Github.Method.ClientSecret,
			Endpoint: oauth2.Endpoint{
				AuthURL:  serverURL + "/login/oauth/authorize",
				TokenURL: serverURL + "/login/oauth/access_token",
			},
			RedirectURL: callbackURL(config.Methods.Github.Method.RedirectAddress),
			Scopes:      config.Methods.Github.Method.Scopes,
		},
	}
}

// RegisterGRPC registers the server as an Server on the provided grpc server.
func (s *Server) RegisterGRPC(server *grpc.Server) {
	auth.RegisterAuthenticationMethodGithubServiceServer(server, s)
}

func (s *Server) SkipsAuthentication(ctx context.Context) bool {
	return true
}

func callbackURL(host string) string {
	// strip trailing slash from host
	host = strings.TrimSuffix(host, "/")
	return host + "/auth/v1/method/github/callback"
}

// AuthorizeURL will return a URL for the client to redirect to for completion of the OAuth flow with GitHub.
func (s *Server) AuthorizeURL(ctx context.Context, req *auth.AuthorizeURLRequest) (*auth.AuthorizeURLResponse, error) {
	u := s.oauth2Config.AuthCodeURL(req.State)

	return &auth.AuthorizeURLResponse{
		AuthorizeUrl: u,
	}, nil
}

// Callback is the OAuth callback method for Github authentication. It will take in a Code
// which is the OAuth grant passed in by the OAuth service, and exchange the grant with an Authentication
// that includes the user information.
func (s *Server) Callback(ctx context.Context, r *auth.CallbackRequest) (*auth.CallbackResponse, error) {
	if r.State != "" {
		if err := method.CallbackValidateState(ctx, r.State); err != nil {
			return nil, err
		}
	}

	token, err := s.oauth2Config.Exchange(ctx, r.Code)
	if err != nil {
		return nil, err
	}

	if !token.Valid() {
		return nil, errors.New("invalid token")
	}

	var githubUserResponse struct {
		Name      string `json:"name,omitempty"`
		Email     string `json:"email,omitempty"`
		AvatarURL string `json:"avatar_url,omitempty"`
		Login     string `json:"login,omitempty"`
		ID        uint64 `json:"id,omitempty"`
	}

	apiURL := githubAPI
	if s.config.Methods.Github.Method.ApiURL != "" {
		apiURL = s.config.Methods.Github.Method.ApiURL
	}

	if err = api(ctx, token, apiURL, githubUser, &githubUserResponse); err != nil {
		return nil, err
	}

	metadata := map[string]string{}
	set := func(key string, s string) {
		if s != "" {
			metadata[key] = s
		}
	}

	set(storageMetadataGithubName, githubUserResponse.Name)
	set(storageMetadataGithubEmail, githubUserResponse.Email)
	set(storageMetadataGithubPicture, githubUserResponse.AvatarURL)

	if githubUserResponse.ID != 0 {
		set(storageMetadataGithubSub, fmt.Sprintf("%d", githubUserResponse.ID))
	}

	set(storageMetadataGitHubPreferredUsername, githubUserResponse.Login)

	// consolidate common fields
	set(method.StorageMetadataEmail, githubUserResponse.Email)
	set(method.StorageMetadataName, githubUserResponse.Name)
	set(method.StorageMetadataPicture, githubUserResponse.AvatarURL)

	if len(s.config.Methods.Github.Method.AllowedOrganizations) != 0 {
		userOrgs, err := getUserOrgs(ctx, token, apiURL)
		if err != nil {
			return nil, err
		}

		var userTeamsByOrg map[string]map[string]bool
		if len(s.config.Methods.Github.Method.AllowedTeams) != 0 {
			userTeamsByOrg, err = getUserTeamsByOrg(ctx, token, apiURL)
			if err != nil {
				return nil, err
			}
		}

		if !slices.ContainsFunc(s.config.Methods.Github.Method.AllowedOrganizations, func(org string) bool {
			if !userOrgs[org] {
				return false
			}

			if userTeamsByOrg == nil {
				return true
			}

			allowedTeams := s.config.Methods.Github.Method.AllowedTeams[org]
			userTeams := userTeamsByOrg[org]

			return slices.ContainsFunc(allowedTeams, func(team string) bool {
				return userTeams[team]
			})
		}) {
			return nil, errors.ErrUnauthenticatedf("request was not authenticated")
		}
	}

	clientToken, a, err := s.store.CreateAuthentication(ctx, &storageauth.CreateAuthenticationRequest{
		Method:    auth.Method_METHOD_GITHUB,
		ExpiresAt: timestamppb.New(time.Now().UTC().Add(s.config.Session.TokenLifetime)),
		Metadata:  metadata,
	})
	if err != nil {
		return nil, err
	}

	return &auth.CallbackResponse{
		ClientToken:    clientToken,
		Authentication: a,
	}, nil
}

type githubSimpleOrganization struct {
	Login string `json:"login"`
}

type githubSimpleTeam struct {
	Slug         string                   `json:"slug"`
	Organization githubSimpleOrganization `json:"organization"`
}

// api calls Github API, decodes and stores successful response in the value pointed to by v.
func api(ctx context.Context, token *oauth2.Token, apiURL string, endpoint endpoint, v any) error {
	c := &http.Client{
		Timeout: 5 * time.Second,
	}

	userReq, err := http.NewRequestWithContext(ctx, "GET", apiURL+string(endpoint), nil)
	if err != nil {
		return err
	}

	userReq.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token.AccessToken))
	userReq.Header.Set("Accept", "application/vnd.github+json")

	resp, err := c.Do(userReq)
	if err != nil {
		return err
	}

	defer func() {
		resp.Body.Close()
	}()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("github %s info response status: %q", userReq.URL.Path, resp.Status)
	}
	return json.NewDecoder(resp.Body).Decode(v)
}

func getUserOrgs(ctx context.Context, token *oauth2.Token, apiURL string) (map[string]bool, error) {
	var response []githubSimpleOrganization
	if err := api(ctx, token, apiURL, githubUserOrganizations, &response); err != nil {
		return nil, err
	}

	orgs := make(map[string]bool)
	for _, org := range response {
		orgs[org.Login] = true
	}

	return orgs, nil
}

func getUserTeamsByOrg(ctx context.Context, token *oauth2.Token, apiURL string) (map[string]map[string]bool, error) {
	var response []githubSimpleTeam
	if err := api(ctx, token, apiURL, githubUserTeams, &response); err != nil {
		return nil, err
	}

	teamsByOrg := make(map[string]map[string]bool)
	for _, team := range response {
		org := team.Organization.Login

		if _, ok := teamsByOrg[org]; !ok {
			teamsByOrg[org] = make(map[string]bool)
		}

		teamsByOrg[org][team.Slug] = true
	}

	return teamsByOrg, nil
}
