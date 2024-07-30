package kafka

import (
	"bytes"
	"context"
	"crypto/tls"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/twmb/franz-go/pkg/kadm"
	"github.com/twmb/franz-go/pkg/kgo"
	"github.com/twmb/franz-go/pkg/sasl/scram"
	"go.flipt.io/flipt/internal/config"
	"go.flipt.io/flipt/internal/server/audit"
	"go.flipt.io/flipt/rpc/flipt"
	"go.uber.org/zap/zaptest"
)

func adminCreateUser(t testing.TB, base string) {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	t.Cleanup(cancel)
	endpoint, err := url.JoinPath(base, "/v1/security/users")
	require.NoError(t, err)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint,
		bytes.NewBuffer([]byte(`{"username": "admin", "password":"password", "algorithm": "SCRAM-SHA-256"}`)),
	)
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")

	res, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, res.StatusCode, "unexpected status code in response")
	res.Body.Close()
}

func adminCreateTopic(t testing.TB, bootstrapServers []string, topic string) {
	t.Helper()
	tlsCfg := &tls.Config{
		MinVersion: tls.VersionTLS12,
		//nolint:gosec
		InsecureSkipVerify: true,
	}

	client, err := kgo.NewClient(
		kgo.SeedBrokers(bootstrapServers...),
		kgo.DialTLSConfig(tlsCfg),
		kgo.SASL(scram.Auth{User: "admin", Pass: "password"}.AsSha256Mechanism()),
	)
	require.NoError(t, err)

	adminClient := kadm.NewClient(client)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	_, err = adminClient.CreateTopic(ctx, 1, 1, nil, topic)
	require.NoError(t, err)
	t.Cleanup(func() {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		_, err := adminClient.DeleteTopic(ctx, topic)
		assert.NoError(t, err)
		adminClient.Close()
	})
}

func TestNewSinkAndSend(t *testing.T) {
	srv := os.Getenv("KAFKA_BOOTSTRAP_SERVER")
	if srv == "" {
		t.Skip("no kafka servers provided")
	}
	adminCreateUser(t, fmt.Sprintf("http://%s:9644", srv))
	bootstrapServers := []string{srv}
	defaultAuth := &config.KafkaAuthenticationConfig{Username: "admin", Password: "password"}
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

	sg := fmt.Sprintf("http://%s:8081", srv)

	tests := []struct {
		name           string
		enc            string
		schemaRegistry string
	}{
		{"avro-basic", encodingAvro, ""},
		{"avro-schema-registry", encodingAvro, sg},
		{"proto-basic", encodingProto, ""},
		{"proto-schema-registry", encodingProto, sg},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			topic := fmt.Sprintf("default-%s", tt.name)
			adminCreateTopic(t, bootstrapServers, topic)
			ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
			t.Cleanup(cancel)
			cfg := config.KafkaSinkConfig{
				BootstrapServers: bootstrapServers,
				Topic:            topic,
				Encoding:         tt.enc,
				RequireTLS:       true,
				//nolint:gosec
				InsecureSkipTLS: true,
				Authentication:  defaultAuth,
			}

			if tt.schemaRegistry != "" {
				cfg.SchemaRegistry = &config.KafkaSchemaRegistryConfig{URL: tt.schemaRegistry}
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

	failTests := []struct {
		name     string
		cfg      config.KafkaSinkConfig
		expected string
	}{
		{
			"unsupported",
			config.KafkaSinkConfig{
				BootstrapServers: bootstrapServers,
				Topic:            "unsupported",
				Encoding:         "unknown",
				RequireTLS:       true,
				InsecureSkipTLS:  true,
				Authentication:   defaultAuth,
			},
			"unsupported encoding:",
		},
		{
			"auth-failure",
			config.KafkaSinkConfig{
				BootstrapServers: bootstrapServers,
				Topic:            "default-auth-failure",
				Encoding:         encodingProto,
				InsecureSkipTLS:  true,
				RequireTLS:       true,
				Authentication:   &config.KafkaAuthenticationConfig{Username: "user", Password: ""},
			},
			"SCRAM-SHA-256 user and pass must be non-empty",
		},
		{
			"tls-failure",
			config.KafkaSinkConfig{
				BootstrapServers: bootstrapServers,
				Topic:            "default-tls-failure",
				Encoding:         encodingProto,
			},
			"is this a plaintext connection speaking to a tls endpoint?",
		},
	}
	for _, tt := range failTests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := NewSink(context.Background(), zaptest.NewLogger(t), tt.cfg)
			require.ErrorContains(t, err, tt.expected)
		})
	}
}
