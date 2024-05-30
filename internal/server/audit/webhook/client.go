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

	"go.flipt.io/flipt/internal/server/audit"
	"go.uber.org/zap"
)

const fliptSignatureHeader = "x-flipt-webhook-signature"

var _ Client = (*webhookClient)(nil)

// webhookClient allows for sending the event payload to a configured webhook service.
type webhookClient struct {
	logger        *zap.Logger
	url           string
	signingSecret string

	maxBackoffDuration time.Duration

	retryableClient audit.Retrier
}

// Client is the client-side contract for sending an audit to a configured sink.
type Client interface {
	SendAudit(ctx context.Context, e audit.Event) error
}

// signPayload takes in a marshalled payload and returns the sha256 hash of it.
func (w *webhookClient) signPayload(payload []byte) []byte {
	h := hmac.New(sha256.New, []byte(w.signingSecret))
	h.Write(payload)
	return h.Sum(nil)
}

// NewHTTPClient is the constructor for a HTTPClient.
func NewWebhookClient(logger *zap.Logger, url, signingSecret string, opts ...ClientOption) Client {
	h := &webhookClient{
		logger:             logger,
		url:                url,
		signingSecret:      signingSecret,
		maxBackoffDuration: 15 * time.Second,
	}

	for _, o := range opts {
		o(h)
	}

	retyableClient := audit.NewRetrier(logger, h.maxBackoffDuration)

	h.retryableClient = retyableClient

	return h
}

type ClientOption func(h *webhookClient)

// WithMaxBackoffDuration allows for a configurable backoff duration configuration.
func WithMaxBackoffDuration(maxBackoffDuration time.Duration) ClientOption {
	return func(h *webhookClient) {
		h.maxBackoffDuration = maxBackoffDuration
	}
}

func (w *webhookClient) createRequest(ctx context.Context, body []byte) (*http.Request, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, w.url, bytes.NewBuffer(body))
	if err != nil {
		return nil, err
	}

	req.Header.Add("Content-Type", "application/json")

	// If the signing secret is configured, add the specific header to it.
	if w.signingSecret != "" {
		signedPayload := w.signPayload(body)

		req.Header.Add(fliptSignatureHeader, fmt.Sprintf("%x", signedPayload))
	}

	return req, nil
}

// SendAudit will send an audit event to a configured server at a URL.
func (w *webhookClient) SendAudit(ctx context.Context, e audit.Event) error {
	body, err := json.Marshal(e)
	if err != nil {
		return err
	}

	err = w.retryableClient.RequestRetry(ctx, body, w.createRequest)
	if err != nil {
		return err
	}

	return nil
}
