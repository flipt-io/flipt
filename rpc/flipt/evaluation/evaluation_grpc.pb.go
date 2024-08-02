// Code generated by protoc-gen-go-grpc. DO NOT EDIT.
// versions:
// - protoc-gen-go-grpc v1.5.1
// - protoc             (unknown)
// source: evaluation/evaluation.proto

package evaluation

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
	EvaluationService_Boolean_FullMethodName = "/flipt.evaluation.EvaluationService/Boolean"
	EvaluationService_Variant_FullMethodName = "/flipt.evaluation.EvaluationService/Variant"
	EvaluationService_Batch_FullMethodName   = "/flipt.evaluation.EvaluationService/Batch"
)

// EvaluationServiceClient is the client API for EvaluationService service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type EvaluationServiceClient interface {
	Boolean(ctx context.Context, in *EvaluationRequest, opts ...grpc.CallOption) (*BooleanEvaluationResponse, error)
	Variant(ctx context.Context, in *EvaluationRequest, opts ...grpc.CallOption) (*VariantEvaluationResponse, error)
	Batch(ctx context.Context, in *BatchEvaluationRequest, opts ...grpc.CallOption) (*BatchEvaluationResponse, error)
}

type evaluationServiceClient struct {
	cc grpc.ClientConnInterface
}

func NewEvaluationServiceClient(cc grpc.ClientConnInterface) EvaluationServiceClient {
	return &evaluationServiceClient{cc}
}

