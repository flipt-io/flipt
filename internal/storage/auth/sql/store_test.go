package sql

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.flipt.io/flipt/errors"
	"go.flipt.io/flipt/internal/storage"
	fliptsql "go.flipt.io/flipt/internal/storage/sql"
	fliptsqltesting "go.flipt.io/flipt/internal/storage/sql/testing"
	storagetesting "go.flipt.io/flipt/internal/storage/testing"
	"go.flipt.io/flipt/rpc/flipt/auth"
	rpcauth "go.flipt.io/flipt/rpc/flipt/auth"
	"go.uber.org/zap"
	"go.uber.org/zap/zaptest"
	"google.golang.org/protobuf/types/known/timestamppb"
)

var (
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

func TestAuthenticationStoreHarness(t *testing.T) {
	storagetesting.TestAuthenticationStoreHarness(t, func(t *testing.T) storage.AuthenticationStore {
		return newTestStore(t)()
	})
}

func TestAuthentication_CreateAuthentication(t *testing.T) {
	// established a store factory with a single seeded auth entry
	storeFn := newTestStore(t, createAuth("create_auth_id", "create_auth_token", auth.Method_TOKEN))

	ctx := context.TODO()
	for _, test := range []struct {
		name                   string
		opts                   func(t *testing.T) []Option
		req                    *storage.CreateAuthenticationRequest
		expectedErrAs          error
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
		{
			name: "successfully creates authentication (no expiration)",
			opts: commonOpts,
			req: &storage.CreateAuthenticationRequest{
				Method: auth.Method_TOKEN,
				Metadata: map[string]string{
					"io.flipt.auth.token.name":        "access_all_areas",
					"io.flipt.auth.token.description": "The keys to the castle",
				},
			},
			expectedToken: "token:TestAuthentication_CreateAuthentication/successfully_creates_authentication_(no_expiration)",
			expectedAuthentication: &auth.Authentication{
				Id:     "id:TestAuthentication_CreateAuthentication/successfully_creates_authentication_(no_expiration)",
				Method: auth.Method_TOKEN,
				Metadata: map[string]string{
					"io.flipt.auth.token.name":        "access_all_areas",
					"io.flipt.auth.token.description": "The keys to the castle",
				},
				CreatedAt: someTimestamp,
				UpdatedAt: someTimestamp,
			},
		},
		{
			name: "fails ID uniqueness constraint",
			opts: func(t *testing.T) []Option {
				return []Option{
					WithIDGeneratorFunc(func() string {
						// return previous tests created ID
						return "create_auth_id"
					}),
				}
			},
			req: &storage.CreateAuthenticationRequest{
				Method:    auth.Method_TOKEN,
				ExpiresAt: timestamppb.New(time.Unix(2, 0)),
				Metadata: map[string]string{
					"io.flipt.auth.token.name":        "access_all_areas",
					"io.flipt.auth.token.description": "The keys to the castle",
				},
			},
			expectedErrAs: errPtr(errors.ErrInvalid("")),
		},
		{
			name: "fails token uniqueness constraint",
			opts: func(t *testing.T) []Option {
				return []Option{
					WithTokenGeneratorFunc(func() string {
						// return previous tests created token
						return "create_auth_token"
					}),
				}
			},
			req: &storage.CreateAuthenticationRequest{
				Method:    auth.Method_TOKEN,
				ExpiresAt: timestamppb.New(time.Unix(2, 0)),
				Metadata: map[string]string{
					"io.flipt.auth.token.name":        "access_all_areas",
					"io.flipt.auth.token.description": "The keys to the castle",
				},
			},
			expectedErrAs: errPtr(errors.ErrInvalid("")),
		},
	} {
		test := test
		t.Run(test.name, func(t *testing.T) {
			store := storeFn(test.opts(t)...)

			clientToken, created, err := store.CreateAuthentication(ctx, test.req)
			if test.expectedErrAs != nil {
				require.ErrorAs(t, err, test.expectedErrAs)
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

	// established a store factory with a single seeded auth entry
	storeFn := newTestStore(t, createAuth("get_auth_id", "get_auth_token", auth.Method_TOKEN))

	// run table tests
	for _, test := range []struct {
		name                   string
		clientToken            string
		expectedErrAs          error
		expectedAuthentication *auth.Authentication
	}{
		{
			name:          "error not found for unexpected clientToken",
			clientToken:   "unknown",
			expectedErrAs: errPtr(errors.ErrNotFound("")),
		},
		{
			name:        "successfully retrieves authentication by clientToken",
			clientToken: "get_auth_token",
			expectedAuthentication: &auth.Authentication{
				Id:     "get_auth_id",
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
			expectedErrAs          = test.expectedErrAs
			expectedAuthentication = test.expectedAuthentication
		)

		t.Run(test.name, func(t *testing.T) {
			retrieved, err := storeFn(commonOpts(t)...).GetAuthenticationByClientToken(ctx, clientToken)
			if expectedErrAs != nil {
				require.ErrorAs(t, err, expectedErrAs)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, expectedAuthentication, retrieved)
		})
	}
}

func TestAuthentication_ListAuthentications_ByMethod(t *testing.T) {
	ctx := context.TODO()
	storeFn := newTestStore(t,
		createAuth("token_id_one", "token_client_token_one", auth.Method_TOKEN),
		createAuth("token_id_two", "token_client_token_two", auth.Method_TOKEN),
		createAuth("token_id_three", "token_client_token_three", auth.Method_TOKEN),
		createAuth("none_id_one", "none_client_token_one", auth.Method_NONE),
		createAuth("none_id_two", "none_client_token_two", auth.Method_NONE),
		createAuth("none_id_three", "none_client_token_three", auth.Method_NONE),
	)

	// list predicated with none auth method
	req := storage.NewListRequest(ListWithMethod(auth.Method_NONE))
	noneMethod, err := storeFn().ListAuthentications(ctx, req)
	require.NoError(t, err)
	assert.Equal(t, storage.ResultSet[*rpcauth.Authentication]{
		Results: []*rpcauth.Authentication{
			{
				Id:     "none_id_one",
				Method: auth.Method_NONE,
				Metadata: map[string]string{
					"io.flipt.auth.token.name":        "access_some_areas",
					"io.flipt.auth.token.description": "The keys to some of the castle",
				},
				ExpiresAt: timestamppb.New(time.Unix(2, 0)),
				CreatedAt: someTimestamp,
				UpdatedAt: someTimestamp,
			},
			{
				Id:     "none_id_two",
				Method: auth.Method_NONE,
				Metadata: map[string]string{
					"io.flipt.auth.token.name":        "access_some_areas",
					"io.flipt.auth.token.description": "The keys to some of the castle",
				},
				ExpiresAt: timestamppb.New(time.Unix(2, 0)),
				CreatedAt: someTimestamp,
				UpdatedAt: someTimestamp,
			},
			{
				Id:     "none_id_three",
				Method: auth.Method_NONE,
				Metadata: map[string]string{
					"io.flipt.auth.token.name":        "access_some_areas",
					"io.flipt.auth.token.description": "The keys to some of the castle",
				},
				ExpiresAt: timestamppb.New(time.Unix(2, 0)),
				CreatedAt: someTimestamp,
				UpdatedAt: someTimestamp,
			},
		},
	}, noneMethod)

	// list predicated with token auth method
	req = storage.NewListRequest(ListWithMethod(auth.Method_TOKEN))
	tokenMethod, err := storeFn().ListAuthentications(ctx, req)
	require.NoError(t, err)
	assert.Equal(t, storage.ResultSet[*rpcauth.Authentication]{
		Results: []*rpcauth.Authentication{
			{
				Id:     "token_id_one",
				Method: auth.Method_TOKEN,
				Metadata: map[string]string{
					"io.flipt.auth.token.name":        "access_some_areas",
					"io.flipt.auth.token.description": "The keys to some of the castle",
				},
				ExpiresAt: timestamppb.New(time.Unix(2, 0)),
				CreatedAt: someTimestamp,
				UpdatedAt: someTimestamp,
			},
			{
				Id:     "token_id_two",
				Method: auth.Method_TOKEN,
				Metadata: map[string]string{
					"io.flipt.auth.token.name":        "access_some_areas",
					"io.flipt.auth.token.description": "The keys to some of the castle",
				},
				ExpiresAt: timestamppb.New(time.Unix(2, 0)),
				CreatedAt: someTimestamp,
				UpdatedAt: someTimestamp,
			},
			{
				Id:     "token_id_three",
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
	}, tokenMethod)
}

type authentication struct {
	id     string
	token  string
	method auth.Method
}

func createAuth(id, token string, method auth.Method) authentication {
	return authentication{id, token, method}
}

func newTestStore(t *testing.T, seed ...authentication) func(...Option) *Store {
	t.Helper()

	db, err := fliptsqltesting.Open()
	if err != nil {
		t.Fatal(err)
	}

	var (
		ctx     = context.TODO()
		logger  = zaptest.NewLogger(t)
		storeFn = func(opts ...Option) *Store {
			return NewStore(
				db.Driver,
				fliptsql.BuilderFor(db.DB, db.Driver),
				logger,
				opts...,
			)
		}
	)

	// seed any authentication fixtures
	for _, a := range seed {
		a := a
		store := storeFn(
			WithNowFunc(func() *timestamppb.Timestamp {
				return someTimestamp
			}),
			WithTokenGeneratorFunc(func() string { return a.token }),
			WithIDGeneratorFunc(func() string { return a.id }),
		)
		clientToken, _, err := store.CreateAuthentication(ctx, &storage.CreateAuthenticationRequest{
			Method:    a.method,
			ExpiresAt: timestamppb.New(time.Unix(2, 0)),
			Metadata: map[string]string{
				"io.flipt.auth.token.name":        "access_some_areas",
				"io.flipt.auth.token.description": "The keys to some of the castle",
			},
		})
		require.NoError(t, err)
		require.Equal(t, a.token, clientToken)

		logger.Debug("seeded authentication", zap.String("id", a.id))
	}

	return storeFn
}

func newStaticGenerator(t *testing.T, purpose string) func() string {
	t.Helper()

	return func() string {
		return fmt.Sprintf("%s:%s", purpose, t.Name())
	}
}

func errPtr[E error](e E) *E {
	return &e
}
