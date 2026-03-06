package method

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
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
			req := httptest.NewRequest("GET", "http://example.com/foo", nil)
			for k, v := range tt.headers {
				req.Header.Add(k, v)
			}
			md := ForwardPrefix(context.Background(), req)
			assert.Equal(t, tt.expected, md.Get(ForwardedPrefixKey))
		})
	}
}

func TestForwardCookies(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "http://example.com/auth/callback", nil)
	req.AddCookie(&http.Cookie{Name: stateCookieKey, Value: "state-value"})
	req.AddCookie(&http.Cookie{Name: tokenCookieKey, Value: "token-value"})

	md := ForwardCookies(context.Background(), req)

	assert.Equal(t, []string{"state-value"}, md.Get(stateCookieKey))
	assert.Equal(t, []string{"token-value"}, md.Get(tokenCookieKey))
}
