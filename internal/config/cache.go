package config

import (
	"encoding/json"
	"errors"
	"time"

	"github.com/spf13/viper"
)

var (
	_ defaulter = (*CacheConfig)(nil)
	_ validator = (*CacheConfig)(nil)
)

// CacheConfig contains fields, which enable and configure
// Flipt's various caching mechanisms.
//
// Currently, flipt support in-memory and redis backed caching.
type CacheConfig struct {
	Enabled bool              `json:"enabled" mapstructure:"enabled" yaml:"enabled"`
	TTL     time.Duration     `json:"ttl,omitempty" mapstructure:"ttl" yaml:"ttl,omitempty"`
	Backend CacheBackend      `json:"backend,omitempty" mapstructure:"backend" yaml:"backend,omitempty"`
	Memory  MemoryCacheConfig `json:"memory,omitempty" mapstructure:"memory" yaml:"memory,omitempty"`
	Redis   RedisCacheConfig  `json:"redis,omitempty" mapstructure:"redis" yaml:"redis,omitempty"`
}

func (c *CacheConfig) setDefaults(v *viper.Viper) error {
	v.SetDefault("cache", map[string]any{
		"enabled": false,
		"backend": CacheBackendMemory,
		"ttl":     1 * time.Minute,
		"redis": map[string]any{
			"host":     "localhost",
			"port":     6379,
			"password": "",
			"db":       0,
		},
		"memory": map[string]any{
			"enabled":           false, // deprecated (see below)
			"eviction_interval": 5 * time.Minute,
		},
	})

	return nil
}

func (c *CacheConfig) validate() error {
	if c.Enabled && c.Backend == CacheBackendRedis {
		return c.Redis.validate()
	}
	return nil
}

// IsZero returns true if the cache config is not enabled.
// This is used for marshalling to YAML for `config init`.
func (c CacheConfig) IsZero() bool {
	return !c.Enabled
}

// CacheBackend is either memory or redis
type CacheBackend string

const (
	CacheBackendMemory CacheBackend = "memory"
	CacheBackendRedis  CacheBackend = "redis"
)

func (c CacheBackend) String() string {
	return string(c)
}

func (c CacheBackend) MarshalJSON() ([]byte, error) {
	return json.Marshal(c.String())
}

func (c CacheBackend) MarshalYAML() (interface{}, error) {
	return c.String(), nil
}

var (
	stringToCacheBackend = map[string]CacheBackend{
		"memory": CacheBackendMemory,
		"redis":  CacheBackendRedis,
	}
)

// MemoryCacheConfig contains fields, which configure in-memory caching.
type MemoryCacheConfig struct {
	EvictionInterval time.Duration `json:"evictionInterval,omitempty" mapstructure:"eviction_interval" yaml:"eviction_interval,omitempty"`
}

// cheers up the unparam linter
var (
	_ defaulter = (*RedisCacheConfig)(nil)
	_ validator = (*RedisCacheConfig)(nil)
)

type RedisCacheMode string

const (
	RedisCacheModeSingle  RedisCacheMode = "single"
	RedisCacheModeCluster RedisCacheMode = "cluster"
)

// RedisCacheConfig contains fields, which configure the connection
// credentials for redis backed caching.
type RedisCacheConfig struct {
	Host            string         `json:"host,omitempty" mapstructure:"host" yaml:"host,omitempty"`
	Port            int            `json:"port,omitempty" mapstructure:"port" yaml:"port,omitempty"`
	RequireTLS      bool           `json:"requireTLS,omitempty" mapstructure:"require_tls" yaml:"require_tls,omitempty"`
	Username        string         `json:"-" mapstructure:"username" yaml:"-"`
	Password        string         `json:"-" mapstructure:"password" yaml:"-"`
	DB              int            `json:"db,omitempty" mapstructure:"db" yaml:"db,omitempty"`
	Prefix          string         `json:"prefix,omitempty" mapstructure:"prefix" yaml:"prefix,omitempty"`
	PoolSize        int            `json:"poolSize" mapstructure:"pool_size" yaml:"pool_size"`
	MinIdleConn     int            `json:"minIdleConn" mapstructure:"min_idle_conn" yaml:"min_idle_conn"`
	ConnMaxIdleTime time.Duration  `json:"connMaxIdleTime" mapstructure:"conn_max_idle_time" yaml:"conn_max_idle_time"`
	NetTimeout      time.Duration  `json:"netTimeout" mapstructure:"net_timeout" yaml:"net_timeout"`
	CaCertBytes     string         `json:"-" mapstructure:"ca_cert_bytes" yaml:"-"`
	CaCertPath      string         `json:"-" mapstructure:"ca_cert_path" yaml:"-"`
	InsecureSkipTLS bool           `json:"-" mapstructure:"insecure_skip_tls" yaml:"-"`
	Mode            RedisCacheMode `json:"mode,omitempty" mapstructure:"mode" yaml:"mode,omitempty"`
}

func (cfg *RedisCacheConfig) setDefaults(v *viper.Viper) error {
	v.SetDefault("redis.mode", RedisCacheModeSingle)
	return nil
}

func (cfg *RedisCacheConfig) validate() error {
	if cfg.CaCertBytes != "" && cfg.CaCertPath != "" {
		return errors.New("please provide exclusively one of ca_cert_bytes or ca_cert_path")
	}
	return nil
}
