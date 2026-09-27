package cmd

import (
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.flipt.io/flipt/internal/config"
	"go.flipt.io/flipt/internal/server/common"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"go.uber.org/zap/zaptest"
	"go.uber.org/zap/zaptest/observer"
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

func TestCrossOriginExemptPatterns(t *testing.T) {
	tests := []struct {
		name         string
		exclude      func(*config.Config)
		wantPatterns []string
	}{
		{
			name:         "no exclusions",
			exclude:      func(_ *config.Config) {},
			wantPatterns: nil,
		},
		{
			name: "evaluation excluded",
			exclude: func(cfg *config.Config) {
				cfg.Authentication.Exclude.Evaluation = true
			},
			wantPatterns: []string{"POST /evaluate/v1/"},
		},
		{
			name: "ofrep excluded",
			exclude: func(cfg *config.Config) {
				cfg.Authentication.Exclude.OFREP = true
			},
			wantPatterns: []string{"POST /ofrep/v1/"},
		},
		{
			name: "evaluation and ofrep excluded",
			exclude: func(cfg *config.Config) {
				cfg.Authentication.Exclude.Evaluation = true
				cfg.Authentication.Exclude.OFREP = true
			},
			wantPatterns: []string{"POST /evaluate/v1/", "POST /ofrep/v1/"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &config.Config{}
			tt.exclude(cfg)
			assert.Equal(t, tt.wantPatterns, crossOriginExemptPatterns(cfg))
		})
	}
}

func TestCrossOriginProtection(t *testing.T) {
	const trustedOrigin = "https://labs.flipt.io"

	tests := []struct {
		name           string
		exemptPatterns []string
		method         string
		path           string
		secFetchSite   string
		origin         string
		expectedCode   int
	}{
		{
			name:         "no sec-fetch-site or origin allowed",
			method:       http.MethodPut,
			path:         "/api/v1/namespaces/default/flags",
			expectedCode: http.StatusOK,
		},
		{
			name:         "trusted origin allowed",
			method:       http.MethodPut,
			path:         "/api/v1/namespaces/default/flags",
			origin:       trustedOrigin,
			expectedCode: http.StatusOK,
		},
		{
			name:         "untrusted origin rejected",
			method:       http.MethodPut,
			path:         "/api/v1/namespaces/default/flags",
			origin:       "https://unknown.flipt.io",
			expectedCode: http.StatusForbidden,
		},
		{
			name:         "same-origin allowed",
			method:       http.MethodPost,
			path:         "/evaluate/v1/boolean",
			secFetchSite: "same-origin",
			expectedCode: http.StatusOK,
		},
		{
			name:         "same-site evaluation rejected without bypass",
			method:       http.MethodPost,
			path:         "/evaluate/v1/boolean",
			secFetchSite: "same-site",
			expectedCode: http.StatusForbidden,
		},
		{
			name:           "same-site evaluation allowed with bypass",
			exemptPatterns: []string{evaluationCrossOriginExemptPattern},
			method:         http.MethodPost,
			path:           "/evaluate/v1/boolean",
			secFetchSite:   "same-site",
			expectedCode:   http.StatusOK,
		},
		{
			name:           "cross-site evaluation allowed with bypass",
			exemptPatterns: []string{evaluationCrossOriginExemptPattern},
			method:         http.MethodPost,
			path:           "/evaluate/v1/variant",
			secFetchSite:   "cross-site",
			expectedCode:   http.StatusOK,
		},
		{
			name:           "bypass does not leak to management routes",
			exemptPatterns: []string{evaluationCrossOriginExemptPattern},
			method:         http.MethodPost,
			path:           "/api/v1/namespaces/default/flags",
			secFetchSite:   "cross-site",
			expectedCode:   http.StatusForbidden,
		},
		{
			name:           "bypass does not leak to other evaluation versions",
			exemptPatterns: []string{evaluationCrossOriginExemptPattern},
			method:         http.MethodPost,
			path:           "/other/v1/boolean",
			secFetchSite:   "cross-site",
			expectedCode:   http.StatusForbidden,
		},
		{
			name:           "cross-site ofrep allowed with bypass",
			exemptPatterns: []string{ofrepCrossOriginExemptPattern},
			method:         http.MethodPost,
			path:           "/ofrep/v1/evaluate/flags",
			secFetchSite:   "cross-site",
			expectedCode:   http.StatusOK,
		},
		{
			name:           "evaluation bypass does not cover ofrep",
			exemptPatterns: []string{evaluationCrossOriginExemptPattern},
			method:         http.MethodPost,
			path:           "/ofrep/v1/evaluate/flags",
			secFetchSite:   "cross-site",
			expectedCode:   http.StatusForbidden,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logger := zaptest.NewLogger(t)
			f := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) { http.Error(w, "OK", http.StatusOK) })
			h := crossOriginProtection(logger, []string{trustedOrigin}, tt.exemptPatterns)(f)

			req := httptest.NewRequestWithContext(t.Context(), tt.method, "https://docs.flipt.io"+tt.path, nil)
			if tt.secFetchSite != "" {
				req.Header.Set("Sec-Fetch-Site", tt.secFetchSite)
			}
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

// TestValidateRequestPath covers the middleware that rejects request paths
// which are not valid UTF-8 after percent-decoding. Such paths are malformed
// client input and must surface as a 4xx (with a log line) rather than being
// forwarded to handlers that turn them into an unlogged 5xx.
//
// See https://github.com/flipt-io/flipt/issues/6337.
func TestValidateRequestPath(t *testing.T) {
	tests := []struct {
		name        string
		path        string // decoded path, as set on r.URL.Path by net/http
		wantStatus  int
		wantReached bool
		wantLogged  bool
	}{
		// valid paths pass through to the downstream handler
		{name: "ascii root", path: "/", wantStatus: http.StatusOK, wantReached: true},
		{name: "ascii path", path: "/nonexistent-path-12345", wantStatus: http.StatusOK, wantReached: true},
		{name: "valid utf8 cafe", path: "/café", wantStatus: http.StatusOK, wantReached: true},
		// invalid UTF-8 paths are rejected with 400 and never reach downstream
		{name: "invalid lone 0xc0", path: "/\xc0", wantStatus: http.StatusBadRequest, wantLogged: true},
		{name: "invalid lone 0xff", path: "/\xff", wantStatus: http.StatusBadRequest, wantLogged: true},
		{name: "invalid overlong", path: "/\xe0\x80\xaf", wantStatus: http.StatusBadRequest, wantLogged: true},
		{name: "invalid surrogate half", path: "/\xed\xa0\x80", wantStatus: http.StatusBadRequest, wantLogged: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			core, recorded := observer.New(zapcore.DebugLevel)
			logger := zap.New(core)

			var reached bool
			next := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
				reached = true
				w.WriteHeader(http.StatusOK)
			})
			h := validateRequestPath(logger)(next)

			req := httptest.NewRequestWithContext(t.Context(), http.MethodGet, "https://flipt.io", nil)
			req.URL.Path = tt.path
			rec := httptest.NewRecorder()
			h.ServeHTTP(rec, req)

			assert.Equal(t, tt.wantStatus, rec.Code, "unexpected status for path %q", tt.path)
			assert.Equal(t, tt.wantReached, reached, "downstream reached=%v for path %q", reached, tt.path)
			if tt.wantLogged {
				assert.NotZero(t, recorded.Len(), "expected a debug log line for invalid path %q", tt.path)
			} else {
				assert.Zero(t, recorded.Len(), "unexpected log line for path %q", tt.path)
			}
		})
	}
}

