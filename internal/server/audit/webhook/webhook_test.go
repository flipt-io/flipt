package webhook

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.flipt.io/flipt/internal/server/audit"
	"go.flipt.io/flipt/rpc/flipt"
	"go.uber.org/zap"
)

type dummy struct{}

func (d *dummy) SendAudit(ctx context.Context, e audit.Event) error {
	return nil
}

func TestSink(t *testing.T) {
	s := NewSink(zap.NewNop(), &dummy{})

	assert.Equal(t, "webhook", s.String())

	err := s.SendAudits(context.TODO(), []audit.Event{
		{
			Version: "0.1",
			Type:    flipt.SubjectFlag,
			Action:  flipt.ActionCreate,
		},
		{
			Version: "0.1",
			Type:    flipt.SubjectConstraint,
			Action:  flipt.ActionUpdate,
		},
	})

	require.NoError(t, err)
	require.NoError(t, s.Close())
}
