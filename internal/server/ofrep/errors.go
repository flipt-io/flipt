package ofrep

import (
	"encoding/json"
	"fmt"
	"strings"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
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
	Key          string `json:"key,omitempty"`
	ErrorCode    string `json:"errorCode,omitempty"`
	ErrorDetails string `json:"errorDetails"`
}

func NewBadRequestError(key string, err error) error {
	msg, err := json.Marshal(errorSchema{
		Key:          key,
		ErrorCode:    string(errorCodeParseGeneral),
		ErrorDetails: err.Error(),
	})
	if err != nil {
		return NewInternalServerError(err)
	}

	return status.Error(codes.InvalidArgument, string(msg))
}

func NewUnauthenticatedError() error {
	msg, err := json.Marshal(errorSchema{ErrorDetails: "unauthenticated error"})
	if err != nil {
		return NewInternalServerError(err)
	}

	return status.Error(codes.Unauthenticated, string(msg))
}

func NewUnauthorizedError() error {
	msg, err := json.Marshal(errorSchema{ErrorDetails: "unauthorized error"})
	if err != nil {
		return NewInternalServerError(err)
	}

	return status.Error(codes.PermissionDenied, string(msg))
}

func NewFlagNotFoundError(key string) error {
	msg, err := json.Marshal(errorSchema{
		Key:       key,
		ErrorCode: string(errorCodeFlagNotFound),
	})
	if err != nil {
		return NewInternalServerError(err)
	}

	return status.Error(codes.NotFound, string(msg))
}

func NewTargetingKeyMissing() error {
	msg, err := json.Marshal(errorSchema{
		ErrorCode:    string(errorCodeTargetingKeyMissing),
		ErrorDetails: "flag key was not provided",
	})
	if err != nil {
		return NewInternalServerError(err)
	}

	return status.Error(codes.InvalidArgument, string(msg))
}

func NewInternalServerError(err error) error {
	return status.Error(
		codes.Internal,
		fmt.Sprintf(`{"errorDetails": "%s"}`, strings.ReplaceAll(err.Error(), `"`, `\"`)),
	)
}
