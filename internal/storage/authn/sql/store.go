package sql

import (
	"context"
	"fmt"
	"strconv"
	"time"

	sq "github.com/Masterminds/squirrel"

	"github.com/gofrs/uuid"
	"go.flipt.io/flipt/internal/storage"
	storageauth "go.flipt.io/flipt/internal/storage/authn"
	storagesql "go.flipt.io/flipt/internal/storage/sql"
	rpcauth "go.flipt.io/flipt/rpc/flipt/auth"
	"go.uber.org/zap"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// Store is the persistent storage layer for Authentications backed by SQL
// based relational database systems.
type Store struct {
	logger  *zap.Logger
	driver  storagesql.Driver
	builder sq.StatementBuilderType

	now func() *timestamppb.Timestamp

	generateID    func() string
	generateToken func() string
}

// Option is a type which configures a *Store
type Option func(*Store)

// NewStore constructs and configures a new instance of *Store.
// Queries are issued to the database via the provided statement builder.
func NewStore(driver storagesql.Driver, builder sq.StatementBuilderType, logger *zap.Logger, opts ...Option) *Store {
	store := &Store{
		logger:  logger,
		driver:  driver,
		builder: builder,
		now: func() *timestamppb.Timestamp {
			// we truncate timestamps to the microsecond to support Postgres/MySQL
			// the lowest common denominators in terms of timestamp precision
			now := time.Now().UTC().Truncate(time.Microsecond)
			return timestamppb.New(now)
		},
		generateID: func() string {
			return uuid.Must(uuid.NewV4()).String()
		},
		generateToken: storageauth.GenerateRandomToken,
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
func (s *Store) CreateAuthentication(ctx context.Context, r *storageauth.CreateAuthenticationRequest) (string, *rpcauth.Authentication, error) {
	var (
		now            = s.now()
		clientToken    = r.ClientToken
		authentication = rpcauth.Authentication{
			Id:        s.generateID(),
			Method:    r.Method,
			Metadata:  r.Metadata,
			ExpiresAt: r.ExpiresAt,
			CreatedAt: now,
			UpdatedAt: now,
		}
	)

	// if no client token is provided, generate a new one
	if clientToken == "" {
		clientToken = s.generateToken()
	}

	hashedToken, err := storageauth.HashClientToken(clientToken)
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
			&storagesql.JSONField[map[string]string]{T: authentication.Metadata},
			&storagesql.NullableTimestamp{Timestamp: authentication.ExpiresAt},
			&storagesql.Timestamp{Timestamp: authentication.CreatedAt},
			&storagesql.Timestamp{Timestamp: authentication.UpdatedAt},
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
	hashedToken, err := storageauth.HashClientToken(clientToken)
	if err != nil {
		return nil, fmt.Errorf("getting authentication by token: %w", err)
	}

	var authentication rpcauth.Authentication

	if err := s.scanAuthentication(
		s.builder.
			Select(
				"id",
				"method",
				"metadata",
				"expires_at",
				"created_at",
				"updated_at",
			).
			From("authentications").
			Where(sq.Eq{"hashed_client_token": hashedToken}).
			QueryRowContext(ctx), &authentication); err != nil {
		return nil, fmt.Errorf(
			"getting authentication by token: %w",
			s.driver.AdaptError(err),
		)
	}

	return &authentication, nil
}

// GetAuthenticationByID retrieves an instance of Authentication from the backing
// store using the provided id string.
func (s *Store) GetAuthenticationByID(ctx context.Context, id string) (*rpcauth.Authentication, error) {
	var authentication rpcauth.Authentication

	if err := s.scanAuthentication(
		s.builder.
			Select(
				"id",
				"method",
				"metadata",
				"expires_at",
				"created_at",
				"updated_at",
			).
			From("authentications").
			Where(sq.Eq{"id": id}).
			QueryRowContext(ctx), &authentication); err != nil {
		return nil, fmt.Errorf(
			"getting authentication by token: %w",
			s.driver.AdaptError(err),
		)
	}

	return &authentication, nil
}

// ListAuthentications lists a page of Authentications from the backing store.
func (s *Store) ListAuthentications(ctx context.Context, req *storage.ListRequest[storageauth.ListAuthenticationsPredicate]) (set storage.ResultSet[*rpcauth.Authentication], err error) {
	defer func() {
		if err != nil {
			err = fmt.Errorf(
				"listing authentications: %w",
				s.driver.AdaptError(err),
			)
		}
	}()

	// adjust the query parameters within normal bounds
	req.QueryParams.Normalize()

	query := s.builder.
		Select(
			"id",
			"method",
			"metadata",
			"expires_at",
			"created_at",
			"updated_at",
		).
		From("authentications").
		Limit(req.QueryParams.Limit + 1).
		OrderBy(fmt.Sprintf("created_at %s", req.QueryParams.Order))

	if req.Predicate.Method != nil {
		query = query.Where(sq.Eq{"method": *req.Predicate.Method})
	}

	var offset int
	if v, err := strconv.ParseInt(req.QueryParams.PageToken, 10, 64); err == nil {
		offset = int(v)
		query = query.Offset(uint64(v))
	}

	rows, err := query.QueryContext(ctx)
	if err != nil {
		return set, err
	}

	defer func() { _ = rows.Close() }()

	for rows.Next() {
		var authentication rpcauth.Authentication
		if err = s.scanAuthentication(rows, &authentication); err != nil {
			return
		}

		if len(set.Results) >= int(req.QueryParams.Limit) {
			// set the next page token to the first
			// row beyond the query limit and break
			set.NextPageToken = fmt.Sprintf("%d", offset+int(req.QueryParams.Limit))
			break
		}

		set.Results = append(set.Results, &authentication)
	}

	if err = rows.Err(); err != nil {
		return
	}

	return
}

func (s *Store) adaptError(fmtStr string, err *error) {
	if *err != nil {
		*err = fmt.Errorf(fmtStr, s.driver.AdaptError(*err))
	}
}

// DeleteAuthentications attempts to delete one or more Authentication instances from the backing store.
// Use auth.DeleteByID to construct a request to delete a single Authentication by ID string.
// Use auth.DeleteByMethod to construct a request to delete 0 or more Authentications by Method and optional expired before constraint.
func (s *Store) DeleteAuthentications(ctx context.Context, req *storageauth.DeleteAuthenticationsRequest) (err error) {
	defer s.adaptError("deleting authentications: %w", &err)

	if err := req.Valid(); err != nil {
		return err
	}

	query := s.builder.
		Delete("authentications")

	if req.ID != nil {
		query = query.Where(sq.Eq{"id": req.ID})
	}

	if req.Method != nil {
		query = query.Where(sq.Eq{"method": req.Method})
	}

	if req.ExpiredBefore != nil {
		query = query.Where(sq.Lt{
			"expires_at": &storagesql.Timestamp{Timestamp: req.ExpiredBefore},
		})
	}

	_, err = query.ExecContext(ctx)

	return
}

// ExpireAuthenticationByID attempts to expire an Authentication by ID string and the provided expiry time.
func (s *Store) ExpireAuthenticationByID(ctx context.Context, id string, expireAt *timestamppb.Timestamp) (err error) {
	defer s.adaptError("expiring authentication by id: %w", &err)

	_, err = s.builder.
		Update("authentications").
		Set("expires_at", &storagesql.Timestamp{Timestamp: expireAt}).
		Where(sq.Eq{"id": id}).
		ExecContext(ctx)

	return
}

func (s *Store) scanAuthentication(scanner sq.RowScanner, authentication *rpcauth.Authentication) error {
	var (
		expiresAt storagesql.NullableTimestamp
		createdAt storagesql.Timestamp
		updatedAt storagesql.Timestamp
	)

	if err := scanner.
		Scan(
			&authentication.Id,
			&authentication.Method,
			&storagesql.JSONField[*map[string]string]{T: &authentication.Metadata},
			&expiresAt,
			&createdAt,
			&updatedAt,
		); err != nil {
		return fmt.Errorf(
			"reading authentication: %w",
			s.driver.AdaptError(err),
		)
	}

	authentication.ExpiresAt = expiresAt.Timestamp
	authentication.CreatedAt = createdAt.Timestamp
	authentication.UpdatedAt = updatedAt.Timestamp

	return nil
}
