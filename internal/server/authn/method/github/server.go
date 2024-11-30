package github

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"slices"
	"strings"
	"time"

	"go.flipt.io/flipt/errors"
	"go.flipt.io/flipt/internal/config"
	"go.flipt.io/flipt/internal/server/authn/method"
	storageauth "go.flipt.io/flipt/internal/storage/authn"
	"go.flipt.io/flipt/rpc/flipt/auth"
	"go.uber.org/zap"
	"golang.org/x/exp/maps"
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
	storageMetadataGitHubOrganiztions      = "io.flipt.auth.github.organizations"
	storageMetadataGithubTeams             = "io.flipt.auth.github.teams"
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
	serverURL := config.Methods.Github.Method.ServerURL
	if serverURL == "" {
		serverURL = githubServer
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
	// nolint:gocritic
	u, _ := url.JoinPath(strings.TrimSuffix(host, "/"), "auth/v1/method/github/callback")
	return u
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

	if slices.Contains(s.config.Methods.Github.Method.Scopes, "read:org") {
		allowedOrgs := s.config.Methods.Github.Method.AllowedOrganizations
		allowedTeams := s.config.Methods.Github.Method.AllowedTeams

		userOrgs, err := getUserOrgs(ctx, token, apiURL)
		if err != nil {
			return nil, err
		}

		var userTeamsByOrg map[string]map[string]bool
		userTeamsByOrg, err = getUserTeamsByOrg(ctx, token, apiURL)
		if err != nil {
			return nil, err
		}

		if err := authnOrgsAndTeams(userOrgs, userTeamsByOrg, s.config.Methods.Github.Method); err != nil {
			return nil, err
		}

		orgs, err := parseOrgsForMetadata(userOrgs, allowedOrgs)
		if err != nil {
			return nil, err
		}
		set(storageMetadataGitHubOrganiztions, orgs)

		teams, err := parseTeamsForMetadata(userTeamsByOrg, allowedOrgs, allowedTeams)
		if err != nil {
			return nil, err
		}
		set(storageMetadataGithubTeams, teams)
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
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, apiURL+string(endpoint), nil)
	if err != nil {
		return err
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token.AccessToken))
	req.Header.Set("Accept", "application/vnd.github+json")

	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("github %s info response status: %q", req.URL.Path, resp.Status)
	}

	return json.NewDecoder(resp.Body).Decode(v)
}

// getUserOrgs returns a map of organizations that the user is in.
func getUserOrgs(ctx context.Context, token *oauth2.Token, apiURL string) (map[string]bool, error) {
	var response []githubSimpleOrganization
	if err := api(ctx, token, apiURL, githubUserOrganizations, &response); err != nil {
		return nil, err
	}

	orgs := make(map[string]bool, len(response))
	for _, org := range response {
		orgs[org.Login] = true
	}

	return orgs, nil
}

// getUserTeamsByOrg returns a map of organizations to teams that the user is in.
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

// authnOrgsAndTeams checks if the user is authenticated based on the allowed organizations and teams.
func authnOrgsAndTeams(orgs map[string]bool, userTeamsByOrg map[string]map[string]bool, config config.AuthenticationMethodGithubConfig) error {
	if len(config.AllowedOrganizations) == 0 {
		return nil
	}

	for _, org := range config.AllowedOrganizations {
		if !orgs[org] {
			continue
		}

		// If no teams are specified, any user in the org is allowed
		if len(config.AllowedTeams) == 0 {
			return nil
		}

		// Check if user is in any of the allowed teams for this org
		var (
			allowedTeams = config.AllowedTeams[org]
			userTeams    = userTeamsByOrg[org]
		)

		for _, team := range allowedTeams {
			if userTeams[team] {
				return nil
			}
		}
	}

	return errors.ErrUnauthenticatedf("request was not authenticated")
}

// parseOrgsForMetadata returns a JSON encoded list of organizations that the user is in.
func parseOrgsForMetadata(orgs map[string]bool, allowedOrgs []string) (string, error) {
	var orgList []string

	if len(allowedOrgs) == 0 {
		orgList = maps.Keys(orgs)
	} else {
		for _, org := range allowedOrgs {
			if orgs[org] {
				orgList = append(orgList, org)
			}
		}
	}

	out, err := json.Marshal(orgList)
	if err != nil {
		return "", err
	}

	return string(out), nil
}

// parseTeamsForMetadata returns a JSON encoded list of teams that the user is in.
func parseTeamsForMetadata(userTeamsByOrg map[string]map[string]bool, allowedOrgs []string, allowedTeams map[string][]string) (string, error) {
	teamsByOrg := make(map[string][]string)

	// Helper to add teams for an org
	addTeamsForOrg := func(org string, filterTeams []string) {
		if teams, exists := userTeamsByOrg[org]; exists {
			if len(filterTeams) == 0 {
				// Add all teams
				teamsByOrg[org] = maps.Keys(teams)
			} else {
				// Add only allowed teams
				for _, team := range filterTeams {
					if teams[team] {
						teamsByOrg[org] = append(teamsByOrg[org], team)
					}
				}
			}
		}
	}

	// nolint:gocritic
	if len(allowedTeams) != 0 {
		// Filter by specific teams within orgs
		for org, teams := range allowedTeams {
			addTeamsForOrg(org, teams)
		}
	} else if len(allowedOrgs) != 0 {
		// Filter by allowed orgs only
		for _, org := range allowedOrgs {
			addTeamsForOrg(org, nil)
		}
	} else {
		// No filtering, add all teams
		for org := range userTeamsByOrg {
			addTeamsForOrg(org, nil)
		}
	}

	out, err := json.Marshal(teamsByOrg)
	if err != nil {
		return "", err
	}

	return string(out), nil
}
