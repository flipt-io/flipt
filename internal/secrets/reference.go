package secrets

import (
	"fmt"
	"regexp"
	"strings"
)

// referencePattern matches a config value that is entirely a secret reference,
// e.g. ${secret:file:github-token}.
var referencePattern = regexp.MustCompile(`^\${secret:([a-zA-Z0-9_:/-]+)}$`)

// ParseReference parses a secret reference of the form ${secret:provider:key}.
// The boolean result reports whether s is a secret reference at all; values
// that are not references are returned with ok set to false and a nil error.
func ParseReference(s string) (ref Reference, ok bool, err error) {
	match := referencePattern.FindStringSubmatch(s)
	if match == nil {
		return Reference{}, false, nil
	}

	reference := match[1]

	// Parse the reference format: either "key" or "provider:key"
	parts := strings.Split(reference, ":")
	switch len(parts) {
	case 1:
		// Simple format: "key-name" - no provider, so Reference.Validate rejects it
		return Reference{Path: parts[0], Key: parts[0]}, true, nil
	case 2:
		return Reference{Provider: parts[0], Path: parts[1], Key: parts[1]}, true, nil
	default:
		return Reference{}, true, fmt.Errorf("invalid secret reference format %q, expected 'key' or 'provider:key'", reference)
	}
}
