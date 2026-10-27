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
	HTTPReadTimeout           time.Duration `json:"httpReadTimeout,omitempty" mapstructure:"http_read_timeout" yaml:"http_read_timeout,omitempty"`
	HTTPIdleTimeout           time.Duration `json:"httpIdleTimeout,omitempty" mapstructure:"http_idle_timeout" yaml:"http_idle_timeout,omitempty"`
}

func (c *ServerConfig) setDefaults(v *viper.Viper) error {
	v.SetDefault("server.host", "0.0.0.0")
	v.SetDefault("server.protocol", HTTP)
	v.SetDefault("server.http_port", 8080)
	v.SetDefault("server.https_port", 443)
	v.SetDefault("server.grpc_port", 9000)
	v.SetDefault("server.http_read_timeout", "10s")
	return nil
}

func (c *ServerConfig) validate() error {
	if c.HTTPReadTimeout < 0 {
		return errString("server", "http_read_timeout must not be negative")
	}

	if c.HTTPIdleTimeout < 0 {
		return errString("server", "http_idle_timeout must not be negative")
	}

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
