package credentials

import (
	"context"
	"fmt"
	"net/http"
	"os"

	"golang.org/x/oauth2"

	"github.com/go-git/go-git/v6/plumbing/client"
	"github.com/go-git/go-git/v6/plumbing/transport"
	githttp "github.com/go-git/go-git/v6/plumbing/transport/http"
	gitssh "github.com/go-git/go-git/v6/plumbing/transport/ssh"
	"github.com/jferrl/go-githubauth"
	"go.flipt.io/flipt/errors"
	"go.flipt.io/flipt/internal/config"
	"go.flipt.io/flipt/internal/secrets"
	"go.uber.org/zap"
	"golang.org/x/crypto/ssh"
)

// SecretResolver resolves secret references to their current values.
type SecretResolver interface {
	GetSecretValue(ctx context.Context, ref secrets.Reference) ([]byte, error)
}

type CredentialSource struct {
	logger  *zap.Logger
	configs config.CredentialsConfig
	secrets SecretResolver
}

// Option configures a CredentialSource.
type Option func(*CredentialSource)

// WithSecretResolver sets the resolver used for access tokens configured as
// ${secret:provider:key} references. Such references are resolved each time
// the token is used, which allows short-lived tokens to be rotated externally.
func WithSecretResolver(r SecretResolver) Option {
	return func(s *CredentialSource) {
		s.secrets = r
	}
}

func New(logger *zap.Logger, configs config.CredentialsConfig, opts ...Option) *CredentialSource {
	s := &CredentialSource{logger: logger, configs: configs}
	for _, opt := range opts {
		opt(s)
	}

	return s
}

func (s *CredentialSource) Get(name string) (*Credential, error) {
	config, ok := s.configs[name]
	if !ok {
		return nil, errors.ErrNotFoundf("credential %q", name)
	}

	return &Credential{logger: s.logger, config: config, secrets: s.secrets}, nil
}

type Credential struct {
	logger  *zap.Logger
	config  *config.CredentialConfig
	secrets SecretResolver
}

// accessToken returns the current value of the configured access token.
// When the token is a ${secret:provider:key} reference it is resolved on
// every call, so a token rotated in the secret store is used straight away.
func (c *Credential) accessToken(ctx context.Context) (string, error) {
	token := *c.config.AccessToken

	ref, ok, err := secrets.ParseReference(token)
	if err != nil {
		return "", fmt.Errorf("access token: %w", err)
	}

	if !ok {
		return token, nil
	}

	if c.secrets == nil {
		return "", fmt.Errorf("access token: no secrets manager available to resolve %q", token)
	}

	value, err := c.secrets.GetSecretValue(ctx, ref)
	if err != nil {
		return "", fmt.Errorf("access token: resolving secret reference %q: %w", token, err)
	}

	if len(value) == 0 {
		return "", fmt.Errorf("access token: secret reference %q resolved to an empty value", token)
	}

	return string(value), nil
}

var _ client.HTTPAuth = (*accessTokenAuth)(nil)

// accessTokenAuth authenticates Git HTTP requests with an access token that
// is looked up for every request rather than once at startup.
type accessTokenAuth struct {
	credential *Credential
}

func (a *accessTokenAuth) Authorizer(r *http.Request) error {
	token, err := a.credential.accessToken(r.Context())
	if err != nil {
		return err
	}

	// Access tokens are sent as HTTP Basic Auth. Different providers have different conventions:
	// - GitLab: username="oauth2", password=token OR user={anynonemptystring}, password=token
	// - GitHub: username=token, password="" OR username={anynonemptystring}, password=token
	// - BitBucket: username="x-token-auth", password=token for repo access tokens
	// We'll use the BitBucket format as it requires a specific string and the others dont
	r.SetBasicAuth("x-token-auth", token)
	return nil
}

// insecureHostKeyAuth wraps an SSH auth method whose HostKeyCallback has already been
// set to ssh.InsecureIgnoreHostKey, guaranteeing HostKeyAlgorithms is non-empty so that
// go-git does not fall back to reading known_hosts. See GitAuthentication for detail.
type insecureHostKeyAuth struct {
	*gitssh.PublicKeys
}

func (a insecureHostKeyAuth) ClientConfig(ctx context.Context, req *transport.Request) (*ssh.ClientConfig, error) {
	cfg, err := a.PublicKeys.ClientConfig(ctx, req)
	if err != nil {
		return nil, err
	}

	if len(cfg.HostKeyAlgorithms) == 0 {
		cfg.HostKeyAlgorithms = ssh.SupportedAlgorithms().HostKeys
	}

	return cfg, nil
}

