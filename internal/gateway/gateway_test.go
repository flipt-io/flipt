package gateway

import (
	"encoding/json"
	"testing"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/stretchr/testify/require"
	"go.flipt.io/flipt/rpc/v2/evaluation"
	"google.golang.org/grpc/codes"
	grpcstatus "google.golang.org/grpc/status"
	"google.golang.org/protobuf/proto"
)

func TestEventSourceMarshalerTypes(t *testing.T) {
	m := &eventSourceMarshaler{JSONPb: runtime.JSONPb{}}

	require.Equal(t, eventStreamContentType, m.ContentType(nil))
	require.Equal(t, eventStreamContentType, m.StreamContentType(nil))
	require.Empty(t, m.Delimiter())
}

func TestEventSourceMarshalerSnapshot(t *testing.T) {
	m := &eventSourceMarshaler{JSONPb: runtime.JSONPb{}}
	snapshot := &evaluation.EvaluationNamespaceSnapshot{Digest: "etag-123"}

	buf, err := m.Marshal(map[string]any{"result": snapshot})
	require.NoError(t, err)

	payload, ok := extractEventPayload(t, buf)
	require.True(t, ok)

	var got map[string]string
	require.NoError(t, json.Unmarshal(payload, &got))
	require.Equal(t, map[string]string{
		"type": "refetchEvaluation",
		"etag": "etag-123",
	}, got)
}

func TestEventSourceMarshalerSnapshotProtoMap(t *testing.T) {
	m := &eventSourceMarshaler{JSONPb: runtime.JSONPb{}}
	snapshot := &evaluation.EvaluationNamespaceSnapshot{Digest: "etag-456"}

	buf, err := m.Marshal(map[string]proto.Message{"result": snapshot})
	require.NoError(t, err)

	payload, ok := extractEventPayload(t, buf)
	require.True(t, ok)

	var got map[string]string
	require.NoError(t, json.Unmarshal(payload, &got))
	require.Equal(t, map[string]string{
		"type": "refetchEvaluation",
		"etag": "etag-456",
	}, got)
}

func TestEventSourceMarshalerError(t *testing.T) {
	m := &eventSourceMarshaler{JSONPb: runtime.JSONPb{}}
	errStatus := grpcstatus.New(codes.Internal, "boom").Proto()

	buf, err := m.Marshal(map[string]any{"error": errStatus})
	require.NoError(t, err)

	payload, ok := extractEventPayload(t, buf)
	require.True(t, ok)

	var got map[string]any
	require.NoError(t, json.Unmarshal(payload, &got))

	require.Equal(t, "error", got["type"])
	require.InDelta(t, float64(codes.Internal), got["code"], 0)
	require.Equal(t, "boom", got["message"])
}

func TestEventSourceMarshalerCanceledErrorIsIgnored(t *testing.T) {
	m := &eventSourceMarshaler{JSONPb: runtime.JSONPb{}}
	errStatus := grpcstatus.New(codes.Canceled, "client disconnected").Proto()

	buf, err := m.Marshal(map[string]any{"error": errStatus})
	require.NoError(t, err)
	require.Nil(t, buf)
}

func TestEventSourceMarshalerUnsupported(t *testing.T) {
	m := &eventSourceMarshaler{JSONPb: runtime.JSONPb{}}

	buf, err := m.Marshal(map[string]any{"result": "nope"})
	require.Error(t, err)
	require.Nil(t, buf)
}

func extractEventPayload(t *testing.T, buf []byte) ([]byte, bool) {
	t.Helper()

	// Handle both with and without id field
	if len(buf) > 0 && string(buf[:3]) == "id:" {
		// Find "data: " after the id line
		const dataPrefix = "data: "
		idx := findSubstring(buf, dataPrefix)
		if idx == -1 {
			return nil, false
		}
		// Find "\n\n" after data
		endIdx := findSubstring(buf[idx:], "\n\n")
		if endIdx == -1 {
			return nil, false
		}
		return buf[idx+len(dataPrefix) : idx+endIdx], true
	}

	const prefix = "data: "
	const suffix = "\n\n"

	if len(buf) < len(prefix)+len(suffix) {
		return nil, false
	}

	if string(buf[:len(prefix)]) != prefix {
		return nil, false
	}
	if string(buf[len(buf)-len(suffix):]) != suffix {
		return nil, false
	}

	return buf[len(prefix) : len(buf)-len(suffix)], true
}

func findSubstring(buf []byte, substr string) int {
	for i := 0; i <= len(buf)-len(substr); i++ {
		if string(buf[i:i+len(substr)]) == substr {
			return i
		}
	}
	return -1
}
