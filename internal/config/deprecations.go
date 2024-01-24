package config

import (
	"fmt"
	"strings"
)

type deprecated string

var (
	deprecateAuthenticationExcludeMetdata deprecated = "authentication.exclude.metadata"
	// fields that are deprecated along with their messages
	deprecatedFields = map[deprecated]string{
		deprecateAuthenticationExcludeMetdata: "This feature never worked as intended. Metadata can no longer be excluded from authentication (when required).",
	}
)

const (
	deprecatedDefaultMessage = `%q is deprecated and will be removed in a future release.`
)

func (d deprecated) Message() string {
	msg, ok := deprecatedFields[d]
	if !ok {
		return strings.TrimSpace(fmt.Sprintf(deprecatedDefaultMessage, d))
	}

	msg = strings.Join([]string{deprecatedDefaultMessage, msg}, " ")
	return strings.TrimSpace(fmt.Sprintf(msg, d))
}
