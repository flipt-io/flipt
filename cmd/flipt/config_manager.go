package main

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"os"

	"go.flipt.io/flipt/internal/config"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// configManager manages loading and building configuration
type configManager struct {
	configFile string
	logger     *zap.Logger
}

// newConfigManager creates a new ConfigManager
func newConfigManager(configFile string) *configManager {
	return &configManager{
		configFile: configFile,
		logger:     defaultLogger,
	}
}

// determineConfig will figure out which file to use for Flipt configuration.
func (cm *configManager) determineConfig() (string, bool) {
	// if config file is provided, use it
	if cm.configFile != "" {
		return cm.configFile, true
	}

	// otherwise, check if user config file exists on filesystem
	_, err := os.Stat(userConfigFile)
	if err == nil {
		return userConfigFile, true
	}

	if !errors.Is(err, fs.ErrNotExist) {
		cm.logger.Warn("unexpected error checking configuration file", zap.String("config_file", userConfigFile), zap.Error(err))
	}

	// finally, check if default config file exists on filesystem
	_, err = os.Stat(defaultConfigFile)
	if err == nil {
		return defaultConfigFile, true
	}

	if !errors.Is(err, fs.ErrNotExist) {
		cm.logger.Warn("unexpected error checking configuration file", zap.String("config_file", defaultConfigFile), zap.Error(err))
	}

	return "", false
}

// build builds the configuration for Flipt
func (cm *configManager) build(ctx context.Context) (*zap.Logger, *config.Config, error) {
	path, found := cm.determineConfig()

	// read in config if it exists
	// otherwise, use defaults
	res, err := config.Load(ctx, path)
	if err != nil {
		return nil, nil, fmt.Errorf("loading configuration: %w", err)
	}

	if !found {
		cm.logger.Info("no configuration file found, using defaults")
	}

	cfg := res.Config

	encoding := defaultEncoding
	encoding.TimeKey = cfg.Log.Keys.Time
	encoding.LevelKey = cfg.Log.Keys.Level
	encoding.MessageKey = cfg.Log.Keys.Message

	loggerConfig := loggerConfig(encoding)

	// log to file if enabled
	if cfg.Log.File != "" {
		loggerConfig.OutputPaths = []string{cfg.Log.File}
	}

	// parse/set log level
	loggerConfig.Level, err = zap.ParseAtomicLevel(cfg.Log.Level)
	if err != nil {
		cm.logger.Warn("parsing log level, defaulting to INFO", zap.String("level", cfg.Log.Level), zap.Error(err))
		// default to info level
		loggerConfig.Level = zap.NewAtomicLevelAt(zap.InfoLevel)
	}

	if cfg.Log.Encoding > config.LogEncodingConsole {
		loggerConfig.Encoding = cfg.Log.Encoding.String()

		// don't encode with colors if not using console log output
		loggerConfig.EncoderConfig.EncodeLevel = zapcore.CapitalLevelEncoder
	}

	logger := zap.Must(loggerConfig.Build())

	// print out any warnings from config parsing
	for _, warning := range res.Warnings {
		logger.Warn("configuration warning", zap.String("message", warning))
	}

	if found {
		logger.Debug("configuration source", zap.String("path", path))
	}

	return logger, cfg, nil
}
