package analytics

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.flipt.io/flipt/errors"
)

func TestValidateRequiresScope(t *testing.T) {
	tests := []struct {
		name    string
		req     interface{ Validate() error }
		wantErr bool
	}{
		{name: "count with scope", req: &GetFlagEvaluationsCountRequest{EnvironmentKey: "env", NamespaceKey: "ns", FlagKey: "flag"}},
		{name: "count without environment", req: &GetFlagEvaluationsCountRequest{NamespaceKey: "ns", FlagKey: "flag"}, wantErr: true},
		{name: "count without namespace", req: &GetFlagEvaluationsCountRequest{EnvironmentKey: "env", FlagKey: "flag"}, wantErr: true},
		{name: "batch with scope", req: &GetBatchFlagEvaluationsCountRequest{EnvironmentKey: "env", NamespaceKey: "ns"}},
		{name: "batch without environment", req: &GetBatchFlagEvaluationsCountRequest{NamespaceKey: "ns"}, wantErr: true},
		{name: "batch without namespace", req: &GetBatchFlagEvaluationsCountRequest{EnvironmentKey: "env"}, wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.req.Validate()
			if !tt.wantErr {
				require.NoError(t, err)
				return
			}

			var invalid errors.ErrInvalid
			assert.ErrorAs(t, err, &invalid)
		})
	}
}
