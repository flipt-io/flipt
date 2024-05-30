package template

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
	"go.flipt.io/flipt/internal/server/audit"
	"go.flipt.io/flipt/rpc/flipt"
	"go.uber.org/zap"
)

type dummyExecuter struct{}

func (d *dummyExecuter) Execute(_ context.Context, _ audit.Event) error {
	return nil
}

func TestSink(t *testing.T) {
	var s audit.Sink = &Sink{
		logger:    zap.NewNop(),
		executers: []Executer{&dummyExecuter{}},
	}

	require.Equal(t, "templates", s.String())

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
