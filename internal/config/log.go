package config

import "encoding/json"

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

// LogConfig contains fields which control, direct and filter
// the logging telemetry produces by Flipt.
type LogConfig struct {
	Level     string      `json:"level,omitempty"`
	File      string      `json:"file,omitempty"`
	Encoding  LogEncoding `json:"encoding,omitempty"`
	GRPCLevel string      `json:"grpc_level,omitempty"`
}

// LogEncoding is either console or JSON
type LogEncoding uint8

func (e LogEncoding) String() string {
	return logEncodingToString[e]
}

func (e LogEncoding) MarshalJSON() ([]byte, error) {
	return json.Marshal(e.String())
}
