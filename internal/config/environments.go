package config

import (
	"fmt"

	"github.com/spf13/viper"
)

var (
	_ validator = (*EnvironmentsConfig)(nil)
	_ defaulter = (*EnvironmentsConfig)(nil)
)

type EnvironmentsConfig map[string]*EnvironmentConfig

func (e *EnvironmentsConfig) validate() error {
	for name, v := range *e {
		if err := v.validate(); err != nil {
			return fmt.Errorf("environment %q: %w", name, err)
		}
	}

	return nil
}

func (e *EnvironmentsConfig) setDefaults(v *viper.Viper) error {
	envs, ok := v.AllSettings()["environments"].(map[string]any)
	if !ok || len(envs) == 0 {
		// Create default environment if no environments are configured
		v.SetDefault("environments.default.name", "default")
		v.SetDefault("environments.default.storage", "default")
		return nil
	}

	for name := range envs {
		getString := func(s string) string {
			return v.GetString("environments." + name + "." + s)
		}

		setDefault := func(k string, val any) {
			v.SetDefault("environments."+name+"."+k, val)
		}

		setDefault("name", name)

		if getString("storage") == "" {
			setDefault("storage", "default")
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
