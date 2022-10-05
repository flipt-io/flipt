package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/spf13/viper"

	jaeger "github.com/uber/jaeger-client-go"
)

const (
	deprecatedMsgMemoryEnabled    = `'cache.memory.enabled' is deprecated and will be removed in a future version. Please use 'cache.backend' and 'cache.enabled' instead.`
	deprecatedMsgMemoryExpiration = `'cache.memory.expiration' is deprecated and will be removed in a future version. Please use 'cache.ttl' instead.`
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

func Default() *Config {
	return &Config{
		Log: LogConfig{
			Level:     "INFO",
			Encoding:  LogEncodingConsole,
			GRPCLevel: "ERROR",
		},

		UI: UIConfig{
			Enabled: true,
		},

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

	// Logging
	if viper.IsSet(logLevel) {
		cfg.Log.Level = viper.GetString(logLevel)
	}

	if viper.IsSet(logFile) {
		cfg.Log.File = viper.GetString(logFile)
	}

	if viper.IsSet(logEncoding) {
		cfg.Log.Encoding = stringToLogEncoding[viper.GetString(logEncoding)]
	}

	if viper.IsSet(logGRPCLevel) {
		cfg.Log.GRPCLevel = viper.GetString(logGRPCLevel)
	}

	// UI
	if viper.IsSet(uiEnabled) {
		cfg.UI.Enabled = viper.GetBool(uiEnabled)
	}

	// CORS
	if viper.IsSet(corsEnabled) {
		cfg.Cors.Enabled = viper.GetBool(corsEnabled)

		if viper.IsSet(corsAllowedOrigins) {
			cfg.Cors.AllowedOrigins = viper.GetStringSlice(corsAllowedOrigins)
		}
	}

	// Cache
	if viper.GetBool(cacheMemoryEnabled) { // handle deprecated memory config
		cfg.Cache.Backend = CacheMemory
		cfg.Cache.Enabled = true

		cfg.Warnings = append(cfg.Warnings, deprecatedMsgMemoryEnabled)

		if viper.IsSet(cacheMemoryExpiration) {
			cfg.Cache.TTL = viper.GetDuration(cacheMemoryExpiration)
			cfg.Warnings = append(cfg.Warnings, deprecatedMsgMemoryExpiration)
		}

	} else if viper.IsSet(cacheEnabled) {
		cfg.Cache.Enabled = viper.GetBool(cacheEnabled)
		if viper.IsSet(cacheBackend) {
			cfg.Cache.Backend = stringToCacheBackend[viper.GetString(cacheBackend)]
		}
		if viper.IsSet(cacheTTL) {
			cfg.Cache.TTL = viper.GetDuration(cacheTTL)
		}
	}

	if cfg.Cache.Enabled {
		switch cfg.Cache.Backend {
		case CacheRedis:
			if viper.IsSet(cacheRedisHost) {
				cfg.Cache.Redis.Host = viper.GetString(cacheRedisHost)
			}
			if viper.IsSet(cacheRedisPort) {
				cfg.Cache.Redis.Port = viper.GetInt(cacheRedisPort)
			}
			if viper.IsSet(cacheRedisPassword) {
				cfg.Cache.Redis.Password = viper.GetString(cacheRedisPassword)
			}
			if viper.IsSet(cacheRedisDB) {
				cfg.Cache.Redis.DB = viper.GetInt(cacheRedisDB)
			}
		case CacheMemory:
			if viper.IsSet(cacheMemoryEvictionInterval) {
				cfg.Cache.Memory.EvictionInterval = viper.GetDuration(cacheMemoryEvictionInterval)
			}
		}
	}

	// Server
	if viper.IsSet(serverHost) {
		cfg.Server.Host = viper.GetString(serverHost)
	}

	if viper.IsSet(serverProtocol) {
		cfg.Server.Protocol = stringToScheme[viper.GetString(serverProtocol)]
	}

	if viper.IsSet(serverHTTPPort) {
		cfg.Server.HTTPPort = viper.GetInt(serverHTTPPort)
	}

	if viper.IsSet(serverHTTPSPort) {
		cfg.Server.HTTPSPort = viper.GetInt(serverHTTPSPort)
	}

	if viper.IsSet(serverGRPCPort) {
		cfg.Server.GRPCPort = viper.GetInt(serverGRPCPort)
	}

	if viper.IsSet(serverCertFile) {
		cfg.Server.CertFile = viper.GetString(serverCertFile)
	}

	if viper.IsSet(serverCertKey) {
		cfg.Server.CertKey = viper.GetString(serverCertKey)
	}

	// Tracing
	if viper.IsSet(tracingJaegerEnabled) {
		cfg.Tracing.Jaeger.Enabled = viper.GetBool(tracingJaegerEnabled)

		if viper.IsSet(tracingJaegerHost) {
			cfg.Tracing.Jaeger.Host = viper.GetString(tracingJaegerHost)
		}

		if viper.IsSet(tracingJaegerPort) {
			cfg.Tracing.Jaeger.Port = viper.GetInt(tracingJaegerPort)
		}
	}

	// DB
	if viper.IsSet(dbURL) {
		cfg.Database.URL = viper.GetString(dbURL)

	} else if viper.IsSet(dbProtocol) || viper.IsSet(dbName) || viper.IsSet(dbUser) || viper.IsSet(dbPassword) || viper.IsSet(dbHost) || viper.IsSet(dbPort) {
		cfg.Database.URL = ""

		if viper.IsSet(dbProtocol) {
			cfg.Database.Protocol = stringToDatabaseProtocol[viper.GetString(dbProtocol)]
		}

		if viper.IsSet(dbName) {
			cfg.Database.Name = viper.GetString(dbName)
		}

		if viper.IsSet(dbUser) {
			cfg.Database.User = viper.GetString(dbUser)
		}

		if viper.IsSet(dbPassword) {
			cfg.Database.Password = viper.GetString(dbPassword)
		}

		if viper.IsSet(dbHost) {
			cfg.Database.Host = viper.GetString(dbHost)
		}

		if viper.IsSet(dbPort) {
			cfg.Database.Port = viper.GetInt(dbPort)
		}

	}

	if viper.IsSet(dbMigrationsPath) {
		cfg.Database.MigrationsPath = viper.GetString(dbMigrationsPath)
	}

	if viper.IsSet(dbMaxIdleConn) {
		cfg.Database.MaxIdleConn = viper.GetInt(dbMaxIdleConn)
	}

	if viper.IsSet(dbMaxOpenConn) {
		cfg.Database.MaxOpenConn = viper.GetInt(dbMaxOpenConn)
	}

	if viper.IsSet(dbConnMaxLifetime) {
		cfg.Database.ConnMaxLifetime = viper.GetDuration(dbConnMaxLifetime)
	}

	// Meta
	if viper.IsSet(metaCheckForUpdates) {
		cfg.Meta.CheckForUpdates = viper.GetBool(metaCheckForUpdates)
	}

	if viper.IsSet(metaTelemetryEnabled) {
		cfg.Meta.TelemetryEnabled = viper.GetBool(metaTelemetryEnabled)
	}

	if viper.IsSet(metaStateDirectory) {
		cfg.Meta.StateDirectory = viper.GetString(metaStateDirectory)
	}

	if err := cfg.validate(); err != nil {
		return &Config{}, err
	}

	return cfg, nil
}

func (c *Config) validate() error {
	if c.Server.Protocol == HTTPS {
		if c.Server.CertFile == "" {
			return errors.New("server.cert_file cannot be empty when using HTTPS")
		}

		if c.Server.CertKey == "" {
			return errors.New("server.cert_key cannot be empty when using HTTPS")
		}

		if _, err := os.Stat(c.Server.CertFile); os.IsNotExist(err) {
			return fmt.Errorf("cannot find TLS server.cert_file at %q", c.Server.CertFile)
		}

		if _, err := os.Stat(c.Server.CertKey); os.IsNotExist(err) {
			return fmt.Errorf("cannot find TLS server.cert_key at %q", c.Server.CertKey)
		}
	}

	if c.Database.URL == "" {
		if c.Database.Protocol == 0 {
			return fmt.Errorf("database.protocol cannot be empty")
		}

		if c.Database.Host == "" {
			return fmt.Errorf("database.host cannot be empty")
		}

		if c.Database.Name == "" {
			return fmt.Errorf("database.name cannot be empty")
		}
	}

	return nil
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
