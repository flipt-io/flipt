package clickhouse

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestGetStepFromDuration(t *testing.T) {
	cases := []struct {
		name     string
		duration time.Duration
		want     *Step
	}{
		{
			name:     "1 hour duration",
			duration: time.Hour,
			want: &Step{
				intervalValue: 1,
				intervalStep:  "MINUTE",
			},
		},
		{
			name:     "3 hour duration",
			duration: 3 * time.Hour,
			want: &Step{
				intervalValue: 1,
				intervalStep:  "MINUTE",
			},
		},
		{
			name:     "24 hour duration",
			duration: 24 * time.Hour,
			want: &Step{
				intervalValue: 15,
				intervalStep:  "MINUTE",
			},
		},
	}

	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			step := getStepFromDuration(tt.duration)
			assert.Equal(t, step.intervalStep, tt.want.intervalStep)
			assert.Equal(t, step.intervalValue, tt.want.intervalValue)
		})
	}
}
