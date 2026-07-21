package oidc

import (
	"context"
	"encoding/json"
	"fmt"
	"maps"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/go-openapi/jsonpointer"
	"go.flipt.io/flipt/errors"
	"go.flipt.io/flipt/internal/config"
	"go.flipt.io/flipt/internal/server/authn/method"
	"go.flipt.io/flipt/internal/storage"
	storageauth "go.flipt.io/flipt/internal/storage/authn"
	"go.flipt.io/flipt/rpc/flipt/auth"
	"go.uber.org/zap"
	"golang.org/x/oauth2"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
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

type (
	optionFunc func(*options)
	options    struct {
		nonceGenerator func() string
	}
)

func (f optionFunc) apply(o *options) { f(o) }

// Option configures the OIDC Server.
type Option interface {
	apply(*options)
}

// WithNonceGenerator sets the function used to generate nonce and PKCE challenge values.
// By default oauth2.GenerateVerifier is used. Intended for tests that need predictable values.
func WithNonceGenerator(f func() string) Option {
	return optionFunc(func(o *options) { o.nonceGenerator = f })
}

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
	logger         *zap.Logger
	store          storageauth.Store
	TokenLifetime  time.Duration
	registry       *Registry
	nonceGenerator func() string

	auth.UnimplementedAuthenticationMethodOIDCServiceServer
}

func NewServer(
	logger *zap.Logger,
	store storageauth.Store,
	registry *Registry,
	config config.AuthenticationConfig,
	opts ...Option,
) *Server {
	o := options{
		nonceGenerator: oauth2.GenerateVerifier,
	}
	for _, opt := range opts {
		opt.apply(&o)
	}

	return &Server{
		logger:         logger,
		store:          store,
		TokenLifetime:  config.Session.TokenLifetime,
		registry:       registry,
		nonceGenerator: o.nonceGenerator,
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
		if errors.AsMatch[errors.ErrNotFound](err) {
			return nil, status.Error(codes.InvalidArgument, "authorize: unknown provider")
		}
		s.logger.Error("failed to get provider", zap.String("provider", req.Provider), zap.Error(err))
		return nil, status.Error(codes.Internal, "authorize: failed get provider")
	}
	challenge := s.nonceGenerator()
	nonce := s.nonceGenerator()

	encoded, err := encodeOAuthChallenge(challenge, nonce)
	if err != nil {
		s.logger.Error("failed to generate challenge", zap.Error(err))
		return nil, status.Error(codes.InvalidArgument, "authorize: failed to generate challenge")
	}

	if err := s.store.PutOAuthChallenge(ctx, req.State, encoded, timestamppb.New(time.Now().UTC().Add(oauthChallengeTTL))); err != nil {
		s.logger.Error("failed to persist challenge", zap.Error(err))
		return nil, status.Error(codes.Internal, "authorize: failed to persist challenge")
	}

	authURL, err := provider.AuthURL(req.State, challenge, nonce)
	if err != nil {
		s.logger.Error("failed to build auth url", zap.Error(err))
		return nil, status.Error(codes.Internal, "authorize: failed to create auth url")
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
		s.logger.Error("callback: failed to get state", zap.Error(err))
		return nil, status.Error(codes.Unauthenticated, "callback: failed to get state")
	}

	provider, err := s.registry.getProvider(ctx, req.Provider)
	if err != nil {
		if errors.AsMatch[errors.ErrNotFound](err) {
			return nil, status.Error(codes.InvalidArgument, "callback: unknown provider")
		}
		s.logger.Error("failed to get provider", zap.String("provider", req.Provider), zap.Error(err))
		return nil, status.Error(codes.Internal, "callback: failed get provider")
	}

	authChallenge, err := s.store.PopOAuthChallenge(ctx, req.State)
	if err != nil {
		if _, ok := errors.As[errors.ErrNotFound](err); ok {
			return nil, status.Error(codes.InvalidArgument, "callback: missing challenge")
		}
		s.logger.Error("getting challenge", zap.Error(err))
		return nil, status.Error(codes.Unknown, "callback: getting challenge")
	}

	challenge, nonce, err := decodeOAuthChallenge(authChallenge)
	if err != nil {
		s.logger.Error("failed to decode challenge", zap.Error(err))
		return nil, status.Error(codes.InvalidArgument, "callback: invalid challenge")
	}
	idToken, err := provider.Exchange(ctx, req.Code, challenge, nonce)
	if err != nil {
		s.logger.Error("failed to exchange code", zap.Error(err))
		return nil, status.Error(codes.Unknown, "callback: code exchange failed")
	}

	metadata, claims, err := s.extractMetadata(ctx, provider, idToken)
	if err != nil {
		s.logger.Error("failed to extract claims metadata", zap.Error(err))
		return nil, status.Error(codes.Unknown, "callback: claim extraction failed")
	}

	authRequest := &storageauth.CreateAuthenticationRequest{
		Method:    auth.Method_METHOD_OIDC,
		ExpiresAt: timestamppb.New(time.Now().UTC().Add(s.TokenLifetime)),
		Metadata:  metadata,
		Provider:  req.Provider,
	}

	if claims.SID != "" {
		authRequest.Sid = claims.SID
	}

	if claims.Sub != nil {
		authRequest.Sub = *claims.Sub
	}

	if provider.cfg.UseEndSessionEndpoint {
		authRequest.IDToken = idToken.rawIDToken
	}

	clientToken, a, err := s.store.CreateAuthentication(ctx, authRequest)
	if err != nil {
		s.logger.Error("failed to create authentication", zap.Error(err))
		return nil, status.Error(codes.Internal, "callback: failed to create authentication")
	}

	return &auth.CallbackResponse{
		ClientToken:    clientToken,
		Authentication: a,
	}, nil
}

