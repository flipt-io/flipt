package environments

import (
	"testing"

	"go.flipt.io/flipt/errors"
)

func TestValidateNamespaceKey(t *testing.T) {
	tests := []struct {
		key     string
		wantErr bool
	}{
		{key: "default"},
		{key: "team_a-1,b"},
		// existing namespaces created before keys were validated stay usable
		{key: "my.namespace"},
		{key: "", wantErr: true},
		{key: "a/b", wantErr: true},
		{key: "../secret", wantErr: true},
		{key: `a\b`, wantErr: true},
		{key: ".", wantErr: true},
		{key: "..", wantErr: true},
		{key: ".git", wantErr: true},
		{key: "a\x00b", wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.key, func(t *testing.T) {
			assertInvalid(t, ValidateNamespaceKey(tt.key), tt.wantErr)
		})
	}
}

func TestValidateNewNamespaceKey(t *testing.T) {
	tests := []struct {
		key     string
		wantErr bool
	}{
		{key: "default"},
		{key: "team_a-1,b"},
		{key: "", wantErr: true},
		{key: "my.namespace", wantErr: true},
		{key: "a b", wantErr: true},
		{key: "a/b", wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.key, func(t *testing.T) {
			assertInvalid(t, ValidateNewNamespaceKey(tt.key), tt.wantErr)
		})
	}
}

func TestRequestsValidateNamespaceKey(t *testing.T) {
	requests := map[string]func(namespace string) interface{ Validate() error }{
		"get namespace":    func(ns string) interface{ Validate() error } { return &GetNamespaceRequest{Key: ns} },
		"update namespace": func(ns string) interface{ Validate() error } { return &UpdateNamespaceRequest{Key: ns} },
		"delete namespace": func(ns string) interface{ Validate() error } { return &DeleteNamespaceRequest{Key: ns} },
		"get resource": func(ns string) interface{ Validate() error } {
			return &GetResourceRequest{NamespaceKey: ns, Key: "flag"}
		},
		"list resources": func(ns string) interface{ Validate() error } { return &ListResourcesRequest{NamespaceKey: ns} },
		"update resource": func(ns string) interface{ Validate() error } {
			return &UpdateResourceRequest{NamespaceKey: ns, Key: "flag"}
		},
		"delete resource": func(ns string) interface{ Validate() error } {
			return &DeleteResourceRequest{NamespaceKey: ns, Key: "flag"}
		},
	}

	for name, newRequest := range requests {
		for _, badKey := range []string{"", "a/b", "team-a/../secret", `a\b`, ".", "..", ".git", "a\x00b"} {
			t.Run(name+"/"+badKey, func(t *testing.T) {
				assertInvalid(t, newRequest(badKey).Validate(), true)
			})
		}

		t.Run(name+"/existing key", func(t *testing.T) {
			assertInvalid(t, newRequest("my.namespace").Validate(), false)
		})
	}

	t.Run("resource requests require a key", func(t *testing.T) {
		for _, req := range []interface{ Validate() error }{
			&GetResourceRequest{NamespaceKey: "default"},
			&UpdateResourceRequest{NamespaceKey: "default"},
			&DeleteResourceRequest{NamespaceKey: "default"},
		} {
			assertInvalid(t, req.Validate(), true)
		}
	})
}

func assertInvalid(t *testing.T, err error, want bool) {
	t.Helper()

	if !want {
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		return
	}

	if !errors.AsMatch[errors.ErrInvalid](err) {
		t.Fatalf("expected invalid error, got %v", err)
	}
}
