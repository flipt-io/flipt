package redis

import (
	"context"
	errs "errors"
	"fmt"
	"strconv"
	"time"

	"github.com/google/uuid"
	goredis "github.com/redis/go-redis/v9"
	"go.flipt.io/flipt/errors"
	"go.flipt.io/flipt/internal/storage"
	"go.flipt.io/flipt/internal/storage/authn"
	rpcflipt "go.flipt.io/flipt/rpc/flipt"
	"go.flipt.io/flipt/rpc/flipt/auth"
	"go.uber.org/zap"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/types/known/timestamppb"
)

var _ authn.Store = (*Store)(nil)

const (
	authIDKeyPrefix    = "auth:id:"
	authTokenKeyPrefix = "auth:token:" //nolint:gosec
	authMethodPrefix   = "auth:method:"
	authAllKey         = "auth:all"

	allPattern = "*"

	authenticationKey = "authentication"
	tokenHashKey      = "token_hash"
	expiresAtKey      = "expires_at"

	batchSize = 1000
)

type Store struct {
	client             *goredis.Client
	logger             *zap.Logger
	now                func() *timestamppb.Timestamp
	generateID         func() string
	generateToken      func() string
	cleanupGracePeriod time.Duration
}

// Helper functions to generate Redis keys
func authIDKey(id string) string {
	return authIDKeyPrefix + id
}

func authTokenKey(token string) string {
	return authTokenKeyPrefix + token
}

func authMethodKey(method auth.Method) string {
	return authMethodPrefix + method.String()
}

// Option is a type which configures a *Store
type Option func(*Store)

func NewStore(c *goredis.Client, logger *zap.Logger, opts ...Option) *Store {
	store := &Store{
		client:             c,
		logger:             logger.With(zap.String("store", "redis")),
		now:                rpcflipt.Now,
		generateID:         uuid.NewString,
		generateToken:      authn.GenerateRandomToken,
		cleanupGracePeriod: 30 * time.Minute,
	}

	for _, opt := range opts {
		opt(store)
	}

	return store
}

