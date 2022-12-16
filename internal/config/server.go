package config

import (
	"encoding/json"
	"os"

	"github.com/spf13/viper"
)

// cheers up the unparam linter
var _ defaulter = (*ServerConfig)(nil)

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

func (c *ServerConfig) setDefaults(v *viper.Viper) {
	v.SetDefault("server", map[string]any{
		"host":       "0.0.0.0",
		"protocol":   HTTP,
		"http_port":  8080,
		"https_port": 443,
		"grpc_port":  9000,
	})
}

func (c *ServerConfig) validate() (err error) {
	// validate configuration is as expected
	if c.Protocol == HTTPS {
		if c.CertFile == "" {
			return errFieldRequired("server.cert_file")
		}

		if c.CertKey == "" {
			return errFieldRequired("server.cert_key")
		}

		if _, err := os.Stat(c.CertFile); err != nil {
			return errFieldWrap("server.cert_file", err)
		}

		if _, err := os.Stat(c.CertKey); err != nil {
			return errFieldWrap("server.cert_key", err)
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
