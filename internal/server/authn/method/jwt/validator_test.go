package jwt

import (
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"maps"
	"testing"
	"time"

	"github.com/hashicorp/cap/jwt"
	"github.com/hashicorp/cap/oidc"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestValidator_CustomClaims(t *testing.T) {
	priv, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)

	ks, err := jwt.NewStaticKeySet([]crypto.PublicKey{priv.Public()})
	require.NoError(t, err)

	cap, err := jwt.NewValidator(ks)
	require.NoError(t, err)

	var (
		expected = jwt.Expected{SigningAlgorithms: []jwt.Alg{jwt.RS256}}
		now      = time.Now()
	)

	for _, tt := range []struct {
		name    string
		custom  []CustomClaim
		claims  map[string]any
		wantErr string
	}{
		{
			name:   "no custom claims",
			claims: map[string]any{},
		},
		{
			name:   "string claim matches",
			custom: []CustomClaim{{Name: "team", Values: []string{"platform"}}},
			claims: map[string]any{"team": "platform"},
		},
		{
			name:   "string claim matches any of the values",
			custom: []CustomClaim{{Name: "team", Values: []string{"platform", "infra"}}},
			claims: map[string]any{"team": "infra"},
		},
		{
			name:    "string claim does not match",
			custom:  []CustomClaim{{Name: "team", Values: []string{"platform"}}},
			claims:  map[string]any{"team": "sales"},
			wantErr: `invalid "team" claim: value is not one of the expected values`,
		},
		{
			name:    "string claim must match exactly",
			custom:  []CustomClaim{{Name: "team", Values: []string{"platform"}}},
			claims:  map[string]any{"team": "platform-eng"},
			wantErr: `invalid "team" claim: value is not one of the expected values`,
		},
		{
			name:    "claim names are case sensitive",
			custom:  []CustomClaim{{Name: "Team", Values: []string{"platform"}}},
			claims:  map[string]any{"team": "platform"},
			wantErr: `invalid "Team" claim: claim is missing`,
		},
		{
			name:    "missing claim",
			custom:  []CustomClaim{{Name: "team", Values: []string{"platform"}}},
			claims:  map[string]any{},
			wantErr: `invalid "team" claim: claim is missing`,
		},
		{
			name:   "namespaced claim name",
			custom: []CustomClaim{{Name: "https://example.com/roles", Values: []string{"admin"}}},
			claims: map[string]any{"https://example.com/roles": "admin"},
		},
		{
			name:   "array claim contains one of the values",
			custom: []CustomClaim{{Name: "groups", Values: []string{"flipt-admins", "flipt-users"}}},
			claims: map[string]any{"groups": []string{"eng", "flipt-users"}},
		},
		{
			name:    "array claim contains none of the values",
			custom:  []CustomClaim{{Name: "groups", Values: []string{"flipt-admins"}}},
			claims:  map[string]any{"groups": []string{"eng", "flipt-users"}},
			wantErr: `invalid "groups" claim: value is not one of the expected values`,
		},
		{
			name:    "empty array claim",
			custom:  []CustomClaim{{Name: "groups", Values: []string{"flipt-admins"}}},
			claims:  map[string]any{"groups": []string{}},
			wantErr: `invalid "groups" claim: value is not one of the expected values`,
		},
		{
			name:   "boolean claim",
			custom: []CustomClaim{{Name: "verified", Values: []string{"true"}}},
			claims: map[string]any{"verified": true},
		},
		{
			name:    "boolean claim mismatch",
			custom:  []CustomClaim{{Name: "verified", Values: []string{"true"}}},
			claims:  map[string]any{"verified": false},
			wantErr: `invalid "verified" claim: value is not one of the expected values`,
		},
		{
			name:   "numeric claim",
			custom: []CustomClaim{{Name: "tier", Values: []string{"2"}}},
			claims: map[string]any{"tier": 2},
		},
		{
			name:    "object claim never matches",
			custom:  []CustomClaim{{Name: "org", Values: []string{"map[id:1]"}}},
			claims:  map[string]any{"org": map[string]any{"id": 1}},
			wantErr: `invalid "org" claim: value is not one of the expected values`,
		},
		{
			name: "every custom claim must be satisfied",
			custom: []CustomClaim{
				{Name: "team", Values: []string{"platform"}},
				{Name: "env", Values: []string{"prod"}},
			},
			claims:  map[string]any{"team": "platform", "env": "dev"},
			wantErr: `invalid "env" claim: value is not one of the expected values`,
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			claims := map[string]any{
				"iss": "flipt.io",
				"iat": now.Unix(),
				"exp": now.Add(time.Hour).Unix(),
			}
			maps.Copy(claims, tt.claims)

			token := oidc.TestSignJWT(t, priv, string(jwt.RS256), claims, []byte("test-key"))

			got, err := NewValidator(cap, expected, WithCustomClaims(tt.custom)).Validate(t.Context(), token)
			if tt.wantErr != "" {
				require.Error(t, err)
				require.EqualError(t, err, tt.wantErr)
				assert.Nil(t, got)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, "flipt.io", got["iss"])
		})
	}
}
