package cmd

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.flipt.io/flipt/internal/config"
	"go.flipt.io/flipt/internal/server/common"
	"go.uber.org/zap/zaptest"
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
	req, err := http.NewRequestWithContext(t.Context(), http.MethodGet, fmt.Sprintf("%s/hello/", s.URL), nil)
	require.NoError(t, err)
	req.Header.Set(tsoHeader, "on")

	res, err := http.DefaultClient.Do(req)
	require.NoError(t, err)

	require.Equal(t, http.StatusOK, res.StatusCode)
	res.Body.Close()

	// Request with the middleware off.
	req, err = http.NewRequestWithContext(t.Context(), http.MethodGet, fmt.Sprintf("%s/hello/", s.URL), nil)
	require.NoError(t, err)

	res, err = http.DefaultClient.Do(req)
	require.NoError(t, err)

	assert.Equal(t, http.StatusNotFound, res.StatusCode)
	res.Body.Close()
}

func TestCORSExposedHeaders(t *testing.T) {
	r := chi.NewRouter()
	r.Use(newCORSHandler(&config.Config{
		Cors: config.CorsConfig{
			Enabled:        true,
			AllowedOrigins: []string{"*"},
			AllowedHeaders: []string{"*"},
		},
	}))
	r.Get("/test", func(w http.ResponseWriter, r *http.Request) {})

	s := httptest.NewServer(r)
	t.Cleanup(s.Close)

	req, err := http.NewRequestWithContext(t.Context(), http.MethodGet, s.URL+"/test", nil)
	require.NoError(t, err)
	req.Header.Set("Origin", "http://localhost:5173")

	res, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer res.Body.Close()

	exposed := res.Header.Get("Access-Control-Expose-Headers")
	require.NotEmpty(t, exposed, "Access-Control-Expose-Headers must be present")
	assert.Contains(t, exposed, "Etag", "Access-Control-Expose-Headers must include Etag for client-side conditional requests")
	assert.Contains(t, exposed, "Link", "Access-Control-Expose-Headers must include Link")
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

func TestOFREPStreamAlias_AcceptsOnlySSE(t *testing.T) {
	tests := []struct {
		name         string
		accept       string
		expectedCode int
	}{
		{name: "no accept header", expectedCode: http.StatusNotAcceptable},
		{name: "application/json", accept: "application/json", expectedCode: http.StatusNotAcceptable},
		{name: "wildcard", accept: "*/*", expectedCode: http.StatusNotAcceptable},
		{name: "compound", accept: "text/event-stream, */*", expectedCode: http.StatusOK},
		{name: "exact sse", accept: "text/event-stream", expectedCode: http.StatusOK},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var called bool
			r := chi.NewRouter()
			r.Get("/ofrep/v1/_stream/{environmentKey}/{namespaceKey}/events", ofrepStreamHandler(http.HandlerFunc(func(_ http.ResponseWriter, r *http.Request) {
				called = true
				assert.True(t, common.IsOFREPStream(r.Context()))
				assert.Equal(t, "text/event-stream", r.Header.Get("Accept"))
				assert.Equal(t, "/client/v2/environments/env/namespaces/ns/stream", r.URL.Path)
			})))

			s := httptest.NewServer(r)
			t.Cleanup(s.Close)

			req, err := http.NewRequestWithContext(t.Context(), http.MethodGet, s.URL+"/ofrep/v1/_stream/env/ns/events", nil)
			require.NoError(t, err)
			if tt.accept != "" {
				req.Header.Set("Accept", tt.accept)
			}

			res, err := http.DefaultClient.Do(req)
			require.NoError(t, err)
			res.Body.Close()

			assert.Equal(t, tt.expectedCode, res.StatusCode)
			assert.Equal(t, tt.expectedCode == http.StatusOK, called)
		})
	}
}
