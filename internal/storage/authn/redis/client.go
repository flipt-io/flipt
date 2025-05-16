package redis

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"os"

	goredis "github.com/redis/go-redis/v9"
	"go.flipt.io/flipt/internal/config"
)

func NewClient(cfg config.AuthenticationSessionStorageRedisConfig) (goredis.UniversalClient, error) {
	var tlsConfig *tls.Config
	if cfg.RequireTLS {
		tlsConfig = &tls.Config{MinVersion: tls.VersionTLS12}
		tlsConfig.InsecureSkipVerify = cfg.InsecureSkipTLS

		caBundle, err := caBundle(cfg)
		if err != nil {
			return nil, err
		}

		if len(caBundle) > 0 {
			rootCAs, err := x509.SystemCertPool()
			if err != nil {
				return nil, err
			}
			rootCAs.AppendCertsFromPEM(caBundle)
			tlsConfig.RootCAs = rootCAs
		}
	}
	rwPoolTimeout := cfg.NetTimeout * 2

	switch cfg.Mode {
	case config.RedisCacheModeSingle:
		return goredis.NewClient(&goredis.Options{
			Addr:            fmt.Sprintf("%s:%d", cfg.Host, cfg.Port),
			TLSConfig:       tlsConfig,
			Username:        cfg.Username,
			Password:        cfg.Password,
			DB:              cfg.DB,
			PoolSize:        cfg.PoolSize,
			MinIdleConns:    cfg.MinIdleConn,
			ConnMaxIdleTime: cfg.ConnMaxIdleTime,
			DialTimeout:     cfg.NetTimeout,
			ReadTimeout:     rwPoolTimeout,
			WriteTimeout:    rwPoolTimeout,
			PoolTimeout:     rwPoolTimeout,
		}), nil
	case config.RedisCacheModeCluster:
		return goredis.NewClusterClient(&goredis.ClusterOptions{
			Addrs:           []string{fmt.Sprintf("%s:%d", cfg.Host, cfg.Port)},
			TLSConfig:       tlsConfig,
			Username:        cfg.Username,
			Password:        cfg.Password,
			PoolSize:        cfg.PoolSize,
			MinIdleConns:    cfg.MinIdleConn,
			ConnMaxIdleTime: cfg.ConnMaxIdleTime,
			DialTimeout:     cfg.NetTimeout,
			ReadTimeout:     rwPoolTimeout,
			WriteTimeout:    rwPoolTimeout,
			PoolTimeout:     rwPoolTimeout,
		}), nil
	default:
		return nil, fmt.Errorf("invalid redis cache mode: %s", cfg.Mode)
	}
}

func caBundle(cfg config.AuthenticationSessionStorageRedisConfig) ([]byte, error) {
	if cfg.CaCertBytes != "" {
		return []byte(cfg.CaCertBytes), nil
	}
	if cfg.CaCertPath != "" {
		bytes, err := os.ReadFile(cfg.CaCertPath)
		if err != nil {
			return nil, err
		}
		return bytes, nil
	}
	return []byte{}, nil
}
