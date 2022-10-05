package config

import (
	"encoding/json"

	"github.com/spf13/viper"
)

const (
	// configuration keys
	logLevel     = "log.level"
	logFile      = "log.file"
	logEncoding  = "log.encoding"
	logGRPCLevel = "log.grpc_level"

	// encoding enum
	_ LogEncoding = iota
	LogEncodingConsole
	LogEncodingJSON
)

// LogConfig contains fields which control, direct and filter
// the logging telemetry produces by Flipt.
type LogConfig struct {
	Level     string      `json:"level,omitempty"`
	File      string      `json:"file,omitempty"`
	Encoding  LogEncoding `json:"encoding,omitempty"`
	GRPCLevel string      `json:"grpc_level,omitempty"`
}

func (c *LogConfig) init() (_ []string, _ error) {
	if viper.IsSet(logLevel) {
		c.Level = viper.GetString(logLevel)
	}

	if viper.IsSet(logFile) {
		c.File = viper.GetString(logFile)
	}

	if viper.IsSet(logEncoding) {
		c.Encoding = stringToLogEncoding[viper.GetString(logEncoding)]
	}

	if viper.IsSet(logGRPCLevel) {
		c.GRPCLevel = viper.GetString(logGRPCLevel)
	}

	return
}

var (
	logEncodingToString = map[LogEncoding]string{
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

func (e LogEncoding) String() string {
	return logEncodingToString[e]
}

func (e LogEncoding) MarshalJSON() ([]byte, error) {
	return json.Marshal(e.String())
}
