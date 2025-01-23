package flipt

import (
	"encoding/json"
	"fmt"

	"go.flipt.io/flipt/errors"
)

const (
	maxJsonStringSizeKB = 1000
	entityPropertyKey   = "entityId"
)

// Validator validates types
type Validator interface {
	Validate() error
}

// Evaluate

func validateJsonParameter(jsonValue string, parameterName string) error {
	if jsonValue == "" {
		return nil
	}

	bytes := []byte(jsonValue)
	if json.Valid(bytes) == false {
		return errors.InvalidFieldError(parameterName, "must be a json string")
	}

	if len(bytes) > (maxJsonStringSizeKB * 1024) {
		return errors.InvalidFieldError(parameterName,
			fmt.Sprintf("must be less than %d KB", maxJsonStringSizeKB),
		)
	}

	return nil
}

func (req *ListFlagRequest) Validate() error {
	if req.Limit == 0 && (req.Offset > 0 || req.PageToken != "") {
		return errors.ErrInvalid("limit must be set when offset or pageToken is set")
	}

	return nil
}
