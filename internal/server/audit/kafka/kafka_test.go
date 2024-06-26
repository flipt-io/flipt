package kafka

import (
	"testing"

	"github.com/stretchr/testify/require"
	"go.flipt.io/flipt/internal/config"
	"go.uber.org/zap/zaptest"
)

func TestNewSink(t *testing.T) {
	for _, enc := range []string{encodingAvro, encodingProto} {
		t.Run(enc, func(t *testing.T) {
			cfg := config.KafkaSinkConfig{
				BootstrapServers: []string{"localhost:9092"},
				Topic:            "default",
				Encoding:         enc,
			}
			s, err := NewSink(zaptest.NewLogger(t), cfg)
			require.NoError(t, err)
			t.Cleanup(func() {
				err := s.Close()
				require.NoError(t, err)
			})
			require.Equal(t, sinkType, s.String())
		})
	}
	t.Run("unsupported", func(t *testing.T) {
		cfg := config.KafkaSinkConfig{
			BootstrapServers: []string{"localhost:9092"},
			Topic:            "default",
			Encoding:         "unknown",
		}
		_, err := NewSink(zaptest.NewLogger(t), cfg)
		require.ErrorContains(t, err, "unsupported encoding:")
	})
}
