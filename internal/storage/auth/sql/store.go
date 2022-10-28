package sql

import (
	"context"
	"fmt"

	sq "github.com/Masterminds/squirrel"

	"github.com/gofrs/uuid"
	"go.flipt.io/flipt/internal/storage"
	"go.flipt.io/flipt/internal/storage/auth"
	fliptsql "go.flipt.io/flipt/internal/storage/sql"
	rpcauth "go.flipt.io/flipt/rpc/flipt/auth"
	"go.uber.org/zap"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// Store is the persistent storage layer for Authentications backed by SQL
// based relational database systems.
type Store struct {
	logger  *zap.Logger
	driver  fliptsql.Driver
	builder sq.StatementBuilderType

	now func() *timestamppb.Timestamp

	generateID    func() string
	generateToken func() string
}

// Option is a type which configures a *Store
type Option func(*Store)

// NewStore constructs and configures a new instance of *Store.
// Queries are issued to the database via the provided statement builder.
func NewStore(driver fliptsql.Driver, builder sq.StatementBuilderType, logger *zap.Logger, opts ...Option) *Store {
	store := &Store{
		logger:  logger,
		driver:  driver,
		builder: builder,
		now:     timestamppb.Now,
		generateID: func() string {
			return uuid.Must(uuid.NewV4()).String()
		},
		generateToken: auth.GenerateRandomToken,
	}

	for _, opt := range opts {
		opt(store)
	}

	return store
}

// WithNowFunc overrides the stores now() function used to obtain
// a protobuf timestamp representative of the current time of evaluation.
func WithNowFunc(fn func() *timestamppb.Timestamp) Option {
	return func(s *Store) {
		s.now = fn
	}
}

// WithTokenGeneratorFunc overrides the stores token generator function
// used to generate new random token strings as client tokens, when
// creating new instances of Authentication.
// The default is a pseudo-random string of bytes base64 encoded.
func WithTokenGeneratorFunc(fn func() string) Option {
	return func(s *Store) {
		s.generateToken = fn
	}
}

// WithIDGeneratorFunc overrides the stores ID generator function
// used to generate new random ID strings, when creating new instances
// of Authentications.
// The default is a string containing a valid UUID (V4).
func WithIDGeneratorFunc(fn func() string) Option {
	return func(s *Store) {
		s.generateID = fn
	}
}

// CreateAuthentication creates and persists an instance of an Authentication.
func (s *Store) CreateAuthentication(ctx context.Context, r *storage.CreateAuthenticationRequest) (string, *rpcauth.Authentication, error) {
	var (
		now            = s.now()
		clientToken    = s.generateToken()
		authentication = rpcauth.Authentication{
			Id:        s.generateID(),
			Method:    r.Method,
			Metadata:  r.Metadata,
			ExpiresAt: r.ExpiresAt,
			CreatedAt: now,
			UpdatedAt: now,
		}
	)

	hashedToken, err := auth.HashClientToken(clientToken)
	if err != nil {
		return "", nil, fmt.Errorf("creating authentication: %w", err)
	}

	if _, err := s.builder.Insert("authentications").
		Columns(
			"id",
			"hashed_client_token",
			"method",
			"metadata",
			"expires_at",
			"created_at",
			"updated_at",
		).
		Values(
			&authentication.Id,
			&hashedToken,
			&authentication.Method,
			&fliptsql.JSONField[map[string]string]{T: authentication.Metadata},
			&fliptsql.NullableTimestamp{Timestamp: authentication.ExpiresAt},
			&fliptsql.Timestamp{Timestamp: authentication.CreatedAt},
			&fliptsql.Timestamp{Timestamp: authentication.UpdatedAt},
		).
		ExecContext(ctx); err != nil {
		return "", nil, fmt.Errorf(
			"inserting authentication %q: %w",
			authentication.Id,
			s.driver.AdaptError(err),
		)
	}

	return clientToken, &authentication, nil
}

// GetAuthenticationByClientToken fetches the associated Authentication for the provided clientToken string.
//
// Given a row is present for the hash of the clientToken then materialize into an Authentication.
// Else, given it cannot be located, a storage.ErrNotFound error is wrapped and returned instead.
func (s *Store) GetAuthenticationByClientToken(ctx context.Context, clientToken string) (*rpcauth.Authentication, error) {
	hashedToken, err := auth.HashClientToken(clientToken)
	if err != nil {
		return nil, fmt.Errorf("getting authentication by token: %w", err)
	}

	var (
		authentication rpcauth.Authentication
		expiresAt      fliptsql.NullableTimestamp
		createdAt      fliptsql.Timestamp
		updatedAt      fliptsql.Timestamp
	)

	if err := s.builder.Select(
		"id",
		"method",
		"metadata",
		"expires_at",
		"created_at",
		"updated_at",
	).
		From("authentications").
		Where(sq.Eq{"hashed_client_token": hashedToken}).
		QueryRowContext(ctx).
		Scan(
			&authentication.Id,
			&authentication.Method,
			&fliptsql.JSONField[*map[string]string]{T: &authentication.Metadata},
			&expiresAt,
			&createdAt,
			&updatedAt,
		); err != nil {
		return nil, fmt.Errorf(
			"getting authentication by token: %w",
			s.driver.AdaptError(err),
		)
	}

	authentication.ExpiresAt = expiresAt.Timestamp
	authentication.CreatedAt = createdAt.Timestamp
	authentication.UpdatedAt = updatedAt.Timestamp

	return &authentication, nil
}
