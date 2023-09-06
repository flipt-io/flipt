package webhook

import (
	"github.com/hashicorp/go-multierror"
	"go.flipt.io/flipt/internal/server/audit"
	"go.uber.org/zap"
)

const sinkType = "webhook"

// Client is the client-side contract for sending an audit to a configured sink.
type Client interface {
	SendAudit(e *audit.Event) error
}

// Sink is a structure in charge of sending Audits to a configured webhook at a URL.
type Sink struct {
	logger        *zap.Logger
	webhookClient Client
}

// NewSink is the constructor for a Sink.
func NewSink(logger *zap.Logger, webhookClient Client) *Sink {
	return &Sink{
		logger:        logger,
		webhookClient: webhookClient,
	}
}

func (w *Sink) SendAudits(events []audit.Event) error {
	var result error

	for _, e := range events {
		err := w.webhookClient.SendAudit(&e)
		if err != nil {
			w.logger.Error("failed to send audit to webhook", zap.Error(err))
			result = multierror.Append(result, err)
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
