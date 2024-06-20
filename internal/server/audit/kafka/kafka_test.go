package kafka

import (
	"testing"

	"github.com/stretchr/testify/require"
	"go.uber.org/zap/zaptest"
)

func TestNewSink(t *testing.T) {
	s, err := NewSink(zaptest.NewLogger(t), []string{"localhost:9092"}, "default", "protobuf")
	require.NoError(t, err)
	t.Cleanup(func() {
		err := s.Close()
		require.NoError(t, err)
	})
	require.Equal(t, sinkType, s.String())
}
