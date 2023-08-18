package oauth

import (
	"context"
	"encoding/json"
	"net/http"
	"net/url"
	"strings"
	"time"

	"go.flipt.io/flipt/internal/config"
	storageauth "go.flipt.io/flipt/internal/storage/auth"
	"go.flipt.io/flipt/rpc/flipt/auth"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/timestamppb"
)

const (
	storageMetadataGithubAccessToken = "io.flipt.auth.oauth.access_token"
)

var (
	hostToAuthorizeBaseURL = map[string]string{
		"github": "https://github.com/login/oauth/authorize",
	}
	hostToAccessTokenBaseURL = map[string]string{
		"github": "https://github.com/login/oauth/access_token",
	}
)

// Server is an OAuth server side handler.
type Server struct {
	logger *zap.Logger
	store  storageauth.Store
	config config.AuthenticationConfig

	auth.UnimplementedAuthenticationMethodOAuthServiceServer
}

// NewServer constructs a Server.
func NewServer(
	logger *zap.Logger,
	store storageauth.Store,
	config config.AuthenticationConfig,
) *Server {
	return &Server{
		logger: logger,
		store:  store,
		config: config,
	}
}

// RegisterGRPC registers the server as an Server on the provided grpc server.
func (s *Server) RegisterGRPC(server *grpc.Server) {
	auth.RegisterAuthenticationMethodOAuthServiceServer(server, s)
}

func callbackURL(host, oauthHost string) string {
	// strip trailing slash from host
	host = strings.TrimSuffix(host, "/")
	return host + "/auth/v1/method/oauth/" + oauthHost + "/callback"
}

// AuthorizeURL will return a URL for the client to redirect to for completion of the OAuth flow with GitHub.
func (s *Server) AuthorizeURL(ctx context.Context, a *auth.OAuthAuthorizeRequest) (*auth.AuthorizeURLResponse, error) {
	oauthConfig := s.config.Methods.OAuth.Method.Hosts[a.Host]

	authorizeBaseURL := hostToAuthorizeBaseURL[a.Host]

	u, err := url.Parse(authorizeBaseURL)
	if err != nil {
		return nil, err
	}

	// Build OAuth authorize url.
	q := u.Query()

	q.Set("client_id", oauthConfig.ClientId)
	q.Set("scope", strings.Join(oauthConfig.Scopes, ":"))
	q.Set("redirect_uri", callbackURL(oauthConfig.RedirectAddress, a.Host))

	u.RawQuery = q.Encode()

	return &auth.AuthorizeURLResponse{
		AuthorizeUrl: u.String(),
	}, nil
}

// OAuthCallback is the OAuth callback method for OAuth authentication. It will take in a SessionCode
// which is the OAuth grant passed in by the OAuth service, and exchange the grant with an Authentication
// that includes the access token.
func (s *Server) OAuthCallback(ctx context.Context, oauth *auth.OAuthCallbackRequest) (*auth.OAuthCallbackResponse, error) {
	oauthConfig := s.config.Methods.OAuth.Method.Hosts[oauth.Host]

	c := &http.Client{
		Timeout: 5 * time.Second,
	}

	accessTokenURL := hostToAccessTokenBaseURL[oauth.Host]

	req, err := http.NewRequestWithContext(ctx, "POST", accessTokenURL, nil)
	if err != nil {
		return nil, err
	}

	q := req.URL.Query()
	q.Set("client_id", oauthConfig.ClientId)
	q.Set("client_secret", oauthConfig.ClientSecret)
	q.Set("code", oauth.Code)
	e := q.Encode()
	req.URL.RawQuery = e

	// We have to accept JSON from the GitHub server when requesting the access token.
	req.Header.Set("Accept", "application/json")

	resp, err := c.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var oauthAccessTokenResponse struct {
		AccessToken string   `json:"access_token,omitempty"`
		Scopes      []string `json:"scopes,omitempty"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&oauthAccessTokenResponse); err != nil {
		return nil, err
	}

	metadata := map[string]string{
		storageMetadataGithubAccessToken: oauthAccessTokenResponse.AccessToken,
	}

	clientToken, a, err := s.store.CreateAuthentication(ctx, &storageauth.CreateAuthenticationRequest{
		Method:    auth.Method_METHOD_OAUTH,
		ExpiresAt: timestamppb.New(time.Now().UTC().Add(s.config.Session.TokenLifetime)),
		Metadata:  metadata,
	})
	if err != nil {
		return nil, err
	}

	return &auth.OAuthCallbackResponse{
		ClientToken:    clientToken,
		Authentication: a,
	}, nil
}
