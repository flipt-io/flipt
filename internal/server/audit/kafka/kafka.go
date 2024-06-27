package kafka

import (
	"context"
	"crypto/tls"
	"fmt"

	"github.com/twmb/franz-go/pkg/kgo"
	"github.com/twmb/franz-go/pkg/sasl/plain"
	"github.com/twmb/franz-go/pkg/sr"
	"go.flipt.io/flipt/internal/config"
	"go.flipt.io/flipt/internal/server/audit"
	"go.uber.org/zap"
)

const (
	sinkType      = "kafka"
	encodingProto = "protobuf"
	encodingAvro  = "avro"
)

// encodingFn represents func which encoding data.
type encodingFn func(v any) ([]byte, error)

// Encoder represents an interface for encoding data and provide the schema.
type Encoder interface {
	Schema() sr.Schema
	Encode(v any) ([]byte, error)
}

type Sink struct {
	logger *zap.Logger
	client *kgo.Client
	encode encodingFn
}

// NewSink is the constructor for a Sink.
func NewSink(logger *zap.Logger, cfg config.KafkaSinkConfig) (audit.Sink, error) {
	logger = logger.With(zap.String("sink", sinkType))
	opts := []kgo.Opt{
		kgo.SeedBrokers(cfg.BootstrapServers...),
		kgo.DefaultProduceTopic(cfg.Topic),
		kgo.WithLogger(&klogger{logger: logger}),
	}

	if cfg.Authentication != nil {
		auth := plain.Auth{User: cfg.Authentication.Username, Pass: cfg.Authentication.Password}
		tlsDialer := &tls.Dialer{}
		opts = append(opts,
			kgo.SASL(auth.AsMechanism()),
			kgo.Dialer(tlsDialer.DialContext),
		)
	}

	client, err := kgo.NewClient(opts...)
	if err != nil {
		return nil, err
	}
	
	if err := client.Ping(context.Background()); err != nil {
		return nil, err
	}
	encoder, err := newEncoder(cfg)
	if err != nil {
		return nil, err
	}
	encodeFn := encoder.Encode
	if cfg.SchemaRegistry != "" {
		encodeFn, err = setupSchemaRegistryEncoder(cfg, encoder)
		if err != nil {
			return nil, err
		}
	}

	return &Sink{
		client: client,
		logger: logger,
		encode: encodeFn,
	}, nil
}

// setupSchemaRegistryEncoder sets up the schema registry client, registers the schema,
// and returns an encoding function.
func setupSchemaRegistryEncoder(cfg config.KafkaSinkConfig, encoder Encoder) (encodingFn, error) {
	rcl, err := sr.NewClient(sr.URLs(cfg.SchemaRegistry))
	if err != nil {
		return nil, err
	}
	ss, err := rcl.CreateSchema(context.Background(), cfg.Topic+"-value", encoder.Schema())
	if err != nil {
		return nil, err
	}
	var serde sr.Serde
	serde.Register(
		ss.ID,
		audit.Event{},
		sr.EncodeFn(encoder.Encode),
	)
	return serde.Encode, nil
}

// newEncoder creates a new Encoder based on the provided KafkaSinkConfig.
// It returns an error if the encoding type specified in the configuration is unsupported.
func newEncoder(cfg config.KafkaSinkConfig) (Encoder, error) {
	switch cfg.Encoding {
	case encodingProto:
		return newProtobufEncoder(), nil
	case encodingAvro:
		return newAvroEncoder(), nil
	default:
		return nil, fmt.Errorf("unsupported encoding: %s", cfg.Encoding)
	}
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
		value, err := s.encode(e)
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
