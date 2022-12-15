package config

import (
	"encoding/json"
	"time"

	"github.com/spf13/viper"
)

// cheers up the unparam linter
var _ defaulter = (*CacheConfig)(nil)

// CacheConfig contains fields, which enable and configure
// Flipt's various caching mechanisms.
//
// Currently, flipt support in-memory and redis backed caching.
type CacheConfig struct {
	Enabled bool              `json:"enabled" mapstructure:"enabled"`
	TTL     time.Duration     `json:"ttl,omitempty" mapstructure:"ttl"`
	Backend CacheBackend      `json:"backend,omitempty" mapstructure:"backend"`
	Memory  MemoryCacheConfig `json:"memory,omitempty" mapstructure:"memory"`
	Redis   RedisCacheConfig  `json:"redis,omitempty" mapstructure:"redis"`
}

func (c *CacheConfig) setDefaults(v *viper.Viper) {
	v.SetDefault("cache", map[string]any{
		"enabled": false,
		"backend": CacheMemory,
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

	if v.GetBool("cache.memory.enabled") {
		// forcibly set top-level `enabled` to true
		v.Set("cache.enabled", true)
		// ensure ttl is mapped to the value at memory.expiration
		v.RegisterAlias("cache.ttl", "cache.memory.expiration")
		// ensure ttl default is set
		v.SetDefault("cache.memory.expiration", 1*time.Minute)
	}
}

func (c *CacheConfig) deprecations(v *viper.Viper) []deprecation {
	var deprecations []deprecation

	if v.GetBool("cache.memory.enabled") {
		deprecations = append(deprecations, deprecation{

			option:            "cache.memory.enabled",
			additionalMessage: deprecatedMsgMemoryEnabled,
		})
	}

	if v.IsSet("cache.memory.expiration") {
		deprecations = append(deprecations, deprecation{
			option:            "cache.memory.expiration",
			additionalMessage: deprecatedMsgMemoryExpiration,
		})
	}

	return deprecations
}

// CacheBackend is either memory or redis
type CacheBackend uint8

func (c CacheBackend) String() string {
	return cacheBackendToString[c]
}

func (c CacheBackend) MarshalJSON() ([]byte, error) {
	return json.Marshal(c.String())
}

const (
	_ CacheBackend = iota
	// CacheMemory ...
	CacheMemory
	// CacheRedis ...
	CacheRedis
)

var (
	cacheBackendToString = map[CacheBackend]string{
		CacheMemory: "memory",
		CacheRedis:  "redis",
	}

	stringToCacheBackend = map[string]CacheBackend{
		"memory": CacheMemory,
		"redis":  CacheRedis,
	}
)

// MemoryCacheConfig contains fields, which configure in-memory caching.
type MemoryCacheConfig struct {
	EvictionInterval time.Duration `json:"evictionInterval,omitempty" mapstructure:"eviction_interval"`
}

// RedisCacheConfig contains fields, which configure the connection
// credentials for redis backed caching.
type RedisCacheConfig struct {
	Host     string `json:"host,omitempty" mapstructure:"host"`
	Port     int    `json:"port,omitempty" mapstructure:"port"`
	Password string `json:"password,omitempty" mapstructure:"password"`
	DB       int    `json:"db,omitempty" mapstructure:"db"`
}
