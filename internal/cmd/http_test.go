package cmd

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/fullstorydev/grpchan/inprocgrpc"
	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.flipt.io/flipt/internal/config"
	"go.flipt.io/flipt/internal/info"
	rpcOfrep "go.flipt.io/flipt/rpc/flipt/ofrep"
	"go.uber.org/zap/zaptest"
	"google.golang.org/grpc/metadata"
	"google.golang.org/protobuf/types/known/structpb"
)

const (
	tsoHeader = "trailing-slash-on"
)

func TestTrailingSlashMiddleware(t *testing.T) {
	r := chi.NewRouter()

	r.Use(func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			tso := r.Header.Get(tsoHeader)
			if tso != "" {
				tsh := removeTrailingSlash(h)

				tsh.ServeHTTP(w, r)
				return
			}

			h.ServeHTTP(w, r)
		})
	})
	r.Get("/hello", func(w http.ResponseWriter, r *http.Request) {
	})

	s := httptest.NewServer(r)

	defer s.Close()

	// Request with the middleware on.
	req, err := http.NewRequestWithContext(context.TODO(), http.MethodGet, fmt.Sprintf("%s/hello/", s.URL), nil)
	require.NoError(t, err)
	req.Header.Set(tsoHeader, "on")

	res, err := http.DefaultClient.Do(req)
	require.NoError(t, err)

	require.Equal(t, http.StatusOK, res.StatusCode)
	res.Body.Close()

	// Request with the middleware off.
	req, err = http.NewRequestWithContext(context.TODO(), http.MethodGet, fmt.Sprintf("%s/hello/", s.URL), nil)
	require.NoError(t, err)

	res, err = http.DefaultClient.Do(req)
	require.NoError(t, err)

	assert.Equal(t, http.StatusNotFound, res.StatusCode)
	res.Body.Close()
}

func TestCrossOriginProtection(t *testing.T) {
	logger := zaptest.NewLogger(t)
	f := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { http.Error(w, "OK", http.StatusOK) })
	h := crossOriginProtection(logger, []string{"https://labs.flipt.io"})(f)

	tests := []struct {
		origin       string
		expectedCode int
	}{
		{origin: "", expectedCode: http.StatusOK},
		{origin: "https://labs.flipt.io", expectedCode: http.StatusOK},
		{origin: "https://unknown.flipt.io", expectedCode: http.StatusForbidden},
	}

	for i, tt := range tests {
		t.Run(fmt.Sprintf("test %d", i), func(t *testing.T) {
			req := httptest.NewRequestWithContext(t.Context(), http.MethodPut, "https://docs.flipt.io", nil)
			if tt.origin != "" {
				req.Header.Set("Origin", tt.origin)
			}
			res := httptest.NewRecorder()
			h.ServeHTTP(res, req)
			assert.Equal(t, tt.expectedCode, res.Code)
		})
	}
}

type referenceCapturingOFREPServer struct {
	rpcOfrep.UnimplementedOFREPServiceServer

	reference string
	namespace string
}

func (s *referenceCapturingOFREPServer) EvaluateFlag(ctx context.Context, req *rpcOfrep.EvaluateFlagRequest) (*rpcOfrep.EvaluatedFlag, error) {
	md, _ := metadata.FromIncomingContext(ctx)
	if values := md.Get("x-flipt-reference"); len(values) > 0 {
		s.reference = values[0]
	}
	if values := md.Get("x-flipt-namespace"); len(values) > 0 {
		s.namespace = values[0]
	}

	return &rpcOfrep.EvaluatedFlag{
		Key:      req.GetKey(),
		Reason:   rpcOfrep.EvaluateReason_DEFAULT,
		Value:    structpb.NewBoolValue(true),
		Metadata: &structpb.Struct{Fields: map[string]*structpb.Value{}},
	}, nil
}

func TestOFREPReferenceHeaderForwarding(t *testing.T) {
	conn := &inprocgrpc.Channel{}
	capturingServer := &referenceCapturingOFREPServer{}
	rpcOfrep.RegisterOFREPServiceServer(conn, capturingServer)

	cfg := config.Default()
	cfg.UI.Enabled = false
	cfg.Cors.Enabled = true

	httpServer, err := NewHTTPServer(
		t.Context(),
		zaptest.NewLogger(t),
		cfg,
		conn,
		info.Flipt{Version: "test"},
	)
	require.NoError(t, err)

	preflight := httptest.NewRequest(
		http.MethodOptions,
		"/ofrep/v1/evaluate/flags/test-flag",
		nil,
	)
	preflight.Header.Set("Origin", "https://client.example")
	preflight.Header.Set("Access-Control-Request-Method", http.MethodPost)
	preflight.Header.Set("Access-Control-Request-Headers", "content-type,x-flipt-namespace,x-flipt-reference")
	preflightResponse := httptest.NewRecorder()
	httpServer.Handler.ServeHTTP(preflightResponse, preflight)

	require.Equal(t, http.StatusOK, preflightResponse.Code)
	assert.Contains(
		t,
		strings.ToLower(preflightResponse.Header().Get("Access-Control-Allow-Headers")),
		"x-flipt-reference",
	)

	request := httptest.NewRequest(
		http.MethodPost,
		"/ofrep/v1/evaluate/flags/test-flag",
		strings.NewReader(`{"context":{"targetingKey":"user-1"}}`),
	)
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("X-Flipt-Namespace", "test-namespace")
	request.Header.Set("X-Flipt-Reference", "release-branch")

	response := httptest.NewRecorder()
	httpServer.Handler.ServeHTTP(response, request)

	require.Equal(t, http.StatusOK, response.Code)
	assert.Equal(t, "release-branch", capturingServer.reference)
	assert.Equal(t, "test-namespace", capturingServer.namespace)
}
