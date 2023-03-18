package config

import (
	"errors"
	"fmt"
	"time"

	"github.com/spf13/viper"
)

var (
	allowedVersions = []string{"v0.1"}
)

// AuditSinkConfig contains fields, which enable and configure
// Flipt's various audit sink mechanisms.
type AuditSinkConfig struct {
	Version string       `json:"version" mapstructure:"version"`
	Sinks   SinksConfig  `json:"sinks,omitempty" mapstructure:"sinks"`
	Buffer  BufferConfig `json:"buffer,omitempty" mapstructure:"buffer"`
}

func (c *AuditSinkConfig) setDefaults(v *viper.Viper) {
	v.SetDefault("audit", map[string]any{
		"version": "v0.1",
		"sinks": map[string]any{
			"logFile": map[string]any{
				"enabled":  "false",
				"filePath": "./path/to/log/file",
			},
		},
		"buffer": map[string]any{
			"capacity":    2,
			"flushPeriod": "2m",
		},
	})
}

func isVersionSupported(version string) bool {
	for _, v := range allowedVersions {
		if version == v {
			return true
		}
	}

	return false
}

func (c *AuditSinkConfig) validate() error {
	if !isVersionSupported(c.Version) {
		return fmt.Errorf("unrecognized version %s", c.Version)
	}

	if c.Buffer.Capacity < 2 || c.Buffer.Capacity > 10 {
		return errors.New("buffer capacity below 2 or above 10")
	}

	if c.Buffer.FlushPeriod < 2*time.Minute || c.Buffer.FlushPeriod > 5*time.Minute {
		return errors.New("flush period below 2 minutes or greater than 5 minutes")
	}

	return nil
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
	Capacity    int           `json:"capacity,omitempty" mapstructure:"capacity"`
	FlushPeriod time.Duration `json:"flushPeriod,omitempty" mapstructure:"flushPeriod"`
}
