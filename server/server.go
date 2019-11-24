package server

import (
	"context"
	"database/sql"

	pb "github.com/markphelps/flipt/rpc"
	"github.com/markphelps/flipt/storage"
	"github.com/markphelps/flipt/storage/cache"
	"github.com/markphelps/flipt/storage/db"

	sq "github.com/Masterminds/squirrel"
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
	storage.Evaluator
}

// New creates a new Server
func New(logger logrus.FieldLogger, builder sq.StatementBuilderType, sql *sql.DB, opts ...Option) *Server {
	var (
		flagStore    = db.NewFlagStore(logger, builder)
		segmentStore = db.NewSegmentStore(logger, builder)
		ruleStore    = db.NewRuleStore(logger, builder, sql)
		evaluator    = db.NewEvaluator(logger, builder)

		s = &Server{
			logger:       logger,
			FlagStore:    flagStore,
			SegmentStore: segmentStore,
			RuleStore:    ruleStore,
			Evaluator:    evaluator,
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

	errorsTotal.Inc()

	switch err.(type) {
	case storage.ErrNotFound:
		err = status.Error(codes.NotFound, err.Error())
	case storage.ErrInvalid:
		err = status.Error(codes.InvalidArgument, err.Error())
	case errInvalidField:
		err = status.Error(codes.InvalidArgument, err.Error())
	default:
		err = status.Error(codes.Internal, err.Error())
	}

	return
}
