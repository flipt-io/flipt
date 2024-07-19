package config

import (
	"errors"
	"time"

	"github.com/spf13/viper"
)

var (
	_ defaulter = (*AuditConfig)(nil)
	_ validator = (*AuditConfig)(nil)
)

// AuditConfig contains fields, which enable and configure
// Flipt's various audit sink mechanisms.
type AuditConfig struct {
	Sinks  SinksConfig  `json:"sinks,omitempty" mapstructure:"sinks" yaml:"sinks,omitempty"`
	Buffer BufferConfig `json:"buffer,omitempty" mapstructure:"buffer" yaml:"buffer,omitempty"`
	Events []string     `json:"events,omitempty" mapstructure:"events" yaml:"events,omitempty"`
}

// Enabled returns true if any nested sink is enabled
func (c AuditConfig) Enabled() bool {
	return c.Sinks.Log.Enabled || c.Sinks.Webhook.Enabled || c.Sinks.Cloud.Enabled || c.Sinks.Kafka.Enabled
}

func (c AuditConfig) IsZero() bool {
	return !c.Enabled()
}

func (c *AuditConfig) setDefaults(v *viper.Viper) error {
	v.SetDefault("audit", map[string]any{
		"sinks": map[string]any{
			"log": map[string]any{
				"enabled": "false",
			},
			"webhook": map[string]any{
				"enabled": "false",
			},
			"cloud": map[string]any{
				"enabled": "false",
			},
			"kafka": map[string]any{
				"enabled":  "false",
				"encoding": "protobuf",
			},
		},
		"buffer": map[string]any{
			"capacity":     2,
			"flush_period": "2m",
		},
		"events": []string{"*:*"},
	})

	return nil
}

func (c *AuditConfig) validate() error {
	if c.Sinks.Webhook.Enabled {
		if err := c.Sinks.Webhook.validate(); err != nil {
			return err
		}
	}

	if c.Sinks.Kafka.Enabled {
		if err := c.Sinks.Kafka.validate(); err != nil {
			return err
		}
	}

	if c.Buffer.Capacity < 2 || c.Buffer.Capacity > 10 {
		return errors.New("buffer capacity below 2 or above 10")
	}

	if c.Buffer.FlushPeriod < 2*time.Minute || c.Buffer.FlushPeriod > 5*time.Minute {
		return errors.New("flush period below 2 minutes or greater than 5 minutes")
	}

	return nil
}

// SinksConfig contains configuration held in structures for the different sinks
// that we will send audits to.
type SinksConfig struct {
	Log     LogSinkConfig     `json:"log,omitempty" mapstructure:"log" yaml:"log,omitempty"`
	Webhook WebhookSinkConfig `json:"webhook,omitempty" mapstructure:"webhook" yaml:"webhook,omitempty"`
	Cloud   CloudSinkConfig   `json:"cloud,omitempty" mapstructure:"cloud" yaml:"cloud,omitempty"`
	Kafka   KafkaSinkConfig   `json:"kafka,omitempty" mapstructure:"kafka" yaml:"kafka,omitempty"`
}

type CloudSinkConfig struct {
	Enabled bool `json:"enabled,omitempty" mapstructure:"enabled" yaml:"enabled,omitempty"`
}

// WebhookSinkConfig contains configuration for sending POST requests to specific
// URL as its configured.
type WebhookSinkConfig struct {
	Enabled            bool              `json:"enabled" mapstructure:"enabled" yaml:"enabled"`
	URL                string            `json:"url,omitempty" mapstructure:"url" yaml:"url,omitempty"`
	MaxBackoffDuration time.Duration     `json:"maxBackoffDuration,omitempty" mapstructure:"max_backoff_duration" yaml:"max_backoff_duration,omitempty"`
	SigningSecret      string            `json:"-" mapstructure:"signing_secret" yaml:"-"`
	Templates          []WebhookTemplate `json:"templates,omitempty" mapstructure:"templates" yaml:"templates,omitempty"`
}

func (w WebhookSinkConfig) validate() error {
	if w.URL == "" && len(w.Templates) == 0 {
		return errors.New("url or template(s) not provided")
	}

	if w.URL != "" && len(w.Templates) > 0 {
		return errors.New("only one of url or template(s) allowed")
	}

	return nil
}

// LogSinkConfig contains fields that hold configuration for sending audits
// to a log.
type LogSinkConfig struct {
	Enabled  bool        `json:"enabled,omitempty" mapstructure:"enabled" yaml:"enabled,omitempty"`
	File     string      `json:"file,omitempty" mapstructure:"file" yaml:"file,omitempty"`
	Encoding LogEncoding `json:"encoding,omitempty" mapstructure:"encoding" yaml:"encoding,omitempty"`
}

// KafkaSinkConfig contains fields that hold configuration for sending audits
// to Kafka.
type KafkaSinkConfig struct {
	Enabled          bool                       `json:"enabled,omitempty" mapstructure:"enabled" yaml:"enabled,omitempty"`
	Topic            string                     `json:"topic,omitempty" mapstructure:"topic" yaml:"topic,omitempty"`
	BootstrapServers []string                   `json:"bootstrapServers,omitempty" mapstructure:"bootstrap_servers" yaml:"bootstrap_servers,omitempty"`
	Encoding         string                     `json:"encoding,omitempty" mapstructure:"encoding" yaml:"encoding,omitempty"`
	Authentication   *KafkaAuthenticationConfig `json:"-" mapstructure:"authentication" yaml:"-"`
	SchemaRegistry   *KafkaSchemaRegistryConfig `json:"-" mapstructure:"schema_registry" yaml:"-"`
	RequireTLS       bool                       `json:"requireTLS,omitempty" mapstructure:"require_tls" yaml:"require_tls,omitempty"`
	InsecureSkipTLS  bool                       `json:"-" mapstructure:"insecure_skip_tls" yaml:"-"`
}

func (c *KafkaSinkConfig) validate() error {
	if c.Topic == "" {
		return errors.New("kafka topic not provided")
	}

	if len(c.BootstrapServers) == 0 {
		return errors.New("kafka bootstrap_servers not provided")
	}

	return nil
}

type KafkaSchemaRegistryConfig struct {
	URL string `json:"-" mapstructure:"url" yaml:"-"`
}

// KafkaAuthenticationConfig contains fields that hold auth configuration for Kafka.
type KafkaAuthenticationConfig struct {
	Username string `json:"-" mapstructure:"username" yaml:"-"`
	Password string `json:"-" mapstructure:"password" yaml:"-"`
}

// BufferConfig holds configuration for the buffering of sending the audit
// events to the sinks.
type BufferConfig struct {
	Capacity    int           `json:"capacity,omitempty" mapstructure:"capacity" yaml:"capacity,omitempty"`
	FlushPeriod time.Duration `json:"flushPeriod,omitempty" mapstructure:"flush_period" yaml:"flush_period,omitempty"`
}

// WebhookTemplate specifies configuration for a user to send a payload a particular destination URL.
type WebhookTemplate struct {
	URL     string            `json:"url,omitempty" mapstructure:"url"`
	Body    string            `json:"body,omitempty" mapstructure:"body"`
	Headers map[string]string `json:"headers,omitempty" mapstructure:"headers"`
}
