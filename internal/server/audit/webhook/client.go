package webhook

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/hashicorp/go-retryablehttp"
	"go.flipt.io/flipt/internal/server/audit"
	"go.flipt.io/flipt/internal/server/audit/template"
	"go.uber.org/zap"
)

const fliptSignatureHeader = "x-flipt-webhook-signature"

var _ Client = (*webhookClient)(nil)

// webhookClient allows for sending the event payload to a configured webhook service.
type webhookClient struct {
	logger        *zap.Logger
	url           string
	signingSecret string

	httpClient *retryablehttp.Client
}

// Client is the client-side contract for sending an audit to a configured sink.
type Client interface {
	SendAudit(ctx context.Context, e audit.Event) (*http.Response, error)
}

// signPayload takes in a marshalled payload and returns the sha256 hash of it.
func (w *webhookClient) signPayload(payload []byte) []byte {
	h := hmac.New(sha256.New, []byte(w.signingSecret))
	h.Write(payload)
	return h.Sum(nil)
}

// NewHTTPClient is the constructor for a HTTPClient.
func NewWebhookClient(logger *zap.Logger, url, signingSecret string, maxBackoffDuration time.Duration) Client {
	httpClient := retryablehttp.NewClient()
	httpClient.Logger = template.NewLeveledLogger(logger)
	httpClient.RetryWaitMax = maxBackoffDuration
	return &webhookClient{
		logger:        logger,
		url:           url,
		signingSecret: signingSecret,
		httpClient:    httpClient,
	}
}

// SendAudit will send an audit event to a configured server at a URL.
func (w *webhookClient) SendAudit(ctx context.Context, e audit.Event) (*http.Response, error) {
	body, err := json.Marshal(e)
	if err != nil {
		return nil, err
	}

	req, err := retryablehttp.NewRequestWithContext(ctx, http.MethodPost, w.url, bytes.NewBuffer(body))
	if err != nil {
		return nil, err
	}

	req.Header.Add("Content-Type", "application/json")

	// If the signing secret is configured, add the specific header to it.
	if w.signingSecret != "" {
		signedPayload := w.signPayload(body)

		req.Header.Add(fliptSignatureHeader, fmt.Sprintf("%x", signedPayload))
	}
	resp, err := w.httpClient.Do(req)
	if err != nil {
		return nil, err
	}

	return resp, nil
}
