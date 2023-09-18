package template

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"
	"text/template"
	"time"

	"github.com/cenkalti/backoff/v4"
	"go.flipt.io/flipt/internal/server/audit"
	"go.uber.org/zap"
)

// WebhookTemplate contains fields that pertain to constructing an HTTP request.
type WebhookTemplate struct {
	logger *zap.Logger

	url    string
	method string

	headers      map[string]string
	bodyTemplate *template.Template

	maxBackoffDuration time.Duration
}

// Executer will try and send the event to the configured URL with all of the parameters.
// It will send the request with retries, according to the backoff strategy.
type Executer interface {
	Execute(context.Context, *http.Client, audit.Event) error
}

// NewWebhookTemplate is the constructor for a WebhookTemplate.
func NewWebhookTemplate(logger *zap.Logger, method, url, body string, headers map[string]string, maxBackoffDuration time.Duration) (Executer, error) {
	tmpl, err := template.New("").Parse(body)
	if err != nil {
		return nil, err
	}

	return &WebhookTemplate{
		logger:             logger,
		url:                url,
		method:             method,
		bodyTemplate:       tmpl,
		headers:            headers,
		maxBackoffDuration: maxBackoffDuration,
	}, nil
}

func (w *WebhookTemplate) createRequest(ctx context.Context, body []byte) (*http.Request, error) {
	req, err := http.NewRequestWithContext(ctx, w.method, w.url, bytes.NewBuffer(body))
	if err != nil {
		return nil, err
	}

	for key, value := range w.headers {
		req.Header.Add(key, value)
	}

	// Add the application/json header if it does not exist in the HTTP headers already.
	value := req.Header.Get("Content-Type")
	if !strings.HasPrefix(value, "application/json") {
		req.Header.Add("Content-Type", "application/json")
	}

	return req, nil
}

// Execute will take in an HTTP client, and the event to send upstream to a destination. It will house the logic
// of doing the exponential backoff for sending the event upon failure to do so.
func (w *WebhookTemplate) Execute(ctx context.Context, client *http.Client, event audit.Event) error {
	buf := &bytes.Buffer{}

	err := w.bodyTemplate.Execute(buf, event)
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
		req, err := w.createRequest(ctx, buf.Bytes())
		if err != nil {
			w.logger.Error("error creating HTTP request", zap.Error(err))
			return err
		}

		resp, err := client.Do(req)
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
		return fmt.Errorf("failed to send event to webhook url: %s", w.url)
	}

	return nil
}
