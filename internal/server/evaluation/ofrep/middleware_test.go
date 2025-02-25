package ofrep

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap/zaptest"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/structpb"
)

func TestErrorHandler(t *testing.T) {
	h := ErrorHandler(zaptest.NewLogger(t))

	tests := []struct {
		input          error
		expectedCode   int
		expectedOutput string
	}{
		{
			input:          newFlagMissingError(),
			expectedCode:   http.StatusBadRequest,
			expectedOutput: `{"errorCode":"INVALID_CONTEXT","errorDetails":"flag key was not provided"}`,
		},
		{
			input:          newFlagNotFoundError("flag1"),
			expectedCode:   http.StatusNotFound,
			expectedOutput: `{"key":"flag1","errorCode":"FLAG_NOT_FOUND","errorDetails":"flag was not found flag1"}`,
		},
		{
			input:          newBadRequestError("flag1", errors.New("custom failure")),
			expectedCode:   http.StatusBadRequest,
			expectedOutput: `{"key":"flag1","errorCode":"INVALID_CONTEXT","errorDetails":"custom failure"}`,
		},
		{
			input:          errors.New("general failure"),
			expectedCode:   http.StatusInternalServerError,
			expectedOutput: `{"errorCode":"GENERAL","errorDetails":"general failure"}`,
		},
		{
			input:          status.Error(codes.Unauthenticated, "unauthenticated"),
			expectedCode:   http.StatusUnauthorized,
			expectedOutput: `{"errorCode":"GENERAL","errorDetails":"unauthenticated"}`,
		},
		{
			input:          status.Error(codes.PermissionDenied, "unauthorized"),
			expectedCode:   http.StatusForbidden,
			expectedOutput: `{"errorCode":"GENERAL","errorDetails":"unauthorized"}`,
		},
	}
	for _, tt := range tests {
		t.Run("error", func(t *testing.T) {
			resp := httptest.NewRecorder()
			h(context.Background(), nil, &runtime.JSONPb{}, resp, nil, tt.input)
			assert.Equal(t, tt.expectedCode, resp.Code)
			assert.Equal(t, tt.expectedOutput+"\n", resp.Body.String())
		})
	}
}

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
