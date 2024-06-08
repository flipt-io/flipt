package audit

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap/zapcore"
)

func TestMarshalLogObject(t *testing.T) {
	var (
		actor = Actor{Authentication: "github"}
		e     = Event{
			Version:  "0.2",
			Type:     "sometype",
			Action:   "modified",
			Metadata: Metadata{Actor: &actor},
			Payload:  "custom payload",
			Status:   "success",
		}
		enc = zapcore.NewMapObjectEncoder()
		err = e.MarshalLogObject(enc)
	)

	require.NoError(t, err)
	assert.Equal(t, map[string]any{
		"version":   "0.2",
		"action":    "modified",
		"type":      "sometype",
		"metadata":  Metadata{Actor: &actor},
		"payload":   "custom payload",
		"timestamp": "",
		"status":    "success",
	}, enc.Fields)
}
