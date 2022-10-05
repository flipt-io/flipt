package config

import (
	"encoding/json"
	"os"

	"github.com/spf13/viper"
)

const (
	// configuration keys
	serverHost      = "server.host"
	serverProtocol  = "server.protocol"
	serverHTTPPort  = "server.http_port"
	serverHTTPSPort = "server.https_port"
	serverGRPCPort  = "server.grpc_port"
	serverCertFile  = "server.cert_file"
	serverCertKey   = "server.cert_key"
)

// ServerConfig contains fields, which configure both HTTP and gRPC
// API serving.
type ServerConfig struct {
	Host      string `json:"host,omitempty"`
	Protocol  Scheme `json:"protocol,omitempty"`
	HTTPPort  int    `json:"httpPort,omitempty"`
	HTTPSPort int    `json:"httpsPort,omitempty"`
	GRPCPort  int    `json:"grpcPort,omitempty"`
	CertFile  string `json:"certFile,omitempty"`
	CertKey   string `json:"certKey,omitempty"`
}

func (c *ServerConfig) init() (warnings []string, _ error) {
	// read in configuration via viper
	if viper.IsSet(serverHost) {
		c.Host = viper.GetString(serverHost)
	}

	if viper.IsSet(serverProtocol) {
		c.Protocol = stringToScheme[viper.GetString(serverProtocol)]
	}

	if viper.IsSet(serverHTTPPort) {
		c.HTTPPort = viper.GetInt(serverHTTPPort)
	}

	if viper.IsSet(serverHTTPSPort) {
		c.HTTPSPort = viper.GetInt(serverHTTPSPort)
	}

	if viper.IsSet(serverGRPCPort) {
		c.GRPCPort = viper.GetInt(serverGRPCPort)
	}

	if viper.IsSet(serverCertFile) {
		c.CertFile = viper.GetString(serverCertFile)
	}

	if viper.IsSet(serverCertKey) {
		c.CertKey = viper.GetString(serverCertKey)
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
