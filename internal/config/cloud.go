package config

import (
	"strings"

	"github.com/spf13/viper"
)

type CloudConfig struct {
	Host           string                    `json:"host,omitempty" mapstructure:"host" yaml:"host,omitempty"`
	Authentication CloudAuthenticationConfig `json:"-" mapstructure:"authentication" yaml:"authentication,omitempty"`
	Organization   string                    `json:"organization,omitempty" mapstructure:"organization" yaml:"organization,omitempty"`
	Gateway        string                    `json:"gateway,omitempty" mapstructure:"gateway" yaml:"gateway,omitempty"`
}

type CloudAuthenticationConfig struct {
	ApiKey string `json:"-" mapstructure:"api_key" yaml:"api_key,omitempty"`
}

//nolint:golint,unparam
func (c *CloudConfig) setDefaults(v *viper.Viper) error {
	v.SetDefault("cloud", map[string]any{
		"host": "flipt.cloud",
	})

	return nil
}

//nolint:golint,unparam
func (c *CloudConfig) validate() error {
	// strip trailing slash if present
	c.Host = strings.TrimSuffix(c.Host, "/")
	return nil
}
