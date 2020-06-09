package server

import (
	"context"

	"github.com/markphelps/flipt/errors"
	flipt "github.com/markphelps/flipt/rpc"
	"github.com/markphelps/flipt/storage"

	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var _ flipt.FliptServer = &Server{}

// Server serves the Flipt backend
type Server struct {
	logger logrus.FieldLogger

	storage.Store
}

// New creates a new Server
func New(logger logrus.FieldLogger, store storage.Store) *Server {
	var (
		s = &Server{
			logger: logger,
			Store:  store,
		}
	)

	return s
}

// ValidationUnaryInterceptor validates incomming requests
func (s *Server) ValidationUnaryInterceptor(ctx context.Context, req interface{}, _ *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
	if v, ok := req.(flipt.Validator); ok {
		if err := v.Validate(); err != nil {
			return nil, err
		}
	}

	return handler(ctx, req)
}

// ErrorUnaryInterceptor intercepts known errors and returns the appropriate GRPC status code
func (s *Server) ErrorUnaryInterceptor(ctx context.Context, req interface{}, _ *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
	resp, err = handler(ctx, req)
	if err == nil {
		return
	}

	errorsTotal.Inc()

	switch err.(type) {
	case errors.ErrNotFound:
		err = status.Error(codes.NotFound, err.Error())
	case errors.ErrInvalid:
		err = status.Error(codes.InvalidArgument, err.Error())
	case errors.ErrValidation:
		err = status.Error(codes.InvalidArgument, err.Error())
	default:
		err = status.Error(codes.Internal, err.Error())
	}

	return
}
