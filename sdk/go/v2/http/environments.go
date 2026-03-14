package http

import (
	context "context"

	environments "go.flipt.io/flipt/rpc/v2/environments"
	grpc "google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
)

func (x *EnvironmentsServiceClient) BranchEnvironment(ctx context.Context, v *environments.BranchEnvironmentRequest, _ ...grpc.CallOption) (*environments.Environment, error) {
	return nil, status.Errorf(codes.Unimplemented, "method BranchEnvironment not implemented")
}

func (x *EnvironmentsServiceClient) DeleteBranchEnvironment(ctx context.Context, v *environments.DeleteBranchEnvironmentRequest, _ ...grpc.CallOption) (*emptypb.Empty, error) {
	return nil, status.Errorf(codes.Unimplemented, "method DeleteBranchEnvironment not implemented")
}

func (x *EnvironmentsServiceClient) ListEnvironmentBranches(ctx context.Context, v *environments.ListEnvironmentBranchesRequest, _ ...grpc.CallOption) (*environments.ListEnvironmentBranchesResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method ListEnvironmentBranches not implemented")
}

func (x *EnvironmentsServiceClient) ProposeEnvironment(ctx context.Context, v *environments.ProposeEnvironmentRequest, _ ...grpc.CallOption) (*environments.EnvironmentProposalDetails, error) {
	return nil, status.Errorf(codes.Unimplemented, "method ProposeEnvironment not implemented")
}

func (x *EnvironmentsServiceClient) ListBranchedEnvironmentChanges(ctx context.Context, v *environments.ListBranchedEnvironmentChangesRequest, _ ...grpc.CallOption) (*environments.ListBranchedEnvironmentChangesResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method ListBranchedEnvironmentChanges not implemented")
}
