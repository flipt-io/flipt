package template

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"text/template"
	"time"

	"github.com/hashicorp/go-retryablehttp"
	"go.flipt.io/flipt/internal/server/audit"
	"go.uber.org/zap"
)

var _ Executer = (*webhookTemplate)(nil)

var funcMap = template.FuncMap{
	"toJson": func(v interface{}) (string, error) {
		jsonData, err := json.Marshal(v)
		if err != nil {
			return "", err
		}
		return string(jsonData), nil
	},
}

// webhookTemplate contains fields that pertain to constructing an HTTP request.
type webhookTemplate struct {
	url string

	headers      map[string]string
	bodyTemplate *template.Template

	httpClient *retryablehttp.Client
}

// Executer will try and send the event to the configured URL with all of the parameters.
// It will send the request with the retyrable client.
type Executer interface {
	Execute(context.Context, audit.Event) error
}

// NewWebhookTemplate is the constructor for a WebhookTemplate.
func NewWebhookTemplate(logger *zap.Logger, url, body string, headers map[string]string, maxBackoffDuration time.Duration) (Executer, error) {
	tmpl, err := template.New("").Funcs(funcMap).Parse(body)
	if err != nil {
		return nil, err
	}

	httpClient := retryablehttp.NewClient()
	httpClient.Logger = NewLeveledLogger(logger)
	httpClient.RetryWaitMax = maxBackoffDuration

	return &webhookTemplate{
		url:          url,
		bodyTemplate: tmpl,
		headers:      headers,
		httpClient:   httpClient,
	}, nil
}

// Execute will take in an HTTP client, and the event to send upstream to a destination. It will house the logic
// of doing the exponential backoff for sending the event upon failure to do so.
func (w *webhookTemplate) Execute(ctx context.Context, event audit.Event) error {
	buf := &bytes.Buffer{}

	err := w.bodyTemplate.Execute(buf, event)
	if err != nil {
		return err
	}

	validJSON := json.Valid(buf.Bytes())

	if !validJSON {
		return fmt.Errorf("invalid JSON: %s", buf.String())
	}

	req, err := retryablehttp.NewRequestWithContext(ctx, http.MethodPost, w.url, buf.Bytes())
	if err != nil {
		return err
	}

	for key, value := range w.headers {
		req.Header.Add(key, value)
	}

	// Add the application/json header if it does not exist in the HTTP headers already.
	value := req.Header.Get("Content-Type")
	if !strings.HasPrefix(value, "application/json") {
		req.Header.Add("Content-Type", "application/json")
	}

	resp, err := w.httpClient.Do(req)
	if err != nil {
		return err
	}

	if resp != nil {
		resp.Body.Close()
	}

	return nil
}
