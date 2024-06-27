package kafka

import (
	"context"
	"testing"

	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
	"github.com/stretchr/testify/require"
	"go.flipt.io/flipt/internal/config"
	"go.uber.org/zap/zaptest"
)

func TestNewSink(t *testing.T) {
	mockCluster, err := kafka.NewMockCluster(1)
	require.NoError(t, err)
	t.Cleanup(mockCluster.Close)
	bootstrapServers := []string{mockCluster.BootstrapServers()}

	for _, enc := range []string{encodingAvro, encodingProto} {
		t.Run(enc, func(t *testing.T) {
			cfg := config.KafkaSinkConfig{
				BootstrapServers: bootstrapServers,
				Topic:            "default",
				Encoding:         enc,
			}
			s, err := NewSink(context.Background(), zaptest.NewLogger(t), cfg)
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
			BootstrapServers: bootstrapServers,
			Topic:            "default",
			Encoding:         "unknown",
		}
		_, err := NewSink(context.Background(), zaptest.NewLogger(t), cfg)
		require.ErrorContains(t, err, "unsupported encoding:")
	})
}
