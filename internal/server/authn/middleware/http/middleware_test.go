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
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestHandler(t *testing.T) {
	handler := func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}

	middleware := NewHTTPMiddleware(config.AuthenticationSession{
		Domain: "localhost",
	})

	srv := middleware.Handler(http.HandlerFunc(handler))

	req := httptest.NewRequest(http.MethodPut, "http://www.your-domain.com/auth/v1/self/expire", nil)
	w := httptest.NewRecorder()

	srv.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)

	res := w.Result()
	defer res.Body.Close()

	cookies := res.Cookies()
	assertCookiesCleared(t, cookies)
}

func TestErrorHandler(t *testing.T) {
	const defaultResponseBody = "default handler called"
	var (
		middleware = NewHTTPMiddleware(config.AuthenticationSession{
			Domain: "localhost",
		})
	)

	middleware.defaultErrHandler = func(ctx context.Context, sm *runtime.ServeMux, m runtime.Marshaler, w http.ResponseWriter, r *http.Request, err error) {
		_, _ = w.Write([]byte(defaultResponseBody))
	}

	req := httptest.NewRequest(http.MethodPut, "http://www.your-domain.com/auth/v1/self/expire", nil)
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
	assertCookiesCleared(t, cookies)
}

func assertCookiesCleared(t *testing.T, cookies []*http.Cookie) {
	t.Helper()

	assert.Len(t, cookies, 2)

	cookiesMap := make(map[string]*http.Cookie)
	for _, cookie := range cookies {
		cookiesMap[cookie.Name] = cookie
	}

	for _, cookieName := range []string{stateCookieKey, middlewarecommon.TokenCookieKey} {
		assert.Contains(t, cookiesMap, cookieName)
		assert.Equal(t, "", cookiesMap[cookieName].Value)
		assert.Equal(t, "localhost", cookiesMap[cookieName].Domain)
		assert.Equal(t, "/", cookiesMap[cookieName].Path)
		assert.Equal(t, -1, cookiesMap[cookieName].MaxAge)
	}
}
