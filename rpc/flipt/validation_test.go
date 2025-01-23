package flipt

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.flipt.io/flipt/errors"
)

func largeJSONString() string {
	prefix := `{"a":"`
	suffix := `"}`

	// adding one for making the string larger than the limit
	b := make([]byte, maxJsonStringSizeKB*1024-len(prefix)-len(suffix)+1)
	for i := range b {
		b[i] = 'a'
	}
	return fmt.Sprintf("%s%s%s", prefix, string(b), suffix)
}

func TestValidate_ListFlagRequest(t *testing.T) {
	tests := []struct {
		name    string
		req     *ListFlagRequest
		wantErr error
	}{
		{
			name:    "noLimitOffset",
			req:     &ListFlagRequest{Offset: 1},
			wantErr: errors.ErrInvalid("limit must be set when offset or pageToken is set"),
		},
		{
			name:    "noLimitPageToken",
			req:     &ListFlagRequest{PageToken: "foo"},
			wantErr: errors.ErrInvalid("limit must be set when offset or pageToken is set"),
		},
		{
			name: "validLimitOnly",
			req:  &ListFlagRequest{Limit: 1},
		},
		{
			name: "validLimitAndOffset",
			req:  &ListFlagRequest{Offset: 1, Limit: 1},
		},
		{
			name: "validLimitAndPageToken",
			req:  &ListFlagRequest{PageToken: "foo", Limit: 1},
		},
	}

	for _, tt := range tests {
		var (
			req     = tt.req
			wantErr = tt.wantErr
		)

		t.Run(tt.name, func(t *testing.T) {
			err := req.Validate()
			assert.Equal(t, wantErr, err)
		})
	}
}
