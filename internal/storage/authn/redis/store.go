package redis

import (
	"context"
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
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/types/known/timestamppb"
)

var _ authn.Store = (*Store)(nil)

type Store struct {
	client *goredis.Client
}

func NewStore(c *goredis.Client) *Store {
	return &Store{client: c}
}

// CreateAuthentication implements authn.Store.
func (s *Store) CreateAuthentication(ctx context.Context, r *authn.CreateAuthenticationRequest) (string, *auth.Authentication, error) {
	if r.ExpiresAt != nil && !r.ExpiresAt.IsValid() {
		return "", nil, errors.ErrInvalidf("invalid expiry time: %v", r.ExpiresAt)
	}

	var (
		now            = rpcflipt.Now()
		clientToken    = r.ClientToken
		authentication = &auth.Authentication{
			Id:        uuid.NewString(),
			Method:    r.Method,
			Metadata:  r.Metadata,
			ExpiresAt: r.ExpiresAt,
			CreatedAt: now,
			UpdatedAt: now,
		}
	)

	// if no client token is provided, generate a new one
	if clientToken == "" {
		clientToken = authn.GenerateRandomToken()
	}

	hashedToken, err := authn.HashClientToken(clientToken)
	if err != nil {
		return "", nil, fmt.Errorf("creating authentication: %w", err)
	}

	// Store authentication data in Redis hash
	pipe := s.client.Pipeline()

	// Store by ID for direct lookups
	idKey := fmt.Sprintf("auth:id:%s", authentication.Id)

	v, err := protojson.Marshal(authentication)
	if err != nil {
		return "", nil, fmt.Errorf("marshalling authentication: %w", err)
	}

	// Store authentication data in Redis hash
	pipe.HSet(ctx, idKey, "authentication", v)
	pipe.HSet(ctx, idKey, "token_hash", hashedToken)

	// If expiry is set, add expiry time and set TTL
	if authentication.ExpiresAt != nil {
		pipe.HSet(ctx, idKey, "expires_at", authentication.ExpiresAt.AsTime().UnixNano())
		pipe.ExpireAt(ctx, idKey, authentication.ExpiresAt.AsTime())
	}

	// Store token hash -> id mapping
	tokenKey := fmt.Sprintf("auth:token:%s", hashedToken)
	pipe.Set(ctx, tokenKey, authentication.Id, 0)
	if authentication.ExpiresAt != nil {
		pipe.ExpireAt(ctx, tokenKey, authentication.ExpiresAt.AsTime())
	}

	// Add to the set of all authentications for listing
	pipe.SAdd(ctx, "auth:all", authentication.Id)

	// Add to method index for filtering
	pipe.SAdd(ctx, fmt.Sprintf("auth:method:%s", authentication.Method.String()), authentication.Id)

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

	// Get IDs to delete based on filters
	var key = "auth:all"
	if req.Method != nil {
		key = fmt.Sprintf("auth:method:%s", req.Method.String())
	}

	var cursor uint64
	var allIDs []string

	// Scan all IDs that match the method filter with larger batch size
	for {
		ids, nextCursor, err := s.client.SScan(ctx, key, cursor, "*", 1000).Result()
		if err != nil {
			return fmt.Errorf("scanning authentications: %w", err)
		}

		// If we have an ID filter, only keep that ID
		if req.ID != nil {
			for _, id := range ids {
				if id == *req.ID {
					allIDs = append(allIDs, id)
					break
				}
			}
		} else {
			// If we have an ExpiredBefore filter, check expiry for each ID
			if req.ExpiredBefore != nil {
				// First pipeline: check expiries in batches
				checkPipe := s.client.Pipeline()
				expiryCmds := make([]*goredis.StringCmd, len(ids))

				for i, id := range ids {
					expiryCmds[i] = checkPipe.HGet(ctx, fmt.Sprintf("auth:id:%s", id), "expires_at")
				}

				if _, err := checkPipe.Exec(ctx); err != nil && err != goredis.Nil {
					return fmt.Errorf("getting expiry times: %w", err)
				}

				// Filter IDs based on expiry
				for i, id := range ids {
					expiresAtStr, err := expiryCmds[i].Result()
					if err == goredis.Nil {
						continue // Skip if no expiry
					}
					if err != nil {
						return fmt.Errorf("getting expiry time: %w", err)
					}

					expiresAt, err := strconv.ParseInt(expiresAtStr, 10, 64)
					if err != nil {
						return fmt.Errorf("parsing expiry time: %w", err)
					}

					if time.Unix(0, expiresAt).Before(req.ExpiredBefore.AsTime()) {
						allIDs = append(allIDs, id)
					}
				}
			} else {
				allIDs = append(allIDs, ids...)
			}
		}

		if nextCursor == 0 {
			break
		}
		cursor = nextCursor
	}

	// Delete in batches of 1000
	for i := 0; i < len(allIDs); i += 1000 {
		end := i + 1000
		if end > len(allIDs) {
			end = len(allIDs)
		}

		batch := allIDs[i:end]
		pipe := s.client.Pipeline()
		tokenCmds := make([]*goredis.StringCmd, len(batch))

		for j, id := range batch {
			// Get the token hash before deletion
			tokenCmds[j] = pipe.HGet(ctx, fmt.Sprintf("auth:id:%s", id), "token_hash")
			// Remove from method set if exists
			if req.Method != nil {
				pipe.SRem(ctx, fmt.Sprintf("auth:method:%s", req.Method.String()), id)
			}
			// Remove from all set
			pipe.SRem(ctx, "auth:all", id)
			// Delete the auth hash
			pipe.Del(ctx, fmt.Sprintf("auth:id:%s", id))
		}

		if _, err := pipe.Exec(ctx); err != nil {
			return fmt.Errorf("deleting authentications: %w", err)
		}

		pipe = s.client.Pipeline()
		for j := range batch {
			if tokenHash, err := tokenCmds[j].Result(); err == nil {
				pipe.Del(ctx, fmt.Sprintf("auth:token:%s", tokenHash))
			}
		}

		if _, err := pipe.Exec(ctx); err != nil {
			return fmt.Errorf("deleting token mappings: %w", err)
		}
	}

	return nil
}

