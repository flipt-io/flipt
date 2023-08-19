package github

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"go.flipt.io/flipt/errors"
	"go.flipt.io/flipt/internal/config"
	storageauth "go.flipt.io/flipt/internal/storage/auth"
	"go.flipt.io/flipt/rpc/flipt/auth"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
	"google.golang.org/protobuf/types/known/timestamppb"
)

const (
	storageMetadataGithubEmail = "io.flipt.auth.github.email"
	storageMetadataGithubName  = "io.flipt.auth.github.name"
)

// Server is an Github server side handler.
type Server struct {
	logger *zap.Logger
	store  storageauth.Store
	config config.AuthenticationConfig

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
	githubConfig := s.config.Methods.Github.Method

	u, err := url.Parse("https://github.com/login/oauth/authorize")
	if err != nil {
		return nil, err
	}

	// Build Github authorize url.
	q := u.Query()

	q.Set("client_id", githubConfig.ClientId)
	q.Set("scope", strings.Join(githubConfig.Scopes, ":"))
	q.Set("redirect_uri", callbackURL(githubConfig.RedirectAddress))
	q.Set("state", req.State)

	u.RawQuery = q.Encode()

	return &auth.AuthorizeURLResponse{
		AuthorizeUrl: u.String(),
	}, nil
}

// Callback is the OAuth callback method for Github authentication. It will take in a Code
// which is the OAuth grant passed in by the OAuth service, and exchange the grant with an Authentication
// that includes the user information.
func (s *Server) Callback(ctx context.Context, r *auth.CallbackRequest) (*auth.CallbackResponse, error) {
	if r.State != "" {
		md, ok := metadata.FromIncomingContext(ctx)
		if !ok {
			return nil, errors.ErrUnauthenticatedf("missing state parameter")
		}

		state, ok := md["flipt_client_state"]
		if !ok || len(state) < 1 {
			return nil, errors.ErrUnauthenticatedf("missing state parameter")
		}

		if r.State != state[0] {
			return nil, errors.ErrUnauthenticatedf("unexpected state parameter")
		}
	}

	githubConfig := s.config.Methods.Github.Method

	c := &http.Client{
		Timeout: 5 * time.Second,
	}

	req, err := http.NewRequestWithContext(ctx, "POST", "https://github.com/login/oauth/access_token", nil)
	if err != nil {
		return nil, err
	}

	q := req.URL.Query()
	q.Set("client_id", githubConfig.ClientId)
	q.Set("client_secret", githubConfig.ClientSecret)
	q.Set("code", r.Code)
	e := q.Encode()
	req.URL.RawQuery = e

	// We have to accept JSON from the GitHub server when requesting the access token.
	req.Header.Set("Accept", "application/json")

	resp, err := c.Do(req)
	if err != nil {
		return nil, err
	}
	defer func() {
		resp.Body.Close()
	}()

	var githubAccessTokenResponse struct {
		AccessToken string   `json:"access_token,omitempty"`
		Scopes      []string `json:"scopes,omitempty"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&githubAccessTokenResponse); err != nil {
		return nil, err
	}

	userReq, err := http.NewRequestWithContext(ctx, "GET", "https://api.github.com/user", nil)
	if err != nil {
		return nil, err
	}

	userReq.Header.Set("Authorization", fmt.Sprintf("Bearer: %s", githubAccessTokenResponse.AccessToken))
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
