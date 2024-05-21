package webhook

import (
	"context"
	"io"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.flipt.io/flipt/internal/server/audit"
	"go.flipt.io/flipt/rpc/flipt"
	"go.uber.org/zap"
)

type dummyRetrier struct{}

func (d *dummyRetrier) RequestRetry(_ context.Context, _ []byte, _ audit.RequestCreator) error {
	return nil
}

func TestConstructorWebhookClient(t *testing.T) {
	client := NewWebhookClient(zap.NewNop(), "https://flipt-webhook.io/webhook", "")

	require.NotNil(t, client)

	whClient, ok := client.(*webhookClient)
	require.True(t, ok, "client should be a webhookClient")

	assert.Equal(t, "https://flipt-webhook.io/webhook", whClient.url)
	assert.Empty(t, whClient.signingSecret)
}

func TestWebhookClient(t *testing.T) {
	whclient := &webhookClient{
		logger:             zap.NewNop(),
		url:                "https://flipt-webhook.io/webhook",
		maxBackoffDuration: 15 * time.Second,
		retryableClient:    &dummyRetrier{},
	}

	err := whclient.SendAudit(context.TODO(), audit.Event{
		Type:   string(flipt.SubjectFlag),
		Action: string(flipt.ActionCreate),
	})

	require.NoError(t, err)
}

func TestWebhookClient_createRequest(t *testing.T) {
	whclient := &webhookClient{
		logger:             zap.NewNop(),
		url:                "https://flipt-webhook.io/webhook",
		maxBackoffDuration: 15 * time.Second,
		retryableClient:    &dummyRetrier{},
		signingSecret:      "s3crET",
	}

	req, err := whclient.createRequest(context.TODO(), []byte(`{"hello": "world"}`))
	require.NoError(t, err)

	assert.Equal(t, "https://flipt-webhook.io/webhook", req.URL.String())

	b, err := io.ReadAll(req.Body)
	require.NoError(t, err)

	assert.Equal(t, `{"hello": "world"}`, string(b))
	assert.Equal(t, "18140945e2c08976ed68adeb0bf80d2fa695d533957cb0e531812798de6b9fad", req.Header.Get(fliptSignatureHeader))
}
