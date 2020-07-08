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
}

type LogConfig struct {
	Level string `json:"level,omitempty"`
	File  string `json:"file,omitempty"`
}

type UIConfig struct {
	Enabled bool `json:"enabled"`
}

type CorsConfig struct {
	Enabled        bool     `json:"enabled"`
	AllowedOrigins []string `json:"allowedOrigins,omitempty"`
}

type MemoryCacheConfig struct {
	Enabled          bool          `json:"enabled"`
	Expiration       time.Duration `json:"expiration,omitempty"`
	EvictionInterval time.Duration `json:"evictionInterval,omitempty"`
}

type CacheConfig struct {
	Memory MemoryCacheConfig `json:"memory,omitempty"`
}

type ServerConfig struct {
	Host      string `json:"host,omitempty"`
	Protocol  Scheme `json:"protocol,omitempty"`
	HTTPPort  int    `json:"httpPort,omitempty"`
	HTTPSPort int    `json:"httpsPort,omitempty"`
	GRPCPort  int    `json:"grpcPort,omitempty"`
	CertFile  string `json:"certFile,omitempty"`
	CertKey   string `json:"certKey,omitempty"`
}

type JaegarTracingConfig struct {
	Enabled bool `json:"enabled,omitempty"`
}

type TracingConfig struct {
	Jaegar JaegarTracingConfig `json:"jaegar,omitempty"`
}

type DatabaseConfig struct {
	MigrationsPath  string        `json:"migrationsPath,omitempty"`
	URL             string        `json:"url,omitempty"`
	MaxIdleConn     int           `json:"maxIdleConn,omitempty"`
	MaxOpenConn     int           `json:"maxOpenConn,omitempty"`
	ConnMaxLifetime time.Duration `json:"connMaxLifetime,omitempty"`
}

type MetaConfig struct {
	CheckForUpdates bool `json:"checkForUpdates"`
}

type Scheme uint

func (s Scheme) String() string {
	return schemeToString[s]
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
			Level: "INFO",
		},

		UI: UIConfig{
			Enabled: true,
		},

		Cors: CorsConfig{
			Enabled:        false,
			AllowedOrigins: []string{"*"},
		},

		Cache: CacheConfig{
			Memory: MemoryCacheConfig{
				Enabled:          false,
				Expiration:       -1,
				EvictionInterval: 10 * time.Minute,
			},
		},

		Server: ServerConfig{
			Host:      "0.0.0.0",
			Protocol:  HTTP,
			HTTPPort:  8080,
			HTTPSPort: 443,
			GRPCPort:  9000,
		},

		Database: DatabaseConfig{
			URL:            "file:/var/opt/flipt/flipt.db",
			MigrationsPath: "/etc/flipt/config/migrations",
			MaxIdleConn:    2,
		},

		Meta: MetaConfig{
			CheckForUpdates: true,
		},
	}
}

