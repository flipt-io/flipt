package main

import (
	"fmt"
	"os"

	"go.uber.org/zap"
)

const (
	initialConfigFile = `
log:
  level: DEBUG

cors:
  enabled: true
  allowed_origins: ["*"]

# cache:
#   enabled: false
#   backend: memory
#   ttl: 60s
#   redis:
#     host: localhost
#     port: 6379
#   memory:
#     eviction_interval: 5m # evict expired items every 5m

# server:
#   protocol: http
#   host: 0.0.0.0
#   https_port: 443
#   http_port: 8080
#   grpc_port: 9000

db:
  url: file:flipt.db
`
)

var (
	userHomeDir, _ = os.UserHomeDir()
	fliptDir       = fmt.Sprintf("%s/.flipt", userHomeDir)
	configFile     = fmt.Sprintf("%s/config.yml", fliptDir)
)

func initCommand(logger *zap.Logger) error {
	if _, err := os.Stat(configFile); err == nil {
		logger.Debug("config file already exists", zap.String("file", configFile))
		return nil
	}

	if err := os.Mkdir(fliptDir, 0750); err != nil {
		logger.Error("failed to create directory", zap.String("dir", fliptDir), zap.Error(err))
		return err
	}

	if err := os.WriteFile(configFile, []byte(initialConfigFile), 0600); err != nil {
		logger.Error("failed to write to file", zap.String("file", configFile), zap.Error(err))
		return err
	}

	logger.Debug("Successfully created config file", zap.String("file", configFile))

	return nil
}
