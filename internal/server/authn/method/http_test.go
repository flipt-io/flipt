package method

import (
	"encoding/base64"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.flipt.io/flipt/internal/config"
	"go.flipt.io/flipt/rpc/flipt/auth"
	"google.golang.org/grpc/metadata"
)

func TestForwardCookies(t *testing.T) {
	req := httptest.NewRequestWithContext(t.Context(), http.MethodGet, "http://example.com/foo", nil)
	req.AddCookie(&http.Cookie{Name: stateCookieKey, Value: "state-value"})
	req.AddCookie(&http.Cookie{Name: tokenCookieKey, Value: "token-value"})

	md := ForwardCookies(t.Context(), req)

	assert.Equal(t, []string{"state-value"}, md.Get(stateCookieKey))
	assert.Equal(t, []string{"token-value"}, md.Get(tokenCookieKey))
}

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
			md := ForwardPrefix(t.Context(), req)
			assert.Equal(t, tt.expected, md.Get(ForwardedPrefixKey))
		})
	}
}

func encodeClientState(t *testing.T, originalState string) string {
	t.Helper()

	const token = "test-security-token"

	v, err := json.Marshal(clientStatePayload{
		SecurityToken: token,
		OriginalState: originalState,
	})
	require.NoError(t, err)

	return base64.URLEncoding.EncodeToString(v)
}

func TestForwardResponseOption(t *testing.T) {
	const token = "client-token"

	tests := []struct {
		name             string
		stateCookie      string
		incomingPrefixes []string
		outgoingPrefixes []string
		expectedLocation string
	}{
		{
			name:             "no state falls back to root",
			expectedLocation: "/",
		},
		{
			name:             "invalid state falls back to root",
			stateCookie:      "not-base64!!!",
			expectedLocation: "/",
		},
		{
			name:             "empty original_state falls back to root",
			stateCookie:      encodeClientState(t, ""),
			expectedLocation: "/",
		},
		{
			name:             "redirects to original_state via hash",
			stateCookie:      encodeClientState(t, "/flags"),
			expectedLocation: "/#/flags",
		},
		{
			name:             "redirects to nested original_state",
			stateCookie:      encodeClientState(t, "/namespaces/default/flags"),
			expectedLocation: "/#/namespaces/default/flags",
		},
		{
			name:             "prefix is preserved with hash redirect",
			stateCookie:      encodeClientState(t, "/flags"),
			outgoingPrefixes: []string{"/my-prefix"},
			expectedLocation: "/my-prefix/#/flags",
		},
		{
			name:             "incoming prefix is preserved with hash redirect",
			stateCookie:      encodeClientState(t, "/flags"),
			incomingPrefixes: []string{"/my-prefix"},
			expectedLocation: "/my-prefix/#/flags",
		},
		{
			name:             "prefix without state falls back to prefixed root",
			outgoingPrefixes: []string{"/my-prefix"},
			expectedLocation: "/my-prefix/",
		},
		{
			name:             "open redirect via protocol is rejected",
			stateCookie:      encodeClientState(t, "https://evil.example.com/flags"),
			expectedLocation: "/",
		},
		{
			name:             "open redirect via protocol-relative is rejected",
			stateCookie:      encodeClientState(t, "//evil.example.com/flags"),
			expectedLocation: "/",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := NewHTTPMiddleware(config.AuthenticationSessionConfig{Domain: "localhost"})

			ctx := t.Context()
			if len(tt.incomingPrefixes) > 0 || tt.stateCookie != "" {
				md := metadata.MD{}
				if tt.stateCookie != "" {
					md.Set(stateCookieKey, tt.stateCookie)
				}
				if len(tt.incomingPrefixes) > 0 {
					md.Set(ForwardedPrefixKey, tt.incomingPrefixes...)
				}
				ctx = metadata.NewIncomingContext(ctx, md)
			}

			if len(tt.outgoingPrefixes) > 0 {
				pairs := make([]string, 0, 2*len(tt.outgoingPrefixes))
				for _, prefix := range tt.outgoingPrefixes {
					pairs = append(pairs, ForwardedPrefixKey, prefix)
				}
				ctx = metadata.NewOutgoingContext(ctx, metadata.Pairs(pairs...))
			}

			w := httptest.NewRecorder()
			resp := &auth.CallbackResponse{ClientToken: token}

			require.NoError(t, m.ForwardResponseOption(ctx, w, resp))

			assert.Empty(t, resp.GetClientToken())
			assert.Equal(t, http.StatusFound, w.Code)
			assert.Equal(t, tt.expectedLocation, w.Header().Get("Location"))

			cookies := w.Result().Cookies()
			require.Len(t, cookies, 1)
			assert.Equal(t, tokenCookieKey, cookies[0].Name)
			assert.Equal(t, token, cookies[0].Value)
		})
	}
}

func TestSanitizeOriginalState(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
		expectOK bool
	}{
		{name: "empty is rejected", input: "", expectOK: false},
		{name: "root is kept", input: "/", expected: "/", expectOK: true},
		{name: "simple path", input: "/flags", expected: "/flags", expectOK: true},
		{name: "missing leading slash is normalized", input: "flags", expected: "/flags", expectOK: true},
		{name: "protocol-relative is rejected", input: "//evil.example.com", expectOK: false},
		{name: "absolute URL is rejected", input: "https://evil.example.com/flags", expectOK: false},
		{name: "backslash is rejected", input: `/flags\evil`, expectOK: false},
		{name: "newline is rejected", input: "/fla\ngs", expectOK: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actual, ok := sanitizeOriginalState(tt.input)
			assert.Equal(t, tt.expectOK, ok)
			if tt.expectOK {
				assert.Equal(t, tt.expected, actual)
			}
		})
	}
}
