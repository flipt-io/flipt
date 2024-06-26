package kafka

import (
	"context"
	"fmt"

	"github.com/twmb/franz-go/pkg/kgo"
	"go.flipt.io/flipt/internal/server/audit"
	"go.uber.org/zap"
)

const (
	sinkType      = "kafka"
	encodingProto = "protobuf"
	encodingAvro  = "avro"
)

type encodingFn func(v any) ([]byte, error)

type Sink struct {
	logger  *zap.Logger
	client  *kgo.Client
	encoder encodingFn
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
	var encoder encodingFn
	switch encoding {
	case encodingProto:
		encoder = toProtobuf()
	case encodingAvro:
		encoder = toAvro()
	default:
		return nil, fmt.Errorf("unsupported encoding: %s", encoding)
	}

	return &Sink{
		client:  client,
		logger:  logger,
		encoder: encoder,
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
		value, err := s.encoder(e)
		if err != nil {
			s.logger.Error("failed to encode event", zap.Error(err))
			continue
		}
		s.client.Produce(ctx, &kgo.Record{
			Value: value,
			Key:   []byte(e.Type),
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
