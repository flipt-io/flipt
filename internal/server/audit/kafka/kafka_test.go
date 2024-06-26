package kafka

import (
	"testing"

	"github.com/stretchr/testify/require"
	"go.uber.org/zap/zaptest"
)

func TestNewSink(t *testing.T) {
	for _, enc := range []string{encodingAvro, encodingProto} {
		t.Run(enc, func(t *testing.T) {
			s, err := NewSink(zaptest.NewLogger(t), []string{"localhost:9092"}, "default", enc)
			require.NoError(t, err)
			t.Cleanup(func() {
				err := s.Close()
				require.NoError(t, err)
			})
			require.Equal(t, sinkType, s.String())
		})
	}
	t.Run("unsupported", func(t *testing.T) {
		_, err := NewSink(zaptest.NewLogger(t), []string{"localhost:9092"}, "default", "unsupported")
		require.ErrorContains(t, err, "unsupported encoding:")
	})
}