// ExpireAuthenticationByID implements authn.Store.
func (s *Store) ExpireAuthenticationByID(ctx context.Context, id string, expireAt *timestamppb.Timestamp) error {
	// Get the token hash first
	idKey := fmt.Sprintf("auth:id:%s", id)
	tokenHash, err := s.client.HGet(ctx, idKey, "token_hash").Result()
	if err != nil {
		if err == goredis.Nil {
			return errors.ErrNotFoundf("getting authentication by id")
		}
		return fmt.Errorf("getting authentication by id: %w", err)
	}

	// Update expiry in a pipeline
	pipe := s.client.Pipeline()

	// Update expiry in hash
	pipe.HSet(ctx, idKey, "expires_at", expireAt.AsTime().UnixNano())

	// Set TTL on both ID and token hash keys
	pipe.ExpireAt(ctx, idKey, expireAt.AsTime())
	pipe.ExpireAt(ctx, fmt.Sprintf("auth:token:%s", tokenHash), expireAt.AsTime())

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
	tokenKey := fmt.Sprintf("auth:token:%s", hashedToken)
	id, err := s.client.Get(ctx, tokenKey).Result()
	if err != nil {
		if err == goredis.Nil {
			return nil, errors.ErrNotFoundf("getting authentication by token")
		}
		return nil, fmt.Errorf("getting authentication by token: %w", err)
	}

	return s.GetAuthenticationByID(ctx, id)
}

// GetAuthenticationByID implements authn.Store.
func (s *Store) GetAuthenticationByID(ctx context.Context, id string) (*auth.Authentication, error) {
	idKey := fmt.Sprintf("auth:id:%s", id)
	result, err := s.client.HGetAll(ctx, idKey).Result()
	if err != nil {
		return nil, fmt.Errorf("getting authentication by id: %w", err)
	}
	if len(result) == 0 {
		return nil, errors.ErrNotFoundf("getting authentication by id")
	}

	auth := &auth.Authentication{}
	if err := protojson.Unmarshal([]byte(result["authentication"]), auth); err != nil {
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
	var key string
	if req.Predicate.Method != nil {
		key = fmt.Sprintf("auth:method:%s", req.Predicate.Method.String())
	} else {
		key = "auth:all"
	}

	// Scan the set with cursor pagination
	ids, nextCursor, err := s.client.SScan(ctx, key, cursor, "*", int64(req.QueryParams.Limit)).Result()
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
		cmds[id] = pipe.HGetAll(ctx, fmt.Sprintf("auth:id:%s", id))
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
		if err := protojson.Unmarshal([]byte(result["authentication"]), auth); err != nil {
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
