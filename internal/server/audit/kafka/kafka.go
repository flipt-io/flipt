package kafka

import (
	"context"

	"github.com/twmb/franz-go/pkg/kgo"
	"go.flipt.io/flipt/internal/server/audit"
	"go.uber.org/zap"
)

const sinkType = "kafka"

type Sink struct {
	logger   *zap.Logger
	client   *kgo.Client
	encodeFn func(v any) ([]byte, error)
}

// NewSink is the constructor for a Sink.
func NewSink(logger *zap.Logger, servers []string, topic string, encoding string) (audit.Sink, error) {
	logger = logger.With(zap.String("sink", sinkType))

	client, err := kgo.NewClient(
		kgo.SeedBrokers(servers...),
		kgo.DefaultProduceTopic(topic),
		kgo.WithLogger(&klogger{logger: logger}),
	)
	if err != nil {
		return nil, err
	}

	return &Sink{
		client:   client,
		logger:   logger,
		encodeFn: toProtobuf,
	}, nil
}

// Close implements audit.Sink.
func (s *Sink) Close() error {
	s.client.Close()
	return nil
}

// SendAudits implements audit.Sink.
func (s *Sink) SendAudits(ctx context.Context, events []audit.Event) error {
	for _, e := range events {
		prm := kgo.AbortingFirstErrPromise(s.client)
		value, err := s.encodeFn(e)
		if err != nil {
			s.logger.Error("failed to encode event", zap.Error(err))
			continue
		}
		s.client.Produce(ctx, &kgo.Record{
			Value: value,
			Key:   []byte(e.Timestamp), // FIXME: it's marked as optional
		}, prm.Promise())
		if err := prm.Err(); err != nil {
			s.logger.Error("failed to publish record", zap.Error(err))
		}
	}
	return nil
}

// String implements audit.Sink.
func (s *Sink) String() string {
	return sinkType
}
