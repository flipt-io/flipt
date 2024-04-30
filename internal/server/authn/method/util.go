package method

import (
	"context"

	"encoding/json"
	"fmt"

	"github.com/jmespath/go-jmespath"
	"go.flipt.io/flipt/errors"
	"google.golang.org/grpc/metadata"
)

// CallbackValidateState validates the state for the callback request on both OIDC and GitHub as
// an OAuth provider.
func CallbackValidateState(ctx context.Context, state string) error {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return errors.ErrUnauthenticatedf("missing state parameter")
	}

	fstate, ok := md["flipt_client_state"]
	if !ok || len(state) < 1 {
		return errors.ErrUnauthenticatedf("missing state parameter")
	}

	if state != fstate[0] {
		return errors.ErrUnauthenticatedf("unexpected state parameter")
	}

	return nil
}

func SearchJSONForAttr[T ~string](attributePath string, data any) (T, error) {
	v, err := searchJSONForAttr(attributePath, data)
	if err != nil {
		return "", err
	}
	return v.(T), nil
}

func searchJSONForAttr(attributePath string, data any) (any, error) {
	if attributePath == "" {
		return "", fmt.Errorf("attribute path: %q", attributePath)
	}

	if data == nil {
		return "", fmt.Errorf("empty json, attribute path: %q", attributePath)
	}

	// Copy the data to a new variable
	var jsonData = data

	// If the data is a byte slice, try to unmarshal it into a JSON object
	if dataBytes, ok := data.([]byte); ok {
		// If the byte slice is empty, return an error
		if len(dataBytes) == 0 {
			return "", fmt.Errorf("empty json, attribute path: %q", attributePath)
		}

		// Try to unmarshal the byte slice
		if err := json.Unmarshal(dataBytes, &jsonData); err != nil {
			return "", fmt.Errorf("%v: %w", "failed to unmarshal user info JSON response", err)
		}
	}

	// Search for the attribute in the JSON object
	value, err := jmespath.Search(attributePath, jsonData)
	if err != nil {
		return "", fmt.Errorf("failed to search user info JSON response with provided path: %q: %w", attributePath, err)
	}

	// Return the value and nil error
	return value, nil
}
