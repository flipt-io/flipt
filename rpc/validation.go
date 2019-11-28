package flipt

import "github.com/markphelps/flipt/errors"

// Validator validates types
type Validator interface {
	Validate() error
}

// Rules

func (req *ListRuleRequest) Validate() error {
	if req.FlagKey == "" {
		return errors.EmptyFieldError("flagKey")
	}

	return nil
}

func (req *GetRuleRequest) Validate() error {
	if req.Id == "" {
		return errors.EmptyFieldError("id")
	}

	if req.FlagKey == "" {
		return errors.EmptyFieldError("flagKey")
	}

	return nil
}

func (req *CreateRuleRequest) Validate() error {
	if req.FlagKey == "" {
		return errors.EmptyFieldError("flagKey")
	}

	if req.SegmentKey == "" {
		return errors.EmptyFieldError("segmentKey")
	}

	if req.Rank <= 0 {
		return errors.InvalidFieldError("rank", "must be greater than 0")
	}

	return nil
}

func (req *DeleteRuleRequest) Validate() error {
	if req.Id == "" {
		return errors.EmptyFieldError("id")
	}

	if req.FlagKey == "" {
		return errors.EmptyFieldError("flagKey")
	}

	return nil
}

func (req *UpdateRuleRequest) Validate() error {
	if req.Id == "" {
		return errors.EmptyFieldError("id")
	}

	if req.FlagKey == "" {
		return errors.EmptyFieldError("flagKey")
	}

	if req.SegmentKey == "" {
		return errors.EmptyFieldError("segmentKey")
	}

	return nil
}

func (req *OrderRulesRequest) Validate() error {
	if req.FlagKey == "" {
		return errors.EmptyFieldError("flagKey")
	}

	if len(req.RuleIds) < 2 {
		return errors.InvalidFieldError("ruleIds", "must contain atleast 2 elements")
	}

	return nil
}

func (req *CreateDistributionRequest) Validate() error {
	if req.FlagKey == "" {
		return errors.EmptyFieldError("flagKey")
	}

	if req.RuleId == "" {
		return errors.EmptyFieldError("ruleId")
	}

	if req.VariantId == "" {
		return errors.EmptyFieldError("variantId")
	}

	if req.Rollout < 0 {
		return errors.InvalidFieldError("rollout", "must be greater than or equal to '0'")
	}

	if req.Rollout > 100 {
		return errors.InvalidFieldError("rollout", "must be less than or equal to '100'")
	}

	return nil
}

func (req *UpdateDistributionRequest) Validate() error {
	if req.Id == "" {
		return errors.EmptyFieldError("id")
	}

	if req.FlagKey == "" {
		return errors.EmptyFieldError("flagKey")
	}

	if req.RuleId == "" {
		return errors.EmptyFieldError("ruleId")
	}

	if req.VariantId == "" {
		return errors.EmptyFieldError("variantId")
	}

	if req.Rollout < 0 {
		return errors.InvalidFieldError("rollout", "must be greater than or equal to '0'")
	}

	if req.Rollout > 100 {
		return errors.InvalidFieldError("rollout", "must be less than or equal to '100'")
	}

	return nil
}

func (req *DeleteDistributionRequest) Validate() error {
	if req.Id == "" {
		return errors.EmptyFieldError("id")
	}

	if req.FlagKey == "" {
		return errors.EmptyFieldError("flagKey")
	}

	if req.RuleId == "" {
		return errors.EmptyFieldError("ruleId")
	}

	if req.VariantId == "" {
		return errors.EmptyFieldError("variantId")
	}

	return nil
}

// Segments

func (req *GetSegmentRequest) Validate() error {
	if req.Key == "" {
		return errors.EmptyFieldError("key")
	}

	return nil
}

func (req *CreateSegmentRequest) Validate() error {
	if req.Key == "" {
		return errors.EmptyFieldError("key")
	}

	if req.Name == "" {
		return errors.EmptyFieldError("name")
	}

	return nil
}

func (req *UpdateSegmentRequest) Validate() error {
	if req.Key == "" {
		return errors.EmptyFieldError("key")
	}

	if req.Name == "" {
		return errors.EmptyFieldError("name")
	}

	return nil
}

func (req *DeleteSegmentRequest) Validate() error {
	if req.Key == "" {
		return errors.EmptyFieldError("key")
	}

	return nil
}

func (req *CreateConstraintRequest) Validate() error {
	if req.SegmentKey == "" {
		return errors.EmptyFieldError("segmentKey")
	}

	if req.Property == "" {
		return errors.EmptyFieldError("property")
	}

	if req.Operator == "" {
		return errors.EmptyFieldError("operator")
	}

	// TODO: test for empty value if operator ! [EMPTY, NOT_EMPTY, PRESENT, NOT_PRESENT]
	return nil
}

func (req *UpdateConstraintRequest) Validate() error {
	if req.Id == "" {
		return errors.EmptyFieldError("id")
	}

	if req.SegmentKey == "" {
		return errors.EmptyFieldError("segmentKey")
	}

	if req.Property == "" {
		return errors.EmptyFieldError("property")
	}

	if req.Operator == "" {
		return errors.EmptyFieldError("operator")
	}

	// TODO: test for empty value if operator ! [EMPTY, NOT_EMPTY, PRESENT, NOT_PRESENT]
	return nil
}

func (req *DeleteConstraintRequest) Validate() error {
	if req.Id == "" {
		return errors.EmptyFieldError("id")
	}

	if req.SegmentKey == "" {
		return errors.EmptyFieldError("segmentKey")
	}

	return nil
}
