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
			eventPairs:    []string{"*:create"},
			expectedError: nil,
			pairs: map[string]bool{
				"flag:create":       true,
				"constraint:create": true,
				"namespace:update":  false,
			},
		},
		{
			name:          "wild card for verbs",
			eventPairs:    []string{"flag:*"},
			expectedError: nil,
			pairs: map[string]bool{
				"flag:create":       true,
				"flag:delete":       true,
				"constraint:update": false,
			},
		},
		{
			name:          "error repeating event pairs",
			eventPairs:    []string{"*:create", "flag:create"},
			expectedError: fmt.Errorf("repeated event pair: %s", "flag:create"),
		},
	}

	for _, tc := range testCases {
		checker, err := NewChecker(tc.eventPairs)
		if tc.expectedError != nil {
			assert.EqualError(t, err, tc.expectedError.Error())
			continue
		}

		for k, v := range tc.pairs {
			actual := checker.Check(k)
			assert.Equal(t, v, actual)
		}
	}
}
