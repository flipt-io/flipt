package config

import (
	"github.com/spf13/viper"
)

type StorageType string

const (
	DatabaseStorageType = StorageType("database")
	LocalStorageType    = StorageType("local")
	GitStorageType      = StorageType("git")
)

// StorageConfig contains fields which will configure the type of backend in which Flipt will serve
// flag state.
type StorageConfig struct {
	Type  StorageType `json:"type,omitempty" mapstructure:"type"`
	Local Local       `json:"local,omitempty" mapstructure:"local"`
	Git   Git         `json:"git,omitempty" mapstructure:"git"`
}

func (c *StorageConfig) setDefaults(v *viper.Viper) {
	v.SetDefault("storage", map[string]any{
		"type": "database",
		"local": map[string]any{
			"path": "",
		},
		"git": map[string]any{
			"repository": "",
			"ref":        "",
		},
	})
}

type Local struct {
	Path string `json:"path,omitempty" mapstructure:"path"`
}

type Git struct {
	Repository string `json:"repository,omitempty" mapstructure:"repository"`
	Ref        string `json:"ref,omitempty" mapstructure:"ref"`
}
