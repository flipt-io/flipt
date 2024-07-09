package cloud

import (
	"context"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.flipt.io/flipt/internal/server/authz/engine/rego/source"
)

func TestFetch(t *testing.T) {
	type want struct {
		hash   source.Hash
		status int
	}
	tests := []struct {
		name        string
		seen        source.Hash
		want        want
		expectedErr error
	}{
		{
			name: "Valid request with matching etag",
			seen: []byte("1234"),
			want: want{
				hash: []byte("1234"),
			},
			expectedErr: source.ErrNotModified,
		},
		{
			name: "Valid request with different etag",
			seen: []byte("1234"),
			want: want{
				hash: []byte("5678"),
			},
		},
		{
			name: "Error status code",
			seen: []byte("1234"),
			want: want{
				status: http.StatusNotFound,
			},
			expectedErr: errors.New("unexpected status code: 404"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svr := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, "application/json", r.Header.Get("Accept"))
				assert.Equal(t, "Bearer apiKey", r.Header.Get("Authorization"))
				assert.NotEmpty(t, r.Header.Get("If-None-Match"))

				w.Header().Set("ETag", string(tt.want.hash))

				if r.Header.Get("If-None-Match") == string(tt.want.hash) {
					w.WriteHeader(http.StatusNotModified)
					return
				}

				if tt.want.status != 0 {
					w.WriteHeader(tt.want.status)
					return
				}

				if _, err := io.WriteString(w, "Hello, world!"); err != nil {
					t.Fatalf("failed to write response: %v", err)
				}
			}))

			t.Cleanup(svr.Close)

			body, hash, err := fetch(context.Background(), tt.seen, svr.URL, "apiKey")

			assert.Equal(t, tt.expectedErr, err)

			if tt.expectedErr == nil {
				assert.Equal(t, tt.want.hash, hash)
				assert.Equal(t, []byte("Hello, world!"), body)
			}
		})
	}
}
