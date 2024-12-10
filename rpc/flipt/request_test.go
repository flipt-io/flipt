package flipt

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestListNamespaceRequest_Request(t *testing.T) {
	req := &ListNamespaceRequest{}
	assert.Equal(t, []Request{NewRequest(ResourceNamespace, ActionRead)}, req.Request())
}
