package config

import (
	"encoding/json"
	"time"
)

const (
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
