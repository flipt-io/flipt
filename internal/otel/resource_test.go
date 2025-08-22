package otel

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel/attribute"
	semconv "go.opentelemetry.io/otel/semconv/v1.34.0"
)

func TestNewResourceDefault(t *testing.T) {
	tests := []struct {
		name string
		envs map[string]string
		want []attribute.KeyValue
	}{
		{
			name: "default",
			envs: map[string]string{},
			want: []attribute.KeyValue{
				semconv.ServiceName("flipt"),
				semconv.ServiceVersion("test"),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			for k, v := range tt.envs {
				t.Setenv(k, v)
			}
			r, err := NewResource(context.Background(), "test")
			require.NoError(t, err)

			want := make(map[attribute.Key]attribute.KeyValue)
			for _, keyValue := range tt.want {
				want[keyValue.Key] = keyValue
			}

			for _, keyValue := range r.Attributes() {
				if wantKeyValue, ok := want[keyValue.Key]; ok {
					assert.Equal(t, wantKeyValue, keyValue)
				}
			}
		})
	}
}
