package server

import (
	"context"
	"errors"

	errs "github.com/markphelps/flipt/errors"
	flipt "github.com/markphelps/flipt/rpc/flipt"
	"github.com/markphelps/flipt/storage"

	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var _ flipt.FliptServer = &Server{}

type Option func(s *Server)

// Server serves the Flipt backend
type Server struct {
	logger logrus.FieldLogger
	store  storage.Store
	flipt.UnimplementedFliptServer
}

// New creates a new Server
func New(logger logrus.FieldLogger, store storage.Store, opts ...Option) *Server {
	var (
		s = &Server{
			logger: logger,
			store:  store,
		}
	)

	for _, fn := range opts {
		fn(s)
	}

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
		return resp, nil
	}

	errorsTotal.Inc()

	var errnf errs.ErrNotFound
	if errors.As(err, &errnf) {
		err = status.Error(codes.NotFound, err.Error())
		return
	}

	var errin errs.ErrInvalid
	if errors.As(err, &errin) {
		err = status.Error(codes.InvalidArgument, err.Error())
		return
	}

	var errv errs.ErrValidation
	if errors.As(err, &errv) {
		err = status.Error(codes.InvalidArgument, err.Error())
		return
	}

	err = status.Error(codes.Internal, err.Error())
	return
}
