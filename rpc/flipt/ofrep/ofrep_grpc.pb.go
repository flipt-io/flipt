// Code generated by protoc-gen-go-grpc. DO NOT EDIT.
// versions:
// - protoc-gen-go-grpc v1.5.1
// - protoc             (unknown)
// source: ofrep/ofrep.proto

package ofrep

import (
	context "context"
	grpc "google.golang.org/grpc"
	codes "google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"
)

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
// Requires gRPC-Go v1.64.0 or later.
const _ = grpc.SupportPackageIsVersion9

const (
	OFREPService_EvaluateFlag_FullMethodName = "/flipt.ofrep.OFREPService/EvaluateFlag"
	OFREPService_EvaluateBulk_FullMethodName = "/flipt.ofrep.OFREPService/EvaluateBulk"
)

// OFREPServiceClient is the client API for OFREPService service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type OFREPServiceClient interface {
	// OFREP single flag evaluation
	EvaluateFlag(ctx context.Context, in *EvaluateFlagRequest, opts ...grpc.CallOption) (*EvaluationResponse, error)
	// OFREP bulk flag evaluation
	EvaluateBulk(ctx context.Context, in *EvaluateBulkRequest, opts ...grpc.CallOption) (*BulkEvaluationResponse, error)
}

type oFREPServiceClient struct {
	cc grpc.ClientConnInterface
}

func NewOFREPServiceClient(cc grpc.ClientConnInterface) OFREPServiceClient {
	return &oFREPServiceClient{cc}
}

func (c *oFREPServiceClient) EvaluateFlag(ctx context.Context, in *EvaluateFlagRequest, opts ...grpc.CallOption) (*EvaluationResponse, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(EvaluationResponse)
	err := c.cc.Invoke(ctx, OFREPService_EvaluateFlag_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *oFREPServiceClient) EvaluateBulk(ctx context.Context, in *EvaluateBulkRequest, opts ...grpc.CallOption) (*BulkEvaluationResponse, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(BulkEvaluationResponse)
	err := c.cc.Invoke(ctx, OFREPService_EvaluateBulk_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// OFREPServiceServer is the server API for OFREPService service.
// All implementations must embed UnimplementedOFREPServiceServer
// for forward compatibility.
type OFREPServiceServer interface {
	// OFREP single flag evaluation
	EvaluateFlag(context.Context, *EvaluateFlagRequest) (*EvaluationResponse, error)
	// OFREP bulk flag evaluation
	EvaluateBulk(context.Context, *EvaluateBulkRequest) (*BulkEvaluationResponse, error)
	mustEmbedUnimplementedOFREPServiceServer()
}

// UnimplementedOFREPServiceServer must be embedded to have
// forward compatible implementations.
//
// NOTE: this should be embedded by value instead of pointer to avoid a nil
// pointer dereference when methods are called.
type UnimplementedOFREPServiceServer struct{}

func (UnimplementedOFREPServiceServer) EvaluateFlag(context.Context, *EvaluateFlagRequest) (*EvaluationResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method EvaluateFlag not implemented")
}
func (UnimplementedOFREPServiceServer) EvaluateBulk(context.Context, *EvaluateBulkRequest) (*BulkEvaluationResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method EvaluateBulk not implemented")
}
func (UnimplementedOFREPServiceServer) mustEmbedUnimplementedOFREPServiceServer() {}
func (UnimplementedOFREPServiceServer) testEmbeddedByValue()                      {}

// UnsafeOFREPServiceServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to OFREPServiceServer will
// result in compilation errors.
type UnsafeOFREPServiceServer interface {
	mustEmbedUnimplementedOFREPServiceServer()
}

func RegisterOFREPServiceServer(s grpc.ServiceRegistrar, srv OFREPServiceServer) {
	// If the following call pancis, it indicates UnimplementedOFREPServiceServer was
	// embedded by pointer and is nil.  This will cause panics if an
	// unimplemented method is ever invoked, so we test this at initialization
	// time to prevent it from happening at runtime later due to I/O.
	if t, ok := srv.(interface{ testEmbeddedByValue() }); ok {
		t.testEmbeddedByValue()
	}
	s.RegisterService(&OFREPService_ServiceDesc, srv)
}

func _OFREPService_EvaluateFlag_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(EvaluateFlagRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(OFREPServiceServer).EvaluateFlag(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: OFREPService_EvaluateFlag_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(OFREPServiceServer).EvaluateFlag(ctx, req.(*EvaluateFlagRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _OFREPService_EvaluateBulk_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(EvaluateBulkRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(OFREPServiceServer).EvaluateBulk(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: OFREPService_EvaluateBulk_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(OFREPServiceServer).EvaluateBulk(ctx, req.(*EvaluateBulkRequest))
	}
	return interceptor(ctx, in, info, handler)
}

// OFREPService_ServiceDesc is the grpc.ServiceDesc for OFREPService service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var OFREPService_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "flipt.ofrep.OFREPService",
	HandlerType: (*OFREPServiceServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "EvaluateFlag",
			Handler:    _OFREPService_EvaluateFlag_Handler,
		},
		{
			MethodName: "EvaluateBulk",
			Handler:    _OFREPService_EvaluateBulk_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "ofrep/ofrep.proto",
}
