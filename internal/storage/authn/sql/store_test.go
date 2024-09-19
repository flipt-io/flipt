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
	storageauth "go.flipt.io/flipt/internal/storage/authn"
	authtesting "go.flipt.io/flipt/internal/storage/authn/testing"
	storagesql "go.flipt.io/flipt/internal/storage/sql"
	sqltesting "go.flipt.io/flipt/internal/storage/sql/testing"
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
	authtesting.TestAuthenticationStoreHarness(t, func(t *testing.T) storageauth.Store {
		return newTestStore(t)()
	})
}

func TestAuthentication_CreateAuthentication(t *testing.T) {
	// established a store factory with a single seeded auth entry
	storeFn := newTestStore(t, createAuth("create_auth_id", "create_auth_token", rpcauth.Method_METHOD_TOKEN))

	ctx := context.TODO()
	for _, test := range []struct {
		name                   string
		opts                   func(t *testing.T) []Option
		req                    *storageauth.CreateAuthenticationRequest
		expectedErrAs          error
		expectedToken          string
		expectedAuthentication *rpcauth.Authentication
	}{
		{
			name: "successfully creates authentication",
			opts: commonOpts,
			req: &storageauth.CreateAuthenticationRequest{
				Method:    rpcauth.Method_METHOD_TOKEN,
				ExpiresAt: timestamppb.New(time.Unix(2, 0)),
				Metadata: map[string]string{
					"io.flipt.auth.token.name":        "access_all_areas",
					"io.flipt.auth.token.description": "The keys to the castle",
				},
			},
			expectedToken: "token:TestAuthentication_CreateAuthentication/successfully_creates_authentication",
			expectedAuthentication: &rpcauth.Authentication{
				Id:     "id:TestAuthentication_CreateAuthentication/successfully_creates_authentication",
				Method: rpcauth.Method_METHOD_TOKEN,
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
			req: &storageauth.CreateAuthenticationRequest{
				Method: rpcauth.Method_METHOD_TOKEN,
				Metadata: map[string]string{
					"io.flipt.auth.token.name":        "access_all_areas",
					"io.flipt.auth.token.description": "The keys to the castle",
				},
			},
			expectedToken: "token:TestAuthentication_CreateAuthentication/successfully_creates_authentication_(no_expiration)",
			expectedAuthentication: &rpcauth.Authentication{
				Id:     "id:TestAuthentication_CreateAuthentication/successfully_creates_authentication_(no_expiration)",
				Method: rpcauth.Method_METHOD_TOKEN,
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
			req: &storageauth.CreateAuthenticationRequest{
				Method:    rpcauth.Method_METHOD_TOKEN,
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
			req: &storageauth.CreateAuthenticationRequest{
				Method:    rpcauth.Method_METHOD_TOKEN,
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
				// nolint:testifylint
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
	storeFn := newTestStore(t, createAuth("get_auth_id", "get_auth_token", rpcauth.Method_METHOD_TOKEN))

	// run table tests
	for _, test := range []struct {
		name                   string
		clientToken            string
		expectedErrAs          error
		expectedAuthentication *rpcauth.Authentication
	}{
		{
			name:          "error not found for unexpected clientToken",
			clientToken:   "unknown",
			expectedErrAs: errPtr(errors.ErrNotFound("")),
		},
		{
			name:        "successfully retrieves authentication by clientToken",
			clientToken: "get_auth_token",
			expectedAuthentication: &rpcauth.Authentication{
				Id:     "get_auth_id",
				Method: rpcauth.Method_METHOD_TOKEN,
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
				// nolint:testifylint
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

	// increment each timestamp by 1 when seeding auths
	var i int64
	seedOpts := func(t *testing.T) []Option {
		return []Option{WithNowFunc(func() *timestamppb.Timestamp {
			i++
			return timestamppb.New(time.Unix(i, 0))
		})}
	}

	storeFn := newTestStore(t,
		createAuth("none_id_one", "none_client_token_one", rpcauth.Method_METHOD_NONE, withOpts(seedOpts)),
		createAuth("none_id_two", "none_client_token_two", rpcauth.Method_METHOD_NONE, withOpts(seedOpts)),
		createAuth("none_id_three", "none_client_token_three", rpcauth.Method_METHOD_NONE, withOpts(seedOpts)),
		createAuth("token_id_one", "token_client_token_one", rpcauth.Method_METHOD_TOKEN, withOpts(seedOpts)),
		createAuth("token_id_two", "token_client_token_two", rpcauth.Method_METHOD_TOKEN, withOpts(seedOpts)),
		createAuth("token_id_three", "token_client_token_three", rpcauth.Method_METHOD_TOKEN, withOpts(seedOpts)),
	)

	t.Run("method == none", func(t *testing.T) {
		// list predicated with none auth method
		req := storage.ListWithOptions(storageauth.ListMethod(rpcauth.Method_METHOD_NONE))
		noneMethod, err := storeFn().ListAuthentications(ctx, req)

		require.NoError(t, err)
		assert.Equal(t, storage.ResultSet[*rpcauth.Authentication]{
			Results: []*rpcauth.Authentication{
				{
					Id:     "none_id_one",
					Method: rpcauth.Method_METHOD_NONE,
					Metadata: map[string]string{
						"io.flipt.auth.token.name":        "access_some_areas",
						"io.flipt.auth.token.description": "The keys to some of the castle",
					},
					ExpiresAt: timestamppb.New(time.Unix(2, 0)),
					CreatedAt: timestamppb.New(time.Unix(1, 0)),
					UpdatedAt: timestamppb.New(time.Unix(1, 0)),
				},
				{
					Id:     "none_id_two",
					Method: rpcauth.Method_METHOD_NONE,
					Metadata: map[string]string{
						"io.flipt.auth.token.name":        "access_some_areas",
						"io.flipt.auth.token.description": "The keys to some of the castle",
					},
					ExpiresAt: timestamppb.New(time.Unix(2, 0)),
					CreatedAt: timestamppb.New(time.Unix(2, 0)),
					UpdatedAt: timestamppb.New(time.Unix(2, 0)),
				},
				{
					Id:     "none_id_three",
					Method: rpcauth.Method_METHOD_NONE,
					Metadata: map[string]string{
						"io.flipt.auth.token.name":        "access_some_areas",
						"io.flipt.auth.token.description": "The keys to some of the castle",
					},
					ExpiresAt: timestamppb.New(time.Unix(2, 0)),
					CreatedAt: timestamppb.New(time.Unix(3, 0)),
					UpdatedAt: timestamppb.New(time.Unix(3, 0)),
				},
			},
		}, noneMethod)
	})

	t.Run("method == token", func(t *testing.T) {
		// list predicated with token auth method
		req := storage.ListWithOptions(storageauth.ListMethod(rpcauth.Method_METHOD_TOKEN))
		tokenMethod, err := storeFn().ListAuthentications(ctx, req)
		require.NoError(t, err)
		assert.Equal(t, storage.ResultSet[*rpcauth.Authentication]{
			Results: []*rpcauth.Authentication{
				{
					Id:     "token_id_one",
					Method: rpcauth.Method_METHOD_TOKEN,
					Metadata: map[string]string{
						"io.flipt.auth.token.name":        "access_some_areas",
						"io.flipt.auth.token.description": "The keys to some of the castle",
					},
					ExpiresAt: timestamppb.New(time.Unix(2, 0)),
					CreatedAt: timestamppb.New(time.Unix(4, 0)),
					UpdatedAt: timestamppb.New(time.Unix(4, 0)),
				},
				{
					Id:     "token_id_two",
					Method: rpcauth.Method_METHOD_TOKEN,
					Metadata: map[string]string{
						"io.flipt.auth.token.name":        "access_some_areas",
						"io.flipt.auth.token.description": "The keys to some of the castle",
					},
					ExpiresAt: timestamppb.New(time.Unix(2, 0)),
					CreatedAt: timestamppb.New(time.Unix(5, 0)),
					UpdatedAt: timestamppb.New(time.Unix(5, 0)),
				},
				{
					Id:     "token_id_three",
					Method: rpcauth.Method_METHOD_TOKEN,
					Metadata: map[string]string{
						"io.flipt.auth.token.name":        "access_some_areas",
						"io.flipt.auth.token.description": "The keys to some of the castle",
					},
					ExpiresAt: timestamppb.New(time.Unix(2, 0)),
					CreatedAt: timestamppb.New(time.Unix(6, 0)),
					UpdatedAt: timestamppb.New(time.Unix(6, 0)),
				},
			},
		}, tokenMethod)
	})
}

type authentication struct {
	id     string
	token  string
	method rpcauth.Method
	optFn  func(t *testing.T) []Option
}

func withOpts(optFn func(t *testing.T) []Option) func(*authentication) {
	return func(a *authentication) {
		a.optFn = optFn
	}
}

func createAuth(id, token string, method rpcauth.Method, opts ...func(*authentication)) authentication {
	a := authentication{id, token, method, nil}
	for _, opt := range opts {
		opt(&a)
	}

	return a
}

func newTestStore(t *testing.T, seed ...authentication) func(...Option) *Store {
	t.Helper()

	db, err := sqltesting.Open()
	if err != nil {
		t.Fatal(err)
	}

	var (
		ctx     = context.TODO()
		logger  = zaptest.NewLogger(t)
		storeFn = func(opts ...Option) *Store {
			return NewStore(
				db.Driver,
				storagesql.BuilderFor(db.DB, db.Driver, true),
				logger,
				opts...,
			)
		}
	)

	// seed any authentication fixtures
	for _, a := range seed {
		a := a
		opts := []Option{
			WithNowFunc(func() *timestamppb.Timestamp {
				return someTimestamp
			}),
			WithTokenGeneratorFunc(func() string { return a.token }),
			WithIDGeneratorFunc(func() string { return a.id }),
		}

		if a.optFn != nil {
			opts = append(opts, a.optFn(t)...)
		}

		clientToken, _, err := storeFn(opts...).CreateAuthentication(ctx, &storageauth.CreateAuthenticationRequest{
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

		time.Sleep(10 * time.Millisecond)
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
