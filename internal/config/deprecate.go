package config

const (
	// deprecation messages
	deprecatedMsgMemoryEnabled      = `'cache.memory.enabled' is deprecated and will be removed in a future version. Please use 'cache.backend' and 'cache.enabled' instead.`
	deprecatedMsgMemoryExpiration   = `'cache.memory.expiration' is deprecated and will be removed in a future version. Please use 'cache.ttl' instead.`
	deprecatedMsgDatabaseMigrations = `'db.migrations_path' is deprecated and will be removed in a future version. Migrations are embedded within Flipt and are no longer required on disk.`
)
