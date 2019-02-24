package server

import (
	"context"
	"database/sql"

	sq "github.com/Masterminds/squirrel"
	pb "github.com/markphelps/flipt/proto"
	"github.com/markphelps/flipt/storage"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var _ pb.FliptServer = &Server{}

type Option func(s *Server)

type Server struct {
	logger logrus.FieldLogger
	storage.FlagStore
	storage.SegmentStore
	storage.RuleStore
}

// New creates a new Server
func New(logger logrus.FieldLogger, db *sql.DB, opts ...Option) *Server {
	var (
		builder = sq.StatementBuilder.RunWith(sq.NewStmtCacher(db))
		tx      = sq.NewStmtCacheProxy(db)
		s       = &Server{
			logger:       logger,
			FlagStore:    storage.NewFlagStorage(logger, builder),
			SegmentStore: storage.NewSegmentStorage(logger, builder),
			RuleStore:    storage.NewRuleStorage(logger, tx, builder),
		}
	)

	for _, opt := range opts {
		opt(s)
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
