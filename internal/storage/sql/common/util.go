package common

import (
	"encoding/base64"
	"encoding/json"

	"go.flipt.io/flipt/errors"
	"go.uber.org/zap"
)

// decodePageToken is a utility function which determines the `PageToken` based on
// input.
func decodePageToken(logger *zap.Logger, pageToken string) (PageToken, error) {
	var token PageToken

	tok, err := base64.StdEncoding.DecodeString(pageToken)
	if err != nil {
		logger.Warn("invalid page token provided", zap.Error(err))
		return token, errors.ErrInvalidf("pageToken is not valid: %q", pageToken)
	}

	if err := json.Unmarshal(tok, &token); err != nil {
		logger.Warn("invalid page token provided", zap.Error(err))

		return token, errors.ErrInvalidf("pageToken is not valid: %q", pageToken)
	}

	return token, nil
}

// removeDuplicates is an inner utility function that will deduplicate a slice of strings.
func removeDuplicates(src []string) []string {
	allKeys := make(map[string]bool)

	dest := []string{}

	for _, item := range src {
		if _, value := allKeys[item]; !value {
			allKeys[item] = true
			dest = append(dest, item)
		}
	}

	return dest
}

// sanitizeSegmentKeys is a utility function that will transform segment keys into the right input.
func sanitizeSegmentKeys(segmentKey string, segmentKeys []string) []string {
	if len(segmentKeys) > 0 {
		return removeDuplicates(segmentKeys)
	} else if segmentKey != "" {
		return []string{segmentKey}
	}

	return nil
}
