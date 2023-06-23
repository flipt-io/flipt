package common

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"

	ferrors "go.flipt.io/flipt/errors"
)

// decodePageToken is a utility function which determines the `PageToken` based on
// input.
func decodePageToken(pageToken string) (PageToken, error) {
	var token PageToken

	tok, err := base64.StdEncoding.DecodeString(pageToken)
	if err != nil {
		return token, ferrors.ErrInvalidf("pageToken is not valid: %q", pageToken)
	}

	if err := json.Unmarshal(tok, &token); err != nil {
		var jerr *json.SyntaxError

		if errors.As(err, &jerr) {
			return token, ferrors.ErrInvalidf("pageToken is not valid: %q", pageToken)
		}

		return token, fmt.Errorf("decoding page token %w", err)
	}

	return token, nil
}
