package evaluation

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	errs "go.flipt.io/flipt/errors"
	"go.flipt.io/flipt/internal/common"
	"go.flipt.io/flipt/internal/server/authz"
	"go.flipt.io/flipt/internal/server/environments"
)

func TestResolveEnvironment(t *testing.T) {
	const header = "header-environment"

	tests := []struct {
		name          string
		key           string
		authzRequired bool
		// fallback is the environment GetFromContext must be asked for, or
		// empty when it must not be called
		fallback string
		wantErr  bool
	}{
		{name: "known key", key: "production"},
		{name: "unknown key falls back to header", key: "typo", fallback: header},
		{name: "blank key falls back to header", key: "", fallback: header},
		{name: "blank key falls back to header with authorization", key: "", authzRequired: true, fallback: header},
		{name: "unknown key is not found with authorization", key: "typo", authzRequired: true, fallback: "typo", wantErr: true},
		{name: "default alias resolves with authorization", key: "default", authzRequired: true, fallback: "default"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var (
				store    = &MockEnvironmentStore{}
				resolved = &environments.MockEnvironment{}
				ctx      = common.WithFliptEnvironment(t.Context(), header)
			)

			if tt.authzRequired {
				ctx = authz.ContextWithAuthorizationRequired(ctx)
			}

			if tt.fallback == "" {
				store.On("Get", mock.Anything, tt.key).Return(resolved, nil)
			} else {
				store.On("Get", mock.Anything, tt.key).Return(nil, errs.ErrNotFoundf("environment: %q", tt.key))

				var fallbackErr error
				if tt.wantErr {
					fallbackErr = errs.ErrNotFoundf("environment: %q", tt.key)
				}
				store.On("GetFromContext", mock.MatchedBy(func(ctx context.Context) bool {
					env, _ := common.FliptEnvironmentFromContext(ctx)
					return env == tt.fallback
				})).Return(resolved, fallbackErr)
			}

			env, err := ResolveEnvironment(ctx, store, tt.key)
			store.AssertExpectations(t)

			if tt.wantErr {
				var notFound errs.ErrNotFound
				require.ErrorAs(t, err, &notFound)
				return
			}

			require.NoError(t, err)
			assert.Same(t, resolved, env)
		})
	}
}