func (*Server) extractMetadata(ctx context.Context, provider *client, idToken *Tk) (map[string]string, claims, error) {
	metadata := map[string]string{
		storageMetadataOIDCProvider: provider.key,
	}

	rawClaims := make(map[string]any)
	if err := idToken.Claims(&rawClaims); err != nil {
		return nil, claims{}, err
	}

	if provider.cfg.FetchExtraUserInfo {
		infoClaims := make(map[string]any)
		err := provider.UserInfo(ctx, oauth2.StaticTokenSource(idToken.oauth2Token), idToken.idToken.Subject, &infoClaims)
		if err != nil {
			return nil, claims{}, fmt.Errorf("failed to get extra user info: %w", err)
		}
		maps.Copy(rawClaims, infoClaims)
	}

	rawClaimsJSON, err := json.Marshal(rawClaims)
	if err != nil {
		return nil, claims{}, err
	}

	if rawClaimsJSON != nil {
		metadata[method.StorageMetadataClaims] = string(rawClaimsJSON)
	}

	var claimsData claims
	if err := idToken.Claims(&claimsData); err != nil {
		return nil, claimsData, err
	}

	claimsData.fallbackFrom(rawClaims)

	claimsData.applyMapping(rawClaims, provider.cfg.ClaimsMapping)

	claimsData.addToMetadata(metadata)
	return metadata, claimsData, nil
}

// applyMapping overrides claim fields using the configured JSON Pointer
// expressions evaluated against the raw claims. This lets providers whose
// claims don't follow the standard OIDC names (e.g. Azure AD B2C's "emails"
// array) be mapped onto Flipt's canonical fields. Mappings take precedence
// over the standard and fallback claim values.
func (c *claims) applyMapping(rawClaims map[string]any, mapping map[string]string) {
	for attribute, expr := range mapping {
		if expr == "" {
			continue
		}

		ptr, err := jsonpointer.New(expr)
		if err != nil {
			continue
		}

		value, _, err := ptr.Get(rawClaims)
		if err != nil {
			continue
		}

		s, ok := value.(string)
		if !ok || s == "" {
			continue
		}

		switch attribute {
		case "email":
			c.Email = &s
		case "name":
			c.Name = &s
		case "picture":
			c.Picture = &s
		case "sub":
			c.Sub = &s
		}
	}
}

