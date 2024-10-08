package evaluation

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_Server_AllowsNamespaceScopedAuthentication(t *testing.T) {
	server := &Server{}
	assert.True(t, server.AllowsNamespaceScopedAuthentication(context.Background()))
}

func Test_Server_SkipsAuthorization(t *testing.T) {
	server := &Server{}
	assert.True(t, server.SkipsAuthorization(context.Background()))
}
