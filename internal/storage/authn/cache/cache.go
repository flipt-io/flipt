package cache

import (
	"context"
	"fmt"

	"go.flipt.io/flipt/internal/cache"
	"go.flipt.io/flipt/internal/storage/authn"
	authrpc "go.flipt.io/flipt/rpc/flipt/auth"
	"go.uber.org/zap"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type Store struct {
	authn.Store
	cacher cache.Cacher
	logger *zap.Logger
}

const (
	// storage:auth:tokenhash:<tokenhash>
	// nolint:gosec
	authTokenCacheKeyFmt = "s:a:t:%s"
	// storage:auth:id:<id>
	authIDCacheKeyFmt = "s:a:i:%s"
)

func NewStore(store authn.Store, cacher cache.Cacher, logger *zap.Logger) *Store {
	return &Store{
		Store:  store,
		cacher: cacher,
		logger: logger,
	}
}

func (s *Store) set(ctx context.Context, key string, value proto.Message) {
	cachePayload, err := proto.Marshal(value)
	if err != nil {
		s.logger.Error("marshalling for storage cache", zap.Error(err))
		return
	}

	err = s.cacher.Set(ctx, key, cachePayload)
	if err != nil {
		s.logger.Error("setting in storage cache", zap.Error(err))
	}
}

func (s *Store) get(ctx context.Context, key string, value proto.Message) bool {
	cachePayload, cacheHit, err := s.cacher.Get(ctx, key)
	if err != nil {
		s.logger.Error("getting from storage cache", zap.Error(err))
		return false
	} else if !cacheHit {
		return false
	}

	err = proto.Unmarshal(cachePayload, value)
	if err != nil {
		s.logger.Error("unmarshalling from storage cache", zap.Error(err))
		return false
	}

	return true
}

func (s *Store) GetAuthenticationByClientToken(ctx context.Context, clientToken string) (*authrpc.Authentication, error) {
	hashedToken, err := authn.HashClientToken(clientToken)
	if err != nil {
		return nil, fmt.Errorf("getting authentication by token: %w", err)
	}

	var (
		tokenCacheKey = fmt.Sprintf(authTokenCacheKeyFmt, hashedToken)
		auth          = &authrpc.Authentication{}
	)

	cacheHit := s.get(ctx, tokenCacheKey, auth)
	if cacheHit {
		return auth, nil
	}

	auth, err = s.Store.GetAuthenticationByClientToken(ctx, clientToken)
	if err != nil {
		return nil, err
	}

	s.set(ctx, tokenCacheKey, auth)

	// set token by id in cache for expiration
	idCacheKey := fmt.Sprintf(authIDCacheKeyFmt, auth.Id)
	if err := s.cacher.Set(ctx, idCacheKey, []byte(hashedToken)); err != nil {
		s.logger.Error("setting in storage cache", zap.Error(err))
	}

	return auth, nil
}

func (s *Store) ExpireAuthenticationByID(ctx context.Context, id string, expireAt *timestamppb.Timestamp) error {
	idCacheKey := fmt.Sprintf(authIDCacheKeyFmt, id)

	// expire token in db
	err := s.Store.ExpireAuthenticationByID(ctx, id, expireAt)
	if err != nil {
		return err
	}

	// lookup token by id to get token, then delete in cache
	cachePayload, cacheHit, err := s.cacher.Get(ctx, idCacheKey)
	if err == nil && cacheHit {
		tokenCacheKey := fmt.Sprintf(authTokenCacheKeyFmt, string(cachePayload))
		_ = s.cacher.Delete(ctx, tokenCacheKey)
		_ = s.cacher.Delete(ctx, idCacheKey)
	}

	return nil
}
