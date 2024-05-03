package ecr

import (
	"io"
	"net/http"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestRoundTripper(t *testing.T) {
	tests := []struct {
		name           string
		tripper        RoundTripperFunc
		wantStatusCode int
		wantErr        error
	}{
		{
			name: "on status ok",
			tripper: func(r *http.Request) (*http.Response, error) {
				return &http.Response{StatusCode: http.StatusOK}, nil
			},
			wantStatusCode: http.StatusOK,
			wantErr:        nil,
		},
		{
			name: "on error",
			tripper: func(r *http.Request) (*http.Response, error) {
				return nil, io.ErrUnexpectedEOF
			},
			wantStatusCode: 0,
			wantErr:        io.ErrUnexpectedEOF,
		},
		{
			name: "authorization token has expired.",
			tripper: func(r *http.Request) (*http.Response, error) {
				return &http.Response{
					StatusCode: http.StatusForbidden,
					Status:     "denied: Your authorization token has expired. Reauthenticate and try again.",
				}, nil
			},
			wantStatusCode: http.StatusUnauthorized,
			wantErr:        nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, err := tt.tripper.RoundTrip(&http.Request{})
			require.ErrorIs(t, err, tt.wantErr)
			if tt.wantErr == nil {
				if resp.Body != nil {
					resp.Body.Close()
				}
				require.Equal(t, tt.wantStatusCode, resp.StatusCode)
			}
		})
	}
}
