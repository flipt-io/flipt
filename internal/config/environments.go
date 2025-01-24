package config

import (
	"fmt"
	"strings"

	"github.com/spf13/viper"
)

var (
	_ validator = (*EnvironmentsConfig)(nil)
	_ defaulter = (*EnvironmentsConfig)(nil)
)

type EnvironmentsConfig map[string]*EnvironmentConfig

//nolint:unparam
func (e *EnvironmentsConfig) setDefaults(v *viper.Viper) error {
	envs, ok := v.AllSettings()["environments"].(map[string]any)
	if !ok {
		return nil
	}

	for name := range envs {
		setDefault := func(k string, val any) {
			v.SetDefault(
				strings.Join([]string{"environments", name, k}, "."),
				val,
			)
		}

		setDefault("name", name)
	}

	return nil
}

func (e *EnvironmentsConfig) validate() error {
	for name, v := range *e {
		if err := v.validate(); err != nil {
			return fmt.Errorf("environment %q: %w", name, err)
		}
	}

	return nil
}

type EnvironmentConfig struct {
	Name      string `json:"name" mapstructure:"name" yaml:"name,omitempty"`
	Storage   string `json:"storage" mapstructure:"storage" yaml:"storage,omitempty"`
	Directory string `json:"directory" mapstructure:"directory" yaml:"directory,omitempty"`
}

func (e *EnvironmentConfig) validate() error {
	if e.Name == "" {
		return errFieldRequired("", "name")
	}
	return nil
}
