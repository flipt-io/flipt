package object

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRemapScheme(t *testing.T) {
	for _, tt := range []struct {
		scheme   string
		expected string
	}{
		{"s3", "s3i"},
		{"azblob", "azblob"},
		{"googlecloud", "gs"},
	} {
		t.Run(tt.scheme, func(t *testing.T) {
			assert.Equal(t, tt.expected, remapScheme(tt.scheme))
		})
	}
}

func TestSupportedSchemes(t *testing.T) {
	for _, tt := range []string{"s3", "s3i", "azblob", "googlecloud", "gs"} {
		t.Run(tt, func(t *testing.T) {
			assert.Contains(t, SupportedSchemes(), tt)
		})
	}
}
