package metadata

import (
	"context"
	"encoding/json"

	"go.flipt.io/flipt/internal/config"
	"go.flipt.io/flipt/internal/info"
	"go.flipt.io/flipt/rpc/flipt/meta"
	"google.golang.org/genproto/googleapis/api/httpbody"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
	"google.golang.org/protobuf/types/known/emptypb"
)

type Server struct {
	cfg  *config.Config
	info info.Flipt

	meta.UnimplementedMetadataServiceServer
}

func New(cfg *config.Config, info info.Flipt) *Server {
	return &Server{
		cfg:  cfg,
		info: info,
	}
}

// RegisterGRPC registers the server on the provided gRPC server instance.
func (s *Server) RegisterGRPC(server *grpc.Server) {
	meta.RegisterMetadataServiceServer(server, s)
}

// GetConfiguration returns a HttpBody instance containing the Flipt instance's
// configuration structure marshalled as JSON.
func (s *Server) GetConfiguration(ctx context.Context, _ *emptypb.Empty) (*httpbody.HttpBody, error) {
	return response(ctx, s.cfg)
}

// GetInfo returns a HttpBody instance containing the Flipt instance's
// runtime information marshalled as JSON.
func (s *Server) GetInfo(ctx context.Context, _ *emptypb.Empty) (*httpbody.HttpBody, error) {
	return response(ctx, s.info)
}

func (s *Server) SkipsAuthorization(ctx context.Context) bool {
	return true
}

func response(ctx context.Context, v any) (*httpbody.HttpBody, error) {
	data, err := marshal(ctx, v)
	if err != nil {
		return nil, err
	}

	return &httpbody.HttpBody{
		ContentType: "application/json",
		Data:        data,
	}, nil
}

func marshal(ctx context.Context, v any) ([]byte, error) {
	if md, ok := metadata.FromIncomingContext(ctx); ok {
		accept := md.Get("grpcgateway-accept")
		if len(accept) > 0 && accept[0] == "application/json+pretty" {
			return json.MarshalIndent(v, "", "  ")
		}
	}

	return json.Marshal(v)
}
