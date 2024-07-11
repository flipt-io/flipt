package ofrep

import (
	"go.flipt.io/flipt/internal/config"

	"go.flipt.io/flipt/rpc/flipt/ofrep"
	"google.golang.org/grpc"
)

// Server servers the methods used by the OpenFeature Remote Evaluation Protocol.
// It will be used only with gRPC Gateway as there's no specification for gRPC itself.
type Server struct {
	cacheCfg config.CacheConfig
	ofrep.UnimplementedOFREPServiceServer
}

// New constructs a new Server.
func New(cacheCfg config.CacheConfig) *Server {
	return &Server{
		cacheCfg: cacheCfg,
	}
}

// RegisterGRPC registers the EvaluateServer onto the provided gRPC Server.
func (s *Server) RegisterGRPC(server *grpc.Server) {
	ofrep.RegisterOFREPServiceServer(server, s)
}
