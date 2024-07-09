package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAuditEnabled(t *testing.T) {
	tests := []struct {
		sink     string
		f        func() AuditConfig
		expected bool
	}{
		{
			sink:     "none",
			f:        func() AuditConfig { return AuditConfig{} },
			expected: false,
		},
		{
			sink:     "log",
			f:        func() AuditConfig { c := AuditConfig{}; c.Sinks.Log.Enabled = true; return c },
			expected: true,
		},
		{
			sink:     "webhook",
			f:        func() AuditConfig { c := AuditConfig{}; c.Sinks.Webhook.Enabled = true; return c },
			expected: true,
		},
		{
			sink:     "cloud",
			f:        func() AuditConfig { c := AuditConfig{}; c.Sinks.Cloud.Enabled = true; return c },
			expected: true,
		},
		{
			sink:     "kafka",
			f:        func() AuditConfig { c := AuditConfig{}; c.Sinks.Kafka.Enabled = true; return c },
			expected: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.sink, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.f().Enabled())
		})
	}
}
