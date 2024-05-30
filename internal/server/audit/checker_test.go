package audit

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestChecker(t *testing.T) {
	testCases := []struct {
		name          string
		eventPairs    []string
		expectedError error
		pairs         map[string]bool
	}{
		{
			name:          "wild card for nouns",
			eventPairs:    []string{"*:created"},
			expectedError: nil,
			pairs: map[string]bool{
				"constraint:created":   true,
				"distribution:created": true,
				"flag:created":         true,
				"namespace:created":    true,
				"rollout:created":      true,
				"rule:created":         true,
				"segment:created":      true,
				"token:created":        true,
				"variant:created":      true,
				"constraint:deleted":   false,
				"distribution:deleted": false,
				"flag:deleted":         false,
				"namespace:deleted":    false,
				"rollout:deleted":      false,
				"rule:deleted":         false,
				"segment:deleted":      false,
				"token:deleted":        false,
				"variant:deleted":      false,
				"constraint:updated":   false,
				"distribution:updated": false,
				"flag:updated":         false,
				"namespace:updated":    false,
				"rollout:updated":      false,
				"rule:updated":         false,
				"segment:updated":      false,
				"variant:updated":      false,
			},
		},
		{
			name:          "wild card for verbs",
			eventPairs:    []string{"flag:*"},
			expectedError: nil,
			pairs: map[string]bool{
				"constraint:created":   false,
				"distribution:created": false,
				"flag:created":         true,
				"namespace:created":    false,
				"rollout:created":      false,
				"rule:created":         false,
				"segment:created":      false,
				"token:created":        false,
				"variant:created":      false,
				"constraint:deleted":   false,
				"distribution:deleted": false,
				"flag:deleted":         true,
				"namespace:deleted":    false,
				"rollout:deleted":      false,
				"rule:deleted":         false,
				"segment:deleted":      false,
				"token:deleted":        false,
				"variant:deleted":      false,
				"constraint:updated":   false,
				"distribution:updated": false,
				"flag:updated":         true,
				"namespace:updated":    false,
				"rollout:updated":      false,
				"rule:updated":         false,
				"segment:updated":      false,
				"variant:updated":      false,
			},
		},
		{
			name:          "single pair",
			eventPairs:    []string{"flag:created"},
			expectedError: nil,
			pairs: map[string]bool{
				"constraint:created":   false,
				"distribution:created": false,
				"flag:created":         true,
				"namespace:created":    false,
				"rollout:created":      false,
				"rule:created":         false,
				"segment:created":      false,
				"token:created":        false,
				"variant:created":      false,
				"constraint:deleted":   false,
				"distribution:deleted": false,
				"flag:deleted":         false,
				"namespace:deleted":    false,
				"rollout:deleted":      false,
				"rule:deleted":         false,
				"segment:deleted":      false,
				"token:deleted":        false,
				"variant:deleted":      false,
				"constraint:updated":   false,
				"distribution:updated": false,
				"flag:updated":         false,
				"namespace:updated":    false,
				"rollout:updated":      false,
				"rule:updated":         false,
				"segment:updated":      false,
				"variant:updated":      false,
			},
		},
		{
			name:          "error repeating event pairs",
			eventPairs:    []string{"*:created", "flag:created"},
			expectedError: fmt.Errorf("repeated event pair: %s", "flag:created"),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			checker, err := NewChecker(tc.eventPairs)
			if tc.expectedError != nil {
				assert.EqualError(t, err, tc.expectedError.Error())
				return
			}

			for k, v := range tc.pairs {
				actual := checker.Check(k)
				assert.Equal(t, v, actual)
			}
		})
	}
}
