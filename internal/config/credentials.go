package config

import (
	"fmt"

	"github.com/spf13/viper"
)

var (
	_ validator = (*CredentialsConfig)(nil)
	_ defaulter = (*CredentialsConfig)(nil)
)

type CredentialsConfig map[string]*CredentialConfig

func (c *CredentialsConfig) validate() error {
	for name, v := range *c {
		if err := v.validate(); err != nil {
			return errFieldWrap("credentials", name, err)
		}
	}

	return nil
}

func (c *CredentialsConfig) setDefaults(v *viper.Viper) error {
	credentials, ok := v.AllSettings()["credentials"].(map[string]any)
	if !ok {
		return nil
	}

	for name := range credentials {
		getString := func(s string) string {
			return v.GetString("credentials." + name + "." + s)
		}

		setDefault := func(k string, s any) {
			v.SetDefault("credentials."+name+"."+k, s)
		}

		if getString("type") == string(CredentialTypeSSH) {
			if getString("ssh.password") != "" ||
				getString("ssh.private_key_path") != "" ||
				getString("ssh.private_key_bytes") != "" {
				setDefault("ssh.user", "git")
			}
		}
	}

	return nil
}

type CredentialType string

const (
	CredentialTypeBasic       = CredentialType("basic")
	CredentialTypeSSH         = CredentialType("ssh")
	CredentialTypeAccessToken = CredentialType("access_token")
)

type CredentialConfig struct {
	Type        CredentialType   `json:"type,omitempty" mapstructure:"type" yaml:"type,omitempty"`
	Basic       *BasicAuthConfig `json:"basic,omitempty" mapstructure:"basic" yaml:"basic,omitempty"`
	SSH         *SSHAuthConfig   `json:"-" mapstructure:"ssh,omitempty" yaml:"-"`
	AccessToken *string          `json:"-" mapstructure:"access_token" yaml:"-"`
}

func (c *CredentialConfig) validate() error {
	switch c.Type {
	case CredentialTypeBasic:
		if c.Basic == nil {
			return errFieldRequired("credentials", "basic auth configuration")
		}

		if err := c.Basic.validate(); err != nil {
			return errFieldWrap("credentials", "basic", err)
		}
	case CredentialTypeSSH:
		if c.SSH == nil {
			return errFieldRequired("credentials", "ssh configuration")
		}

		if err := c.SSH.validate(); err != nil {
			return errFieldWrap("credentials", "ssh", err)
		}
	case CredentialTypeAccessToken:
		if c.AccessToken == nil || *c.AccessToken == "" {
			return errFieldRequired("credentials", "access token")
		}
	default:
		return errFieldWrap("credentials", "type", fmt.Errorf("unexpected credential type %q", c.Type))
	}

	return nil
}

// BasicAuthConfig has configuration for authenticating with private git repositories
// with basic auth.
type BasicAuthConfig struct {
	Username string `json:"-" mapstructure:"username" yaml:"-"`
	Password string `json:"-" mapstructure:"password" yaml:"-"`
}

func (b BasicAuthConfig) validate() error {
	if (b.Username != "" && b.Password == "") || (b.Username == "" && b.Password != "") {
		return errFieldRequired("credentials", "both username and password for basic auth")
	}

	return nil
}

// SSHAuthConfig provides configuration support for SSH private key credentials when
// authenticating with private git repositories
type SSHAuthConfig struct {
	User                  string `json:"-" mapstructure:"user" yaml:"-" `
	Password              string `json:"-" mapstructure:"password" yaml:"-" `
	PrivateKeyBytes       string `json:"-" mapstructure:"private_key_bytes" yaml:"-" `
	PrivateKeyPath        string `json:"-" mapstructure:"private_key_path" yaml:"-" `
	InsecureIgnoreHostKey bool   `json:"-" mapstructure:"insecure_ignore_host_key" yaml:"-"`
}

func (a *SSHAuthConfig) validate() (err error) {
	defer func() {
		if err != nil {
			err = errFieldWrap("credentials", "ssh authentication", err)
		}
	}()

	if a == nil {
		return errFieldRequired("credentials", "ssh configuration is missing")
	}

	if a.Password == "" {
		return errFieldRequired("credentials", "password")
	}

	if (a.PrivateKeyBytes == "" && a.PrivateKeyPath == "") || (a.PrivateKeyBytes != "" && a.PrivateKeyPath != "") {
		return errFieldWrap("credentials", "private_key", fmt.Errorf("please provide exclusively one of private_key_bytes or private_key_path"))
	}

	return nil
}
