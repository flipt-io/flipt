package config

import "github.com/spf13/viper"

var (
	_ defaulter = (*EvaluationConfig)(nil)
)

// EvaluationConfig contains fields which configure flag evaluation behavior.
type EvaluationConfig struct {
	IncludeFlagMetadata bool `json:"includeFlagMetadata,omitempty" mapstructure:"include_flag_metadata" yaml:"include_flag_metadata,omitempty"`
}

func (c *EvaluationConfig) setDefaults(v *viper.Viper) error {
	v.SetDefault("evaluation.include_flag_metadata", false)
	return nil
}
