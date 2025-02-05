// Code generated by protoc-gen-go-grpc. DO NOT EDIT.
// versions:
// - protoc-gen-go-grpc v1.5.1
// - protoc             (unknown)
// source: meta/meta.proto

package meta

import (
	context "context"
	httpbody "google.golang.org/genproto/googleapis/api/httpbody"
	grpc "google.golang.org/grpc"
	codes "google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"
	emptypb "google.golang.org/protobuf/types/known/emptypb"
)

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
// Requires gRPC-Go v1.64.0 or later.
const _ = grpc.SupportPackageIsVersion9

const (
	MetadataService_GetInfo_FullMethodName = "/flipt.meta.MetadataService/GetInfo"
)

// MetadataServiceClient is the client API for MetadataService service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type MetadataServiceClient interface {
	GetInfo(ctx context.Context, in *emptypb.Empty, opts ...grpc.CallOption) (*httpbody.HttpBody, error)
}

type metadataServiceClient struct {
	cc grpc.ClientConnInterface
}

func NewMetadataServiceClient(cc grpc.ClientConnInterface) MetadataServiceClient {
	return &metadataServiceClient{cc}
}

func (c *metadataServiceClient) GetInfo(ctx context.Context, in *emptypb.Empty, opts ...grpc.CallOption) (*httpbody.HttpBody, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(httpbody.HttpBody)
	err := c.cc.Invoke(ctx, MetadataService_GetInfo_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// MetadataServiceServer is the server API for MetadataService service.
// All implementations must embed UnimplementedMetadataServiceServer
// for forward compatibility.
type MetadataServiceServer interface {
	GetInfo(context.Context, *emptypb.Empty) (*httpbody.HttpBody, error)
	mustEmbedUnimplementedMetadataServiceServer()
}

// UnimplementedMetadataServiceServer must be embedded to have
// forward compatible implementations.
//
// NOTE: this should be embedded by value instead of pointer to avoid a nil
// pointer dereference when methods are called.
type UnimplementedMetadataServiceServer struct{}

func (UnimplementedMetadataServiceServer) GetInfo(context.Context, *emptypb.Empty) (*httpbody.HttpBody, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetInfo not implemented")
}
func (UnimplementedMetadataServiceServer) mustEmbedUnimplementedMetadataServiceServer() {}
func (UnimplementedMetadataServiceServer) testEmbeddedByValue()                         {}

// UnsafeMetadataServiceServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to MetadataServiceServer will
// result in compilation errors.
type UnsafeMetadataServiceServer interface {
	mustEmbedUnimplementedMetadataServiceServer()
}

func RegisterMetadataServiceServer(s grpc.ServiceRegistrar, srv MetadataServiceServer) {
	// If the following call pancis, it indicates UnimplementedMetadataServiceServer was
	// embedded by pointer and is nil.  This will cause panics if an
	// unimplemented method is ever invoked, so we test this at initialization
	// time to prevent it from happening at runtime later due to I/O.
	if t, ok := srv.(interface{ testEmbeddedByValue() }); ok {
		t.testEmbeddedByValue()
	}
	s.RegisterService(&MetadataService_ServiceDesc, srv)
}

func _MetadataService_GetInfo_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(emptypb.Empty)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(MetadataServiceServer).GetInfo(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: MetadataService_GetInfo_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(MetadataServiceServer).GetInfo(ctx, req.(*emptypb.Empty))
	}
	return interceptor(ctx, in, info, handler)
}

// MetadataService_ServiceDesc is the grpc.ServiceDesc for MetadataService service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var MetadataService_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "flipt.meta.MetadataService",
	HandlerType: (*MetadataServiceServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "GetInfo",
			Handler:    _MetadataService_GetInfo_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "meta/meta.proto",
}
