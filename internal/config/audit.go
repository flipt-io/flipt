package config

import (
	"errors"
	"time"

	"github.com/spf13/viper"
)

// AuditConfig contains fields, which enable and configure
// Flipt's various audit sink mechanisms.
type AuditConfig struct {
	Sinks  SinksConfig  `json:"sinks,omitempty" mapstructure:"sinks"`
	Buffer BufferConfig `json:"buffer,omitempty" mapstructure:"buffer"`
}

// Enabled returns true if any nested sink is enabled
func (c *AuditConfig) Enabled() bool {
	return c.Sinks.LogFile.Enabled
}

func (c *AuditConfig) setDefaults(v *viper.Viper) {
	v.SetDefault("audit", map[string]any{
		"sinks": map[string]any{
			"log": map[string]any{
				"enabled": "false",
				"file":    "",
			},
		},
		"buffer": map[string]any{
			"capacity":     2,
			"flush_period": "2m",
		},
	})
}

func (c *AuditConfig) validate() error {
	if c.Sinks.LogFile.Enabled && c.Sinks.LogFile.File == "" {
		return errors.New("file not specified")
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
	LogFile LogFileSinkConfig `json:"log,omitempty" mapstructure:"log"`
}

// LogFileSinkConfig contains fields that hold configuration for sending audits
// to a log file.
type LogFileSinkConfig struct {
	Enabled bool   `json:"enabled,omitempty" mapstructure:"enabled"`
	File    string `json:"file,omitempty" mapstructure:"file"`
}

// BufferConfig holds configuration for the buffering of sending the audit
// events to the sinks.
type BufferConfig struct {
	Capacity    int           `json:"capacity,omitempty" mapstructure:"capacity"`
	FlushPeriod time.Duration `json:"flushPeriod,omitempty" mapstructure:"flush_period"`
}
