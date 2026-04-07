package config

import (
	"fmt"
	"strings"
)

type deprecated string

// fields that are deprecated along with their messages
var deprecatedFields = map[deprecated]string{}

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
