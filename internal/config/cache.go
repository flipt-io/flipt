package config

import (
	"encoding/json"
	"time"

	"github.com/mitchellh/mapstructure"
	"github.com/spf13/viper"
)

var cacheDecodeHooks = mapstructure.ComposeDecodeHookFunc(
	mapstructure.StringToTimeDurationHookFunc(),
	mapstructure.StringToSliceHookFunc(","),
	StringToEnumHookFunc(stringToCacheBackend),
)

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

func (c *CacheConfig) viperKey() string {
	return "cache"
}

func (c *CacheConfig) unmarshalViper(v *viper.Viper) (warnings []string, err error) {
	if mem := v.Sub("memory"); mem != nil {
		// handle legacy memory structure
		if mem.GetBool("enabled") {
			// forcibly set top-level `enabled` to true
			v.Set("enabled", true)
			// ensure ttl is mapped to the value at memory.expiration
			v.RegisterAlias("ttl", "memory.expiration")
		} else {
			mem.SetDefault("eviction_interval", 5*time.Minute)
		}
	}

	if redis := v.Sub("redis"); redis != nil {
		redis.SetDefault("host", "localhost")
		redis.SetDefault("port", 6379)
		redis.SetDefault("password", "")
		redis.SetDefault("db", 0)
	}

	if v.GetBool("enabled") {
		v.SetDefault("backend", CacheMemory)
		v.SetDefault("ttl", 1*time.Minute)
	}

	if err = v.Unmarshal(c, viper.DecodeHook(cacheDecodeHooks)); err != nil {
		return
	}

	if v.GetBool("memory.enabled") { // handle deprecated memory config
		warnings = append(warnings, deprecatedMsgMemoryEnabled)

		if v.IsSet("memory.expiration") {
			warnings = append(warnings, deprecatedMsgMemoryExpiration)
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
