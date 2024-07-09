package webhook

import (
	"context"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.flipt.io/flipt/internal/server/audit"
	"go.flipt.io/flipt/rpc/flipt"
	"go.uber.org/zap"
)

type dummy struct{}

func (d *dummy) SendAudit(ctx context.Context, e audit.Event) (*http.Response, error) {
	return nil, nil
}

func TestSink(t *testing.T) {
	s := NewSink(zap.NewNop(), &dummy{})

	assert.Equal(t, "webhook", s.String())

	err := s.SendAudits(context.TODO(), []audit.Event{
		{
			Version: "0.1",
			Type:    string(flipt.SubjectFlag),
			Action:  string(flipt.ActionCreate),
		},
		{
			Version: "0.1",
			Type:    string(flipt.SubjectConstraint),
			Action:  string(flipt.ActionUpdate),
		},
	})

	require.NoError(t, err)
	require.NoError(t, s.Close())
}
