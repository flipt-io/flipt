package gateway

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.flipt.io/flipt/rpc/flipt/auth"
	"go.flipt.io/flipt/rpc/v2/evaluation"
	"google.golang.org/grpc/codes"
	grpcstatus "google.golang.org/grpc/status"
	"google.golang.org/protobuf/proto"
)

func TestEventSourceMarshalerTypes(t *testing.T) {
	m := &eventSourceMarshaler{JSONPb: runtime.JSONPb{}}

	require.Equal(t, MIMEEventStream, m.ContentType(nil))
	require.Equal(t, MIMEEventStream, m.StreamContentType(nil))
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

func TestEventSourceMarshalerStatus(t *testing.T) {
	m := &eventSourceMarshaler{JSONPb: runtime.JSONPb{}}
	errStatus := grpcstatus.New(codes.Internal, "boom").Proto()

	buf, err := m.Marshal(errStatus)
	require.NoError(t, err)

	payload, ok := extractEventPayload(t, buf)
	require.True(t, ok)

	var got map[string]any
	require.NoError(t, json.Unmarshal(payload, &got))

	require.Equal(t, "error", got["type"])
	require.InDelta(t, float64(codes.Internal), got["code"], 0)
	require.Equal(t, "boom", got["message"])
}

func TestEventSourceMarshalerCanceledStatus(t *testing.T) {
	m := &eventSourceMarshaler{JSONPb: runtime.JSONPb{}}
	errStatus := grpcstatus.New(codes.Canceled, "disconnected").Proto()

	buf, err := m.Marshal(errStatus)
	require.NoError(t, err)
	require.Nil(t, buf)
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

func TestFormURLEncodedMarshaler(t *testing.T) {
	t.Run("ContentType", func(t *testing.T) {
		m := &formURLEncodedMarshaler{runtime.JSONBuiltin{}}
		assert.Equal(t, "application/json", m.ContentType(nil))
	})

	t.Run("Marshal", func(t *testing.T) {
		m := &formURLEncodedMarshaler{runtime.JSONBuiltin{}}
		req := &auth.RevokeOIDCRequest{
			Provider:    "google",
			LogoutToken: "logout-token-value",
		}
		buf, err := m.Marshal(req)
		require.NoError(t, err)
		assert.Contains(t, string(buf), `"provider":"google"`)
		assert.Contains(t, string(buf), `"logout_token":"logout-token-value"`)
	})

	t.Run("Marshal nil", func(t *testing.T) {
		m := &formURLEncodedMarshaler{runtime.JSONBuiltin{}}
		buf, err := m.Marshal(nil)
		require.NoError(t, err)
		assert.Equal(t, "null", string(buf))
	})

	t.Run("NewDecoder parses RevokeOIDCRequest from form data", func(t *testing.T) {
		m := &formURLEncodedMarshaler{runtime.JSONBuiltin{}}
		formData := "provider=google&logout_token=my-logout-token"
		decoder := m.NewDecoder(strings.NewReader(formData))

		var req auth.RevokeOIDCRequest
		err := decoder.Decode(&req)
		require.NoError(t, err)
		assert.Equal(t, "google", req.Provider)
		assert.Equal(t, "my-logout-token", req.LogoutToken)
	})

	t.Run("NewDecoder rejects non-RevokeOIDCRequest", func(t *testing.T) {
		m := &formURLEncodedMarshaler{runtime.JSONBuiltin{}}
		decoder := m.NewDecoder(strings.NewReader(""))

		var invalid auth.GetAuthenticationRequest
		err := decoder.Decode(&invalid)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "not proto message")
	})

	t.Run("NewFormURLEncodedMarshaler returns a ServeMuxOption", func(t *testing.T) {
		opt := NewFormURLEncodedMarshaler()
		assert.NotNil(t, opt)
	})

	t.Run("NewDecoder rejects body larger than limit", func(t *testing.T) {
		m := &formURLEncodedMarshaler{runtime.JSONBuiltin{}}
		largeValue := strings.Repeat("a", maxFormURLEncodedBodySize)
		prefix := "provider=google&logout_token="
		formData := prefix + largeValue
		decoder := m.NewDecoder(strings.NewReader(formData))

		var req auth.RevokeOIDCRequest
		err := decoder.Decode(&req)
		require.NoError(t, err)
		require.Len(t, req.LogoutToken, maxFormURLEncodedBodySize-len(prefix))
	})
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
