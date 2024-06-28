package kafka

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/twmb/franz-go/pkg/kadm"
	"github.com/twmb/franz-go/pkg/kgo"
	"go.flipt.io/flipt/internal/config"
	"go.flipt.io/flipt/internal/server/audit"
	"go.flipt.io/flipt/rpc/flipt"
	"go.uber.org/zap/zaptest"
)

func createTopic(t testing.TB, bootstrapServers []string, topic string) {
	t.Helper()
	client, err := kgo.NewClient(
		kgo.SeedBrokers(bootstrapServers...),
	)
	require.NoError(t, err)
	defer client.Close()

	adminClient := kadm.NewClient(client)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err = adminClient.CreateTopic(ctx, 1, 1, nil, topic)
	require.NoError(t, err)
}

func TestNewSink(t *testing.T) {
	srv := os.Getenv("KAFKA_BOOTSTRAP_SERVER")
	if srv == "" {
		t.Skip("no kafka servers provided")
	}

	bootstrapServers := []string{srv}
	topic := fmt.Sprintf("default-%d", time.Now().Unix())
	createTopic(t, bootstrapServers, topic)
	e := audit.NewEvent(
		flipt.NewRequest(flipt.ResourceFlag, flipt.ActionCreate, flipt.WithSubject(flipt.SubjectRule)),
		&audit.Actor{
			Authentication: "token",
			IP:             "127.0.0.1",
		},
		audit.NewFlag(&flipt.Flag{
			Key:          "this-flag",
			Name:         "this-flag",
			Description:  "this description",
			Enabled:      false,
			NamespaceKey: "default",
		}),
	)

	for _, enc := range []string{encodingAvro, encodingProto} {
		t.Run(enc, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
			t.Cleanup(cancel)
			cfg := config.KafkaSinkConfig{
				BootstrapServers: bootstrapServers,
				Topic:            topic,
				Encoding:         enc,
			}
			s, err := NewSink(context.Background(), zaptest.NewLogger(t), cfg)
			require.NoError(t, err)
			t.Cleanup(func() {
				err := s.Close()
				require.NoError(t, err)
			})
			require.Equal(t, sinkType, s.String())
			err = s.SendAudits(ctx, []audit.Event{*e})
			require.NoError(t, err)
			require.NoError(t, ctx.Err())
		})
	}
	t.Run("unsupported", func(t *testing.T) {
		cfg := config.KafkaSinkConfig{
			BootstrapServers: bootstrapServers,
			Topic:            topic,
			Encoding:         "unknown",
		}
		_, err := NewSink(context.Background(), zaptest.NewLogger(t), cfg)
		require.ErrorContains(t, err, "unsupported encoding:")
	})
}
