package config

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestWithForwardPrefix(t *testing.T) {
	ctx := context.Background()
	assert.Empty(t, getForwardPrefix(ctx))
	ctx = WithForwardPrefix(ctx, "/some/prefix")
	assert.Equal(t, "/some/prefix", getForwardPrefix(ctx))
}

func TestRequiresDatabase(t *testing.T) {
	tests := []struct {
		name   string
		f      func(m *AuthenticationMethods)
		expect bool
	}{
		{
			"token",
			func(m *AuthenticationMethods) { m.Token.Enabled = true },
			true,
		},
		{
			"github",
			func(m *AuthenticationMethods) { m.Github.Enabled = true },
			true,
		},
		{
			"oidc",
			func(m *AuthenticationMethods) { m.OIDC.Enabled = true },
			true,
		},
		{
			"kubernetes",
			func(m *AuthenticationMethods) { m.Kubernetes.Enabled = true },
			true,
		},
		{
			"jwt",
			func(m *AuthenticationMethods) { m.JWT.Enabled = true },
			false,
		},
		{
			"cloud",
			func(m *AuthenticationMethods) { m.Cloud.Enabled = true },
			false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := AuthenticationConfig{}
			tt.f(&c.Methods)
			assert.Equal(t, tt.expect, c.RequiresDatabase())
			enabledMethogs := c.Methods.EnabledMethods()
			assert.Len(t, enabledMethogs, 1)
			fmt.Println(enabledMethogs[0].Name())
			assert.Equal(t, tt.name, enabledMethogs[0].Name())
		})
	}
}
