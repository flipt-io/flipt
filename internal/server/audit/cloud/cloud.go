package cloud

import (
	"context"
	"fmt"
	"time"

	"github.com/hashicorp/go-multierror"
	"go.flipt.io/flipt/internal/server/audit"
	"go.flipt.io/flipt/internal/server/audit/template"
	"go.uber.org/zap"
)

const sinkType = "cloud"

type Sink struct {
	logger *zap.Logger

	executer template.Executer
}

// NewSink is the constructor for a Sink.
func NewSink(logger *zap.Logger, apiKey string, url string) (audit.Sink, error) {
	const body = `{
		"type": "{{ .Type }}",
		"action": "{{ .Action }}",
		"actor": {{ toJson .Metadata.Actor }},
		"payload": {{ toJson .Payload }}
	}`

	headers := map[string]string{
		"Authorization": fmt.Sprintf("Bearer %s", apiKey),
		"Content-Type":  "application/json",
	}

	executer, err := template.NewWebhookTemplate(logger, url, body, headers, 15*time.Second)
	if err != nil {
		return nil, fmt.Errorf("failed to create webhook template sink: %w", err)
	}

	return &Sink{
		logger:   logger.With(zap.String("sink", sinkType)),
		executer: executer,
	}, nil
}

// Close implements audit.Sink.
func (s *Sink) Close() error {
	return nil
}

// SendAudits implements audit.Sink.
func (s *Sink) SendAudits(ctx context.Context, events []audit.Event) error {
	var result error

	for _, e := range events {
		err := s.executer.Execute(ctx, e)
		if err != nil {
			s.logger.Error("failed to send audit to webhook", zap.Error(err))
			result = multierror.Append(result, err)
		}
	}

	return result
}

// String implements audit.Sink.
func (s *Sink) String() string {
	return sinkType
}
