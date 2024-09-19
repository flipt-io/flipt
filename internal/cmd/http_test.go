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
