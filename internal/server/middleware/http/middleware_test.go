package http_middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEtag(t *testing.T) {
	req, _ := http.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()

	r := chi.NewRouter()
	r.Use(Etag)

	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("Hello World"))
	})

	r.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "0a4d55a8d778e5022fab701977c5d840bbc486d0", w.Header().Get("ETag"))

	req, _ = http.NewRequest("GET", "/", nil)
	req.Header.Set("If-None-Match", "0a4d55a8d778e5022fab701977c5d840bbc486d0")
	w = httptest.NewRecorder()

	r.ServeHTTP(w, req)
	require.Equal(t, http.StatusNotModified, w.Code)
}
