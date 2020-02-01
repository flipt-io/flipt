package server

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

type noopCacher struct{}

func (n *noopCacher) Get(key string) (interface{}, bool) {
	return nil, false
}

func (n *noopCacher) Set(key string, value interface{}) {
}

func (n *noopCacher) Delete(key string) {
}

func (n *noopCacher) Flush() {
}

func TestWithCache(t *testing.T) {
	var (
		opt    = WithCache(&noopCacher{})
		server = &Server{}
	)

	opt(server)

	assert.NotNil(t, server.cache)
}
