package template

import (
	"context"
	"fmt"
	"time"

	"github.com/hashicorp/go-multierror"
	"go.flipt.io/flipt/internal/config"
	"go.flipt.io/flipt/internal/server/audit"
	"go.uber.org/zap"
)

const sinkType = "templates"

type Sink struct {
	logger *zap.Logger

	executers []Executer
}

// NewSink is the constructor for a Sink.
func NewSink(logger *zap.Logger, webhookTemplates []config.WebhookTemplate, maxBackoffDuration time.Duration) (audit.Sink, error) {
	executers := make([]Executer, 0, len(webhookTemplates))

	for _, wht := range webhookTemplates {
		executer, err := NewWebhookTemplate(logger, wht.URL, wht.Body, wht.Headers, maxBackoffDuration)
		if err != nil {
			return nil, fmt.Errorf("failed to create webhook template sink: %w", err)
		}

		executers = append(executers, executer)
	}

	return &Sink{
		logger:    logger.With(zap.String("sink", sinkType)),
		executers: executers,
	}, nil
}

func (t *Sink) SendAudits(ctx context.Context, events []audit.Event) error {
	var result error

	for _, e := range events {
		for _, executer := range t.executers {
			err := executer.Execute(ctx, e)
			if err != nil {
				t.logger.Error("failed to send audit to webhook", zap.Error(err))
				result = multierror.Append(result, err)
			}
		}
	}

	return result
}

func (t *Sink) Close() error {
	return nil
}

func (t *Sink) String() string {
	return sinkType
}
