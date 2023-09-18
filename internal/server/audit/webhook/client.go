package webhook

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"io"
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

	maxBackoffDuration time.Duration
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
		logger:             logger,
		Client:             c,
		url:                url,
		signingSecret:      signingSecret,
		maxBackoffDuration: 15 * time.Second,
	}

	for _, o := range opts {
		o(h)
	}

	return h
}

type ClientOption func(h *HTTPClient)

// WithMaxBackoffDuration allows for a configurable backoff duration configuration.
func WithMaxBackoffDuration(maxBackoffDuration time.Duration) ClientOption {
	return func(h *HTTPClient) {
		h.maxBackoffDuration = maxBackoffDuration
	}
}

func (w *HTTPClient) createRequest(ctx context.Context, body []byte) (*http.Request, error) {
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
func (w *HTTPClient) SendAudit(ctx context.Context, e audit.Event) error {
	body, err := json.Marshal(e)
	if err != nil {
		return err
	}

	be := backoff.NewExponentialBackOff()
	be.MaxElapsedTime = w.maxBackoffDuration

	ticker := backoff.NewTicker(be)
	defer ticker.Stop()

	successfulRequest := false

	// Make requests with configured retries.
	for range ticker.C {
		req, err := w.createRequest(ctx, body)
		if err != nil {
			w.logger.Error("error creating HTTP request", zap.Error(err))
			return err
		}

		resp, err := w.Client.Do(req)
		if err != nil {
			w.logger.Debug("webhook request failed, retrying...", zap.Error(err))
			continue
		}

		if resp.StatusCode != http.StatusOK {
			w.logger.Debug("webhook request failed, retrying...", zap.Int("status_code", resp.StatusCode))
			_, _ = io.Copy(io.Discard, resp.Body)
			_ = resp.Body.Close()
			continue
		}

		_, _ = io.Copy(io.Discard, resp.Body)
		_ = resp.Body.Close()

		w.logger.Debug("successful request to webhook")
		successfulRequest = true
		break
	}

	if !successfulRequest {
		return fmt.Errorf("failed to send event to webhook url: %s after %s", w.url, w.maxBackoffDuration)
	}

	return nil
}
