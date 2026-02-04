package credentials

import (
	"fmt"
	"net/http"
	"os"

	"golang.org/x/oauth2"

	"github.com/go-git/go-git/v6/plumbing/transport"
	githttp "github.com/go-git/go-git/v6/plumbing/transport/http"
	gitssh "github.com/go-git/go-git/v6/plumbing/transport/ssh"
	"github.com/jferrl/go-githubauth"
	"go.flipt.io/flipt/errors"
	"go.flipt.io/flipt/internal/config"
	"go.uber.org/zap"
	"golang.org/x/crypto/ssh"
)

type CredentialSource struct {
	logger  *zap.Logger
	configs config.CredentialsConfig
}

func New(logger *zap.Logger, configs config.CredentialsConfig) *CredentialSource {
	return &CredentialSource{logger, configs}
}

func (s *CredentialSource) Get(name string) (*Credential, error) {
	config, ok := s.configs[name]
	if !ok {
		return nil, errors.ErrNotFoundf("credential %q", name)
	}

	return &Credential{logger: s.logger, config: config}, nil
}

type Credential struct {
	logger *zap.Logger
	config *config.CredentialConfig
}

// GitAuthentication returns the appropriate transport.AuthMethod for Git operations.
// This method handles the complexity of converting different credential types
// to the format expected by Git operations.
func (c *Credential) GitAuthentication() (auth transport.AuthMethod, err error) {
	switch c.config.Type {
	case config.CredentialTypeBasic:
		return &githttp.BasicAuth{
			Username: c.config.Basic.Username,
			Password: c.config.Basic.Password,
		}, nil
	case config.CredentialTypeSSH:
		var method *gitssh.PublicKeys
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
		}

		return method, nil
	case config.CredentialTypeAccessToken:
		// For Git operations, access tokens need to be converted to HTTP Basic Auth
		// Different providers have different conventions:
		// - GitLab: username="oauth2", password=token OR user={anynonemptystring}, password=token
		// - GitHub: username=token, password="" OR username={anynonemptystring}, password=token
		// - BitBucket: username="x-token-auth", password=token for repo access tokens
		// We'll use the BitBucket format as it requires a specific string and the others dont
		return &githttp.BasicAuth{
			Username: "x-token-auth",
			Password: *c.config.AccessToken,
		}, nil
	case config.CredentialTypeGithubApp:
		return NewGitHubAppCredentials(c.logger, c.config.GitHubApp)
	}

	return nil, fmt.Errorf("unexpected credential type: %q", c.config.Type)
}

var _ githttp.AuthMethod = (*GitHubAppCredentials)(nil)

// GitHubAppCredentials holds the configuration and token source for GitHub App authentication.
// This centralizes the GitHub App auth logic for both Git operations and API operations.
type GitHubAppCredentials struct {
	logger         *zap.Logger
	clientID       string
	installationID int64
	privateKey     []byte
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
		privateKey:     privateKey,
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

func (g *GitHubAppCredentials) SetAuth(r *http.Request) {
	token, err := g.TokenSource().Token()
	if err != nil {
		g.logger.Error("failed to get GitHub App token", zap.Error(err))
		return
	}
	if !token.Valid() {
		g.logger.Error("GitHub App token is invalid", zap.String("token_type", token.Type()))
		return
	}
	r.SetBasicAuth("x-token-auth", token.AccessToken)
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
func (c *Credential) APIAuthentication() (*APIAuth, error) {
	switch c.config.Type {
	case config.CredentialTypeBasic:
		return &APIAuth{
			typ:      c.config.Type,
			Username: c.config.Basic.Username,
			Password: c.config.Basic.Password,
		}, nil
	case config.CredentialTypeAccessToken:
		return &APIAuth{
			typ:   c.config.Type,
			Token: *c.config.AccessToken,
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
