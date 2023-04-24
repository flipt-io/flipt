package config

import "github.com/spf13/viper"

// cheers up the unparam linter
var _ defaulter = (*StorageConfig)(nil)

type StorageType string

const (
	DatabaseStorageType = StorageType("database")
	LocalStorageType    = StorageType("local")
)

type StorageConfig struct {
	Type  StorageType `json:"type,omitempty" mapstructure:"type"`
	Local struct {
		Path string `json:"path,omitempty" mapstructure:"path"`
	} `json:"local,omitempty" mapstructure:"local"`
}

func (c *StorageConfig) setDefaults(v *viper.Viper) {
	v.SetDefault("type", "database")

	if v.GetString("type") == "local" {
		v.SetDefault("local", map[string]any{
			"path": ".",
		})
	}
}
