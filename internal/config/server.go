package config

import (
	"encoding/json"
	"os"
	"time"

	"github.com/spf13/viper"
)

// cheers up the unparam linter
var _ defaulter = (*ServerConfig)(nil)

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
	Cloud                     CloudConfig   `json:"cloud,omitempty" mapstructure:"cloud" yaml:"cloud,omitempty"`
}

func (c *ServerConfig) setDefaults(v *viper.Viper) error {
	v.SetDefault("server", map[string]any{
		"host":       "0.0.0.0",
		"protocol":   HTTP,
		"http_port":  8080,
		"https_port": 443,
		"grpc_port":  9000,
	})

	return c.Cloud.setDefaults(v)
}

func (c *ServerConfig) validate() error {
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

	return c.Cloud.validate()
}

// Scheme is either HTTP or HTTPS.
// TODO: can we use a string instead?
type Scheme uint

func (s Scheme) String() string {
	return schemeToString[s]
}

func (s Scheme) MarshalJSON() ([]byte, error) {
	return json.Marshal(s.String())
}

func (s Scheme) MarshalYAML() (interface{}, error) {
	return s.String(), nil
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

type CloudConfig struct {
	Enabled        bool                      `json:"enabled,omitempty" mapstructure:"enabled" yaml:"enabled"`
	Authentication CloudAuthenticationConfig `json:"authentication,omitempty" mapstructure:"authentication" yaml:"authentication,omitempty"`
	Address        string                    `json:"address,omitempty" mapstructure:"address" yaml:"address,omitempty"`
	Port           int                       `json:"port,omitempty" mapstructure:"port" yaml:"port,omitempty"`
	Organization   string                    `json:"organization,omitempty" mapstructure:"organization" yaml:"organization,omitempty"`
	Instance       string                    `json:"instance,omitempty" mapstructure:"instance" yaml:"instance,omitempty"`
}

type CloudAuthenticationConfig struct {
	ApiKey string `json:"-" mapstructure:"api_key" yaml:"api_key,omitempty"`
}

func (c *CloudAuthenticationConfig) validate() error {
	if c.ApiKey == "" {
		return errFieldRequired("server.cloud.authentication.api_key")
	}

	return nil
}

func (c *CloudConfig) setDefaults(v *viper.Viper) error {
	v.SetDefault("server.cloud", map[string]any{
		"enabled": false,
		"address": "flipt.cloud",
		"port":    8443,
	})

	return nil
}

func (c *CloudConfig) validate() error {
	// validate configuration is as expected
	if c.Enabled {
		if c.Address == "" {
			return errFieldRequired("server.cloud.address")
		}

		if c.Port == 0 {
			return errFieldRequired("server.cloud.port")
		}

		if c.Organization == "" {
			return errFieldRequired("server.cloud.organization")
		}

		if c.Instance == "" {
			return errFieldRequired("server.cloud.instance")
		}

		if c.Authentication.ApiKey == "" {
			return errFieldRequired("server.cloud.authentication.api_key")
		}
	}

	return nil
}
