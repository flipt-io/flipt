package config

import (
	"os"
	"time"

	"github.com/spf13/viper"
)

var (
	_ defaulter = (*ServerConfig)(nil)
	_ validator = (*ServerConfig)(nil)
)

// ServerConfig contains fields, which configure both HTTP and gRPC
// API serving.
type ServerConfig struct {
	Host                      string        `json:"host,omitempty" mapstructure:"host" yaml:"host,omitempty"`
	Protocol                  Scheme        `json:"protocol,omitempty" mapstructure:"protocol" yaml:"protocol,omitempty"`
	HTTPPort                  int           `json:"httpPort,omitempty" mapstructure:"http_port" yaml:"http_port,omitempty"`
	HTTPSPort                 int           `json:"httpsPort,omitempty" mapstructure:"https_port" yaml:"https_port,omitempty"`
	GRPCPort                  int           `json:"grpcPort,omitempty" mapstructure:"grpc_port" yaml:"grpc_port,omitempty"`
	CertFile                  string        `json:"-" mapstructure:"cert_file" yaml:"-"`
	CertKey                   string        `json:"-" mapstructure:"cert_key" yaml:"-"`
	GRPCConnectionMaxIdleTime time.Duration `json:"-" mapstructure:"grpc_conn_max_idle_time" yaml:"-"`
	GRPCConnectionMaxAge      time.Duration `json:"-" mapstructure:"grpc_conn_max_age" yaml:"-"`
	GRPCConnectionMaxAgeGrace time.Duration `json:"-" mapstructure:"grpc_conn_max_age_grace" yaml:"-"`
	IncludeFlagMetadata       bool          `json:"-" mapstructure:"include_flag_metadata" yaml:"include_flag_metadata,omitempty"`
}

func (c *ServerConfig) setDefaults(v *viper.Viper) error {
	v.SetDefault("server.host", "0.0.0.0")
	v.SetDefault("server.protocol", HTTP)
	v.SetDefault("server.http_port", 8080)
	v.SetDefault("server.https_port", 443)
	v.SetDefault("server.grpc_port", 9000)
	return nil
}

func (c *ServerConfig) validate() error {
	// validate configuration is as expected
	if c.Protocol == HTTPS {
		if c.CertFile == "" {
			return errFieldRequired("server", "cert_file")
		}

		if c.CertKey == "" {
			return errFieldRequired("server", "cert_key")
		}

		if _, err := os.Stat(c.CertFile); err != nil {
			return errFieldWrap("server", "cert_file", err)
		}

		if _, err := os.Stat(c.CertKey); err != nil {
			return errFieldWrap("server", "cert_key", err)
		}
	}

	return nil
}

// Scheme is either HTTP or HTTPS.
type Scheme string

func (s Scheme) String() string {
	return string(s)
}

const (
	HTTP  Scheme = "http"
	HTTPS Scheme = "https"
)
