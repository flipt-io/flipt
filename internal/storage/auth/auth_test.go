package auth

import (
	"encoding/base64"
	"testing"

	"github.com/stretchr/testify/require"
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
