package webhook

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/cenkalti/backoff/v4"
	"go.flipt.io/flipt/internal/server/audit"
	"go.uber.org/zap"
)

const (
	fliptSignatureHeader = "x-flipt-webhook-signature"
)

// HTTPClient allows for sending the event payload to a configured webhook service.
type HTTPClient struct {
	logger *zap.Logger
	*http.Client
	url           string
	signingSecret string

	backoffDuration time.Duration
}

// signPayload takes in a marshalled payload and returns the sha256 hash of it.
func (w *HTTPClient) signPayload(payload []byte) []byte {
	h := hmac.New(sha256.New, []byte(w.signingSecret))
	h.Write(payload)
	return h.Sum(nil)
}

// NewHTTPClient is the constructor for a HTTPClient.
func NewHTTPClient(logger *zap.Logger, url, signingSecret string, opts ...ClientOption) *HTTPClient {
	c := &http.Client{
		Timeout: 5 * time.Second,
	}

	h := &HTTPClient{
		logger:          logger,
		Client:          c,
		url:             url,
		signingSecret:   signingSecret,
		backoffDuration: 15 * time.Second,
	}

	for _, o := range opts {
		o(h)
	}

	return h
}

type ClientOption func(h *HTTPClient)

// WithBackoffDuration allows for a configurable backoff duration configuration.
func WithBackoffDuration(backoffDuration time.Duration) ClientOption {
	return func(h *HTTPClient) {
		h.backoffDuration = backoffDuration
	}
}

// SendAudit will send an audit event to a configured server at a URL.
func (w *HTTPClient) SendAudit(e *audit.Event) error {
	body, err := json.Marshal(e)
	if err != nil {
		return err
	}

	req, err := http.NewRequest(http.MethodPost, w.url, bytes.NewBuffer(body))
	if err != nil {
		return err
	}

	// If the signing secret is configured, add the specific header to it.
	if w.signingSecret != "" {
		signedPayload := w.signPayload(body)

		req.Header.Add(fliptSignatureHeader, fmt.Sprintf("%x", signedPayload))
	}

	be := backoff.NewExponentialBackOff()
	be.MaxElapsedTime = w.backoffDuration

	ticker := backoff.NewTicker(be)
	defer ticker.Stop()

	// Make requests with configured retries.
	for range ticker.C {
		resp, err := w.Client.Do(req)
		if err != nil {
			w.logger.Debug("webhook request failed, retrying...", zap.Error(err))
			continue
		}
		if resp.StatusCode != http.StatusOK {
			w.logger.Debug("webhook request failed, retrying...", zap.Int("status_code", resp.StatusCode), zap.Error(err))
			continue
		}

		w.logger.Debug("successful request to webhook")
		break
	}

	return nil
}
