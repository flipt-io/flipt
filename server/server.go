package server

import (
	"context"
	"database/sql"

	sq "github.com/Masterminds/squirrel"
	pb "github.com/markphelps/flipt/rpc"
	"github.com/markphelps/flipt/storage"
	"github.com/markphelps/flipt/storage/cache"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var _ pb.FliptServer = &Server{}

// Server serves the Flipt backend
type Server struct {
	logger logrus.FieldLogger
	cache  cache.Cacher
	storage.FlagStore
	storage.SegmentStore
	storage.RuleStore
}

// New creates a new Server
func New(logger logrus.FieldLogger, db *sql.DB, driver storage.Driver, opts ...Option) *Server {

	// TODO: clean this up with a different abstraction
	var (
		builder sq.StatementBuilderType
		cacher  = sq.NewStmtCacher(db)
	)

	switch driver {
	case storage.SQLite:
		builder = sq.StatementBuilder.RunWith(cacher)
	case storage.Postgres:
		builder = sq.StatementBuilder.PlaceholderFormat(sq.Dollar).RunWith(cacher)
	}

	var (
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

	if s.cache != nil {
		// wrap flagStore with lru cache
		s.FlagStore = cache.NewFlagCache(logger, s.cache, flagStore)
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
