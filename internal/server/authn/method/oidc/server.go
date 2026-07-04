package oidc

import (
	"context"
	"encoding/json"
	"fmt"
	"maps"
	"net/url"
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
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

const (
	storageMetadataOIDCProvider      = "io.flipt.auth.oidc.provider"
	storageMetadataOIDCEmail         = "io.flipt.auth.oidc.email"
	storageMetadataOIDCEmailVerified = "io.flipt.auth.oidc.email_verified"
	storageMetadataOIDCName          = "io.flipt.auth.oidc.name"
	storageMetadataOIDCProfile       = "io.flipt.auth.oidc.profile"
	storageMetadataOIDCPicture       = "io.flipt.auth.oidc.picture"
	storageMetadataOIDCSub           = "io.flipt.auth.oidc.sub"
	oauthChallengeTTL                = 2 * time.Minute
)

// errProviderWithNoEndSessionEndpoint is returned when a provider has no end_session_endpoint.
var errProviderWithNoEndSessionEndpoint = errors.New("provider doesn't have end_session_endpoint")

// Server is the core OIDC server implementation for Flipt.
// It supports two primary operations:
//   - AuthorizeURL
//   - Callback
//
// These are two legs of the OIDC/OAuth flow.
// Step 1 is Flipt establishes a URL directed at the delegated authentication service (e.g. Google).
// The URL is configured using the client ID configured for the provided, a state parameter used to
// prevent CSRF attacks and a callback URL directing back to the Callback operation.
// Step 2 the user-agent navigates to the authorizer and establishes authenticity with them.
// Once established they're redirected to the Callback operation with an authenticity code.
// Step 3 the Callback operation uses this "code" and exchanges with the authorization service
// for an ID Token. The validity of the response is checked (signature verified) and then the identity
// details contained in this response are used to create a temporary Flipt client token.
// This client token can be used to access the rest of the Flipt API.
// Given the user-agent is requestin using HTTP the token is instead established as an HTTP cookie.
type Server struct {
	logger        *zap.Logger
	store         storageauth.Store
	TokenLifetime time.Duration
	registry      *Registry

	auth.UnimplementedAuthenticationMethodOIDCServiceServer
}

func NewServer(
	logger *zap.Logger,
	store storageauth.Store,
	registry *Registry,
	config config.AuthenticationConfig,
) *Server {
	return &Server{
		logger:        logger,
		store:         store,
		TokenLifetime: config.Session.TokenLifetime,
		registry:      registry,
	}
}

// RegisterGRPC registers the server as a Server on the provided grpc server.
func (s *Server) RegisterGRPC(server *grpc.Server) {
	auth.RegisterAuthenticationMethodOIDCServiceServer(server, s)
}

func (s *Server) SkipsAuthentication(ctx context.Context) bool {
	return true
}

// AuthorizeURL constructs and returns a URL directed at the requested OIDC provider
// based on our internal oauth2 client configuration.
// The operation is configured to return a URL which ultimately redirects to the
// callback operation below.
func (s *Server) AuthorizeURL(ctx context.Context, req *auth.AuthorizeURLRequest) (*auth.AuthorizeURLResponse, error) {
	provider, err := s.registry.getProvider(ctx, req.Provider)
	if err != nil {
		return nil, fmt.Errorf("authorize: failed to get provider: %w", err)
	}
	challenge := oauth2.GenerateVerifier()
	nonce := oauth2.GenerateVerifier()
	encoded, err := encodeOAuthChallenge(challenge, nonce)
	if err != nil {
		return nil, err
	}
	if err := s.store.PutOAuthChallenge(ctx, req.State, encoded, timestamppb.New(time.Now().UTC().Add(oauthChallengeTTL))); err != nil {
		return nil, err
	}

	authURL, err := provider.AuthURL(req.State, challenge, nonce)
	if err != nil {
		return nil, fmt.Errorf("authorize: failed to build auth url: %w", err)
	}
	return &auth.AuthorizeURLResponse{AuthorizeUrl: authURL}, nil
}

// Callback attempts to authenticate a callback request from a delegated authorization service.
// Given the request includes a "state" parameter then the requests metadata is interrogated
// for the "flipt_client_state" metadata key.
// This entry must exist and the value match the request state.
// The provided code is exchanged with the associated authorization service provider for an "id_token".
// We verify the retrieved "id_token" is valid and for our client.
// Once verified we extract the users associated email address.
// Given all this completes successfully then we established an associated clientToken in
// the backing authentication store with the identity information retrieved as metadata.
func (s *Server) Callback(ctx context.Context, req *auth.CallbackRequest) (_ *auth.CallbackResponse, err error) {
	if err := method.CallbackValidateState(ctx, req.State); err != nil {
		return nil, fmt.Errorf("callback: failed to get state %w", err)
	}

	provider, err := s.registry.getProvider(ctx, req.Provider)
	if err != nil {
		return nil, fmt.Errorf("callback: failed to get provider: %w", err)
	}

	authChallenge, err := s.store.PopOAuthChallenge(ctx, req.State)
	if err != nil {
		if _, ok := errors.As[errors.ErrNotFound](err); ok {
			return nil, errors.ErrUnauthenticatedf("missing or expired oauth challenge")
		}
		return nil, fmt.Errorf("callback: getting oauth challenge: %w", err)
	}
	challenge, nonce, err := decodeOAuthChallenge(authChallenge)
	if err != nil {
		return nil, err
	}
	idToken, err := provider.Exchange(ctx, req.Code, challenge, nonce)
	if err != nil {
		return nil, err
	}

	metadata := map[string]string{
		storageMetadataOIDCProvider: req.Provider,
	}

	rawClaims := make(map[string]any)
	if err := idToken.Claims(&rawClaims); err != nil {
		return nil, err
	}

	if provider.cfg.FetchExtraUserInfo {
		infoClaims := make(map[string]any)
		err := provider.UserInfo(ctx, oauth2.StaticTokenSource(idToken.oauth2Token), idToken.idToken.Subject, &infoClaims)
		if err != nil {
			return nil, fmt.Errorf("failed to get extra user info: %w", err)
		}
		maps.Copy(rawClaims, infoClaims)
	}

	rawClaimsJSON, err := json.Marshal(rawClaims)
	if err != nil {
		return nil, err
	}

	if rawClaimsJSON != nil {
		metadata[method.StorageMetadataClaims] = string(rawClaimsJSON)
	}

	var claims claims
	if err := idToken.Claims(&claims); err != nil {
		return nil, err
	}

	claims.fallbackFrom(rawClaims)

	claims.addToMetadata(metadata)

	authRequest := &storageauth.CreateAuthenticationRequest{
		Method:    auth.Method_METHOD_OIDC,
		ExpiresAt: timestamppb.New(time.Now().UTC().Add(s.TokenLifetime)),
		Metadata:  metadata,
	}

	if provider.cfg.UseEndSessionEndpoint {
		authRequest.IDToken = idToken.rawIDToken
		if claims.SID != "" {
			authRequest.SessionID = fmt.Sprintf("%s:%s", req.Provider, claims.SID)
		}
	}

	clientToken, a, err := s.store.CreateAuthentication(ctx, authRequest)
	if err != nil {
		return nil, err
	}

	return &auth.CallbackResponse{
		ClientToken:    clientToken,
		Authentication: a,
	}, nil
}

// Revoke handles back-channel logout from the OIDC provider.
// It verifies the logout token, looks up the corresponding authentication record
// via the session ID (sid), and expires it.
func (s *Server) Revoke(ctx context.Context, req *auth.RevokeOIDCRequest) (*auth.RevokeOIDCResponse, error) {
	provider, err := s.registry.getProvider(ctx, req.Provider)
	if err != nil {
		return nil, fmt.Errorf("revoke: failed to get provider: %w", err)
	}
	logoutToken, err := provider.VerifyLogout(ctx, req.LogoutToken)
	if err != nil {
		s.logger.Error("failed to verify logout token", zap.Error(err))
		return nil, status.Error(codes.InvalidArgument, "invalid logout token")
	}

	if logoutToken.SessionID == "" {
		s.logger.Debug("logout token has no session ID, skipping")
		return &auth.RevokeOIDCResponse{}, nil
	}

	sid := fmt.Sprintf("%s:%s", req.Provider, logoutToken.SessionID)
	authID, err := s.store.GetAuthenticationIDBySID(ctx, sid)
	if err != nil {
		s.logger.Debug("failed to find logout token", zap.String("sid", logoutToken.SessionID), zap.Error(err))
		return &auth.RevokeOIDCResponse{}, nil
	}

	err = s.store.DeleteAuthentications(ctx, storageauth.Delete(storageauth.WithID(authID)))
	if err != nil {
		s.logger.Error("failed to expire auth token", zap.String("sid", logoutToken.SessionID), zap.Error(err))
		return nil, status.Error(codes.Unknown, "operation failed")
	}
	return &auth.RevokeOIDCResponse{}, nil
}

type claims struct {
	Email    *string `json:"email"`
	Verified *bool   `json:"email_verified"`
	Name     *string `json:"name"`
	Profile  *string `json:"profile"`
	Picture  *string `json:"picture"`
	Sub      *string `json:"sub"`
	SID      string  `json:"sid"`
}

func (c *claims) fallbackFrom(extra map[string]any) {
	for _, k := range []string{"name", "email", "picture"} {
		if v, ok := extra[k].(string); ok && v != "" {
			switch k {
			case "name":
				if c.Name == nil || *c.Name == "" {
					c.Name = new(v)
				}
			case "email":
				if c.Email == nil || *c.Email == "" {
					c.Email = new(v)
				}
			case "picture":
				if c.Picture == nil || *c.Picture == "" {
					c.Picture = new(v)
				}
			}
		}
	}
}

func EndSessionURI(ctx context.Context, r *Registry, authentication *auth.Authentication, idToken string) (string, error) {
	providerName := authentication.Metadata[storageMetadataOIDCProvider]
	p, ok := r.providers[providerName]
	if !ok {
		return "", fmt.Errorf("endsessionuri: failed to get provider: no oidc provider %q", providerName)
	}

	if !p.cfg.UseEndSessionEndpoint {
		return "", nil
	}

	provider, err := r.getProvider(ctx, providerName)
	if err != nil {
		return "", fmt.Errorf("endsessionuri: failed to get provider: %w", err)
	}

	if provider.endSessionEndpoint == "" {
		return "", errProviderWithNoEndSessionEndpoint
	}

	u, err := url.Parse(provider.endSessionEndpoint)
	if err != nil {
		return "", err
	}

	q := u.Query()
	q.Set("id_token_hint", idToken)
	q.Set("post_logout_redirect_uri", provider.cfg.RedirectAddress)
	u.RawQuery = q.Encode()

	return u.String(), nil
}

func (c claims) addToMetadata(m map[string]string) {
	set := func(key string, s *string) {
		if s != nil && *s != "" {
			m[key] = *s
		}
	}

	set(storageMetadataOIDCEmail, c.Email)
	set(storageMetadataOIDCName, c.Name)
	set(storageMetadataOIDCProfile, c.Profile)
	set(storageMetadataOIDCPicture, c.Picture)
	set(storageMetadataOIDCSub, c.Sub)
	set(method.StorageMetadataEmail, c.Email)
	set(method.StorageMetadataName, c.Name)
	set(method.StorageMetadataPicture, c.Picture)

	if c.Verified != nil {
		m[storageMetadataOIDCEmailVerified] = fmt.Sprintf("%v", *c.Verified)
	}
}

// encodeOAuthChallenge encodes a challenge and nonce into a single string for storage.
func encodeOAuthChallenge(challenge, nonce string) (string, error) {
	if challenge == "" || nonce == "" {
		return "", fmt.Errorf("encodeOAuthChallenge: challenge and nonce must not be empty")
	}
	return challenge + "." + nonce, nil
}

// decodeOAuthChallenge decodes a stored challenge string back into challenge and nonce.
func decodeOAuthChallenge(s string) (challenge, nonce string, err error) {
	parts := strings.SplitN(s, ".", 2)
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		return "", "", fmt.Errorf("decodeOAuthChallenge: invalid challenge data: %q", s)
	}
	return parts[0], parts[1], nil
}
