package kafka

import (
	"context"
	"crypto/tls"
	"fmt"

	"github.com/twmb/franz-go/pkg/kgo"
	"github.com/twmb/franz-go/pkg/sasl/scram"
	"github.com/twmb/franz-go/pkg/sr"
	"github.com/twmb/franz-go/plugin/kzap"
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
func NewSink(ctx context.Context, logger *zap.Logger, cfg config.KafkaSinkConfig) (audit.Sink, error) {
	logger = logger.With(zap.String("sink", sinkType))
	logLevel := logger.Level()
	if logLevel == zap.DebugLevel {
		// kgo produces a lot of debug logs
		logLevel = zap.InfoLevel
	}
	opts := []kgo.Opt{
		kgo.SeedBrokers(cfg.BootstrapServers...),
		kgo.DefaultProduceTopic(cfg.Topic),
		kgo.WithLogger(kzap.New(logger, kzap.AtomicLevel(zap.NewAtomicLevelAt(logLevel)))),
	}

	if cfg.RequireTLS {
		opts = append(opts, kgo.DialTLSConfig(&tls.Config{
			MinVersion: tls.VersionTLS12,
			//nolint:gosec
			InsecureSkipVerify: cfg.InsecureSkipTLS,
		}))
	}

	if cfg.Authentication != nil {
		auth := scram.Auth{User: cfg.Authentication.Username, Pass: cfg.Authentication.Password}
		opts = append(opts, kgo.SASL(auth.AsSha256Mechanism()))
	}

	client, err := kgo.NewClient(opts...)
	if err != nil {
		return nil, err
	}

	if err := client.Ping(ctx); err != nil {
		return nil, err
	}

	encoder, err := newEncoder(cfg)
	if err != nil {
		return nil, err
	}
	encodeFn := encoder.Encode
	if cfg.SchemaRegistry != nil {
		encodeFn, err = setupSchemaRegistryEncoder(ctx, cfg, encoder)
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
func setupSchemaRegistryEncoder(ctx context.Context, cfg config.KafkaSinkConfig, encoder Encoder) (encodingFn, error) {
	rcl, err := sr.NewClient(sr.URLs(cfg.SchemaRegistry.URL))
	if err != nil {
		return nil, err
	}
	ss, err := rcl.CreateSchema(ctx, cfg.Topic+"-value", encoder.Schema())
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
	prm := kgo.AbortingFirstErrPromise(s.client)
	for _, e := range events {
		value, err := s.encode(e)
		if err != nil {
			s.logger.Error("failed to encode event", zap.Error(err))
			continue
		}
		s.client.Produce(ctx, &kgo.Record{
			Value: value,
			Key:   []byte(e.Type),
		}, prm.Promise())
	}
	s.client.Flush(ctx)
	if err := prm.Err(); err != nil {
		s.logger.Error("failed to publish record", zap.Error(err))
	}
	return nil
}

// String implements audit.Sink.
func (s *Sink) String() string {
	return sinkType
}
