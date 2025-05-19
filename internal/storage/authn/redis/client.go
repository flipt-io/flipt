package redis

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"os"

	"github.com/redis/go-redis/extra/redisotel/v9"
	goredis "github.com/redis/go-redis/v9"
	"go.flipt.io/flipt/internal/config"
)

func NewClient(cfg config.Config) (goredis.UniversalClient, error) {
	var (
		redisCfg  = cfg.Authentication.Session.Storage.Redis
		tlsConfig *tls.Config
	)

	if redisCfg.RequireTLS {
		tlsConfig = &tls.Config{MinVersion: tls.VersionTLS12}
		tlsConfig.InsecureSkipVerify = redisCfg.InsecureSkipTLS

		caBundle, err := caBundle(redisCfg)
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

	var (
		rwPoolTimeout = redisCfg.NetTimeout * 2
		rdb           goredis.UniversalClient
	)

	switch redisCfg.Mode {
	case config.RedisCacheModeSingle:
		rdb = goredis.NewClient(&goredis.Options{
			Addr:            fmt.Sprintf("%s:%d", redisCfg.Host, redisCfg.Port),
			TLSConfig:       tlsConfig,
			Username:        redisCfg.Username,
			Password:        redisCfg.Password,
			DB:              redisCfg.DB,
			PoolSize:        redisCfg.PoolSize,
			MinIdleConns:    redisCfg.MinIdleConn,
			ConnMaxIdleTime: redisCfg.ConnMaxIdleTime,
			DialTimeout:     redisCfg.NetTimeout,
			ReadTimeout:     rwPoolTimeout,
			WriteTimeout:    rwPoolTimeout,
			PoolTimeout:     rwPoolTimeout,
		})
	case config.RedisCacheModeCluster:
		rdb = goredis.NewClusterClient(&goredis.ClusterOptions{
			Addrs:           []string{fmt.Sprintf("%s:%d", redisCfg.Host, redisCfg.Port)}, // TODO: maybe support multiple addresses in the future
			TLSConfig:       tlsConfig,
			Username:        redisCfg.Username,
			Password:        redisCfg.Password,
			PoolSize:        redisCfg.PoolSize,
			MinIdleConns:    redisCfg.MinIdleConn,
			ConnMaxIdleTime: redisCfg.ConnMaxIdleTime,
			DialTimeout:     redisCfg.NetTimeout,
			ReadTimeout:     rwPoolTimeout,
			WriteTimeout:    rwPoolTimeout,
			PoolTimeout:     rwPoolTimeout,
		})
	default:
		return nil, fmt.Errorf("invalid redis cache mode: %s", redisCfg.Mode)
	}

	if cfg.Metrics.Enabled {
		if err := redisotel.InstrumentMetrics(rdb); err != nil {
			return nil, fmt.Errorf("instrumenting redis: %w", err)
		}
	}

	if cfg.Tracing.Enabled {
		if err := redisotel.InstrumentTracing(rdb); err != nil {
			return nil, fmt.Errorf("instrumenting redis: %w", err)
		}
	}

	return rdb, nil
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