const (
	// Logging
	cfgLogLevel = "log.level"
	cfgLogFile  = "log.file"

	// UI
	cfgUIEnabled = "ui.enabled"

	// CORS
	cfgCorsEnabled        = "cors.enabled"
	cfgCorsAllowedOrigins = "cors.allowed_origins"

	// Cache
	cfgCacheMemoryEnabled          = "cache.memory.enabled"
	cfgCacheMemoryExpiration       = "cache.memory.expiration"
	cfgCacheMemoryEvictionInterval = "cache.memory.eviction_interval"

	// Server
	cfgServerHost      = "server.host"
	cfgServerProtocol  = "server.protocol"
	cfgServerHTTPPort  = "server.http_port"
	cfgServerHTTPSPort = "server.https_port"
	cfgServerGRPCPort  = "server.grpc_port"
	cfgServerCertFile  = "server.cert_file"
	cfgServerCertKey   = "server.cert_key"

	// Tracing
	cfgTracingJaegarEnabled = "tracing.jaegar.enabled"

	// DB
	cfgDBURL             = "db.url"
	cfgDBMigrationsPath  = "db.migrations.path"
	cfgDBMaxIdleConn     = "db.max_idle_conn"
	cfgDBMaxOpenConn     = "db.max_open_conn"
	cfgDBConnMaxLifetime = "db.conn_max_lifetime"

	// Meta
	cfgMetaCheckForUpdates = "meta.check_for_updates"
)

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
	if viper.IsSet(cfgLogLevel) {
		cfg.Log.Level = viper.GetString(cfgLogLevel)
	}

	if viper.IsSet(cfgLogFile) {
		cfg.Log.File = viper.GetString(cfgLogFile)
	}

	// UI
	if viper.IsSet(cfgUIEnabled) {
		cfg.UI.Enabled = viper.GetBool(cfgUIEnabled)
	}

	// CORS
	if viper.IsSet(cfgCorsEnabled) {
		cfg.Cors.Enabled = viper.GetBool(cfgCorsEnabled)

		if viper.IsSet(cfgCorsAllowedOrigins) {
			cfg.Cors.AllowedOrigins = viper.GetStringSlice(cfgCorsAllowedOrigins)
		}
	}

	// Cache
	if viper.IsSet(cfgCacheMemoryEnabled) {
		cfg.Cache.Memory.Enabled = viper.GetBool(cfgCacheMemoryEnabled)

		if viper.IsSet(cfgCacheMemoryExpiration) {
			cfg.Cache.Memory.Expiration = viper.GetDuration(cfgCacheMemoryExpiration)
		}
		if viper.IsSet(cfgCacheMemoryEvictionInterval) {
			cfg.Cache.Memory.EvictionInterval = viper.GetDuration(cfgCacheMemoryEvictionInterval)
		}
	}

	// Server
	if viper.IsSet(cfgServerHost) {
		cfg.Server.Host = viper.GetString(cfgServerHost)
	}

	if viper.IsSet(cfgServerProtocol) {
		cfg.Server.Protocol = stringToScheme[viper.GetString(cfgServerProtocol)]
	}

	if viper.IsSet(cfgServerHTTPPort) {
		cfg.Server.HTTPPort = viper.GetInt(cfgServerHTTPPort)
	}

	if viper.IsSet(cfgServerHTTPSPort) {
		cfg.Server.HTTPSPort = viper.GetInt(cfgServerHTTPSPort)
	}

	if viper.IsSet(cfgServerGRPCPort) {
		cfg.Server.GRPCPort = viper.GetInt(cfgServerGRPCPort)
	}

	if viper.IsSet(cfgServerCertFile) {
		cfg.Server.CertFile = viper.GetString(cfgServerCertFile)
	}

	if viper.IsSet(cfgServerCertKey) {
		cfg.Server.CertKey = viper.GetString(cfgServerCertKey)
	}

	// Tracing
	if viper.IsSet(cfgTracingJaegarEnabled) {
		cfg.Tracing.Jaegar.Enabled = viper.GetBool(cfgTracingJaegarEnabled)
	}

	// DB
	if viper.IsSet(cfgDBURL) {
		cfg.Database.URL = viper.GetString(cfgDBURL)
	}

	if viper.IsSet(cfgDBMigrationsPath) {
		cfg.Database.MigrationsPath = viper.GetString(cfgDBMigrationsPath)
	}

	if viper.IsSet(cfgDBMaxIdleConn) {
		cfg.Database.MaxIdleConn = viper.GetInt(cfgDBMaxIdleConn)
	}

	if viper.IsSet(cfgDBMaxOpenConn) {
		cfg.Database.MaxOpenConn = viper.GetInt(cfgDBMaxOpenConn)
	}

	if viper.IsSet(cfgDBConnMaxLifetime) {
		cfg.Database.ConnMaxLifetime = viper.GetDuration(cfgDBConnMaxLifetime)
	}

	// Meta
	if viper.IsSet(cfgMetaCheckForUpdates) {
		cfg.Meta.CheckForUpdates = viper.GetBool(cfgMetaCheckForUpdates)
	}

	if err := cfg.validate(); err != nil {
		return &Config{}, err
	}

	return cfg, nil
}

func (c *Config) validate() error {
	if c.Server.Protocol == HTTPS {
		if c.Server.CertFile == "" {
			return errors.New("cert_file cannot be empty when using HTTPS")
		}

		if c.Server.CertKey == "" {
			return errors.New("cert_key cannot be empty when using HTTPS")
		}

		if _, err := os.Stat(c.Server.CertFile); os.IsNotExist(err) {
			return fmt.Errorf("cannot find TLS cert_file at %q", c.Server.CertFile)
		}

		if _, err := os.Stat(c.Server.CertKey); os.IsNotExist(err) {
			return fmt.Errorf("cannot find TLS cert_key at %q", c.Server.CertKey)
		}
	}

	return nil
}

func (c *Config) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	out, err := json.Marshal(c)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if _, err = w.Write(out); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}