// GitAuthentication returns the appropriate client.Option for Git operations.
// This method handles the complexity of converting different credential types
// to the format expected by Git operations.
func (c *Credential) GitAuthentication(ctx context.Context) (client.Option, error) {
	switch c.config.Type {
	case config.CredentialTypeBasic:
		return client.WithHTTPAuth(&githttp.BasicAuth{
			Username: c.config.Basic.Username,
			Password: c.config.Basic.Password,
		}), nil
	case config.CredentialTypeSSH:
		var (
			method *gitssh.PublicKeys
			err    error
		)
		if c.config.SSH.PrivateKeyBytes != "" {
			method, err = gitssh.NewPublicKeys(
				c.config.SSH.User,
				[]byte(c.config.SSH.PrivateKeyBytes),
				c.config.SSH.Password,
			)
		} else {
			method, err = gitssh.NewPublicKeysFromFile(
				c.config.SSH.User,
				c.config.SSH.PrivateKeyPath,
				c.config.SSH.Password,
			)
		}
		if err != nil {
			return nil, err
		}

		// we're protecting against this explicitly so we can disable
		// the gosec linting rule
		if c.config.SSH.InsecureIgnoreHostKey {
			// nolint:gosec
			method.HostKeyCallback = ssh.InsecureIgnoreHostKey()

			// Setting HostKeyCallback alone is not enough. go-git still reads
			// known_hosts in order to pre-populate HostKeyAlgorithms, and returns an
			// error when no known_hosts file exists -- even though a HostKeyCallback
			// has been supplied:
			//
			//	} else if len(config.HostKeyAlgorithms) == 0 {
			//		db, err := newKnownHostsDb()
			//		if err != nil {
			//			return nil, err
			//		}
			//
			// (go-git plumbing/transport/ssh/ssh.go, Transport.connect)
			//
			// So on a host with no known_hosts file -- a container, typically --
			// insecure_ignore_host_key still fails with "unable to find any valid
			// known_hosts file, set SSH_KNOWN_HOSTS env variable". Pre-populating the
			// algorithms skips that lookup. Their ordering is a negotiation preference
			// rather than a correctness requirement, so the defaults are fine here.
			return client.WithSSHAuth(insecureHostKeyAuth{PublicKeys: method}), nil
		}

		return client.WithSSHAuth(method), nil
	case config.CredentialTypeAccessToken:
		// Resolve the token once up front so that a misconfigured secret
		// reference fails at startup rather than on the first fetch.
		if _, err := c.accessToken(ctx); err != nil {
			return nil, err
		}

		return client.WithHTTPAuth(&accessTokenAuth{credential: c}), nil
	case config.CredentialTypeGithubApp:
		gha, err := NewGitHubAppCredentials(c.logger, c.config.GitHubApp)
		if err != nil {
			return nil, err
		}
		return client.WithHTTPAuth(gha), err
	}

	return nil, fmt.Errorf("unexpected credential type: %q", c.config.Type)
}

var _ client.HTTPAuth = (*GitHubAppCredentials)(nil)

// GitHubAppCredentials holds the configuration and token source for GitHub App authentication.
// This centralizes the GitHub App auth logic for both Git operations and API operations.
type GitHubAppCredentials struct {
	logger         *zap.Logger
	clientID       string
	installationID int64
	apiURL         string
	tokenSource    oauth2.TokenSource
}

// NewGitHubAppCredentials creates a GitHubAppCredentials from a GitHubAppConfig.
// This function centralizes the private key reading and token source creation.
func NewGitHubAppCredentials(logger *zap.Logger, c *config.GitHubAppConfig) (*GitHubAppCredentials, error) {
	privateKey := []byte(c.PrivateKeyBytes)
	if c.PrivateKeyPath != "" {
		var err error
		privateKey, err = os.ReadFile(c.PrivateKeyPath)
		if err != nil {
			return nil, fmt.Errorf("failed to read private key: %w", err)
		}
	}

	appTokenSource, err := githubauth.NewApplicationTokenSource(c.ClientID, privateKey)
	if err != nil {
		return nil, fmt.Errorf("failed to create GitHub App application token source: %w", err)
	}

	ops := make([]githubauth.InstallationTokenSourceOpt, 0, 1)
	if c.ApiURL != "" {
		ops = append(ops, githubauth.WithEnterpriseURL(c.ApiURL))
	}

	installationTokenSource := githubauth.NewInstallationTokenSource(c.InstallationID, appTokenSource, ops...)

	return &GitHubAppCredentials{
		logger:         logger,
		clientID:       c.ClientID,
		installationID: c.InstallationID,
		apiURL:         c.ApiURL,
		tokenSource:    installationTokenSource,
	}, nil
}

