package auth

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.flipt.io/flipt/internal/storage"
	fliptsql "go.flipt.io/flipt/internal/storage/sql"
	fliptsqltesting "go.flipt.io/flipt/internal/storage/sql/testing"
	"go.flipt.io/flipt/rpc/flipt/auth"
	"go.uber.org/zap/zaptest"
	"google.golang.org/protobuf/types/known/timestamppb"
)

var (
	db *fliptsqltesting.Database

	someTimestamp = timestamppb.New(time.Date(2022, 10, 25, 18, 0, 0, 0, time.UTC))
	commonOpts    = func(t *testing.T) []Option {
		return []Option{
			WithNowFunc(func() *timestamppb.Timestamp {
				return someTimestamp
			}),
			// tokens created will be "token:<test_name>"
			WithTokenGeneratorFunc(newStaticGenerator(t, "token")),
			// ids created will be "id:<test_name>"
			WithIDGeneratorFunc(newStaticGenerator(t, "id")),
		}
	}
)

func TestMain(m *testing.M) {
	var err error
	db, err = fliptsqltesting.Open()
	if err != nil {
		panic(err)
	}
	os.Exit(m.Run())
}

func TestAuthentication_CreateAuthentication(t *testing.T) {
	ctx := context.TODO()
	for _, test := range []struct {
		name                   string
		opts                   func(t *testing.T) []Option
		req                    *storage.CreateAuthenticationRequest
		expectedErr            error
		expectedToken          string
		expectedAuthentication *auth.Authentication
	}{
		{
			name: "successfully creates authentication",
			opts: commonOpts,
			req: &storage.CreateAuthenticationRequest{
				Method:    auth.Method_TOKEN,
				ExpiresAt: timestamppb.New(time.Unix(2, 0)),
				Metadata: map[string]string{
					"io.flipt.auth.token.name":        "access_all_areas",
					"io.flipt.auth.token.description": "The keys to the castle",
				},
			},
			expectedToken: "token:TestAuthentication_CreateAuthentication/successfully_creates_authentication",
			expectedAuthentication: &auth.Authentication{
				Id:     "id:TestAuthentication_CreateAuthentication/successfully_creates_authentication",
				Method: auth.Method_TOKEN,
				Metadata: map[string]string{
					"io.flipt.auth.token.name":        "access_all_areas",
					"io.flipt.auth.token.description": "The keys to the castle",
				},
				ExpiresAt: timestamppb.New(time.Unix(2, 0)),
				CreatedAt: someTimestamp,
				UpdatedAt: someTimestamp,
			},
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			store := newTestStore(t, test.opts(t)...)

			clientToken, created, err := store.CreateAuthentication(ctx, test.req)
			if test.expectedErr != nil {
				require.Equal(t, test.expectedErr, err)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, test.expectedToken, clientToken)
			assert.Equal(t, test.expectedAuthentication, created)
		})
	}
}

func TestAuthentication_GetAuthenticationByClientToken(t *testing.T) {
	// seed database state
	ctx := context.TODO()
	store := newTestStore(t, commonOpts(t)...)
	clientToken, _, err := store.CreateAuthentication(ctx, &storage.CreateAuthenticationRequest{
		Method:    auth.Method_TOKEN,
		ExpiresAt: timestamppb.New(time.Unix(2, 0)),
		Metadata: map[string]string{
			"io.flipt.auth.token.name":        "access_some_areas",
			"io.flipt.auth.token.description": "The keys to some of the castle",
		},
	})
	require.NoError(t, err)
	require.Equal(t, "token:TestAuthentication_GetAuthenticationByClientToken", clientToken)

	// run table tests
	for _, test := range []struct {
		name                   string
		clientToken            string
		expectedErrIs          error
		expectedAuthentication *auth.Authentication
	}{
		{
			name:          "error not found for unexpected clientToken",
			clientToken:   "unknown",
			expectedErrIs: storage.ErrNotFound,
		},
		{
			name:        "successfully retrieves authentication by clientToken",
			clientToken: "token:TestAuthentication_GetAuthenticationByClientToken",
			expectedAuthentication: &auth.Authentication{
				Id:     "id:TestAuthentication_GetAuthenticationByClientToken",
				Method: auth.Method_TOKEN,
				Metadata: map[string]string{
					"io.flipt.auth.token.name":        "access_some_areas",
					"io.flipt.auth.token.description": "The keys to some of the castle",
				},
				ExpiresAt: timestamppb.New(time.Unix(2, 0)),
				CreatedAt: someTimestamp,
				UpdatedAt: someTimestamp,
			},
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			retrieved, err := store.GetAuthenticationByClientToken(ctx, test.clientToken)
			if test.expectedErrIs != nil {
				require.ErrorIs(t, err, test.expectedErrIs)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, test.expectedAuthentication, retrieved)
		})
	}
}

func newTestStore(t *testing.T, opts ...Option) *Store {
	t.Helper()

	var (
		builder = fliptsql.BuilderFor(db.DB, db.Driver)
		logger  = zaptest.NewLogger(t)
	)

	return NewStore(builder, logger, opts...)
}

func newStaticGenerator(t *testing.T, purpose string) func() string {
	t.Helper()

	return func() string {
		return fmt.Sprintf("%s:%s", purpose, t.Name())
	}
}
