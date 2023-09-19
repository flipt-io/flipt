package template

import (
	"bytes"
	"context"
	"net/http"
	"strings"
	"text/template"
	"time"

	"go.flipt.io/flipt/internal/server/audit"
	"go.uber.org/zap"
)

// webhookTemplate contains fields that pertain to constructing an HTTP request.
type webhookTemplate struct {
	method string
	url    string

	headers      map[string]string
	bodyTemplate *template.Template

	maxBackoffDuration time.Duration

	retryableClient audit.Retrier
}

// Executer will try and send the event to the configured URL with all of the parameters.
// It will send the request with the retyrable client.
type Executer interface {
	Execute(context.Context, audit.Event) error
}

// NewWebhookTemplate is the constructor for a WebhookTemplate.
func NewWebhookTemplate(logger *zap.Logger, method, url, body string, headers map[string]string, maxBackoffDuration time.Duration) (Executer, error) {
	tmpl, err := template.New("").Parse(body)
	if err != nil {
		return nil, err
	}

	return &webhookTemplate{
		url:                url,
		method:             method,
		bodyTemplate:       tmpl,
		headers:            headers,
		maxBackoffDuration: maxBackoffDuration,
		retryableClient:    audit.NewRetrier(logger, maxBackoffDuration),
	}, nil
}

func (w *webhookTemplate) createRequest(ctx context.Context, body []byte) (*http.Request, error) {
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
func (w *webhookTemplate) Execute(ctx context.Context, event audit.Event) error {
	buf := &bytes.Buffer{}

	err := w.bodyTemplate.Execute(buf, event)
	if err != nil {
		return err
	}

	err = w.retryableClient.RequestRetry(ctx, buf.Bytes(), w.createRequest)
	if err != nil {
		return err
	}

	return nil
}
