package ext

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
)

func TestDistribution_ZeroRollout_YAML(t *testing.T) {
	tests := []struct {
		name     string
		rollout  float32
		expected string
	}{
		{
			name:     "zero rollout",
			rollout:  0,
			expected: "variant: test\nrollout: 0\n",
		},
		{
			name:     "zero float rollout",
			rollout:  0.0,
			expected: "variant: test\nrollout: 0\n",
		},
		{
			name:     "non-zero rollout",
			rollout:  50.0,
			expected: "variant: test\nrollout: 50\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dist := &Distribution{
				VariantKey: "test",
				Rollout:    tt.rollout,
			}

			data, err := yaml.Marshal(dist)
			require.NoError(t, err)
			assert.Equal(t, tt.expected, string(data))

			// Test unmarshaling back
			var unmarshaled Distribution
			err = yaml.Unmarshal(data, &unmarshaled)
			require.NoError(t, err)
			assert.Equal(t, dist.VariantKey, unmarshaled.VariantKey)
			assert.InDelta(t, dist.Rollout, unmarshaled.Rollout, 0.001)
		})
	}
}

func TestDistribution_ZeroRollout_JSON(t *testing.T) {
	tests := []struct {
		name     string
		rollout  float32
		expected string
	}{
		{
			name:     "zero rollout",
			rollout:  0,
			expected: `{"variant":"test","rollout":0}`,
		},
		{
			name:     "zero float rollout",
			rollout:  0.0,
			expected: `{"variant":"test","rollout":0}`,
		},
		{
			name:     "non-zero rollout",
			rollout:  50.0,
			expected: `{"variant":"test","rollout":50}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dist := &Distribution{
				VariantKey: "test",
				Rollout:    tt.rollout,
			}

			data, err := json.Marshal(dist)
			require.NoError(t, err)
			assert.Equal(t, tt.expected, string(data))

			// Test unmarshaling back
			var unmarshaled Distribution
			err = json.Unmarshal(data, &unmarshaled)
			require.NoError(t, err)
			assert.Equal(t, dist.VariantKey, unmarshaled.VariantKey)
			assert.InDelta(t, dist.Rollout, unmarshaled.Rollout, 0.001)
		})
	}
}

func TestThresholdRule_ZeroPercentage_YAML(t *testing.T) {
	tests := []struct {
		name       string
		percentage float32
		value      bool
		expected   string
	}{
		{
			name:       "zero percentage with true value",
			percentage: 0,
			value:      true,
			expected:   "percentage: 0\nvalue: true\n",
		},
		{
			name:       "zero float percentage with false value",
			percentage: 0.0,
			value:      false,
			expected:   "percentage: 0\n",
		},
		{
			name:       "non-zero percentage",
			percentage: 50.0,
			value:      true,
			expected:   "percentage: 50\nvalue: true\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			threshold := &ThresholdRule{
				Percentage: tt.percentage,
				Value:      tt.value,
			}

			data, err := yaml.Marshal(threshold)
			require.NoError(t, err)
			assert.Equal(t, tt.expected, string(data))

			// Test unmarshaling back
			var unmarshaled ThresholdRule
			err = yaml.Unmarshal(data, &unmarshaled)
			require.NoError(t, err)
			assert.InDelta(t, threshold.Percentage, unmarshaled.Percentage, 0.001)
			assert.Equal(t, threshold.Value, unmarshaled.Value)
		})
	}
}

func TestThresholdRule_ZeroPercentage_JSON(t *testing.T) {
	tests := []struct {
		name       string
		percentage float32
		value      bool
		expected   string
	}{
		{
			name:       "zero percentage with true value",
			percentage: 0,
			value:      true,
			expected:   `{"percentage":0,"value":true}`,
		},
		{
			name:       "zero float percentage with false value",
			percentage: 0.0,
			value:      false,
			expected:   `{"percentage":0}`,
		},
		{
			name:       "non-zero percentage",
			percentage: 50.0,
			value:      true,
			expected:   `{"percentage":50,"value":true}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			threshold := &ThresholdRule{
				Percentage: tt.percentage,
				Value:      tt.value,
			}

			data, err := json.Marshal(threshold)
			require.NoError(t, err)
			assert.Equal(t, tt.expected, string(data))

			// Test unmarshaling back
			var unmarshaled ThresholdRule
			err = json.Unmarshal(data, &unmarshaled)
			require.NoError(t, err)
			assert.InDelta(t, threshold.Percentage, unmarshaled.Percentage, 0.001)
			assert.Equal(t, threshold.Value, unmarshaled.Value)
		})
	}
}
