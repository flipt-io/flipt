package config

import (
	"fmt"
	"strings"
)

const (
	// additional deprecation messages
	deprecatedMsgMemoryEnabled      = `Please use 'cache.backend' and 'cache.enabled' instead.`
	deprecatedMsgMemoryExpiration   = `Please use 'cache.ttl' instead.`
	deprecatedMsgDatabaseMigrations = `Migrations are now embedded within Flipt and are no longer required on disk.`
)

// deprecation represents a deprecated configuration option
type deprecation struct {
	// the deprecated option
	option string
	// the (optional) additionalMessage to display
	additionalMessage string
}

func (d deprecation) String() string {
	return strings.TrimSpace(fmt.Sprintf("%q is deprecated and will be removed in a future version. %s", d.option, d.additionalMessage))
}
