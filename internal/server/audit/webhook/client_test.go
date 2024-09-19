package webhook

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/hashicorp/go-retryablehttp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.flipt.io/flipt/internal/server/audit"
	"go.flipt.io/flipt/rpc/flipt"
	"go.uber.org/zap"
)

func TestConstructorWebhookClient(t *testing.T) {
	client := NewWebhookClient(zap.NewNop(), "https://flipt-webhook.io/webhook", "", 8*time.Second)

	require.NotNil(t, client)

	whClient, ok := client.(*webhookClient)
	require.True(t, ok, "client should be a webhookClient")

	assert.Equal(t, "https://flipt-webhook.io/webhook", whClient.url)
	assert.Empty(t, whClient.signingSecret)
	assert.Equal(t, 8*time.Second, whClient.httpClient.RetryWaitMax)
	_, ok = whClient.httpClient.Logger.(retryablehttp.LeveledLogger)
	assert.True(t, ok)
}

func TestWebhookClient(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		b, err := io.ReadAll(r.Body)
		assert.NoError(t, err)

		assert.JSONEq(t, `{"version":"","type":"flag","action":"create","metadata":{},"payload":null,"timestamp":"","status":""}`, string(b))
		assert.Equal(t, "6bc16d7434cbd746abfd25e3350a3f43fba04883b7f0b83f09bbc373116f84b0", r.Header.Get(fliptSignatureHeader))
		w.WriteHeader(200)
	}))

	t.Cleanup(ts.Close)

	client := NewWebhookClient(zap.NewNop(), ts.URL, "s3crET", time.Second)

	resp, err := client.SendAudit(context.TODO(), audit.Event{
		Type:   string(flipt.SubjectFlag),
		Action: string(flipt.ActionCreate),
	})

	require.NoError(t, err)
	resp.Body.Close()
}
