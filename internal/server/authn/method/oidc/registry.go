package oidc

import (
	"context"
	"fmt"
	"net/url"
	"reflect"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/coreos/go-oidc/v3/oidc"
	"go.flipt.io/flipt/internal/config"
	"golang.org/x/oauth2"
)

// Registry manages OIDC provider clients with lazy initialization and periodic refresh.
type Registry struct {
	providers map[string]*entry
	cacheTTL  time.Duration
}

// NewRegistry creates a Registry from the provided authentication configuration.
func NewRegistry(config config.AuthenticationConfig) *Registry {
	interval := config.Methods.OIDC.Method.CacheTTL
	if interval <= 0 {
		interval = 15 * time.Minute
	}
	var providers map[string]*entry
	if config.Methods.OIDC.Enabled {
		providers = make(map[string]*entry, len(config.Methods.OIDC.Method.Providers))
		for k, v := range config.Methods.OIDC.Method.Providers {
			providers[k] = &entry{
				cfg: v,
				cur: atomic.Pointer[client]{},
			}
		}
	}
	return &Registry{providers: providers, cacheTTL: interval}
}

func (r *Registry) getProvider(ctx context.Context, key string) (*client, error) {
	p, ok := r.providers[key]
	if !ok {
		return nil, fmt.Errorf("no oidc provider %q", key)
	}

	if xp := p.cur.Load(); xp != nil && time.Since(xp.createdAt) < r.cacheTTL {
		return xp, nil
	}

	p.mu.Lock()
	defer p.mu.Unlock()

	if xp := p.cur.Load(); xp != nil && time.Since(xp.createdAt) < r.cacheTTL {
		return xp, nil
	}

	c, err := newOIDCClient(ctx, p.cfg, key)
	if err != nil {
		return nil, err
	}
	var providerMeta struct {
		EndSessionEndpoint string `json:"end_session_endpoint"`
	}
	if err := c.provider.Claims(&providerMeta); err != nil {
		return nil, err
	}
	c.endSessionEndpoint = providerMeta.EndSessionEndpoint

	p.cur.Store(c)
	return c, nil
}

type entry struct {
	mu  sync.Mutex
	cur atomic.Pointer[client]
	cfg config.AuthenticationMethodOIDCProvider
}

// Tk represents an ID token received from an OIDC provider
// after a successful token exchange.
type Tk struct {
	idToken     *oidc.IDToken
	rawIDToken  string
	oauth2Token *oauth2.Token
}

// String returns a redacted representation of the token for logging.
func (i *Tk) String() string {
	if i == nil {
		return "<nil>"
	}
	return "Tk{idToken:<redacted>, rawIDToken:<redacted>, oauth2Token:<redacted>}"
}

// Claims unmarshals the ID token claims into the provided value.
func (i *Tk) Claims(v any) error {
	return i.idToken.Claims(v)
}

// client wraps an OIDC provider with its OAuth2 config and verifier.
type client struct {
	cfg                config.AuthenticationMethodOIDCProvider
	provider           *oidc.Provider
	oauth2Cfg          *oauth2.Config
	oidcCfg            *oidc.Config
	endSessionEndpoint string
	createdAt          time.Time
}

func callbackURL(host, provider string) string {
	return fmt.Sprintf("%s/auth/v1/method/oidc/%s/callback", strings.TrimSuffix(host, "/"), provider)
}

func newOIDCClient(ctx context.Context, cfg config.AuthenticationMethodOIDCProvider, provider string) (*client, error) {
	p, err := oidc.NewProvider(ctx, cfg.IssuerURL)
	if err != nil {
		return nil, fmt.Errorf("creating OIDC provider: %w", err)
	}

	scopes := []string{oidc.ScopeOpenID}
	scopes = append(scopes, cfg.Scopes...)

	oauth2Config := &oauth2.Config{
		ClientID:     cfg.ClientID,
		ClientSecret: cfg.ClientSecret,
		RedirectURL:  callbackURL(cfg.RedirectAddress, provider),
		Endpoint:     p.Endpoint(),
		Scopes:       scopes,
	}

	supportedAlgs := cfg.Algorithms
	if len(supportedAlgs) == 0 {
		supportedAlgs = []string{oidc.RS256}
	}

	oidcConfig := &oidc.Config{
		ClientID:             cfg.ClientID,
		SupportedSigningAlgs: supportedAlgs,
	}

	return &client{
		cfg:       cfg,
		provider:  p,
		oauth2Cfg: oauth2Config,
		oidcCfg:   oidcConfig,
		createdAt: time.Now(),
	}, nil
}