func (c *evaluationServiceClient) Boolean(ctx context.Context, in *EvaluationRequest, opts ...grpc.CallOption) (*BooleanEvaluationResponse, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(BooleanEvaluationResponse)
	err := c.cc.Invoke(ctx, EvaluationService_Boolean_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *evaluationServiceClient) Variant(ctx context.Context, in *EvaluationRequest, opts ...grpc.CallOption) (*VariantEvaluationResponse, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(VariantEvaluationResponse)
	err := c.cc.Invoke(ctx, EvaluationService_Variant_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *evaluationServiceClient) Batch(ctx context.Context, in *BatchEvaluationRequest, opts ...grpc.CallOption) (*BatchEvaluationResponse, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(BatchEvaluationResponse)
	err := c.cc.Invoke(ctx, EvaluationService_Batch_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// EvaluationServiceServer is the server API for EvaluationService service.
// All implementations must embed UnimplementedEvaluationServiceServer
// for forward compatibility.
type EvaluationServiceServer interface {
	Boolean(context.Context, *EvaluationRequest) (*BooleanEvaluationResponse, error)
	Variant(context.Context, *EvaluationRequest) (*VariantEvaluationResponse, error)
	Batch(context.Context, *BatchEvaluationRequest) (*BatchEvaluationResponse, error)
	mustEmbedUnimplementedEvaluationServiceServer()
}

// UnimplementedEvaluationServiceServer must be embedded to have
// forward compatible implementations.
//
// NOTE: this should be embedded by value instead of pointer to avoid a nil
// pointer dereference when methods are called.
type UnimplementedEvaluationServiceServer struct{}

func (UnimplementedEvaluationServiceServer) Boolean(context.Context, *EvaluationRequest) (*BooleanEvaluationResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Boolean not implemented")
}
func (UnimplementedEvaluationServiceServer) Variant(context.Context, *EvaluationRequest) (*VariantEvaluationResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Variant not implemented")
}
func (UnimplementedEvaluationServiceServer) Batch(context.Context, *BatchEvaluationRequest) (*BatchEvaluationResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Batch not implemented")
}
func (UnimplementedEvaluationServiceServer) mustEmbedUnimplementedEvaluationServiceServer() {}
func (UnimplementedEvaluationServiceServer) testEmbeddedByValue()                           {}

// UnsafeEvaluationServiceServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to EvaluationServiceServer will
// result in compilation errors.
type UnsafeEvaluationServiceServer interface {
	mustEmbedUnimplementedEvaluationServiceServer()
}

func RegisterEvaluationServiceServer(s grpc.ServiceRegistrar, srv EvaluationServiceServer) {
	// If the following call pancis, it indicates UnimplementedEvaluationServiceServer was
	// embedded by pointer and is nil.  This will cause panics if an
	// unimplemented method is ever invoked, so we test this at initialization
	// time to prevent it from happening at runtime later due to I/O.
	if t, ok := srv.(interface{ testEmbeddedByValue() }); ok {
		t.testEmbeddedByValue()
	}
	s.RegisterService(&EvaluationService_ServiceDesc, srv)
}

func _EvaluationService_Boolean_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(EvaluationRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(EvaluationServiceServer).Boolean(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: EvaluationService_Boolean_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(EvaluationServiceServer).Boolean(ctx, req.(*EvaluationRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _EvaluationService_Variant_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(EvaluationRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(EvaluationServiceServer).Variant(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: EvaluationService_Variant_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(EvaluationServiceServer).Variant(ctx, req.(*EvaluationRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _EvaluationService_Batch_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(BatchEvaluationRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(EvaluationServiceServer).Batch(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: EvaluationService_Batch_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(EvaluationServiceServer).Batch(ctx, req.(*BatchEvaluationRequest))
	}
	return interceptor(ctx, in, info, handler)
}

// EvaluationService_ServiceDesc is the grpc.ServiceDesc for EvaluationService service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var EvaluationService_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "flipt.evaluation.EvaluationService",
	HandlerType: (*EvaluationServiceServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "Boolean",
			Handler:    _EvaluationService_Boolean_Handler,
		},
		{
			MethodName: "Variant",
			Handler:    _EvaluationService_Variant_Handler,
		},
		{
			MethodName: "Batch",
			Handler:    _EvaluationService_Batch_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "evaluation/evaluation.proto",
}

const (
	DataService_EvaluationSnapshotNamespace_FullMethodName = "/flipt.evaluation.DataService/EvaluationSnapshotNamespace"
)

// DataServiceClient is the client API for DataService service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
//
// flipt:sdk:ignore
type DataServiceClient interface {
	EvaluationSnapshotNamespace(ctx context.Context, in *EvaluationNamespaceSnapshotRequest, opts ...grpc.CallOption) (*EvaluationNamespaceSnapshot, error)
}

type dataServiceClient struct {
	cc grpc.ClientConnInterface
}

func NewDataServiceClient(cc grpc.ClientConnInterface) DataServiceClient {
	return &dataServiceClient{cc}
}

func (c *dataServiceClient) EvaluationSnapshotNamespace(ctx context.Context, in *EvaluationNamespaceSnapshotRequest, opts ...grpc.CallOption) (*EvaluationNamespaceSnapshot, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(EvaluationNamespaceSnapshot)
	err := c.cc.Invoke(ctx, DataService_EvaluationSnapshotNamespace_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// DataServiceServer is the server API for DataService service.
// All implementations must embed UnimplementedDataServiceServer
// for forward compatibility.
//
// flipt:sdk:ignore
type DataServiceServer interface {
	EvaluationSnapshotNamespace(context.Context, *EvaluationNamespaceSnapshotRequest) (*EvaluationNamespaceSnapshot, error)
	mustEmbedUnimplementedDataServiceServer()
}

// UnimplementedDataServiceServer must be embedded to have
// forward compatible implementations.
//
// NOTE: this should be embedded by value instead of pointer to avoid a nil
// pointer dereference when methods are called.
type UnimplementedDataServiceServer struct{}

func (UnimplementedDataServiceServer) EvaluationSnapshotNamespace(context.Context, *EvaluationNamespaceSnapshotRequest) (*EvaluationNamespaceSnapshot, error) {
	return nil, status.Errorf(codes.Unimplemented, "method EvaluationSnapshotNamespace not implemented")
}
func (UnimplementedDataServiceServer) mustEmbedUnimplementedDataServiceServer() {}
func (UnimplementedDataServiceServer) testEmbeddedByValue()                     {}

// UnsafeDataServiceServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to DataServiceServer will
// result in compilation errors.
type UnsafeDataServiceServer interface {
	mustEmbedUnimplementedDataServiceServer()
}

func RegisterDataServiceServer(s grpc.ServiceRegistrar, srv DataServiceServer) {
	// If the following call pancis, it indicates UnimplementedDataServiceServer was
	// embedded by pointer and is nil.  This will cause panics if an
	// unimplemented method is ever invoked, so we test this at initialization
	// time to prevent it from happening at runtime later due to I/O.
	if t, ok := srv.(interface{ testEmbeddedByValue() }); ok {
		t.testEmbeddedByValue()
	}
	s.RegisterService(&DataService_ServiceDesc, srv)
}

func _DataService_EvaluationSnapshotNamespace_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(EvaluationNamespaceSnapshotRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(DataServiceServer).EvaluationSnapshotNamespace(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: DataService_EvaluationSnapshotNamespace_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(DataServiceServer).EvaluationSnapshotNamespace(ctx, req.(*EvaluationNamespaceSnapshotRequest))
	}
	return interceptor(ctx, in, info, handler)
}

// DataService_ServiceDesc is the grpc.ServiceDesc for DataService service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var DataService_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "flipt.evaluation.DataService",
	HandlerType: (*DataServiceServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "EvaluationSnapshotNamespace",
			Handler:    _DataService_EvaluationSnapshotNamespace_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "evaluation/evaluation.proto",
}
