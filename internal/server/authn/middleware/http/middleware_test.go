package http_middleware

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.flipt.io/flipt/internal/config"
	middlewarecommon "go.flipt.io/flipt/internal/server/authn/middleware/common"
	"go.flipt.io/flipt/rpc/flipt/auth"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/proto"
)

func TestHandler(t *testing.T) {
	for _, tt := range []struct {
		name, method string
	}{
		{"expire", http.MethodPut},
		{"revoke", http.MethodDelete},
	} {
		t.Run(tt.name, func(t *testing.T) {
			handler := func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
			}

			middleware := NewHTTPMiddleware(config.AuthenticationSessionConfig{
				Domain: "localhost",
			})

			srv := middleware.Handler(http.HandlerFunc(handler))

			req := httptest.NewRequestWithContext(t.Context(), tt.method, "http://www.your-domain.com/auth/v1/self/"+tt.name, nil)
			w := httptest.NewRecorder()

			srv.ServeHTTP(w, req)
			assert.Equal(t, http.StatusOK, w.Code)

			res := w.Result()
			defer res.Body.Close()

			cookies := res.Cookies()
			assertCookiesCleared(t, cookies, http.SameSiteStrictMode)
		})
	}
}

func TestErrorHandler(t *testing.T) {
	for _, tt := range []struct {
		name, method string
	}{
		{"expire", http.MethodPut},
		{"revoke", http.MethodDelete},
	} {
		t.Run(tt.name, func(t *testing.T) {
			const defaultResponseBody = "default handler called"
			middleware := NewHTTPMiddleware(config.AuthenticationSessionConfig{
				Domain: "localhost",
			})

			middleware.defaultErrHandler = func(ctx context.Context, sm *runtime.ServeMux, m runtime.Marshaler, w http.ResponseWriter, r *http.Request, err error) {
				_, _ = w.Write([]byte(defaultResponseBody))
			}

			req := httptest.NewRequestWithContext(t.Context(), tt.method, "http://www.your-domain.com/auth/v1/self/"+tt.name, nil)
			req.Header.Add("Cookie", "flipt_client_token=expired")
			w := httptest.NewRecorder()

			err := status.Errorf(codes.Unauthenticated, "token expired")
			middleware.ErrorHandler(context.TODO(), nil, nil, w, req, err)

			res := w.Result()
			defer res.Body.Close()

			body, err := io.ReadAll(res.Body)
			require.NoError(t, err)
			assert.Equal(t, []byte(defaultResponseBody), body)

			cookies := res.Cookies()
			assertCookiesCleared(t, cookies, http.SameSiteStrictMode)
		})
	}
}

func assertCookiesCleared(t *testing.T, cookies []*http.Cookie, sameSite http.SameSite) {
	t.Helper()

	assert.Len(t, cookies, 2)

	cookiesMap := make(map[string]*http.Cookie)
	for _, cookie := range cookies {
		cookiesMap[cookie.Name] = cookie
	}

	for _, cookieName := range []string{stateCookieKey, middlewarecommon.TokenCookieKey} {
		assert.Contains(t, cookiesMap, cookieName)
		assert.Empty(t, cookiesMap[cookieName].Value)
		assert.Equal(t, "localhost", cookiesMap[cookieName].Domain)
		assert.Equal(t, "/", cookiesMap[cookieName].Path)
		assert.Equal(t, -1, cookiesMap[cookieName].MaxAge)
		assert.True(t, cookiesMap[cookieName].HttpOnly)
		assert.Equal(t, sameSite, cookiesMap[cookieName].SameSite)

		if sameSite == http.SameSiteNoneMode {
			assert.True(t, cookiesMap[cookieName].Secure)
		}
	}
}

func TestForwardRevokeOIDCResponseOption(t *testing.T) {
	middleware := NewHTTPMiddleware(config.AuthenticationSessionConfig{
		Domain: "localhost",
		Secure: true,
	})

	t.Run("revoke OIDC success via GET clears cookies", func(t *testing.T) {
		ctx := metadata.NewOutgoingContext(t.Context(), metadata.Pairs("x-http-method", http.MethodGet))
		w := httptest.NewRecorder()

		err := middleware.ForwardRevokeOIDCResponseOption(ctx, w, &auth.RevokeOIDCResponse{})
		require.NoError(t, err)

		res := w.Result()
		defer res.Body.Close()
		assertCookiesCleared(t, res.Cookies(), http.SameSiteNoneMode)
	})

	t.Run("revoke OIDC success via POST does not clear cookies", func(t *testing.T) {
		ctx := metadata.NewOutgoingContext(t.Context(), metadata.Pairs("x-http-method", http.MethodPost))
		w := httptest.NewRecorder()

		err := middleware.ForwardRevokeOIDCResponseOption(ctx, w, &auth.RevokeOIDCResponse{})
		require.NoError(t, err)

		res := w.Result()
		defer res.Body.Close()
		assert.Empty(t, res.Cookies())
	})

	t.Run("non-revoke response does not clear cookies", func(t *testing.T) {
		ctx := metadata.NewOutgoingContext(t.Context(), metadata.Pairs("x-http-method", http.MethodGet))
		w := httptest.NewRecorder()

		err := middleware.ForwardRevokeOIDCResponseOption(ctx, w, &struct{ proto.Message }{})
		require.NoError(t, err)

		res := w.Result()
		defer res.Body.Close()
		assert.Empty(t, res.Cookies())
	})

	t.Run("no x-http-method metadata does not clear cookies", func(t *testing.T) {
		ctx := t.Context()
		w := httptest.NewRecorder()

		err := middleware.ForwardRevokeOIDCResponseOption(ctx, w, &auth.RevokeOIDCResponse{})
		require.NoError(t, err)

		res := w.Result()
		defer res.Body.Close()
		assert.Empty(t, res.Cookies())
	})

	t.Run("revoke OIDC success via GET clears cookies using incoming context", func(t *testing.T) {
		ctx := metadata.NewIncomingContext(t.Context(), metadata.Pairs("x-http-method", http.MethodGet))
		w := httptest.NewRecorder()

		err := middleware.ForwardRevokeOIDCResponseOption(ctx, w, &auth.RevokeOIDCResponse{})
		require.NoError(t, err)

		res := w.Result()
		defer res.Body.Close()
		assertCookiesCleared(t, res.Cookies(), http.SameSiteNoneMode)
	})
}