// UsePKCE returns whether PKCE is enabled for this provider.
func (c *client) UsePKCE() bool {
	return c.cfg.UsePKCE
}

// AuthURL builds the authorization URL for the OIDC provider.
func (c *client) AuthURL(state, challenge, nonce string) (string, error) {
	opts := []oauth2.AuthCodeOption{
		oidc.Nonce(nonce),
	}

	if c.cfg.UsePKCE {
		opts = append(opts, oauth2.S256ChallengeOption(challenge))
	}

	authURL := c.oauth2Cfg.AuthCodeURL(state, opts...)
	if len(c.cfg.AuthorizeParameters) != 0 {
		at, err := url.Parse(authURL)
		if err != nil {
			return "", err
		}
		query := at.Query()
		for k, v := range c.cfg.AuthorizeParameters {
			query.Add(k, v)
		}
		at.RawQuery = query.Encode()
		authURL = at.String()
	}
	return authURL, nil
}

// Exchange exchanges an authorization code for an ID token.
func (c *client) Exchange(ctx context.Context, code, challenge, nonce string) (*Tk, error) {
	opts := []oauth2.AuthCodeOption{}
	if c.cfg.UsePKCE {
		opts = append(opts, oauth2.VerifierOption(challenge))
	}

	oauth2Token, err := c.oauth2Cfg.Exchange(ctx, code, opts...)
	if err != nil {
		return nil, fmt.Errorf("exchange id_token: %w", err)
	}
	rawIDToken, ok := oauth2Token.Extra("id_token").(string)
	if !ok {
		return nil, fmt.Errorf("no id_token in token response")
	}

	idToken, err := c.provider.Verifier(c.oidcCfg).Verify(ctx, rawIDToken)
	if err != nil {
		return nil, fmt.Errorf("verifying id_token: %w", err)
	}

	if idToken.Nonce != nonce {
		return nil, fmt.Errorf("invalid nonce in id_token")
	}

	return &Tk{
		idToken:     idToken,
		rawIDToken:  rawIDToken,
		oauth2Token: oauth2Token,
	}, nil
}

// UserInfo fetches additional claims from the provider's UserInfo endpoint.
func (c *client) UserInfo(ctx context.Context, tokenSource oauth2.TokenSource, validSubject string, claims any) error {
	if tokenSource == nil {
		return fmt.Errorf("token source is nil")
	}
	if claims == nil {
		return fmt.Errorf("claims interface is nil")
	}
	if reflect.ValueOf(claims).Kind() != reflect.Pointer {
		return fmt.Errorf("claims must to be a pointer")
	}
	userInfo, err := c.provider.UserInfo(ctx, tokenSource)
	if err != nil {
		return fmt.Errorf("request to get user info failed: %w", err)
	}
	if userInfo.Subject != validSubject {
		return fmt.Errorf("invalid subject")
	}
	err = userInfo.Claims(claims)
	if err != nil {
		return fmt.Errorf("failed to get claims: %w", err)
	}
	return nil
}

// VerifyLogout verifies a back-channel logout token from the OIDC provider.
func (c *client) VerifyLogout(ctx context.Context, rawlogoutToken string) (*oidc.LogoutToken, error) {
	verifier := c.provider.Verifier(c.oidcCfg)
	logoutToken, err := verifier.VerifyLogout(ctx, rawlogoutToken)
	if err != nil {
		return nil, fmt.Errorf("failed verify raw logout token: %w", err)
	}

	return logoutToken, nil
}
