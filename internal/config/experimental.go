package config

import (
	"github.com/spf13/viper"
)

// ExperimentalConfig allows for experimental features to be enabled
// and disabled.
type ExperimentalConfig struct {
}

func (c *ExperimentalConfig) deprecations(v *viper.Viper) []deprecated {
	return nil
}

// ExperimentalFlag is a structure which has properties to configure
// experimental feature enablement.
type ExperimentalFlag struct {
	Enabled bool `json:"enabled,omitempty" mapstructure:"enabled" yaml:"enabled,omitempty"`
}