// Revoke handles back-channel logout from the OIDC provider.
// It verifies the logout token, looks up the corresponding authentication record
// via the session ID (sid), and deletes it.
func (s *Server) Revoke(ctx context.Context, req *auth.RevokeOIDCRequest) (*auth.RevokeOIDCResponse, error) {
	provider, err := s.registry.getProvider(ctx, req.Provider)
	if err != nil {
		if errors.AsMatch[errors.ErrNotFound](err) {
			return nil, status.Error(codes.InvalidArgument, "revoke: unknown provider")
		}
		s.logger.Error("failed to get provider", zap.String("provider", req.Provider), zap.Error(err))
		return nil, status.Error(codes.Internal, "revoke: failed get provider")
	}

	httpMethod := getHTTPMethod(ctx)

	listQuery := storage.ListRequest[storageauth.ListAuthenticationsPredicate]{
		Predicate: storageauth.ListAuthenticationsPredicate{
			Method:   new(auth.Method_METHOD_OIDC),
			Provider: &req.Provider,
		},
		QueryParams: storage.QueryParams{Limit: storage.MaxListLimit},
	}

	switch httpMethod {
	case http.MethodPost, "":
		logoutToken, err := provider.VerifyLogout(ctx, req.LogoutToken)
		if err != nil {
			s.logger.Debug("failed to verify logout token", zap.Error(err))
			return nil, status.Error(codes.InvalidArgument, "revoke: invalid logout token")
		}
		switch {
		case logoutToken.SessionID != "" && logoutToken.Subject != "":
			listQuery.Predicate.Sid = &logoutToken.SessionID
			listQuery.Predicate.Sub = &logoutToken.Subject

		case logoutToken.SessionID != "":
			listQuery.Predicate.Sid = &logoutToken.SessionID

		case logoutToken.Subject != "":
			listQuery.Predicate.Sub = &logoutToken.Subject

		default:
			s.logger.Error("logout token is verified as valid but missing both sid and sub")
		}

	case http.MethodGet:
		if !provider.cfg.AllowFrontChannelLogout {
			return nil, status.Error(codes.InvalidArgument, "revoke: operation disabled")
		}

		// Per spec, a present iss MUST match the provider's issuer exactly;
		// mismatch indicates a misconfigured or spoofed request, so error
		// rather than silently ignoring it.
		if req.Iss != nil && *req.Iss != provider.cfg.IssuerURL {
			return nil, status.Error(codes.InvalidArgument, "revoke: issuer mismatch")
		}

		// Without sid or iss the session cannot be identified, so treat it as a no-op
		// success per spec rather than erroring.
		if req.Sid == nil || req.Iss == nil {
			return &auth.RevokeOIDCResponse{}, nil
		}

		listQuery.Predicate.Sid = req.Sid

	default:
		return nil, status.Error(codes.InvalidArgument, "revoke: unsupported http method")
	}

	if listQuery.Predicate.Sid == nil && listQuery.Predicate.Sub == nil {
		return nil, status.Error(codes.InvalidArgument, "revoke: operation requires sid or sub")
	}

	var ids []string

	for {
		authentications, err := s.store.ListAuthentications(ctx, &listQuery)
		if err != nil {
			s.logger.Error("failed to lookup authentication", zap.Error(err))
			return nil, status.Error(codes.Internal, "revoke: lookup failed")
		}

		for _, authentication := range authentications.Results {
			ids = append(ids, authentication.Id)
		}

		if authentications.NextPageToken == "" {
			break
		}

		listQuery.QueryParams.PageToken = authentications.NextPageToken
	}

	s.logger.Debug("attempting to delete authentications", zap.Int("size", len(ids)))

	for _, id := range ids {
		if err := s.store.DeleteAuthentications(ctx, storageauth.Delete(storageauth.WithID(id))); err != nil {
			s.logger.Error("failed to delete authentication", zap.String("id", id), zap.Error(err))
			return nil, status.Error(codes.Internal, "revoke: operation failed")
		}
	}
	return &auth.RevokeOIDCResponse{}, nil
}

func getHTTPMethod(ctx context.Context) string {
	var httpMethod string
	md, ok := metadata.FromIncomingContext(ctx)
	if ok {
		methods := md.Get("x-http-method")
		if len(methods) > 0 {
			httpMethod = methods[0]
		}
	}
	return httpMethod
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
	// consolidate common fields
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
