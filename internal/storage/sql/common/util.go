package common

import (
	"encoding/base64"
	"encoding/json"

	ferrors "go.flipt.io/flipt/errors"
	"go.uber.org/zap"
)

// decodePageToken is a utility function which determines the `PageToken` based on
// input.
func decodePageToken(logger *zap.Logger, pageToken string) (PageToken, error) {
	var token PageToken

	tok, err := base64.StdEncoding.DecodeString(pageToken)
	if err != nil {
		logger.Warn("invalid page token provided", zap.Error(err))
		return token, ferrors.ErrInvalidf("pageToken is not valid: %q", pageToken)
	}

	if err := json.Unmarshal(tok, &token); err != nil {
		logger.Warn("invalid page token provided", zap.Error(err))

		return token, ferrors.ErrInvalidf("pageToken is not valid: %q", pageToken)
	}

	return token, nil
}
