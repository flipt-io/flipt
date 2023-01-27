package config

import (
	"encoding/json"

	"github.com/spf13/viper"
)

// cheers up the unparam linter
var _ defaulter = (*LogConfig)(nil)

// LogConfig contains fields which control, direct and filter
// the logging telemetry produces by Flipt.
type LogConfig struct {
	Level     string      `json:"level,omitempty" mapstructure:"level"`
	File      string      `json:"file,omitempty" mapstructure:"file"`
	Encoding  LogEncoding `json:"encoding,omitempty" mapstructure:"encoding"`
	GRPCLevel string      `json:"grpcLevel,omitempty" mapstructure:"grpc_level"`
	Keys      LogKeys     `json:"keys" mapstructure:"keys"`
}

type LogKeys struct {
	Time    string `json:"time" mapstructure:"time"`
	Level   string `json:"level" mapstructure:"level"`
	Message string `json:"message" mapstructure:"message"`
}

func (c *LogConfig) setDefaults(v *viper.Viper) {
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
}

var (
	logEncodingToString = [...]string{
		LogEncodingConsole: "console",
		LogEncodingJSON:    "json",
	}

	stringToLogEncoding = map[string]LogEncoding{
		"console": LogEncodingConsole,
		"json":    LogEncodingJSON,
	}
)

// LogEncoding is either console or JSON
type LogEncoding uint8

const (
	_ LogEncoding = iota
	LogEncodingConsole
	LogEncodingJSON
)

func (e LogEncoding) String() string {
	return logEncodingToString[e]
}

func (e LogEncoding) MarshalJSON() ([]byte, error) {
	return json.Marshal(e.String())
}
