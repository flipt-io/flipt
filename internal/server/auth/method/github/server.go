package github

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"go.flipt.io/flipt/errors"
	"go.flipt.io/flipt/internal/config"
	"go.flipt.io/flipt/internal/server/auth/method"
	storageauth "go.flipt.io/flipt/internal/storage/auth"
	"go.flipt.io/flipt/rpc/flipt/auth"
	"go.uber.org/zap"
	"golang.org/x/oauth2"
	oauth2GitHub "golang.org/x/oauth2/github"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/timestamppb"
)

const (
	githubUserAPI = "https://api.github.com/user"
)

// OAuth2Client is our abstraction of communication with an OAuth2 Provider.
type OAuth2Client interface {
	AuthCodeURL(state string, opts ...oauth2.AuthCodeOption) string
	Exchange(ctx context.Context, code string, opts ...oauth2.AuthCodeOption) (*oauth2.Token, error)
	Client(ctx context.Context, t *oauth2.Token) *http.Client
}

const (
	storageMetadataGithubEmail = "io.flipt.auth.github.email"
	storageMetadataGithubName  = "io.flipt.auth.github.name"
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
	return &Server{
		logger: logger,
		store:  store,
		config: config,
		oauth2Config: &oauth2.Config{
			ClientID:     config.Methods.Github.Method.ClientId,
			ClientSecret: config.Methods.Github.Method.ClientSecret,
			Endpoint:     oauth2GitHub.Endpoint,
			RedirectURL:  callbackURL(config.Methods.Github.Method.RedirectAddress),
			Scopes:       config.Methods.Github.Method.Scopes,
		},
	}
}

// RegisterGRPC registers the server as an Server on the provided grpc server.
func (s *Server) RegisterGRPC(server *grpc.Server) {
	auth.RegisterAuthenticationMethodGithubServiceServer(server, s)
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

	c := &http.Client{
		Timeout: 5 * time.Second,
	}

	userReq, err := http.NewRequestWithContext(ctx, "GET", githubUserAPI, nil)
	if err != nil {
		return nil, err
	}

	userReq.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token.AccessToken))
	userReq.Header.Set("Accept", "application/vnd.github+json")

	var githubUserResponse struct {
		Name  string `json:"name,omitempty"`
		Email string `json:"email,omitempty"`
	}

	userResp, err := c.Do(userReq)
	if err != nil {
		return nil, err
	}

	defer func() {
		userResp.Body.Close()
	}()

	if err := json.NewDecoder(userResp.Body).Decode(&githubUserResponse); err != nil {
		return nil, err
	}

	metadata := map[string]string{
		storageMetadataGithubEmail: githubUserResponse.Email,
		storageMetadataGithubName:  githubUserResponse.Name,
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
