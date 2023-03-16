package config

import (
	"github.com/spf13/viper"
)

// AuditSinkConfig contains fields, which enable and configure
// Flipt's various audit sink mechanisms.
type AuditSinkConfig struct {
	Sinks  SinksConfig  `json:"sinks,omitempty" mapstructure:"sinks"`
	Buffer BufferConfig `json:"buffer,omitempty" mapstructure:"buffer"`
}

func (c *AuditSinkConfig) setDefaults(v *viper.Viper) {
	v.SetDefault("audit", map[string]any{
		"sinks": map[string]any{
			"logFile": map[string]any{
				"enabled":  "false",
				"filePath": "./path/to/log/file",
			},
		},
		"buffer": map[string]any{
			"capacity": 2,
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

// BufferConfig holds configuration for the buffering of sending the audit
// events to the sinks.
type BufferConfig struct {
	Capacity int `json:"capacity,omitempty" mapstructure:"capacity"`
}
