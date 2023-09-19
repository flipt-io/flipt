package template

import (
	"context"
	"testing"
	"text/template"
	"time"

	"github.com/stretchr/testify/require"
	"go.flipt.io/flipt/internal/server/audit"
)

type dummyRetrier struct{}

func (d *dummyRetrier) RequestRetry(_ context.Context, _ []byte, _ audit.RequestCreator) error {
	return nil
}

func TestExecuter(t *testing.T) {
	tmpl, err := template.New("").Parse(`{
		"type": "{{ .Type }}",
		"action": "{{ .Action }}"
	}`)

	require.NoError(t, err)

	whTemplate := &webhookTemplate{
		url:                "https://flipt-webhook.io/webhook",
		maxBackoffDuration: 15 * time.Second,
		retryableClient:    &dummyRetrier{},
		bodyTemplate:       tmpl,
	}

	err = whTemplate.Execute(context.TODO(), audit.Event{
		Type:   audit.FlagType,
		Action: audit.Create,
	})

	require.NoError(t, err)
}
