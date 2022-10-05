package config

import (
	"encoding/json"
	"time"

	"github.com/spf13/viper"
)

const (
	// deprecation messages
	deprecatedMsgMemoryEnabled    = `'cache.memory.enabled' is deprecated and will be removed in a future version. Please use 'cache.backend' and 'cache.enabled' instead.`
	deprecatedMsgMemoryExpiration = `'cache.memory.expiration' is deprecated and will be removed in a future version. Please use 'cache.ttl' instead.`

	// configuration keys
	cacheBackend                = "cache.backend"
	cacheEnabled                = "cache.enabled"
	cacheTTL                    = "cache.ttl"
	cacheMemoryEnabled          = "cache.memory.enabled"    // deprecated in v1.10.0
	cacheMemoryExpiration       = "cache.memory.expiration" // deprecated in v1.10.0
	cacheMemoryEvictionInterval = "cache.memory.eviction_interval"
	cacheRedisHost              = "cache.redis.host"
	cacheRedisPort              = "cache.redis.port"
	cacheRedisPassword          = "cache.redis.password"
	cacheRedisDB                = "cache.redis.db"
)

// CacheConfig contains fields, which enable and configure
// Flipt's various caching mechanisms.
//
// Currently, flipt support in-memory and redis backed caching.
type CacheConfig struct {
	Enabled bool              `json:"enabled"`
	TTL     time.Duration     `json:"ttl,omitempty"`
	Backend CacheBackend      `json:"backend,omitempty"`
	Memory  MemoryCacheConfig `json:"memory,omitempty"`
	Redis   RedisCacheConfig  `json:"redis,omitempty"`
}

func (c *CacheConfig) init() (warnings []string, _ error) {
	if viper.GetBool(cacheMemoryEnabled) { // handle deprecated memory config
		c.Backend = CacheMemory
		c.Enabled = true

		warnings = append(warnings, deprecatedMsgMemoryEnabled)

		if viper.IsSet(cacheMemoryExpiration) {
			c.TTL = viper.GetDuration(cacheMemoryExpiration)
			warnings = append(warnings, deprecatedMsgMemoryExpiration)
		}

	} else if viper.IsSet(cacheEnabled) {
		c.Enabled = viper.GetBool(cacheEnabled)
		if viper.IsSet(cacheBackend) {
			c.Backend = stringToCacheBackend[viper.GetString(cacheBackend)]
		}
		if viper.IsSet(cacheTTL) {
			c.TTL = viper.GetDuration(cacheTTL)
		}
	}

	if c.Enabled {
		switch c.Backend {
		case CacheRedis:
			if viper.IsSet(cacheRedisHost) {
				c.Redis.Host = viper.GetString(cacheRedisHost)
			}
			if viper.IsSet(cacheRedisPort) {
				c.Redis.Port = viper.GetInt(cacheRedisPort)
			}
			if viper.IsSet(cacheRedisPassword) {
				c.Redis.Password = viper.GetString(cacheRedisPassword)
			}
			if viper.IsSet(cacheRedisDB) {
				c.Redis.DB = viper.GetInt(cacheRedisDB)
			}
		case CacheMemory:
			if viper.IsSet(cacheMemoryEvictionInterval) {
				c.Memory.EvictionInterval = viper.GetDuration(cacheMemoryEvictionInterval)
			}
		}
	}

	return
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
	EvictionInterval time.Duration `json:"evictionInterval,omitempty"`
}

// RedisCacheConfig contains fields, which configure the connection
// credentials for redis backed caching.
type RedisCacheConfig struct {
	Host     string `json:"host,omitempty"`
	Port     int    `json:"port,omitempty"`
	Password string `json:"password,omitempty"`
	DB       int    `json:"db,omitempty"`
}
