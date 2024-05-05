package config

import (
	"strings"

	"github.com/spf13/viper"
)

type CloudConfig struct {
	Host         string `json:"host,omitempty" mapstructure:"host" yaml:"host,omitempty"`
	Organization string `json:"organization,omitempty" mapstructure:"organization" yaml:"organization,omitempty"`
	Instance     string `json:"instance,omitempty" mapstructure:"instance" yaml:"instance,omitempty"`
}

func (c *CloudConfig) setDefaults(v *viper.Viper) error {
	v.SetDefault("cloud", map[string]any{
		"host": "flipt.cloud",
	})

	return nil
}

func (c *CloudConfig) validate() error {
	// strip trailing slash if present
	c.Host = strings.TrimSuffix(c.Host, "/")
	return nil
}
