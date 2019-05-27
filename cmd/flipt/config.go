package main

import (
	"strings"

	"github.com/pkg/errors"
	"github.com/spf13/viper"
)

type config struct {
	logLevel string
	ui       uiConfig
	cache    cacheConfig
	server   serverConfig
	database databaseConfig
}

type uiConfig struct {
	enabled bool
}

type memoryCacheConfig struct {
	enabled bool
	items   int
}

type cacheConfig struct {
	memory memoryCacheConfig
}

type serverConfig struct {
	host     string
	httpPort int
	grpcPort int
}

type databaseConfig struct {
	autoMigrate    bool
	migrationsPath string
	url            string
}

func defaultConfig() *config {
	return &config{
		logLevel: "INFO",

		ui: uiConfig{
			enabled: true,
		},

		cache: cacheConfig{
			memory: memoryCacheConfig{
				enabled: false,
				items:   500,
			},
		},

		server: serverConfig{
			host:     "0.0.0.0",
			httpPort: 8080,
			grpcPort: 9000,
		},

		database: databaseConfig{
			url:            "file:/var/opt/flipt/flipt.db",
			migrationsPath: "/etc/flipt/config/migrations",
			autoMigrate:    true,
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
		cfg.logLevel = viper.GetString(cfgLogLevel)
	}

	// UI
	if viper.IsSet(cfgUIEnabled) {
		cfg.ui.enabled = viper.GetBool(cfgUIEnabled)
	}

	// Cache
	if viper.IsSet(cfgCacheMemoryEnabled) {
		cfg.cache.memory.enabled = viper.GetBool(cfgCacheMemoryEnabled)

		if viper.IsSet(cfgCacheMemoryItems) {
			cfg.cache.memory.items = viper.GetInt(cfgCacheMemoryItems)
		}
	}

	// Server
	if viper.IsSet(cfgServerHost) {
		cfg.server.host = viper.GetString(cfgServerHost)
	}
	if viper.IsSet(cfgServerHTTPPort) {
		cfg.server.httpPort = viper.GetInt(cfgServerHTTPPort)
	}
	if viper.IsSet(cfgServerGRPCPort) {
		cfg.server.grpcPort = viper.GetInt(cfgServerGRPCPort)
	}

	// DB
	if viper.IsSet(cfgDBURL) {
		cfg.database.url = viper.GetString(cfgDBURL)
	}
	if viper.IsSet(cfgDBMigrationsPath) {
		cfg.database.migrationsPath = viper.GetString(cfgDBMigrationsPath)
	}
	if viper.IsSet(cfgDBAutoMigrate) {
		cfg.database.autoMigrate = viper.GetBool(cfgDBAutoMigrate)
	}

	return cfg, nil
}
