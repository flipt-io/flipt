package credentials

import (
	"fmt"

	"github.com/go-git/go-git/v5/plumbing/transport"
	githttp "github.com/go-git/go-git/v5/plumbing/transport/http"
	gitssh "github.com/go-git/go-git/v5/plumbing/transport/ssh"
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

func (c *Credential) Authentication() (auth transport.AuthMethod, err error) {
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
		return &githttp.TokenAuth{Token: *c.config.AccessToken}, nil
	}

	return nil, fmt.Errorf("unxpected credential type: %q", c.config.Type)
}