// TestInvalidUTF8PathRejected exercises the UI mount (the chi router with the
// static file server mounted at "/") to ensure that validateRequestPath,
// scoped to the UI FileServer, rejects percent-encoded paths decoding to
// invalid UTF-8 with a 400 Bad Request, while valid paths reach the file
// server normally and non-UI (API) routes are left untouched by the guard.
//
// This is a regression test for https://github.com/flipt-io/flipt/issues/6337
// where such paths produced a 500 Internal Server Error with no log output.
func TestInvalidUTF8PathRejected(t *testing.T) {
	core, recorded := observer.New(zapcore.DebugLevel)
	logger := zap.New(core)

	r := chi.NewRouter()

	// Non-UI routes are mounted on the root router without validateRequestPath,
	// mirroring how API routes are mounted in NewHTTPServer. An invalid UTF-8
	// path on such a route must reach the handler, not be rejected by the guard
	// that is scoped to the UI FileServer below.
	var apiReached bool
	r.Mount("/api/v1", http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		apiReached = true
		w.WriteHeader(http.StatusOK)
	}))

	err := mountUI(r, logger)
	require.NoError(t, err)

	s := httptest.NewServer(r)
	t.Cleanup(s.Close)

	// fetch sends a request preserving the raw percent-encoded path (like
	// `curl --path-as-is`), so the server percent-decodes it itself.
	fetch := func(t *testing.T, path string) (int, []byte) {
		t.Helper()
		req, err := http.NewRequestWithContext(t.Context(), http.MethodGet, s.URL+path, nil)
		require.NoError(t, err)
		res, err := http.DefaultClient.Do(req)
		require.NoError(t, err)
		body, err := io.ReadAll(res.Body)
		require.NoError(t, err)
		err = res.Body.Close()
		require.NoError(t, err)
		return res.StatusCode, body
	}

	t.Run("invalid UTF-8 UI paths return 400 and are logged", func(t *testing.T) {
		for _, p := range []string{"/%c0", "/%c1", "/%ff", "/%e0%80%af", "/%ed%a0%80"} {
			before := recorded.Len()
			statusCode, body := fetch(t, p)
			assert.Equalf(t, http.StatusBadRequest, statusCode,
				"path %s: want 400, got %d (body=%q)", p, statusCode, string(body))
			assert.Greater(t, recorded.Len(), before, "path %s: want a debug log line", p)
		}
	})

	t.Run("invalid UTF-8 on non-UI route is not intercepted", func(t *testing.T) {
		// The guard is scoped to the UI FileServer; an invalid UTF-8 path on a
		// non-UI route must reach that route's handler instead of being rejected
		// with 400 by validateRequestPath.
		apiReached = false
		before := recorded.Len()
		statusCode, _ := fetch(t, "/api/v1/%c0")
		assert.Equal(t, http.StatusOK, statusCode,
			"non-UI route should be reached, not rejected with 400")
		assert.True(t, apiReached, "non-UI handler should have been reached")
		assert.Equal(t, before, recorded.Len(),
			"validateRequestPath must not fire for non-UI routes")
	})

	t.Run("valid unknown path returns 404 and is not logged", func(t *testing.T) {
		before := recorded.Len()
		statusCode, _ := fetch(t, "/nonexistent-path-12345")
		assert.Equal(t, http.StatusNotFound, statusCode)
		assert.Equal(t, before, recorded.Len())
	})

	t.Run("valid utf8 unknown path returns 404", func(t *testing.T) {
		statusCode, _ := fetch(t, "/caf%c3%a9")
		assert.Equal(t, http.StatusNotFound, statusCode)
	})

	t.Run("root serves index", func(t *testing.T) {
		statusCode, _ := fetch(t, "/")
		assert.Equal(t, http.StatusOK, statusCode)
	})
}
