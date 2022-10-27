package auth

import (
	"context"
	"encoding/base64"
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
		var (
			opts                   = test.opts
			req                    = test.req
			expectedErr            = test.expectedErr
			expectedToken          = test.expectedToken
			expectedAuthentication = test.expectedAuthentication
		)

		t.Run(test.name, func(t *testing.T) {
			store := newTestStore(t, opts(t)...)

			clientToken, created, err := store.CreateAuthentication(ctx, req)
			if expectedErr != nil {
				require.Equal(t, expectedErr, err)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, expectedToken, clientToken)
			assert.Equal(t, expectedAuthentication, created)
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
		var (
			clientToken            = test.clientToken
			expectedErrIs          = test.expectedErrIs
			expectedAuthentication = test.expectedAuthentication
		)

		t.Run(test.name, func(t *testing.T) {
			retrieved, err := store.GetAuthenticationByClientToken(ctx, clientToken)
			if expectedErrIs != nil {
				require.ErrorIs(t, err, expectedErrIs)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, expectedAuthentication, retrieved)
		})
	}
}

func Fuzz_hashClientToken(f *testing.F) {
	for _, seed := range []string{
		"hello, world",
		"supersecretstring",
		"egGpvIxtdG6tI3OIJjXOrv7xZW3hRMYg/Lt/G6X/UEwC",
	} {
		f.Add(seed)
	}
	for _, seed := range [][]byte{{}, {0}, {9}, {0xa}, {0xf}, {1, 2, 3, 4}} {
		f.Add(string(seed))
	}
	f.Fuzz(func(t *testing.T, token string) {
		hashed, err := hashClientToken(token)
		require.NoError(t, err)
		require.NotEmpty(t, hashed, "hashed result is empty")

		_, err = base64.URLEncoding.DecodeString(hashed)
		require.NoError(t, err)
	})
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
