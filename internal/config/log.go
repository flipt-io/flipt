package config

import (
	"encoding/json"

	"github.com/mitchellh/mapstructure"
	"github.com/spf13/viper"
)

// LogConfig contains fields which control, direct and filter
// the logging telemetry produces by Flipt.
type LogConfig struct {
	Level     string      `json:"level,omitempty" mapstructure:"level"`
	File      string      `json:"file,omitempty" mapstructure:"file"`
	Encoding  LogEncoding `json:"encoding,omitempty" mapstructure:"encoding"`
	GRPCLevel string      `json:"grpc_level,omitempty" mapstructure:"grpc_level"`
}

var (
	logDecodeHooks = mapstructure.ComposeDecodeHookFunc(
		mapstructure.StringToTimeDurationHookFunc(),
		mapstructure.StringToSliceHookFunc(","),
		StringToEnumHookFunc(stringToLogEncoding),
	)
)

func (c *LogConfig) viperKey() string {
	return "log"
}

func (c *LogConfig) unmarshalViper(v *viper.Viper) (_ []string, _ error) {
	v.SetDefault("level", "INFO")
	v.SetDefault("encoding", "console")
	v.SetDefault("grpc_level", "ERROR")

	return nil, v.Unmarshal(c, viper.DecodeHook(logDecodeHooks))
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
