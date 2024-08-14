package ofrep

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap/zaptest"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestErrorHandler(t *testing.T) {
	h := ErrorHandler(zaptest.NewLogger(t))

	tests := []struct {
		input          error
		expectedCode   int
		expectedOutput string
	}{
		{
			input:          newFlagsMissingError(),
			expectedCode:   http.StatusBadRequest,
			expectedOutput: `{"errorCode":"INVALID_CONTEXT","errorDetails":"flags were not provided in context"}`,
		},
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
