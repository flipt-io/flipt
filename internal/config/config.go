package config

import (
	"encoding/json"
	"fmt"
	"net/http"
	"reflect"
	"strings"
	"time"

	"github.com/mitchellh/mapstructure"
	"github.com/spf13/viper"
	"golang.org/x/exp/constraints"

	jaeger "github.com/uber/jaeger-client-go"
)

type Config struct {
	Log      LogConfig      `json:"log,omitempty"`
	UI       UIConfig       `json:"ui,omitempty"`
	Cors     CorsConfig     `json:"cors,omitempty"`
	Cache    CacheConfig    `json:"cache,omitempty"`
	Server   ServerConfig   `json:"server,omitempty"`
	Tracing  TracingConfig  `json:"tracing,omitempty"`
	Database DatabaseConfig `json:"database,omitempty"`
	Meta     MetaConfig     `json:"meta,omitempty"`
	Warnings []string       `json:"warnings,omitempty"`
}

func Default() *Config {
	return &Config{
		Cors: CorsConfig{
			Enabled:        false,
			AllowedOrigins: []string{"*"},
		},

		Cache: CacheConfig{
			Enabled: false,
			Backend: CacheMemory,
			TTL:     1 * time.Minute,
			Memory: MemoryCacheConfig{
				EvictionInterval: 5 * time.Minute,
			},
			Redis: RedisCacheConfig{
				Host:     "localhost",
				Port:     6379,
				Password: "",
				DB:       0,
			},
		},

		Server: ServerConfig{
			Host:      "0.0.0.0",
			Protocol:  HTTP,
			HTTPPort:  8080,
			HTTPSPort: 443,
			GRPCPort:  9000,
		},

		Tracing: TracingConfig{
			Jaeger: JaegerTracingConfig{
				Enabled: false,
				Host:    jaeger.DefaultUDPSpanServerHost,
				Port:    jaeger.DefaultUDPSpanServerPort,
			},
		},

		Database: DatabaseConfig{
			URL:            "file:/var/opt/flipt/flipt.db",
			MigrationsPath: "/etc/flipt/config/migrations",
			MaxIdleConn:    2,
		},

		Meta: MetaConfig{
			CheckForUpdates:  true,
			TelemetryEnabled: true,
			StateDirectory:   "",
		},
	}
}

func Load(path string) (*Config, error) {
	viper.SetEnvPrefix("FLIPT")
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	viper.AutomaticEnv()

	viper.SetConfigFile(path)

	if err := viper.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("loading configuration: %w", err)
	}

	cfg := Default()
	for _, initializer := range []interface {
		init() (warnings []string, err error)
	}{
		&cfg.Cors,
		&cfg.Cache,
		&cfg.Server,
		&cfg.Tracing,
		&cfg.Database,
		&cfg.Meta,
	} {
		warnings, err := initializer.init()
		if err != nil {
			return nil, err
		}

		cfg.Warnings = append(cfg.Warnings, warnings...)
	}

	for _, unmarshaller := range []interface {
		viperKey() string
		unmarshalViper(*viper.Viper) (warnings []string, err error)
	}{
		&cfg.Log,
		&cfg.UI,
	} {
		if v := viper.Sub(unmarshaller.viperKey()); v != nil {
			warnings, err := unmarshaller.unmarshalViper(v)
			if err != nil {
				return nil, err
			}

			cfg.Warnings = append(cfg.Warnings, warnings...)
		}
	}

	return cfg, nil
}

func (c *Config) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var (
		out []byte
		err error
	)

	if r.Header.Get("Accept") == "application/json+pretty" {
		out, err = json.MarshalIndent(c, "", "  ")
	} else {
		out, err = json.Marshal(c)
	}

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if _, err = w.Write(out); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}

// StringToEnumHookFunc returns a DecodeHookFunc that converts strings to a target enum
func StringToEnumHookFunc[T constraints.Integer](mappings map[string]T) mapstructure.DecodeHookFunc {
	return func(
		f reflect.Type,
		t reflect.Type,
		data interface{}) (interface{}, error) {
		if f.Kind() != reflect.String {
			return data, nil
		}
		if t != reflect.TypeOf(T(0)) {
			return data, nil
		}

		enum, _ := mappings[data.(string)]

		return enum, nil
	}
}
