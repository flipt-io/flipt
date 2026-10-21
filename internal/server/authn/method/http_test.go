package method

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.flipt.io/flipt/internal/config"
	"google.golang.org/grpc/metadata"
)

func TestForwardPrefix(t *testing.T) {
	tests := []struct {
		name     string
		headers  map[string]string
		expected []string
	}{
		{"forward", map[string]string{"X-Forwarded-Prefix": "/my-prefix"}, []string{"/my-prefix"}},
		{"none", map[string]string{}, nil},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequestWithContext(t.Context(), "GET", "http://example.com/foo", nil)
			for k, v := range tt.headers {
				req.Header.Add(k, v)
			}
			md := ForwardPrefix(context.Background(), req)
			assert.Equal(t, tt.expected, md.Get(ForwardedPrefixKey))
		})
	}
}

func TestForwardCookies(t *testing.T) {
	req := httptest.NewRequestWithContext(t.Context(), http.MethodGet, "http://example.com/auth/callback", nil)
	req.AddCookie(&http.Cookie{Name: stateCookieKey, Value: "state-value"})
	req.AddCookie(&http.Cookie{Name: tokenCookieKey, Value: "token-value"})

	req.AddCookie(&http.Cookie{Name: VerifierCookieKey, Value: "verifier-value"})

	md := ForwardCookies(context.Background(), req)

	assert.Equal(t, []string{"state-value"}, md.Get(stateCookieKey))
	assert.Equal(t, []string{"token-value"}, md.Get(tokenCookieKey))
	assert.Equal(t, []string{"verifier-value"}, md.Get(VerifierCookieKey))

	t.Run("verifier generated for authorize takes precedence", func(t *testing.T) {
		req := req.WithContext(context.WithValue(req.Context(), verifierContextKey{}, "generated"))
		md := ForwardCookies(context.Background(), req)
		assert.Equal(t, []string{"generated"}, md.Get(VerifierCookieKey))
	})
}

func TestHandlerVerifierCookie(t *testing.T) {
	m := NewHTTPMiddleware(config.AuthenticationSession{Domain: "example.com", Secure: true, StateLifetime: 10 * time.Minute})

	authorize := func(t *testing.T, path string) (*httptest.ResponseRecorder, metadata.MD) {
		t.Helper()
		var md metadata.MD
		handler := m.Handler(http.HandlerFunc(func(_ http.ResponseWriter, r *http.Request) {
			md = ForwardCookies(r.Context(), r)
		}))
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, httptest.NewRequestWithContext(t.Context(), http.MethodGet, "http://example.com"+path, nil))
		return rec, md
	}

	cookie := func(rec *httptest.ResponseRecorder, name string) *http.Cookie {
		for _, c := range rec.Result().Cookies() {
			if c.Name == name {
				return c
			}
		}
		return nil
	}

	t.Run("oidc authorize binds a verifier to the callback", func(t *testing.T) {
		rec, md := authorize(t, "/auth/v1/method/oidc/google/authorize")

		state, verifier := cookie(rec, stateCookieKey), cookie(rec, VerifierCookieKey)
		require.NotNil(t, state)
		require.NotNil(t, verifier)
		assert.NotEmpty(t, verifier.Value)
		assert.NotEqual(t, state.Value, verifier.Value)
		assert.Equal(t, "/auth/v1/method/oidc/google/callback", verifier.Path)
		assert.Equal(t, state.Path, verifier.Path)
		assert.Equal(t, state.Domain, verifier.Domain)
		assert.Equal(t, state.Expires, verifier.Expires)
		assert.True(t, verifier.HttpOnly)
		assert.True(t, verifier.Secure)
		assert.Equal(t, http.SameSiteLaxMode, verifier.SameSite)

		// the verifier reaches AuthorizeURL on the same request
		assert.Equal(t, []string{verifier.Value}, md.Get(VerifierCookieKey))
	})

	t.Run("github authorize does not set a verifier", func(t *testing.T) {
		rec, md := authorize(t, "/auth/v1/method/github/authorize")

		assert.NotNil(t, cookie(rec, stateCookieKey))
		assert.Nil(t, cookie(rec, VerifierCookieKey))
		assert.Empty(t, md.Get(VerifierCookieKey))
	})
}
