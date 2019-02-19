package server

import (
	"context"
	"database/sql"

	pb "github.com/markphelps/flipt/proto"
	"github.com/markphelps/flipt/storage"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var _ pb.FliptServer = &Server{}

type Server struct {
	logger logrus.FieldLogger
	storage.FlagRepository
	storage.SegmentRepository
	storage.RuleRepository
}

// New creates a new Server
func New(logger logrus.FieldLogger, db *sql.DB) (*Server, error) {
	return &Server{
		logger:            logger,
		FlagRepository:    storage.NewFlagStorage(logger, db),
		SegmentRepository: storage.NewSegmentStorage(logger, db),
		RuleRepository:    storage.NewRuleStorage(logger, db),
	}, nil
}

// ErrorInterceptor intercepts known errors and returns the appropriate GRPC status code
func ErrorInterceptor(ctx context.Context, req interface{}, _ *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
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
