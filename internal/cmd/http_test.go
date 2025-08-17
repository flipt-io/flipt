package cmd

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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
			req := httptest.NewRequest(http.MethodPut, "https://docs.flipt.io", nil)
			if tt.origin != "" {
				req.Header.Set("Origin", tt.origin)
			}
			res := httptest.NewRecorder()
			h.ServeHTTP(res, req)
			assert.Equal(t, tt.expectedCode, res.Code)
		})
	}
}
