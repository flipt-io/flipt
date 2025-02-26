package config

import (
	"errors"
	"strings"

	"github.com/spf13/viper"
)

// ExperimentalConfig allows for experimental features to be enabled
// and disabled.
type ExperimentalConfig struct {
	Jira Jira `json:"jira,omitempty" mapstructure:"jira" yaml:"jira,omitempty"`
}

func (c *ExperimentalConfig) deprecations(v *viper.Viper) []deprecated {
	return nil
}

func (c *ExperimentalConfig) validate() error {
	return c.Jira.validate()
}

// ExperimentalFlag is a structure which has properties to configure
// experimental feature enablement.
type ExperimentalFlag struct {
	Enabled bool `json:"enabled,omitempty" mapstructure:"enabled" yaml:"enabled,omitempty"`
}

type Jira struct {
	Enabled        bool               `json:"enabled,omitempty" mapstructure:"enabled" yaml:"enabled,omitempty"`
	InstanceName   string             `json:"instanceName,omitempty" mapstructure:"instance_name" yaml:"instance_name,omitempty"`
	Authentication JiraAuthentication `json:"authentication,omitempty" mapstructure:"authentication" yaml:"authentication,omitempty"`
}

func (j *Jira) validate() error {
	if j.Enabled {
		if strings.TrimSpace(j.InstanceName) == "" {
			return errors.New("instance name cannot be empty when jira integration is enabled")
		}

		if strings.TrimSpace(j.Authentication.OAuth.ClientID) == "" || strings.TrimSpace(j.Authentication.OAuth.ClientSecret) == "" {
			return errors.New("invalid oauth parameters for jira integration")
		}
	}

	return nil
}

type JiraAuthentication struct {
	OAuth JiraOauth `json:"oauth,omitempty" mapstructure:"oauth" yaml:"oauth,omitempty"`
}

type JiraOauth struct {
	ClientID     string `json:"-" mapstructure:"client_id" yaml:"-"`
	ClientSecret string `json:"-" mapstructure:"client_secret" yaml:"-"`
}
