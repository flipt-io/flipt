package server

import (
	"context"
	"errors"
	"testing"

	"github.com/markphelps/flipt/storage"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestErrorUnaryInterceptor(t *testing.T) {
	tests := []struct {
		name     string
		wantErr  error
		wantCode codes.Code
	}{
		{
			name:     "storage not found error",
			wantErr:  storage.ErrNotFound("foo"),
			wantCode: codes.NotFound,
		},
		{
			name:     "storage invalid error",
			wantErr:  storage.ErrInvalid("foo"),
			wantCode: codes.InvalidArgument,
		},
		{
			name:     "server invalid field",
			wantErr:  invalidFieldError("bar", "is wrong"),
			wantCode: codes.InvalidArgument,
		},
		{
			name:     "server empty field",
			wantErr:  emptyFieldError("bar"),
			wantCode: codes.InvalidArgument,
		},
		{
			name:     "other error",
			wantErr:  errors.New("foo"),
			wantCode: codes.Internal,
		},
		{
			name: "no error",
		},
	}

	for _, tt := range tests {
		var (
			wantErr  = tt.wantErr
			wantCode = tt.wantCode
		)

		t.Run(tt.name, func(t *testing.T) {
			var (
				subject = &Server{}

				spyHandler = grpc.UnaryHandler(func(ctx context.Context, req interface{}) (interface{}, error) {
					return nil, wantErr
				})
			)

			_, err := subject.ErrorUnaryInterceptor(context.Background(), nil, nil, spyHandler)
			if wantErr != nil {
				require.Error(t, err)
				status := status.Convert(err)
				assert.Equal(t, wantCode, status.Code())
				return
			}

			require.NoError(t, err)
		})
	}
}
