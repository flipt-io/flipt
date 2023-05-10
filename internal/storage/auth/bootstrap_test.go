package auth_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"go.flipt.io/flipt/internal/storage/auth"
	"go.flipt.io/flipt/internal/storage/auth/memory"
)

func TestBootstrap(t *testing.T) {
	store := memory.NewStore()
	s, err := auth.Bootstrap(context.TODO(), store, auth.WithExpiration(time.Minute), auth.WithToken("this-is-a-token"))
	assert.NoError(t, err)
	assert.NotEmpty(t, s)
}
