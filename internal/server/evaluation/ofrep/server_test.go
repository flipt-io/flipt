package ofrep

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_Server_SkipsAuthorization(t *testing.T) {
	server := &Server{}
	assert.True(t, server.SkipsAuthorization(t.Context()))

	server = New(nil, nil, nil, WithAuthorizationEnabled(true))
	assert.False(t, server.SkipsAuthorization(t.Context()))
}
