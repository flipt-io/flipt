package config

import (
	"github.com/spf13/viper"
)

// ExperimentalConfig allows for experimental features to be enabled
// and disabled.
type ExperimentalConfig struct {
	Cloud ExperimentalFlag `json:"cloud,omitempty" mapstructure:"cloud" yaml:"cloud,omitempty"`
}

func (c *ExperimentalConfig) deprecations(v *viper.Viper) []deprecated {
	var deprecations []deprecated

	if v.InConfig("experimental.filesystem_storage") {
		deprecations = append(deprecations, "experimental.filesystem_storage")
	}

	return deprecations
}

// ExperimentalFlag is a structure which has properties to configure
// experimental feature enablement.
type ExperimentalFlag struct {
	Enabled bool `json:"enabled,omitempty" mapstructure:"enabled" yaml:"enabled,omitempty"`
}
