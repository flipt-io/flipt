package webhook

import (
	"context"

	"github.com/hashicorp/go-multierror"
	"go.flipt.io/flipt/internal/server/audit"
	"go.uber.org/zap"
)

const sinkType = "webhook"

// Sink is a structure in charge of sending Audits to a configured webhook at a URL.
type Sink struct {
	logger        *zap.Logger
	webhookClient Client
}

// NewSink is the constructor for a Sink.
func NewSink(logger *zap.Logger, webhookClient Client) audit.Sink {
	return &Sink{
		logger:        logger.With(zap.String("sink", sinkType)),
		webhookClient: webhookClient,
	}
}

func (w *Sink) SendAudits(ctx context.Context, events []audit.Event) error {
	var result error

	for _, e := range events {
		resp, err := w.webhookClient.SendAudit(ctx, e)
		if err != nil {
			w.logger.Error("failed to send audit to webhook", zap.Error(err))
			result = multierror.Append(result, err)
		}

		if resp != nil {
			resp.Body.Close()
		}
	}

	return result
}

func (w *Sink) Close() error {
	return nil
}

func (w *Sink) String() string {
	return sinkType
}
