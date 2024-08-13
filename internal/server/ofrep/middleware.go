package ofrep

import (
	"context"
	"net/http"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/structpb"
)

type errorCode string

const (
	errorCodeFlagNotFound        errorCode = "FLAG_NOT_FOUND"
	errorCodeParseError          errorCode = "PARSE_ERROR"
	errorCodeTargetingKeyMissing errorCode = "TARGETING_KEY_MISSING"
	errorCodeInvalidContext      errorCode = "INVALID_CONTEXT"
	errorCodeParseGeneral        errorCode = "GENERAL"
)

type errorSchema struct {
	Key          string    `json:"key,omitempty"`
	ErrorCode    errorCode `json:"errorCode,omitempty"`
	ErrorDetails string    `json:"errorDetails"`
}

func ErrorHandler(logger *zap.Logger) runtime.ErrorHandlerFunc {
	return func(ctx context.Context, mux *runtime.ServeMux, marshaler runtime.Marshaler, w http.ResponseWriter, r *http.Request, err error) {
		st, ok := status.FromError(err)
		if !ok {
			st = status.New(codes.Unknown, err.Error())
		}
		response := errorSchema{
			ErrorDetails: st.Message(),
		}
		details := st.Details()
		if len(details) == 1 {
			if values, ok := details[0].(*structpb.Struct); ok {
				if key, ok := values.AsMap()[statusFlagKeyPointer]; ok {
					response.Key = key.(string)
				}
			}
		}
		switch st.Code() {
		case codes.InvalidArgument:
			response.ErrorCode = errorCodeInvalidContext
		case codes.NotFound:
			response.ErrorCode = errorCodeFlagNotFound
		default:
			response.ErrorCode = errorCodeParseGeneral
		}
		w.Header().Set("Content-Type", marshaler.ContentType(response))
		w.WriteHeader(runtime.HTTPStatusFromCode(st.Code()))

		eerr := marshaler.NewEncoder(w).Encode(response)
		if eerr != nil {
			logger.Error("failed to encode error response", zap.Error(eerr))
		}
	}
}
