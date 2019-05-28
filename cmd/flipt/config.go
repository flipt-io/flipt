package main

import (
	"strings"

	"github.com/pkg/errors"
	"github.com/spf13/viper"
)

type config struct {
	LogLevel string         `json:"logLevel,omitempty"`
	UI       uiConfig       `json:"ui,omitempty"`
	Cache    cacheConfig    `json:"cache,omitempty"`
	Server   serverConfig   `json:"server,omitempty"`
	Database databaseConfig `json:"database,omitempty"`
}

type uiConfig struct {
	Enabled bool `json:"enabled"`
}

type memoryCacheConfig struct {
	Enabled bool `json:"enabled"`
	Items   int  `json:"items,omitempty"`
}

type cacheConfig struct {
	Memory memoryCacheConfig `json:"memory,omitempty"`
}

type serverConfig struct {
	Host     string `json:"host,omitempty"`
	HTTPPort int    `json:"httpPort,omitempty"`
	GRPCPort int    `json:"grpcPort,omitempty"`
}

type databaseConfig struct {
	AutoMigrate    bool   `json:"autoMigrate"`
	MigrationsPath string `json:"migrationsPath,omitempty"`
	URL            string `json:"url,omitempty"`
}

func defaultConfig() *config {
	return &config{
		LogLevel: "INFO",

		UI: uiConfig{
			Enabled: true,
		},

		Cache: cacheConfig{
			Memory: memoryCacheConfig{
				Enabled: false,
				Items:   500,
			},
		},

		Server: serverConfig{
			Host:     "0.0.0.0",
			HTTPPort: 8080,
			GRPCPort: 9000,
		},

		Database: databaseConfig{
			URL:            "file:/var/opt/flipt/flipt.db",
			MigrationsPath: "/etc/flipt/config/migrations",
			AutoMigrate:    true,
		},
	}
}

const (
	// Logging
	cfgLogLevel = "log.level"

	// UI
	cfgUIEnabled = "ui.enabled"

	// Cache
	cfgCacheMemoryEnabled = "cache.memory.enabled"
	cfgCacheMemoryItems   = "cache.memory.items"

	// Server
	cfgServerHost     = "server.host"
	cfgServerHTTPPort = "server.http_port"
	cfgServerGRPCPort = "server.grpc_port"

	// DB
	cfgDBURL            = "db.url"
	cfgDBMigrationsPath = "db.migrations.path"
	cfgDBAutoMigrate    = "db.migrations.auto"
)

func configure() (*config, error) {
	viper.SetEnvPrefix("FLIPT")
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	viper.AutomaticEnv()

	viper.SetConfigFile(cfgPath)

	if err := viper.ReadInConfig(); err != nil {
		return nil, errors.Wrap(err, "loading config")
	}

	cfg := defaultConfig()

	// Logging
	if viper.IsSet(cfgLogLevel) {
		cfg.LogLevel = viper.GetString(cfgLogLevel)
	}

	// UI
	if viper.IsSet(cfgUIEnabled) {
		cfg.UI.Enabled = viper.GetBool(cfgUIEnabled)
	}

	// Cache
	if viper.IsSet(cfgCacheMemoryEnabled) {
		cfg.Cache.Memory.Enabled = viper.GetBool(cfgCacheMemoryEnabled)

		if viper.IsSet(cfgCacheMemoryItems) {
			cfg.Cache.Memory.Items = viper.GetInt(cfgCacheMemoryItems)
		}
	}

	// Server
	if viper.IsSet(cfgServerHost) {
		cfg.Server.Host = viper.GetString(cfgServerHost)
	}
	if viper.IsSet(cfgServerHTTPPort) {
		cfg.Server.HTTPPort = viper.GetInt(cfgServerHTTPPort)
	}
	if viper.IsSet(cfgServerGRPCPort) {
		cfg.Server.GRPCPort = viper.GetInt(cfgServerGRPCPort)
	}

	// DB
	if viper.IsSet(cfgDBURL) {
		cfg.Database.URL = viper.GetString(cfgDBURL)
	}
	if viper.IsSet(cfgDBMigrationsPath) {
		cfg.Database.MigrationsPath = viper.GetString(cfgDBMigrationsPath)
	}
	if viper.IsSet(cfgDBAutoMigrate) {
		cfg.Database.AutoMigrate = viper.GetBool(cfgDBAutoMigrate)
	}

	return cfg, nil
}
