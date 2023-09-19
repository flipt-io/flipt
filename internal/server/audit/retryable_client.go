package audit

import (
	"context"
	"errors"
	"io"
	"net/http"
	"time"

	"github.com/cenkalti/backoff/v4"
	"go.uber.org/zap"
)

type retryClient struct {
	logger *zap.Logger
	*http.Client
	maxBackoffDuration time.Duration
}

// Retrier contains a minimal API for retrying a request onto an endpoint upon failure or non-200 status.
type Retrier interface {
	RequestRetry(ctx context.Context, body []byte, fn RequestCreator) error
}

// NewRetrier is a constructor for a retryClient, but returns a contract for the consumer
// to just execute retryable HTTP requests.
func NewRetrier(logger *zap.Logger, maxBackoffDuration time.Duration) Retrier {
	return &retryClient{
		logger: logger,
		Client: &http.Client{
			Timeout: 5 * time.Second,
		},
		maxBackoffDuration: maxBackoffDuration,
	}
}

// RequestCreator provides a basic function for rewinding a request with a body, upon
// request failure.
type RequestCreator func(ctx context.Context, body []byte) (*http.Request, error)

func (r *retryClient) RequestRetry(ctx context.Context, body []byte, fn RequestCreator) error {
	be := backoff.NewExponentialBackOff()
	be.MaxElapsedTime = r.maxBackoffDuration

	ticker := backoff.NewTicker(be)
	defer ticker.Stop()

	successfulRequest := false

	// Make requests with configured retries.
	for range ticker.C {
		req, err := fn(ctx, body)
		if err != nil {
			r.logger.Error("error creating HTTP request", zap.Error(err))
			return err
		}

		resp, err := r.Client.Do(req)
		if err != nil {
			r.logger.Debug("webhook request failed, retrying...", zap.Error(err))
			continue
		}

		if resp.StatusCode != http.StatusOK {
			r.logger.Debug("webhook request failed, retrying...", zap.Int("status_code", resp.StatusCode))
			_, _ = io.Copy(io.Discard, resp.Body)
			_ = resp.Body.Close()
			continue
		}

		_, _ = io.Copy(io.Discard, resp.Body)
		_ = resp.Body.Close()

		r.logger.Debug("successful request to webhook")
		successfulRequest = true
		break
	}

	if !successfulRequest {
		return errors.New("failed to send event to webhook")
	}

	return nil
}
