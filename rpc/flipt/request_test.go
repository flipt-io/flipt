package flipt

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestListNamespaceRequest_Request(t *testing.T) {
	req := &ListNamespaceRequest{}
	expected := NewRequest(ResourceNamespace, ActionRead)
	expected.Namespace = ""
	assert.Equal(t, []Request{expected}, req.Request())
}
