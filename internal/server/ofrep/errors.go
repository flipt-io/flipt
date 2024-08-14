package ofrep

import (
	"fmt"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/structpb"
)

const statusFlagKeyPointer = "ofrep-flag-key"

func statusWithKey(st *status.Status, key string) (*status.Status, error) {
	return st.WithDetails(&structpb.Struct{
		Fields: map[string]*structpb.Value{
			statusFlagKeyPointer: structpb.NewStringValue(key),
		},
	})
}

func newBadRequestError(key string, err error) error {
	v := status.New(codes.InvalidArgument, err.Error())
	v, derr := statusWithKey(v, key)
	if derr != nil {
		return status.Errorf(codes.Internal, "failed to encode not bad request error")
	}
	return v.Err()
}

func newFlagNotFoundError(key string) error {
	v := status.New(codes.NotFound, fmt.Sprintf("flag was not found %s", key))
	v, derr := statusWithKey(v, key)
	if derr != nil {
		return status.Errorf(codes.Internal, "failed to encode not found error")
	}
	return v.Err()
}

func newFlagMissingError() error {
	return status.Error(codes.InvalidArgument, "flag key was not provided")
}

func newFlagsMissingError() error {
	return status.Error(codes.InvalidArgument, "flags were not provided in context")
}
