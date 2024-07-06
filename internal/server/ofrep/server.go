package ofrep

import (
	"go.flipt.io/flipt/internal/config"

	"go.flipt.io/flipt/rpc/flipt/ofrep"
	"google.golang.org/grpc"
)

type Server struct {
	cacheCfg config.CacheConfig
	ofrep.UnimplementedOFREPServiceServer
}

func New(cacheCfg config.CacheConfig) *Server {
	return &Server{
		cacheCfg: cacheCfg,
	}
}

func (s *Server) RegisterGRPC(server *grpc.Server) {
	ofrep.RegisterOFREPServiceServer(server, s)
}
