package config

import (
	"encoding/json"
	"os"

	"github.com/mitchellh/mapstructure"
	"github.com/spf13/viper"
)

var serverDecodeHooks = mapstructure.ComposeDecodeHookFunc(
	mapstructure.StringToTimeDurationHookFunc(),
	mapstructure.StringToSliceHookFunc(","),
	StringToEnumHookFunc(stringToScheme),
)

// ServerConfig contains fields, which configure both HTTP and gRPC
// API serving.
type ServerConfig struct {
	Host      string `json:"host,omitempty" mapstructure:"host"`
	Protocol  Scheme `json:"protocol,omitempty" mapstructure:"protocol"`
	HTTPPort  int    `json:"httpPort,omitempty" mapstructure:"http_port"`
	HTTPSPort int    `json:"httpsPort,omitempty" mapstructure:"https_port"`
	GRPCPort  int    `json:"grpcPort,omitempty" mapstructure:"grpc_port"`
	CertFile  string `json:"certFile,omitempty" mapstructure:"cert_file"`
	CertKey   string `json:"certKey,omitempty" mapstructure:"cert_key"`
}

func (c *ServerConfig) viperKey() string {
	return "server"
}

func (c *ServerConfig) unmarshalViper(v *viper.Viper) (warnings []string, err error) {
	v.SetDefault("host", "0.0.0.0")
	v.SetDefault("protocol", HTTP)
	v.SetDefault("http_port", 8080)
	v.SetDefault("https_port", 443)
	v.SetDefault("grpc_port", 9000)

	if err = v.Unmarshal(c, viper.DecodeHook(serverDecodeHooks)); err != nil {
		return
	}

	// validate configuration is as expected
	if c.Protocol == HTTPS {
		if c.CertFile == "" {
			return nil, errFieldRequired("server.cert_file")
		}

		if c.CertKey == "" {
			return nil, errFieldRequired("server.cert_key")
		}

		if _, err := os.Stat(c.CertFile); err != nil {
			return nil, errFieldWrap("server.cert_file", err)
		}

		if _, err := os.Stat(c.CertKey); err != nil {
			return nil, errFieldWrap("server.cert_key", err)
		}
	}

	return
}

type Scheme uint

func (s Scheme) String() string {
	return schemeToString[s]
}

func (s Scheme) MarshalJSON() ([]byte, error) {
	return json.Marshal(s.String())
}

const (
	HTTP Scheme = iota
	HTTPS
)

var (
	schemeToString = map[Scheme]string{
		HTTP:  "http",
		HTTPS: "https",
	}

	stringToScheme = map[string]Scheme{
		"http":  HTTP,
		"https": HTTPS,
	}
)
