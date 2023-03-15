package config

import (
	"github.com/spf13/viper"
)

// AuditSinkConfig contains fields, which enable and configure
// Flipt's various audit sink mechanisms.
type AuditSinkConfig struct {
	Sinks    SinksConfig    `json:"sinks,omitempty" mapstructure:"sinks"`
	Advanced AdvancedConfig `json:"advanced,omitempty" mapstructure:"advanced"`
}

func (c *AuditSinkConfig) setDefaults(v *viper.Viper) {
	v.SetDefault("audit", map[string]any{
		"sinks": map[string]any{
			"logFile": map[string]any{
				"enabled":  "false",
				"filePath": "./path/to/log/file",
			},
		},
		"advanced": map[string]any{
			"bufferSize": 2,
		},
	})
}

// SinksConfig contains configuration held in structures for the different sinks
// that we will send audits to.
type SinksConfig struct {
	LogFile LogFileSinkConfig `json:"logFile,omitempty" mapstructure:"logFile"`
}

// LogFileSinkConfig contains fields that hold configuration for sending audits
// to a log file.
type LogFileSinkConfig struct {
	Enabled  bool   `json:"enabled,omitempty" mapstructure:"enabled"`
	FilePath string `json:"filePath,omitempty" mapstructure:"filePath"`
}

// AdvancedConfig holds configuration for advanced configuration of sending the audit
// events to the sinks.
type AdvancedConfig struct {
	BufferSize int `json:"bufferSize,omitempty" mapstructure:"bufferSize"`
}