func (s *Store) String() string {
	return "redis"
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

// WithCleanupGracePeriod overrides the stores cleanup grace period
// used to set the TTL on authentication records.
func WithCleanupGracePeriod(t time.Duration) Option {
	return func(s *Store) {
		s.cleanupGracePeriod = t
	}
}

// CreateAuthentication implements authn.Store.
func (s *Store) CreateAuthentication(ctx context.Context, r *authn.CreateAuthenticationRequest) (string, *auth.Authentication, error) {
	if r.ExpiresAt != nil && !r.ExpiresAt.IsValid() {
		return "", nil, errors.ErrInvalidf("invalid expiry time: %v", r.ExpiresAt)
	}

	var (
		now            = s.now()
		clientToken    = r.ClientToken
		authentication = &auth.Authentication{
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

	hashedToken, err := authn.HashClientToken(clientToken)
	if err != nil {
		return "", nil, fmt.Errorf("creating authentication: %w", err)
	}

	var (
		pipe  = s.client.Pipeline()
		idKey = authIDKey(authentication.Id)
	)

	v, err := protojson.Marshal(authentication)
	if err != nil {
		return "", nil, fmt.Errorf("marshalling authentication: %w", err)
	}

	// Store authentication data in Redis hash
	pipe.HSet(ctx, idKey, authenticationKey, v)
	pipe.HSet(ctx, idKey, tokenHashKey, hashedToken)

	// If expiry is set, add expiry time and set TTL
	if authentication.ExpiresAt != nil {
		pipe.HSet(ctx, idKey, expiresAtKey, authentication.ExpiresAt.AsTime().Unix())
		pipe.ExpireAt(ctx, idKey, authentication.ExpiresAt.AsTime().Add(s.cleanupGracePeriod))
	}

	// Store token hash -> id mapping
	tokenKey := authTokenKey(hashedToken)
	pipe.Set(ctx, tokenKey, authentication.Id, 0)
	if authentication.ExpiresAt != nil {
		pipe.ExpireAt(ctx, tokenKey, authentication.ExpiresAt.AsTime().Add(s.cleanupGracePeriod))
	}

	// Add to the set of all authentications for listing
	pipe.SAdd(ctx, authAllKey, authentication.Id)

	// Add to method index for filtering
	pipe.SAdd(ctx, authMethodKey(authentication.Method), authentication.Id)

	if _, err := pipe.Exec(ctx); err != nil {
		return "", nil, fmt.Errorf("storing authentication: %w", err)
	}

	return clientToken, authentication, nil
}

// DeleteAuthentications implements authn.Store.
func (s *Store) DeleteAuthentications(ctx context.Context, req *authn.DeleteAuthenticationsRequest) error {
	if err := req.Valid(); err != nil {
		return fmt.Errorf("deleting authentications: %w", err)
	}

	// Determine source set based on method filter
	sourceKey := authAllKey
	if req.Method != nil {
		sourceKey = authMethodKey(*req.Method)
	}

	ids, err := s.scanMatchingIDs(ctx, sourceKey, req)
	if err != nil {
		return err
	}

	return s.deleteAuthenticationBatches(ctx, ids, req.Method)
}

func (s *Store) scanMatchingIDs(ctx context.Context, sourceKey string, req *authn.DeleteAuthenticationsRequest) ([]string, error) {
	var (
		cursor uint64
		allIDs []string
	)

	for {
		ids, nextCursor, err := s.client.SScan(ctx, sourceKey, cursor, allPattern, batchSize).Result()
		if err != nil {
			return nil, fmt.Errorf("scanning authentications: %w", err)
		}

		matchingIDs, err := s.filterIDs(ctx, ids, req)
		if err != nil {
			return nil, err
		}
		allIDs = append(allIDs, matchingIDs...)

		if nextCursor == 0 {
			break
		}
		cursor = nextCursor
	}

	return allIDs, nil
}

func (s *Store) filterIDs(ctx context.Context, ids []string, req *authn.DeleteAuthenticationsRequest) ([]string, error) {
	if req.ID != nil {
		for _, id := range ids {
			if id == *req.ID {
				return []string{id}, nil
			}
		}
		return nil, nil
	}

	if req.ExpiredBefore == nil {
		return ids, nil
	}

	return s.filterExpiredIDs(ctx, ids, req.ExpiredBefore)
}

func (s *Store) filterExpiredIDs(ctx context.Context, ids []string, expiredBefore *timestamppb.Timestamp) ([]string, error) {
	var (
		pipe       = s.client.Pipeline()
		expiryCmds = make([]*goredis.StringCmd, len(ids))
	)

	for i, id := range ids {
		expiryCmds[i] = pipe.HGet(ctx, authIDKey(id), expiresAtKey)
	}

	if _, err := pipe.Exec(ctx); err != nil && !errs.Is(err, goredis.Nil) {
		return nil, fmt.Errorf("checking expiry times: %w", err)
	}

	var matchingIDs []string
	for i, id := range ids {
		expiresAtStr, err := expiryCmds[i].Result()
		if errs.Is(err, goredis.Nil) {
			continue
		}
		if err != nil {
			return nil, fmt.Errorf("getting expiry time: %w", err)
		}

		expiresAt, err := strconv.ParseInt(expiresAtStr, 10, 64)
		if err != nil {
			return nil, fmt.Errorf("parsing expiry time: %w", err)
		}

		if time.Unix(expiresAt, 0).Before(expiredBefore.AsTime()) {
			matchingIDs = append(matchingIDs, id)
		}
	}

	return matchingIDs, nil
}

func (s *Store) deleteAuthenticationBatches(ctx context.Context, allIDs []string, method *auth.Method) error {
	for i := 0; i < len(allIDs); i += batchSize {
		end := i + batchSize
		if end > len(allIDs) {
			end = len(allIDs)
		}

		if err := s.deleteAuthenticationBatch(ctx, allIDs[i:end], method); err != nil {
			return err
		}
	}
	return nil
}

func (s *Store) deleteAuthenticationBatch(ctx context.Context, ids []string, method *auth.Method) error {
	var (
		pipe      = s.client.Pipeline()
		tokenCmds = make([]*goredis.StringCmd, len(ids))
	)

	for i, id := range ids {
		tokenCmds[i] = pipe.HGet(ctx, authIDKey(id), tokenHashKey)
		if method != nil {
			pipe.SRem(ctx, authMethodKey(*method), id)
		}
		pipe.SRem(ctx, authAllKey, id)
		pipe.Del(ctx, authIDKey(id))
	}

	if _, err := pipe.Exec(ctx); err != nil && !errs.Is(err, goredis.Nil) {
		return fmt.Errorf("deleting authentications: %w", err)
	}

	pipe = s.client.Pipeline()
	for i := range ids {
		if tokenHash, err := tokenCmds[i].Result(); err == nil {
			pipe.Del(ctx, authTokenKey(tokenHash))
		}
	}

	if _, err := pipe.Exec(ctx); err != nil && !errs.Is(err, goredis.Nil) {
		return fmt.Errorf("deleting token mappings: %w", err)
	}

	return nil
}

// ExpireAuthenticationByID implements authn.Store.
func (s *Store) ExpireAuthenticationByID(ctx context.Context, id string, expireAt *timestamppb.Timestamp) error {
	// Get the token hash first
	idKey := authIDKey(id)
	tokenHash, err := s.client.HGet(ctx, idKey, tokenHashKey).Result()
	if err != nil {
		if errs.Is(err, goredis.Nil) {
			return errors.ErrNotFoundf("getting authentication by id")
		}
		return fmt.Errorf("getting authentication by id: %w", err)
	}

	// Update expiry in a pipeline
	pipe := s.client.Pipeline()

	// Update expiry in hash
	pipe.HSet(ctx, idKey, expiresAtKey, expireAt.AsTime().UnixNano())

	// Set TTL on both ID and token hash keys, adding the grace period
	pipe.ExpireAt(ctx, idKey, expireAt.AsTime().Add(s.cleanupGracePeriod))
	pipe.ExpireAt(ctx, authTokenKey(tokenHash), expireAt.AsTime().Add(s.cleanupGracePeriod))

	if _, err := pipe.Exec(ctx); err != nil {
		return fmt.Errorf("updating authentication expiry: %w", err)
	}

	return nil
}

// GetAuthenticationByClientToken implements authn.Store.
func (s *Store) GetAuthenticationByClientToken(ctx context.Context, clientToken string) (*auth.Authentication, error) {
	hashedToken, err := authn.HashClientToken(clientToken)
	if err != nil {
		return nil, fmt.Errorf("getting authentication by token: %w", err)
	}

	// Get ID from token hash
	tokenKey := authTokenKey(hashedToken)
	id, err := s.client.Get(ctx, tokenKey).Result()
	if err != nil {
		if errs.Is(err, goredis.Nil) {
			return nil, errors.ErrNotFoundf("getting authentication by token")
		}
		return nil, fmt.Errorf("getting authentication by token: %w", err)
	}

	return s.GetAuthenticationByID(ctx, id)
}

// GetAuthenticationByID implements authn.Store.
func (s *Store) GetAuthenticationByID(ctx context.Context, id string) (*auth.Authentication, error) {
	idKey := authIDKey(id)
	result, err := s.client.HGetAll(ctx, idKey).Result()
	if err != nil {
		return nil, fmt.Errorf("getting authentication by id: %w", err)
	}
	if len(result) == 0 {
		return nil, errors.ErrNotFoundf("getting authentication by id")
	}

	auth := &auth.Authentication{}
	if err := protojson.Unmarshal([]byte(result[authenticationKey]), auth); err != nil {
		return nil, fmt.Errorf("unmarshalling authentication: %w", err)
	}

	return auth, nil
}

// ListAuthentications implements authn.Store.
func (s *Store) ListAuthentications(ctx context.Context, req *storage.ListRequest[authn.ListAuthenticationsPredicate]) (storage.ResultSet[*auth.Authentication], error) {
	// Normalize query parameters
	req.QueryParams.Normalize()

	var (
		set    storage.ResultSet[*auth.Authentication]
		cursor uint64
	)

	// Parse page token as cursor if provided
	if req.QueryParams.PageToken != "" {
		var err error
		cursor, err = strconv.ParseUint(req.QueryParams.PageToken, 10, 64)
		if err != nil {
			return set, fmt.Errorf("parsing page token: %w", err)
		}
	}

	// Determine which set to scan based on method filter
	var key = authAllKey
	if req.Predicate.Method != nil {
		key = authMethodKey(*req.Predicate.Method)
	}

	// Scan the set with cursor pagination
	ids, nextCursor, err := s.client.SScan(ctx, key, cursor, allPattern, int64(req.QueryParams.Limit)).Result()
	if err != nil {
		return set, fmt.Errorf("scanning authentications: %w", err)
	}

	// Set next page token if there are more results
	if nextCursor > 0 {
		set.NextPageToken = strconv.FormatUint(nextCursor, 10)
	}

	// Get authentication details for each ID
	pipe := s.client.Pipeline()
	cmds := make(map[string]*goredis.MapStringStringCmd, len(ids))

	for _, id := range ids {
		cmds[id] = pipe.HGetAll(ctx, authIDKey(id))
	}

	if _, err := pipe.Exec(ctx); err != nil {
		return set, fmt.Errorf("getting authentications: %w", err)
	}

	// Process results
	set.Results = make([]*auth.Authentication, 0, len(ids))
	for _, id := range ids {
		result := cmds[id].Val()
		if len(result) == 0 {
			continue // Skip if authentication was deleted
		}

		auth := &auth.Authentication{}
		if err := protojson.Unmarshal([]byte(result[authenticationKey]), auth); err != nil {
			return set, fmt.Errorf("unmarshalling authentication: %w", err)
		}

		set.Results = append(set.Results, auth)
	}

	return set, nil
}

// Shutdown implements authn.Store.
func (s *Store) Shutdown(context.Context) error {
	return s.client.Close()
}
