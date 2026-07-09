package names

import (
	"crypto/rand"
	"fmt"
	"math/big"
)

var adjectives = []string{
	"brave",
	"bright",
	"calm",
	"clever",
	"curious",
	"eager",
	"gentle",
	"happy",
	"kind",
	"lively",
	"quiet",
	"swift",
}

var nouns = []string{
	"badger",
	"falcon",
	"fox",
	"heron",
	"otter",
	"panda",
	"rabbit",
	"raven",
	"tiger",
	"turtle",
	"whale",
	"wolf",
}

// Random returns a human-readable, URL-safe random name.
func Random() string {
	return fmt.Sprintf("%s-%s-%s", choose(adjectives), choose(nouns), randomHex(2))
}

func choose(values []string) string {
	n, err := rand.Int(rand.Reader, big.NewInt(int64(len(values))))
	if err != nil {
		return values[0]
	}

	return values[n.Int64()]
}

func randomHex(size int) string {
	b := make([]byte, size)
	if _, err := rand.Read(b); err != nil {
		return "0000"
	}

	return fmt.Sprintf("%x", b)
}