// TokenSource returns the oauth2.TokenSource for GitHub App authentication.
func (g *GitHubAppCredentials) TokenSource() oauth2.TokenSource {
	return g.tokenSource
}

// ClientID returns the GitHub App client ID.
func (g *GitHubAppCredentials) ClientID() string {
	return g.clientID
}

// InstallationID returns the GitHub App installation ID.
func (g *GitHubAppCredentials) InstallationID() int64 {
	return g.installationID
}

func (*GitHubAppCredentials) Name() string {
	return "http-github-app"
}

func (g *GitHubAppCredentials) String() string {
	return fmt.Sprintf("%s - %s/%d", g.Name(), g.ClientID(), g.InstallationID())
}

func (g *GitHubAppCredentials) Authorizer(r *http.Request) error {
	token, err := g.TokenSource().Token()
	if err != nil {
		g.logger.Error("failed to get GitHub App token", zap.Error(err))
		return err
	}
	if !token.Valid() {
		g.logger.Error("GitHub App token is invalid", zap.String("token_type", token.Type()))
		return fmt.Errorf("token is invalid")
	}
	r.SetBasicAuth("x-token-auth", token.AccessToken)
	return nil
}

// APIAuth represents different ways to authenticate with SCM APIs.
// For GitHub App authentication, use the GitHubAppCredentials method.
type APIAuth struct {
	// Type of authentication
	typ config.CredentialType
	// Token for Bearer token authentication (GitHub, GitLab, Gitea)
	Token string
	// Username and Password for basic authentication
	Username string
	Password string
	// GitHubAppCreds holds GitHub App credentials for CredentialTypeGithubApp
	githubAppCreds *GitHubAppCredentials
}

// Type returns the credential type.
func (a *APIAuth) Type() config.CredentialType {
	return a.typ
}

// GitHubAppTokenSource returns the oauth2.TokenSource for GitHub App authentication.
// Returns nil if the auth type is not CredentialTypeGithubApp.
func (a *APIAuth) GitHubAppTokenSource() oauth2.TokenSource {
	if a.githubAppCreds == nil {
		return &failingTokenSource{}
	}
	return a.githubAppCreds.TokenSource()
}

// failingTokenSource is a mock token source that always returns an error
type failingTokenSource struct{}

func (f *failingTokenSource) Token() (*oauth2.Token, error) {
	return nil, fmt.Errorf("token source error")
}

// APIAuthentication returns authentication information for SCM API operations.
// This provides a clean abstraction for different authentication methods.
// An access token configured as a secret reference is resolved when this is
// called, so SCM clients built from the result keep the value read at that time.
func (c *Credential) APIAuthentication(ctx context.Context) (*APIAuth, error) {
	switch c.config.Type {
	case config.CredentialTypeBasic:
		return &APIAuth{
			typ:      c.config.Type,
			Username: c.config.Basic.Username,
			Password: c.config.Basic.Password,
		}, nil
	case config.CredentialTypeAccessToken:
		token, err := c.accessToken(ctx)
		if err != nil {
			return nil, err
		}

		return &APIAuth{
			typ:   c.config.Type,
			Token: token,
		}, nil
	case config.CredentialTypeSSH:
		// SSH is not used for API operations, return empty auth
		return &APIAuth{
			typ: c.config.Type,
		}, nil
	case config.CredentialTypeGithubApp:
		creds, err := NewGitHubAppCredentials(c.logger, c.config.GitHubApp)
		if err != nil {
			return nil, err
		}
		return &APIAuth{
			typ:            c.config.Type,
			githubAppCreds: creds,
		}, nil
	}
	return &APIAuth{}, nil
}

// Type returns the credential type.
func (c *Credential) Type() config.CredentialType {
	return c.config.Type
}

// SSHUser returns the SSH user configured in the credential.
// Returns empty string if the credential is not SSH type.
func (c *Credential) SSHUser() string {
	if c.config.Type != config.CredentialTypeSSH || c.config.SSH == nil {
		return ""
	}
	return c.config.SSH.User
}
