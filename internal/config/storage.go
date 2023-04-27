package config

import "github.com/spf13/viper"

// cheers up the unparam linter
var _ defaulter = (*StorageConfig)(nil)

type StorageType string

const (
	DatabaseStorageType = StorageType("database")
	LocalStorageType    = StorageType("local")
	GitStorageType      = StorageType("git")
)

type StorageConfig struct {
	Type  StorageType `json:"type,omitempty" mapstructure:"type"`
	Local struct {
		Path string `json:"path,omitempty" mapstructure:"path"`
	} `json:"local,omitempty" mapstructure:"local"`
	Git struct {
		Repository string `json:"repository,omitempty" mapstructure:"repository"`
	} `json:"git,omitempty" mapstructure:"git"`
}

func (c *StorageConfig) setDefaults(v *viper.Viper) {
	v.SetDefault("type", "database")

	switch v.GetString("type") {
	case string(LocalStorageType):
		v.SetDefault("local", map[string]any{
			"path": ".",
		})
	}
}
