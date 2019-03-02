package server

import (
	"context"
	"database/sql"

	sq "github.com/Masterminds/squirrel"
	lru "github.com/hashicorp/golang-lru"
	pb "github.com/markphelps/flipt/proto"
	"github.com/markphelps/flipt/storage"
	"github.com/markphelps/flipt/storage/cache"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var _ pb.FliptServer = &Server{}

// Option is a server option
type Option func(s *Server)

// WithCacheSize sets the cache size for the server
func WithCacheSize(size int) Option {
	return func(s *Server) {
		s.cacheSize = size
	}
}

// Server serves the Flipt backend
type Server struct {
	cacheSize int

	logger logrus.FieldLogger
	storage.FlagStore
	storage.SegmentStore
	storage.RuleStore
}

// New creates a new Server
func New(logger logrus.FieldLogger, db *sql.DB, opts ...Option) *Server {
	var (
		builder = sq.StatementBuilder.RunWith(sq.NewStmtCacher(db))

		flagStore    = storage.NewFlagStorage(logger, builder)
		segmentStore = storage.NewSegmentStorage(logger, builder)
		ruleStore    = storage.NewRuleStorage(logger, sq.NewStmtCacheProxy(db), builder)

		s = &Server{
			logger:       logger,
			FlagStore:    flagStore,
			SegmentStore: segmentStore,
			RuleStore:    ruleStore,
		}
	)

	for _, opt := range opts {
		opt(s)
	}

	if s.cacheSize > 0 {
		lru, _ := lru.New(s.cacheSize)

		// wrap flagStore with lru cache
		s.FlagStore = cache.NewFlagCache(logger, lru, flagStore)
	}

	return s
}

// ErrorUnaryInterceptor intercepts known errors and returns the appropriate GRPC status code
func (s *Server) ErrorUnaryInterceptor(ctx context.Context, req interface{}, _ *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
	resp, err = handler(ctx, req)
	if err == nil {
		return
	}

	switch err.(type) {
	case storage.ErrNotFound:
		err = status.Error(codes.NotFound, err.Error())
	case storage.ErrInvalid:
		err = status.Error(codes.InvalidArgument, err.Error())
	case ErrInvalidField:
		err = status.Error(codes.InvalidArgument, err.Error())
	default:
		err = status.Error(codes.Internal, err.Error())
	}
	return
}
