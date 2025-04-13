package config

import (
	"fmt"
	"sort"
	"strings"

	"github.com/spf13/viper"
)

var (
	_ validator = (*EnvironmentsConfig)(nil)
	_ defaulter = (*EnvironmentsConfig)(nil)
)

type EnvironmentsConfig map[string]*EnvironmentConfig

func (e *EnvironmentsConfig) validate() error {
	// map of storage key to directories used
	storageDirectories := make(map[string]map[string][]string) // storage -> directory -> []environment names

	for name, v := range *e {
		if err := v.validate(); err != nil {
			return fmt.Errorf("environment %q: %w", name, err)
		}

		// Initialize map for this storage if not exists
		if _, ok := storageDirectories[v.Storage]; !ok {
			storageDirectories[v.Storage] = make(map[string][]string)
		}

		// Use empty string if directory not specified
		dir := v.Directory
		if dir == "" {
			dir = ""
		}

		// Add this environment to the map
		storageDirectories[v.Storage][dir] = append(storageDirectories[v.Storage][dir], name)
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

	// Check for duplicate directory usage within same storage
	for storage, directories := range storageDirectories {
		// If only one environment uses this storage, no need to check
		totalEnvs := 0
		for _, envs := range directories {
			totalEnvs += len(envs)
		}
		if totalEnvs <= 1 {
			continue
		}

		// If multiple environments use this storage and share empty directory
		if envs, exists := directories[""]; exists && len(envs) > 1 {
			sort.Strings(envs) // Sort for deterministic output
			return fmt.Errorf("environments [%s] share the same storage %q but have no distinct directory values", strings.Join(envs, ", "), storage)
		}

		// Check for any directory shared by multiple environments
		for dir, envs := range directories {
			if len(envs) > 1 {
				sort.Strings(envs) // Sort for deterministic output
				return fmt.Errorf("environments [%s] share the same storage %q and directory %q", strings.Join(envs, ", "), storage, dir)
			}
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
