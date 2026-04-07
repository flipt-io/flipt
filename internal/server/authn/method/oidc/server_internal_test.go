package oidc

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCallbackURL(t *testing.T) {
	tests := []struct {
		name string
		host string
		want string
	}{
		{
			name: "plain",
			host: "localhost",
			want: "localhost/auth/v1/method/oidc/foo/callback",
		},
		{
			name: "no trailing slash",
			host: "localhost:8080",
			want: "localhost:8080/auth/v1/method/oidc/foo/callback",
		},
		{
			name: "with trailing slash",
			host: "localhost:8080/",
			want: "localhost:8080/auth/v1/method/oidc/foo/callback",
		},
		{
			name: "with protocol",
			host: "http://localhost:8080",
			want: "http://localhost:8080/auth/v1/method/oidc/foo/callback",
		},
		{
			name: "with protocol and trailing slash",
			host: "http://localhost:8080/",
			want: "http://localhost:8080/auth/v1/method/oidc/foo/callback",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := callbackURL(tt.host, "foo")
			assert.Equal(t, tt.want, got)
		})
	}
}

func Test_Server_SkipsAuthentication(t *testing.T) {
	server := &Server{}
	assert.True(t, server.SkipsAuthentication(t.Context()))
}

func TestClaims_FallbackFrom(t *testing.T) {
	tests := []struct {
		name    string
		initial claims
		extra   map[string]any
		want    claims
	}{
		{
			name:    "no extra claims - unchanged",
			initial: claims{Name: new("John"), Email: new("john@example.com"), Picture: new("http://img.jpg")},
			extra:   nil,
			want:    claims{Name: new("John"), Email: new("john@example.com"), Picture: new("http://img.jpg")},
		},
		{
			name:    "extra claims but all fields already populated - unchanged",
			initial: claims{Name: new("John"), Email: new("john@example.com"), Picture: new("http://img.jpg")},
			extra:   map[string]any{"name": "Jane", "email": "jane@example.com", "picture": "http://jane.jpg"},
			want:    claims{Name: new("John"), Email: new("john@example.com"), Picture: new("http://img.jpg")},
		},
		{
			name:    "all fields nil - fallback from extra",
			initial: claims{},
			extra:   map[string]any{"name": "Jane", "email": "jane@example.com", "picture": "http://jane.jpg"},
			want:    claims{Name: new("Jane"), Email: new("jane@example.com"), Picture: new("http://jane.jpg")},
		},
		{
			name:    "some fields nil - partial fallback",
			initial: claims{Name: new("John")},
			extra:   map[string]any{"name": "Jane", "email": "jane@example.com", "picture": "http://jane.jpg"},
			want:    claims{Name: new("John"), Email: new("jane@example.com"), Picture: new("http://jane.jpg")},
		},
		{
			name:    "empty string fields - fallback from extra",
			initial: claims{Name: new(""), Email: new(""), Picture: new("")},
			extra:   map[string]any{"name": "Jane", "email": "jane@example.com", "picture": "http://jane.jpg"},
			want:    claims{Name: new("Jane"), Email: new("jane@example.com"), Picture: new("http://jane.jpg")},
		},
		{
			name:    "partial empty string - fallback only empty",
			initial: claims{Name: new("John"), Email: new(""), Picture: new("http://existing.jpg")},
			extra:   map[string]any{"name": "Jane", "email": "jane@example.com", "picture": "http://jane.jpg"},
			want:    claims{Name: new("John"), Email: new("jane@example.com"), Picture: new("http://existing.jpg")},
		},
		{
			name:    "non-string values in extra - ignored",
			initial: claims{},
			extra:   map[string]any{"name": 123, "email": true, "picture": nil},
			want:    claims{},
		},
		{
			name:    "empty string in extra - ignored",
			initial: claims{},
			extra:   map[string]any{"name": "", "email": "", "picture": ""},
			want:    claims{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.initial.fallbackFrom(tt.extra)
			assert.Equal(t, tt.want, tt.initial)
		})
	}
}
