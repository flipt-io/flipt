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

	defaults := 0
	for _, v := range *e {
		if v.Default {
			defaults++
		}
	}

	if defaults == 0 {
		return fmt.Errorf("no default environment configured")
	}

	if defaults > 1 {
		return fmt.Errorf("only one environment can be default")
	}

	return nil
}

func (e *EnvironmentsConfig) setDefaults(v *viper.Viper) error {
	envs, ok := v.AllSettings()["environments"].(map[string]any)
	if !ok || len(envs) == 0 {
		// Create default environment if no environments are configured
		v.SetDefault("environments.default.name", "default")
		v.SetDefault("environments.default.storage", "default")
		v.SetDefault("environments.default.default", true)
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

	// if there is only one environment, set it as the default
	if len(envs) == 1 {
		for name := range envs {
			v.SetDefault("environments."+name+".default", true)
		}
	}

	return nil
}

type EnvironmentConfig struct {
	Name      string `json:"name" mapstructure:"name" yaml:"name,omitempty"`
	Default   bool   `json:"default" mapstructure:"default" yaml:"default,omitempty"`
	Storage   string `json:"storage" mapstructure:"storage" yaml:"storage,omitempty"`
	Directory string `json:"directory" mapstructure:"directory" yaml:"directory,omitempty"`
}

func (e *EnvironmentConfig) validate() error {
	if e.Name == "" {
		return errFieldRequired("", "name")
	}
	return nil
}
