package config

import (
	"os"
)

type LicenseConfig struct {
	Key  string `json:"key,omitempty" mapstructure:"key" yaml:"key,omitempty"`
	File string `json:"file,omitempty" mapstructure:"file" yaml:"file,omitempty"`
}

func (c *LicenseConfig) validate() error {
	// key is required if file is provided as it is used to decrypt the license file
	if c.File != "" && c.Key == "" {
		return errFieldRequired("license", "key")
	}

	if c.File != "" {
		if _, err := os.Stat(c.File); err != nil {
			return errFieldWrap("license", "file", err)
		}
	}
	return nil
}
