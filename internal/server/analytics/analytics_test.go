package analytics

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestGetStepFromDuration(t *testing.T) {
	cases := []struct {
		name     string
		duration time.Duration
		want     int
	}{
		{
			name:     "1 hour duration",
			duration: time.Hour,
			want:     1,
		},
		{
			name:     "4 hour duration",
			duration: 4 * time.Hour,
			want:     5,
		},
		{
			name:     "12 hour duration",
			duration: 12 * time.Hour,
			want:     15,
		},
		{
			name:     "24 hour duration",
			duration: 24 * time.Hour,
			want:     15,
		},
	}

	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			step := getStepFromDuration(tt.duration)
			assert.Equal(t, step, tt.want)
		})
	}
}
