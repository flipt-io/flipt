package template

import (
	"context"
	"io"
	"testing"
	"text/template"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.flipt.io/flipt/internal/server/audit"
)

type dummyRetrier struct{}

func (d *dummyRetrier) RequestRetry(_ context.Context, _ []byte, _ audit.RequestCreator) error {
	return nil
}

func TestExecuter_JSON_Failure(t *testing.T) {
	tmpl, err := template.New("").Parse(`this is invalid JSON {{ .Type }}, {{ .Action }}`)
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

	assert.EqualError(t, err, "invalid JSON: this is invalid JSON flag, created")
}

func TestExecuter_Execute(t *testing.T) {
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

func TestExecuter_createRequest(t *testing.T) {
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

	req, err := whTemplate.createRequest(context.TODO(), []byte(`{"hello": "world"}`))
	require.NoError(t, err)

	assert.Equal(t, "https://flipt-webhook.io/webhook", req.URL.String())

	b, err := io.ReadAll(req.Body)
	require.NoError(t, err)

	assert.Equal(t, `{"hello": "world"}`, string(b))
}
