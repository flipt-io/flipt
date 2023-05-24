package config

import (
	"errors"

	"github.com/spf13/viper"
)

type StorageType string

const (
	DatabaseStorageType = StorageType("database")
	LocalStorageType    = StorageType("local")
	GitStorageType      = StorageType("git")
)

// StorageConfig contains fields which will configure the type of backend in which Flipt will serve
// flag state.
type StorageConfig struct {
	Type  StorageType `json:"type,omitempty" mapstructure:"type"`
	Local Local       `json:"local,omitempty" mapstructure:"local"`
	Git   Git         `json:"git,omitempty" mapstructure:"git"`
}

func (c *StorageConfig) setDefaults(v *viper.Viper) {
	switch v.GetString("storage.type") {
	case string(LocalStorageType):
		v.SetDefault("storage.local.path", ".")
	case string(GitStorageType):
		v.SetDefault("storage.git.ref", "main")
	default:
		v.SetDefault("storage.type", "database")
	}
}

func (c *StorageConfig) validate() error {
	if c.Type == GitStorageType {
		if c.Git.Ref == "" {
			return errors.New("git ref must be specified")
		}
		if c.Git.Repository == "" {
			return errors.New("git repository must be specified")
		}
	}

	if c.Type == LocalStorageType {
		if c.Local.Path == "" {
			return errors.New("local path must be specified")
		}
	}

	return nil
}

// Local contains configuration for referencing a local filesystem.
type Local struct {
	Path string `json:"path,omitempty" mapstructure:"path"`
}

// Git contains configuration for referencing a git repository.
type Git struct {
	Repository string    `json:"repository,omitempty" mapstructure:"repository"`
	Ref        string    `json:"ref,omitempty" mapstructure:"ref"`
	BasicAuth  BasicAuth `json:"basicAuth,omitempty" mapstructure:"basic_auth"`
	TokenAuth  TokenAuth `json:"tokenAuth,omitempty" mapstructure:"token_auth"`
}

// BasicAuth has configuration for authenticating with private git repositories
// with basic auth.
type BasicAuth struct {
	Username string `json:"username,omitempty" mapstructure:"username"`
	Password string `json:"password,omitempty" mapstructure:"password"`
}

// TokenAuth has configuration for authenticating with private git repositories
// with token auth.
type TokenAuth struct {
	Token string `json:"token,omitempty" mapstructure:"token"`
}
