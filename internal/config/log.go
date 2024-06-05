package config

import (
	"encoding/json"

	"github.com/spf13/viper"
)

var (
	_ defaulter = (*LogConfig)(nil)
)

// LogConfig contains fields which control, direct and filter
// the logging telemetry produces by Flipt.
type LogConfig struct {
	Level     string      `json:"level,omitempty" mapstructure:"level" yaml:"level,omitempty"`
	File      string      `json:"file,omitempty" mapstructure:"file" yaml:"file,omitempty"`
	Encoding  LogEncoding `json:"encoding,omitempty" mapstructure:"encoding" yaml:"encoding,omitempty"`
	GRPCLevel string      `json:"grpcLevel,omitempty" mapstructure:"grpc_level" yaml:"grpc_level,omitempty"`
	Keys      LogKeys     `json:"keys,omitempty" mapstructure:"keys" yaml:"-"`
}

type LogKeys struct {
	Time    string `json:"time,omitempty" mapstructure:"time" yaml:"time,omitempty"`
	Level   string `json:"level,omitempty" mapstructure:"level" yaml:"level,omitempty"`
	Message string `json:"message,omitempty" mapstructure:"message" yaml:"message,omitempty"`
}

func (c *LogConfig) setDefaults(v *viper.Viper) error {
	v.SetDefault("log", map[string]any{
		"level":      "INFO",
		"encoding":   "console",
		"grpc_level": "ERROR",
		"keys": map[string]any{
			"time":    "T",
			"level":   "L",
			"message": "M",
		},
	})

	return nil
}

// LogEncoding is either console or JSON.
type LogEncoding string

const (
	LogEncodingConsole = LogEncoding("console")
	LogEncodingJSON    = LogEncoding("json")
)

func (e LogEncoding) String() string {
	return string(e)
}

func (e LogEncoding) MarshalJSON() ([]byte, error) {
	return json.Marshal(e.String())
}

func (e LogEncoding) MarshalYAML() (interface{}, error) {
	return e.String(), nil
}
