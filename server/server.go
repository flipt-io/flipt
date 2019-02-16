package server

import (
	"context"
	"database/sql"

	"github.com/markphelps/flipt"
	"github.com/markphelps/flipt/storage"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var _ flipt.FliptServer = &Server{}

type Server struct {
	logger logrus.FieldLogger
	flipt.FlagService
	flipt.SegmentService
	flipt.RuleService
}

// New creates a new Server
func New(logger logrus.FieldLogger, db *sql.DB) (*Server, error) {
	return &Server{
		logger:         logger,
		FlagService:    storage.NewFlagService(logger, db),
		SegmentService: storage.NewSegmentService(logger, db),
		RuleService:    storage.NewRuleService(logger, db),
	}, nil
}

// ErrorInterceptor intercepts known errors and returns the appropriate GRPC status code
func ErrorInterceptor(ctx context.Context, req interface{}, _ *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
	resp, err = handler(ctx, req)
	if err == nil {
		return
	}

	switch err.(type) {
	case flipt.ErrNotFound:
		err = status.Error(codes.NotFound, err.Error())
	case flipt.ErrInvalid:
		err = status.Error(codes.InvalidArgument, err.Error())
	case flipt.ErrInvalidField:
		err = status.Error(codes.InvalidArgument, err.Error())
	default:
		err = status.Error(codes.Internal, err.Error())
	}
	return
}
