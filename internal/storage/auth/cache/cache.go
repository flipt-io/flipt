package cache

import (
	"context"
	"fmt"

	"go.flipt.io/flipt/internal/cache"
	"go.flipt.io/flipt/internal/storage/auth"
	authrpc "go.flipt.io/flipt/rpc/flipt/auth"
	"go.uber.org/zap"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
)

type Store struct {
	auth.Store
	cacher cache.Cacher
	logger *zap.Logger
}

// storage:auth:token:<token>
// nolint:gosec
const authTokenCacheKeyFmt = "s:a:t:%s"

func NewStore(store auth.Store, cacher cache.Cacher, logger *zap.Logger) *Store {
	return &Store{
		Store:  store,
		cacher: cacher,
		logger: logger,
	}
}

func (s *Store) setCache(ctx context.Context, key string, value protoreflect.ProtoMessage) {
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

func (s *Store) getCache(ctx context.Context, key string, value protoreflect.ProtoMessage) bool {
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

func (s *Store) GetAuthenticationByClientToken(ctx context.Context, token string) (*authrpc.Authentication, error) {
	var (
		cacheKey = fmt.Sprintf(authTokenCacheKeyFmt, token)
		auth     = &authrpc.Authentication{}
	)

	cacheHit := s.getCache(ctx, cacheKey, auth)
	if cacheHit {
		s.logger.Debug("auth client token storage cache hit")
		return auth, nil
	}

	auth, err := s.Store.GetAuthenticationByClientToken(ctx, token)
	if err != nil {
		return nil, err
	}

	s.setCache(ctx, cacheKey, auth)

	return auth, nil
}
