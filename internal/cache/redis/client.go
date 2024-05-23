package redis

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"os"

	goredis "github.com/redis/go-redis/v9"
	"go.flipt.io/flipt/internal/config"
)

func NewClient(cfg config.RedisCacheConfig) (*goredis.Client, error) {
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

	rdb := goredis.NewClient(&goredis.Options{
		Addr:            fmt.Sprintf("%s:%d", cfg.Host, cfg.Port),
		TLSConfig:       tlsConfig,
		Username:        cfg.Username,
		Password:        cfg.Password,
		DB:              cfg.DB,
		PoolSize:        cfg.PoolSize,
		MinIdleConns:    cfg.MinIdleConn,
		ConnMaxIdleTime: cfg.ConnMaxIdleTime,
		DialTimeout:     cfg.NetTimeout,
		ReadTimeout:     cfg.NetTimeout * 2,
		WriteTimeout:    cfg.NetTimeout * 2,
		PoolTimeout:     cfg.NetTimeout * 2,
	})
	return rdb, nil
}

func caBundle(cfg config.RedisCacheConfig) ([]byte, error) {
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
