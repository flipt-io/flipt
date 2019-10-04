package server

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

type noopCacher struct{}

func (n *noopCacher) Get(key interface{}) (interface{}, bool) {
	return nil, false
}

func (n *noopCacher) Add(key interface{}, value interface{}) bool {
	return false
}

func (n *noopCacher) Remove(key interface{}) bool {
	return false
}

func TestWithCache(t *testing.T) {
	var (
		opt    = WithCache(&noopCacher{})
		server = &Server{}
	)

	opt(server)

	assert.NotNil(t, server.cache)
}
