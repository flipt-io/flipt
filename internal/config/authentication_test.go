package config

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestWithForwardPrefix(t *testing.T) {
	ctx := context.Background()
	assert.Empty(t, getForwardPrefix(ctx))
	ctx = WithForwardPrefix(ctx, "/some/prefix")
	assert.Equal(t, "/some/prefix", getForwardPrefix(ctx))
}
