package credentials

import (
	"fmt"

	"github.com/go-git/go-git/v6/plumbing/transport"
	githttp "github.com/go-git/go-git/v6/plumbing/transport/http"
	gitssh "github.com/go-git/go-git/v6/plumbing/transport/ssh"
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
		// - GitLab: username="oauth2", password=token
		// - GitHub: username=token, password="" OR username="x-access-token", password=token
		// We'll use the GitLab format as it's more widely supported
		return &githttp.BasicAuth{
			Username: "oauth2",
			Password: *c.config.AccessToken,
		}, nil
	}

	return nil, fmt.Errorf("unexpected credential type: %q", c.config.Type)
}

// APIAuth represents different ways to authenticate with SCM APIs
type APIAuth struct {
	// Type of authentication
	typ config.CredentialType
	// Token for Bearer token authentication (GitHub, GitLab, Gitea)
	Token string
	// Username and Password for basic authentication
	Username string
	Password string
}

// Type returns the credential type.
func (a *APIAuth) Type() config.CredentialType {
	return a.typ
}

// APIAuthentication returns authentication information for SCM API operations.
// This provides a clean abstraction for different authentication methods.
func (c *Credential) APIAuthentication() *APIAuth {
	switch c.config.Type {
	case config.CredentialTypeBasic:
		return &APIAuth{
			typ:      c.config.Type,
			Username: c.config.Basic.Username,
			Password: c.config.Basic.Password,
		}
	case config.CredentialTypeAccessToken:
		return &APIAuth{
			typ:   c.config.Type,
			Token: *c.config.AccessToken,
		}
	case config.CredentialTypeSSH:
		// SSH is not used for API operations, return empty auth
		return &APIAuth{
			typ: c.config.Type,
		}
	}
	return &APIAuth{}
}

// Type returns the credential type.
func (c *Credential) Type() config.CredentialType {
	return c.config.Type
}
