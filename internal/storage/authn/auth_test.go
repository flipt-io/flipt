package authn

import (
	"encoding/base64"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.flipt.io/flipt/rpc/flipt/auth"
)

func FuzzHashClientToken(f *testing.F) {
	for _, seed := range []string{
		"hello, world",
		"supersecretstring",
		"egGpvIxtdG6tI3OIJjXOrv7xZW3hRMYg/Lt/G6X/UEwC",
	} {
		f.Add(seed)
	}
	for _, seed := range [][]byte{{}, {0}, {9}, {0xa}, {0xf}, {1, 2, 3, 4}} {
		f.Add(string(seed))
	}
	f.Fuzz(func(t *testing.T, token string) {
		hashed, err := HashClientToken(token)
		require.NoError(t, err)
		require.NotEmpty(t, hashed, "hashed result is empty")

		_, err = base64.URLEncoding.DecodeString(hashed)
		require.NoError(t, err)
	})
}

func TestDelete_Valid(t *testing.T) {
	dreq := Delete(WithID("987654321"), WithMethod(auth.Method_METHOD_TOKEN), WithExpiredBefore(time.Now()))
	err := dreq.Valid()

	assert.NoError(t, err)
}

func TestDelete_Invalid(t *testing.T) {
	dreq := Delete()
	err := dreq.Valid()

	assert.Error(t, err)
}

func TestGenerateRandomToken(t *testing.T) {
	s := GenerateRandomToken()

	assert.NotEmpty(t, s)
}
