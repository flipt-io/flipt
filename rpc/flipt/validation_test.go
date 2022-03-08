package flipt

import (
	"fmt"
	"testing"

	"github.com/markphelps/flipt/errors"
	"github.com/stretchr/testify/assert"
)

func largeJsonString() string {
	prefix := `{"a":"`
	suffix := `"}`

	//adding one for making the string larger than the limit
	b := make([]byte, maxVariantAttachmentSize-len(prefix)-len(suffix)+1)
	for i := range b {
		b[i] = 'a'
	}
	return fmt.Sprintf("%s%s%s", prefix, string(b), suffix)
}

func TestValidate_EvaluationRequest(t *testing.T) {
	tests := []struct {
		name    string
		req     *EvaluationRequest
		wantErr error
	}{
		{
			name:    "emptyFlagKey",
			req:     &EvaluationRequest{FlagKey: "", EntityId: "entityID"},
			wantErr: errors.EmptyFieldError("flagKey"),
		},
		{
			name:    "emptyEntityId",
			req:     &EvaluationRequest{FlagKey: "flagKey", EntityId: ""},
			wantErr: errors.EmptyFieldError("entityId"),
		},
		{
			name: "valid",
			req:  &EvaluationRequest{FlagKey: "flagKey", EntityId: "entityId"},
		},
	}

	for _, tt := range tests {
		var (
			req     = tt.req
			wantErr = tt.wantErr
		)

		t.Run(tt.name, func(t *testing.T) {
			err := req.Validate()
			assert.Equal(t, wantErr, err)
		})
	}
}

func TestValidate_GetFlagRequest(t *testing.T) {
	tests := []struct {
		name    string
		req     *GetFlagRequest
		wantErr error
	}{
		{
			name:    "emptyKey",
			req:     &GetFlagRequest{Key: ""},
			wantErr: errors.EmptyFieldError("key"),
		},
		{
			name: "valid",
			req:  &GetFlagRequest{Key: "key"},
		},
	}

	for _, tt := range tests {
		var (
			req     = tt.req
			wantErr = tt.wantErr
		)

		t.Run(tt.name, func(t *testing.T) {
			err := req.Validate()
			assert.Equal(t, wantErr, err)
		})
	}
}

func TestValidate_CreateFlagRequest(t *testing.T) {
	tests := []struct {
		name    string
		req     *CreateFlagRequest
		wantErr error
	}{
		{
			name: "emptyKey",
			req: &CreateFlagRequest{
				Key:         "",
				Name:        "name",
				Description: "desc",
				Enabled:     true,
			},
			wantErr: errors.EmptyFieldError("key"),
		},
		{
			name: "invalidKey",
			req: &CreateFlagRequest{
				Key:         "foo:bar",
				Name:        "name",
				Description: "desc",
				Enabled:     true,
			},
			wantErr: errors.InvalidFieldError("key", "contains invalid characters"),
		},
		{
			name: "emptyName",
			req: &CreateFlagRequest{
				Key:         "key",
				Name:        "",
				Description: "desc",
				Enabled:     true,
			},
			wantErr: errors.EmptyFieldError("name"),
		},
		{
			name: "valid",
			req: &CreateFlagRequest{
				Key:         "key",
				Name:        "name",
				Description: "desc",
				Enabled:     true,
			},
		},
	}

	for _, tt := range tests {
		var (
			req     = tt.req
			wantErr = tt.wantErr
		)

		t.Run(tt.name, func(t *testing.T) {
			err := req.Validate()
			assert.Equal(t, wantErr, err)
		})
	}
}

func TestValidate_UpdateFlagRequest(t *testing.T) {
	tests := []struct {
		name    string
		req     *UpdateFlagRequest
		wantErr error
	}{
		{
			name: "emptyKey",
			req: &UpdateFlagRequest{
				Key:         "",
				Name:        "name",
				Description: "desc",
				Enabled:     true,
			},
			wantErr: errors.EmptyFieldError("key"),
		},
		{
			name: "emptyName",
			req: &UpdateFlagRequest{
				Key:         "key",
				Name:        "",
				Description: "desc",
				Enabled:     true,
			},
			wantErr: errors.EmptyFieldError("name"),
		},
		{
			name: "valid",
			req: &UpdateFlagRequest{
				Key:         "key",
				Name:        "name",
				Description: "desc",
				Enabled:     true,
			},
		},
	}

	for _, tt := range tests {
		var (
			req     = tt.req
			wantErr = tt.wantErr
		)

		t.Run(tt.name, func(t *testing.T) {
			err := req.Validate()
			assert.Equal(t, wantErr, err)
		})
	}
}

func TestValidate_DeleteFlagRequest(t *testing.T) {
	tests := []struct {
		name    string
		req     *DeleteFlagRequest
		wantErr error
	}{
		{
			name: "emptyKey",
			req: &DeleteFlagRequest{
				Key: "",
			},
			wantErr: errors.EmptyFieldError("key"),
		},
		{
			name: "valid",
			req: &DeleteFlagRequest{
				Key: "key",
			},
		},
	}

	for _, tt := range tests {
		var (
			req     = tt.req
			wantErr = tt.wantErr
		)

		t.Run(tt.name, func(t *testing.T) {
			err := req.Validate()
			assert.Equal(t, wantErr, err)
		})
	}
}

func TestValidate_CreateVariantRequest(t *testing.T) {
	tests := []struct {
		name    string
		req     *CreateVariantRequest
		wantErr error
	}{
		{
			name: "emptyFlagKey",
			req: &CreateVariantRequest{
				FlagKey:     "",
				Key:         "key",
				Name:        "name",
				Description: "desc",
			},
			wantErr: errors.EmptyFieldError("flagKey"),
		},
		{
			name: "emptyKey",
			req: &CreateVariantRequest{
				FlagKey:     "flagKey",
				Key:         "",
				Name:        "name",
				Description: "desc",
			},
			wantErr: errors.EmptyFieldError("key"),
		},
		{
			name: "malformedJsonAttachment",
			req: &CreateVariantRequest{
				FlagKey:     "flagKey",
				Key:         "key",
				Name:        "name",
				Description: "desc",
				Attachment:  "attachment",
			},
			wantErr: errors.InvalidFieldError("attachment", "must be a json string"),
		},
		{
			name: "attachmentExceededLimit",
			req: &CreateVariantRequest{
				FlagKey:     "flagKey",
				Key:         "key",
				Name:        "name",
				Description: "desc",
				Attachment:  largeJsonString(),
			},
			wantErr: errors.InvalidFieldError(
				"attachment",
				fmt.Sprintf("must be less than %d KB", maxVariantAttachmentSize),
			),
		},
		{
			name: "valid",
			req: &CreateVariantRequest{
				FlagKey:     "flagKey",
				Key:         "key",
				Name:        "name",
				Description: "desc",
			},
		},
		{
			name: "validWithAttachment",
			req: &CreateVariantRequest{
				FlagKey:     "flagKey",
				Key:         "key",
				Name:        "name",
				Description: "desc",
				Attachment:  `{"key":"value"}`,
			},
		},
	}

	for _, tt := range tests {
		var (
			req     = tt.req
			wantErr = tt.wantErr
		)

		t.Run(tt.name, func(t *testing.T) {
			err := req.Validate()
			assert.Equal(t, wantErr, err)
		})
	}
}

func TestValidate_UpdateVariantRequest(t *testing.T) {
	tests := []struct {
		name    string
		req     *UpdateVariantRequest
		wantErr error
	}{
		{
			name: "emptyId",
			req: &UpdateVariantRequest{
				Id:          "",
				FlagKey:     "flagKey",
				Key:         "key",
				Name:        "name",
				Description: "desc",
			},
			wantErr: errors.EmptyFieldError("id"),
		},
		{
			name: "emptyFlagKey",
			req: &UpdateVariantRequest{
				Id:          "id",
				FlagKey:     "",
				Key:         "key",
				Name:        "name",
				Description: "desc",
			},
			wantErr: errors.EmptyFieldError("flagKey"),
		},
		{
			name: "emptyKey",
			req: &UpdateVariantRequest{
				Id:          "id",
				FlagKey:     "flagKey",
				Key:         "",
				Name:        "name",
				Description: "desc",
			},
			wantErr: errors.EmptyFieldError("key"),
		},
		{
			name: "malformedJsonAttachment",
			req: &UpdateVariantRequest{
				Id:          "id",
				FlagKey:     "flagKey",
				Key:         "key",
				Name:        "name",
				Description: "desc",
				Attachment:  "attachment",
			},
			wantErr: errors.InvalidFieldError("attachment", "must be a json string"),
		},
		{
			name: "attachmentExceededLimit",
			req: &UpdateVariantRequest{
				Id:          "id",
				FlagKey:     "flagKey",
				Key:         "key",
				Name:        "name",
				Description: "desc",
				Attachment:  largeJsonString(),
			},
			wantErr: errors.InvalidFieldError(
				"attachment",
				fmt.Sprintf("must be less than %d KB", maxVariantAttachmentSize),
			),
		},
		{
			name: "valid",
			req: &UpdateVariantRequest{
				Id:          "id",
				FlagKey:     "flagKey",
				Key:         "key",
				Name:        "name",
				Description: "desc",
			},
		},
		{
			name: "validWithAttachment",
			req: &UpdateVariantRequest{
				Id:          "id",
				FlagKey:     "flagKey",
				Key:         "key",
				Name:        "name",
				Description: "desc",
				Attachment:  `{"key":"value"}`,
			},
		},
	}

	for _, tt := range tests {
		var (
			req     = tt.req
			wantErr = tt.wantErr
		)

		t.Run(tt.name, func(t *testing.T) {
			err := req.Validate()
			assert.Equal(t, wantErr, err)
		})
	}
}

func TestValidate_DeleteVariantRequest(t *testing.T) {
	tests := []struct {
		name    string
		req     *DeleteVariantRequest
		wantErr error
	}{
		{
			name: "emptyId",
			req: &DeleteVariantRequest{
				Id:      "",
				FlagKey: "flagKey",
			},
			wantErr: errors.EmptyFieldError("id"),
		},
		{
			name: "emptyFlagKey",
			req: &DeleteVariantRequest{
				Id:      "id",
				FlagKey: "",
			},
			wantErr: errors.EmptyFieldError("flagKey"),
		},
		{
			name: "valid",
			req: &DeleteVariantRequest{
				Id:      "id",
				FlagKey: "flagKey",
			},
		},
	}

	for _, tt := range tests {
		var (
			req     = tt.req
			wantErr = tt.wantErr
		)

		t.Run(tt.name, func(t *testing.T) {
			err := req.Validate()
			assert.Equal(t, wantErr, err)
		})
	}
}

func TestValidate_ListRuleRequest(t *testing.T) {
	tests := []struct {
		name    string
		req     *ListRuleRequest
		wantErr error
	}{
		{
			name:    "emptyFlagKey",
			req:     &ListRuleRequest{FlagKey: ""},
			wantErr: errors.EmptyFieldError("flagKey"),
		},
		{
			name: "valid",
			req:  &ListRuleRequest{FlagKey: "flagKey"},
		},
	}

	for _, tt := range tests {
		var (
			req     = tt.req
			wantErr = tt.wantErr
		)

		t.Run(tt.name, func(t *testing.T) {
			err := req.Validate()
			assert.Equal(t, wantErr, err)
		})
	}
}

func TestValidate_GetRuleRequest(t *testing.T) {
	tests := []struct {
		name    string
		req     *GetRuleRequest
		wantErr error
	}{
		{
			name:    "emptyId",
			req:     &GetRuleRequest{Id: ""},
			wantErr: errors.EmptyFieldError("id"),
		},
		{
			name:    "emptyFlagKey",
			req:     &GetRuleRequest{Id: "id", FlagKey: ""},
			wantErr: errors.EmptyFieldError("flagKey"),
		},
		{
			name: "valid",
			req:  &GetRuleRequest{Id: "id", FlagKey: "flagKey"},
		},
	}

	for _, tt := range tests {
		var (
			req     = tt.req
			wantErr = tt.wantErr
		)

		t.Run(tt.name, func(t *testing.T) {
			err := req.Validate()
			assert.Equal(t, wantErr, err)
		})
	}
}

func TestValidate_CreateRuleRequest(t *testing.T) {
	tests := []struct {
		name    string
		req     *CreateRuleRequest
		wantErr error
	}{
		{
			name: "emptyFlagKey",
			req: &CreateRuleRequest{
				FlagKey:    "",
				SegmentKey: "segmentKey",
				Rank:       1,
			},
			wantErr: errors.EmptyFieldError("flagKey"),
		},
		{
			name: "emptySegmentKey",
			req: &CreateRuleRequest{
				FlagKey:    "flagKey",
				SegmentKey: "",
				Rank:       1,
			},
			wantErr: errors.EmptyFieldError("segmentKey"),
		},
		{
			name: "rankLessThanZero",
			req: &CreateRuleRequest{
				FlagKey:    "flagKey",
				SegmentKey: "segmentKey",
				Rank:       -1,
			},
			wantErr: errors.InvalidFieldError("rank", "must be greater than 0"),
		},
		{
			name: "valid",
			req: &CreateRuleRequest{
				FlagKey:    "flagKey",
				SegmentKey: "segmentKey",
				Rank:       1,
			},
		},
	}

	for _, tt := range tests {
		var (
			req     = tt.req
			wantErr = tt.wantErr
		)

		t.Run(tt.name, func(t *testing.T) {
			err := req.Validate()
			assert.Equal(t, wantErr, err)
		})
	}
}

func TestValidate_UpdateRuleRequest(t *testing.T) {
	tests := []struct {
		name    string
		req     *UpdateRuleRequest
		wantErr error
	}{
		{
			name: "emptyID",
			req: &UpdateRuleRequest{
				Id:         "",
				FlagKey:    "flagKey",
				SegmentKey: "segmentKey",
			},
			wantErr: errors.EmptyFieldError("id"),
		},
		{
			name: "emptyFlagKey",
			req: &UpdateRuleRequest{
				Id:         "id",
				FlagKey:    "",
				SegmentKey: "segmentKey",
			},
			wantErr: errors.EmptyFieldError("flagKey"),
		},
		{
			name: "emptySegmentKey",
			req: &UpdateRuleRequest{
				Id:         "id",
				FlagKey:    "flagKey",
				SegmentKey: "",
			},
			wantErr: errors.EmptyFieldError("segmentKey"),
		},
		{
			name: "valid",
			req: &UpdateRuleRequest{
				Id:         "id",
				FlagKey:    "flagKey",
				SegmentKey: "segmentKey",
			},
		},
	}

	for _, tt := range tests {
		var (
			req     = tt.req
			wantErr = tt.wantErr
		)

		t.Run(tt.name, func(t *testing.T) {
			err := req.Validate()
			assert.Equal(t, wantErr, err)
		})
	}
}

func TestValidate_DeleteRuleRequest(t *testing.T) {
	tests := []struct {
		name    string
		req     *DeleteRuleRequest
		wantErr error
	}{
		{
			name: "emptyID",
			req: &DeleteRuleRequest{
				Id:      "",
				FlagKey: "flagKey",
			},
			wantErr: errors.EmptyFieldError("id"),
		},
		{
			name: "emptyFlagKey",
			req: &DeleteRuleRequest{
				Id:      "id",
				FlagKey: "",
			},
			wantErr: errors.EmptyFieldError("flagKey"),
		},
		{
			name: "valid",
			req: &DeleteRuleRequest{
				Id:      "id",
				FlagKey: "flagKey",
			},
		},
	}

	for _, tt := range tests {
		var (
			req     = tt.req
			wantErr = tt.wantErr
		)

		t.Run(tt.name, func(t *testing.T) {
			err := req.Validate()
			assert.Equal(t, wantErr, err)
		})
	}
}

func TestValidate_OrderRulesRequest(t *testing.T) {
	tests := []struct {
		name    string
		req     *OrderRulesRequest
		wantErr error
	}{
		{
			name:    "emptyFlagKey",
			req:     &OrderRulesRequest{FlagKey: "", RuleIds: []string{"1", "2"}},
			wantErr: errors.EmptyFieldError("flagKey"),
		},
		{
			name:    "ruleIds length lesser than 2",
			req:     &OrderRulesRequest{FlagKey: "flagKey", RuleIds: []string{"1"}},
			wantErr: errors.InvalidFieldError("ruleIds", "must contain atleast 2 elements"),
		},
		{
			name: "valid",
			req:  &OrderRulesRequest{FlagKey: "flagKey", RuleIds: []string{"1", "2"}},
		},
	}

	for _, tt := range tests {
		var (
			req     = tt.req
			wantErr = tt.wantErr
		)

		t.Run(tt.name, func(t *testing.T) {
			err := req.Validate()
			assert.Equal(t, wantErr, err)
		})
	}
}

func TestValidate_CreateDistributionRequest(t *testing.T) {
	tests := []struct {
		name    string
		req     *CreateDistributionRequest
		wantErr error
	}{
		{
			name:    "emptyFlagKey",
			req:     &CreateDistributionRequest{FlagKey: "", RuleId: "ruleID", VariantId: "variantID"},
			wantErr: errors.EmptyFieldError("flagKey"),
		},
		{
			name:    "emptyRuleID",
			req:     &CreateDistributionRequest{FlagKey: "flagKey", RuleId: "", VariantId: "variantID"},
			wantErr: errors.EmptyFieldError("ruleId"),
		},
		{
			name:    "emptyVariantID",
			req:     &CreateDistributionRequest{FlagKey: "flagKey", RuleId: "ruleID", VariantId: ""},
			wantErr: errors.EmptyFieldError("variantId"),
		},
		{
			name:    "rollout is less than 0",
			req:     &CreateDistributionRequest{FlagKey: "flagKey", RuleId: "ruleID", VariantId: "variantID", Rollout: -1},
			wantErr: errors.InvalidFieldError("rollout", "must be greater than or equal to '0'"),
		},
		{
			name:    "rollout is more than 100",
			req:     &CreateDistributionRequest{FlagKey: "flagKey", RuleId: "ruleID", VariantId: "variantID", Rollout: 101},
			wantErr: errors.InvalidFieldError("rollout", "must be less than or equal to '100'"),
		},
		{
			name: "valid",
			req:  &CreateDistributionRequest{FlagKey: "flagKey", RuleId: "ruleID", VariantId: "variantID", Rollout: 100},
		},
	}

	for _, tt := range tests {
		var (
			req     = tt.req
			wantErr = tt.wantErr
		)

		t.Run(tt.name, func(t *testing.T) {
			err := req.Validate()
			assert.Equal(t, wantErr, err)
		})
	}
}

func TestValidate_UpdateDistributionRequest(t *testing.T) {
	tests := []struct {
		name    string
		req     *UpdateDistributionRequest
		wantErr error
	}{
		{
			name:    "emptyID",
			req:     &UpdateDistributionRequest{Id: "", FlagKey: "flagKey", RuleId: "ruleID", VariantId: "variantID"},
			wantErr: errors.EmptyFieldError("id"),
		},
		{
			name:    "emptyFlagKey",
			req:     &UpdateDistributionRequest{Id: "id", FlagKey: "", RuleId: "ruleID", VariantId: "variantID"},
			wantErr: errors.EmptyFieldError("flagKey"),
		},
		{
			name:    "emptyRuleID",
			req:     &UpdateDistributionRequest{Id: "id", FlagKey: "flagKey", RuleId: "", VariantId: "variantID"},
			wantErr: errors.EmptyFieldError("ruleId"),
		},
		{
			name:    "emptyVariantID",
			req:     &UpdateDistributionRequest{Id: "id", FlagKey: "flagKey", RuleId: "ruleID", VariantId: ""},
			wantErr: errors.EmptyFieldError("variantId"),
		},
		{
			name:    "rollout is less than 0",
			req:     &UpdateDistributionRequest{Id: "id", FlagKey: "flagKey", RuleId: "ruleID", VariantId: "variantID", Rollout: -1},
			wantErr: errors.InvalidFieldError("rollout", "must be greater than or equal to '0'"),
		},
		{
			name:    "rollout is more than 100",
			req:     &UpdateDistributionRequest{Id: "id", FlagKey: "flagKey", RuleId: "ruleID", VariantId: "variantID", Rollout: 101},
			wantErr: errors.InvalidFieldError("rollout", "must be less than or equal to '100'"),
		},
		{
			name: "valid",
			req:  &UpdateDistributionRequest{Id: "id", FlagKey: "flagKey", RuleId: "ruleID", VariantId: "variantID", Rollout: 100},
		},
	}

	for _, tt := range tests {
		var (
			req     = tt.req
			wantErr = tt.wantErr
		)

		t.Run(tt.name, func(t *testing.T) {
			err := req.Validate()
			assert.Equal(t, wantErr, err)
		})
	}
}

func TestValidate_DeleteDistributionRequest(t *testing.T) {
	tests := []struct {
		name    string
		req     *DeleteDistributionRequest
		wantErr error
	}{
		{
			name:    "emptyID",
			req:     &DeleteDistributionRequest{Id: "", FlagKey: "flagKey", RuleId: "ruleID", VariantId: "variantID"},
			wantErr: errors.EmptyFieldError("id"),
		},
		{
			name:    "emptyFlagKey",
			req:     &DeleteDistributionRequest{Id: "id", FlagKey: "", RuleId: "ruleID", VariantId: "variantID"},
			wantErr: errors.EmptyFieldError("flagKey"),
		},
		{
			name:    "emptyRuleID",
			req:     &DeleteDistributionRequest{Id: "id", FlagKey: "flagKey", RuleId: "", VariantId: "variantID"},
			wantErr: errors.EmptyFieldError("ruleId"),
		},
		{
			name:    "emptyVariantID",
			req:     &DeleteDistributionRequest{Id: "id", FlagKey: "flagKey", RuleId: "ruleID", VariantId: ""},
			wantErr: errors.EmptyFieldError("variantId"),
		},
		{
			name: "emptyVariantID",
			req:  &DeleteDistributionRequest{Id: "id", FlagKey: "flagKey", RuleId: "ruleID", VariantId: "variantID"},
		},
	}

	for _, tt := range tests {
		var (
			req     = tt.req
			wantErr = tt.wantErr
		)

		t.Run(tt.name, func(t *testing.T) {
			err := req.Validate()
			assert.Equal(t, wantErr, err)
		})
	}
}

func TestValidate_GetSegmentRequest(t *testing.T) {
	tests := []struct {
		name    string
		req     *GetSegmentRequest
		wantErr error
	}{
		{
			name:    "emptyKey",
			req:     &GetSegmentRequest{Key: ""},
			wantErr: errors.EmptyFieldError("key"),
		},
		{
			name: "valid",
			req:  &GetSegmentRequest{Key: "key"},
		},
	}

	for _, tt := range tests {
		var (
			req     = tt.req
			wantErr = tt.wantErr
		)

		t.Run(tt.name, func(t *testing.T) {
			err := req.Validate()
			assert.Equal(t, wantErr, err)
		})
	}
}

func TestValidate_CreateSegmentRequest(t *testing.T) {
	tests := []struct {
		name    string
		req     *CreateSegmentRequest
		wantErr error
	}{
		{
			name: "emptyKey",
			req: &CreateSegmentRequest{
				Key:         "",
				Name:        "name",
				Description: "desc",
			},
			wantErr: errors.EmptyFieldError("key"),
		},
		{
			name: "invalidKey",
			req: &CreateSegmentRequest{
				Key:         "foo:bar",
				Name:        "name",
				Description: "desc",
			},
			wantErr: errors.InvalidFieldError("key", "contains invalid characters"),
		},
		{
			name: "emptyName",
			req: &CreateSegmentRequest{
				Key:         "key",
				Name:        "",
				Description: "desc",
			},
			wantErr: errors.EmptyFieldError("name"),
		},
		{
			name: "valid",
			req: &CreateSegmentRequest{
				Key:         "key",
				Name:        "name",
				Description: "desc",
			},
		},
	}

	for _, tt := range tests {
		var (
			req     = tt.req
			wantErr = tt.wantErr
		)

		t.Run(tt.name, func(t *testing.T) {
			err := req.Validate()
			assert.Equal(t, wantErr, err)
		})
	}
}

func TestValidate_UpdateSegmentRequest(t *testing.T) {
	tests := []struct {
		name    string
		req     *UpdateSegmentRequest
		wantErr error
	}{
		{
			name: "emptyKey",
			req: &UpdateSegmentRequest{
				Key:         "",
				Name:        "name",
				Description: "desc",
			},
			wantErr: errors.EmptyFieldError("key"),
		},
		{
			name: "emptyName",
			req: &UpdateSegmentRequest{
				Key:         "key",
				Name:        "",
				Description: "desc",
			},
			wantErr: errors.EmptyFieldError("name"),
		},
		{
			name: "valid",
			req: &UpdateSegmentRequest{
				Key:         "key",
				Name:        "name",
				Description: "desc",
			},
		},
	}

	for _, tt := range tests {
		var (
			req     = tt.req
			wantErr = tt.wantErr
		)

		t.Run(tt.name, func(t *testing.T) {
			err := req.Validate()
			assert.Equal(t, wantErr, err)
		})
	}
}

func TestValidate_DeleteSegmentRequest(t *testing.T) {
	tests := []struct {
		name    string
		req     *DeleteSegmentRequest
		wantErr error
	}{
		{
			name:    "emptyKey",
			req:     &DeleteSegmentRequest{Key: ""},
			wantErr: errors.EmptyFieldError("key"),
		},
		{
			name: "valid",
			req:  &DeleteSegmentRequest{Key: "key"},
		},
	}

	for _, tt := range tests {
		var (
			req     = tt.req
			wantErr = tt.wantErr
		)

		t.Run(tt.name, func(t *testing.T) {
			err := req.Validate()
			assert.Equal(t, wantErr, err)
		})
	}
}

func TestValidate_CreateConstraintRequest(t *testing.T) {
	tests := []struct {
		name    string
		req     *CreateConstraintRequest
		wantErr error
	}{
		{
			name: "emptySegmentKey",
			req: &CreateConstraintRequest{
				SegmentKey: "",
				Type:       ComparisonType_BOOLEAN_COMPARISON_TYPE,
				Property:   "foo",
				Operator:   "EQ",
				Value:      "bar",
			},
			wantErr: errors.EmptyFieldError("segmentKey"),
		},
		{
			name: "emptyProperty",
			req: &CreateConstraintRequest{
				SegmentKey: "segmentKey",
				Type:       ComparisonType_BOOLEAN_COMPARISON_TYPE,
				Property:   "",
				Operator:   "EQ",
				Value:      "bar",
			},
			wantErr: errors.EmptyFieldError("property"),
		},
		{
			name: "emptyOperator",
			req: &CreateConstraintRequest{
				SegmentKey: "segmentKey",
				Type:       ComparisonType_BOOLEAN_COMPARISON_TYPE,
				Property:   "foo",
				Operator:   "",
				Value:      "bar",
			},
			wantErr: errors.EmptyFieldError("operator"),
		},
		{
			name: "invalidStringType",
			req: &CreateConstraintRequest{
				SegmentKey: "segmentKey",
				Type:       ComparisonType_STRING_COMPARISON_TYPE,
				Property:   "foo",
				Operator:   "false",
				Value:      "bar",
			},
			wantErr: errors.ErrInvalid("constraint operator \"false\" is not valid for type string"),
		},
		{
			name: "invalidNumberType",
			req: &CreateConstraintRequest{
				SegmentKey: "segmentKey",
				Type:       ComparisonType_NUMBER_COMPARISON_TYPE,
				Property:   "foo",
				Operator:   "false",
				Value:      "bar",
			},
			wantErr: errors.ErrInvalid("constraint operator \"false\" is not valid for type number"),
		},
		{
			name: "invalidBooleanType",
			req: &CreateConstraintRequest{
				SegmentKey: "segmentKey",
				Type:       ComparisonType_BOOLEAN_COMPARISON_TYPE,
				Property:   "foo",
				Operator:   "eq",
				Value:      "bar",
			},
			wantErr: errors.ErrInvalid("constraint operator \"eq\" is not valid for type boolean"),
		},
		{
			name: "invalidType",
			req: &CreateConstraintRequest{
				SegmentKey: "segmentKey",
				Type:       ComparisonType_UNKNOWN_COMPARISON_TYPE,
				Property:   "foo",
				Operator:   "eq",
				Value:      "bar",
			},
			wantErr: errors.ErrInvalid("invalid constraint type: \"UNKNOWN_COMPARISON_TYPE\""),
		},
		{
			name: "valid",
			req: &CreateConstraintRequest{
				SegmentKey: "segmentKey",
				Type:       ComparisonType_STRING_COMPARISON_TYPE,
				Property:   "foo",
				Operator:   "eq",
				Value:      "bar",
			},
		},
		{
			name: "emptyValue string valid",
			req: &CreateConstraintRequest{
				SegmentKey: "segmentKey",
				Type:       ComparisonType_STRING_COMPARISON_TYPE,
				Property:   "foo",
				Operator:   "notempty",
			},
		},
		{
			name: "emptyValue string not allowed",
			req: &CreateConstraintRequest{
				SegmentKey: "segmentKey",
				Type:       ComparisonType_STRING_COMPARISON_TYPE,
				Property:   "foo",
				Operator:   "eq",
			},
			wantErr: errors.EmptyFieldError("value"),
		},
		{
			name: "emptyValue boolean valid",
			req: &CreateConstraintRequest{
				SegmentKey: "segmentKey",
				Type:       ComparisonType_BOOLEAN_COMPARISON_TYPE,
				Property:   "foo",
				Operator:   "true",
			},
		},
		{
			name: "emptyValue number valid",
			req: &CreateConstraintRequest{
				SegmentKey: "segmentKey",
				Type:       ComparisonType_NUMBER_COMPARISON_TYPE,
				Property:   "foo",
				Operator:   "present",
			},
		},
	}

	for _, tt := range tests {
		var (
			req     = tt.req
			wantErr = tt.wantErr
		)

		t.Run(tt.name, func(t *testing.T) {
			err := req.Validate()
			assert.Equal(t, wantErr, err)
		})
	}
}

func TestValidate_UpdateConstraintRequest(t *testing.T) {
	tests := []struct {
		name    string
		req     *UpdateConstraintRequest
		wantErr error
	}{
		{
			name: "emptyID",
			req: &UpdateConstraintRequest{
				Id:         "",
				SegmentKey: "segmentKey",
				Type:       ComparisonType_BOOLEAN_COMPARISON_TYPE,
				Property:   "foo",
				Operator:   "EQ",
				Value:      "bar",
			},
			wantErr: errors.EmptyFieldError("id"),
		},
		{
			name: "emptySegmentKey",
			req: &UpdateConstraintRequest{
				Id:         "1",
				SegmentKey: "",
				Type:       ComparisonType_BOOLEAN_COMPARISON_TYPE,
				Property:   "foo",
				Operator:   "EQ",
				Value:      "bar",
			},
			wantErr: errors.EmptyFieldError("segmentKey"),
		},
		{
			name: "emptyProperty",
			req: &UpdateConstraintRequest{
				Id:         "1",
				SegmentKey: "segmentKey",
				Type:       ComparisonType_BOOLEAN_COMPARISON_TYPE,
				Property:   "",
				Operator:   "EQ",
				Value:      "bar",
			},
			wantErr: errors.EmptyFieldError("property"),
		},
		{
			name: "emptyOperator",
			req: &UpdateConstraintRequest{
				Id:         "1",
				SegmentKey: "segmentKey",
				Type:       ComparisonType_BOOLEAN_COMPARISON_TYPE,
				Property:   "foo",
				Operator:   "",
				Value:      "bar",
			},
			wantErr: errors.EmptyFieldError("operator"),
		},
		{
			name: "invalidStringType",
			req: &UpdateConstraintRequest{
				Id:         "1",
				SegmentKey: "segmentKey",
				Type:       ComparisonType_STRING_COMPARISON_TYPE,
				Property:   "foo",
				Operator:   "false",
				Value:      "bar",
			},
			wantErr: errors.ErrInvalid("constraint operator \"false\" is not valid for type string"),
		},
		{
			name: "invalidNumberType",
			req: &UpdateConstraintRequest{
				Id:         "1",
				SegmentKey: "segmentKey",
				Type:       ComparisonType_NUMBER_COMPARISON_TYPE,
				Property:   "foo",
				Operator:   "false",
				Value:      "bar",
			},
			wantErr: errors.ErrInvalid("constraint operator \"false\" is not valid for type number"),
		},
		{
			name: "invalidBooleanType",
			req: &UpdateConstraintRequest{
				Id:         "1",
				SegmentKey: "segmentKey",
				Type:       ComparisonType_BOOLEAN_COMPARISON_TYPE,
				Property:   "foo",
				Operator:   "eq",
				Value:      "bar",
			},
			wantErr: errors.ErrInvalid("constraint operator \"eq\" is not valid for type boolean"),
		},
		{
			name: "invalidType",
			req: &UpdateConstraintRequest{
				Id:         "1",
				SegmentKey: "segmentKey",
				Type:       ComparisonType_UNKNOWN_COMPARISON_TYPE,
				Property:   "foo",
				Operator:   "eq",
				Value:      "bar",
			},
			wantErr: errors.ErrInvalid("invalid constraint type: \"UNKNOWN_COMPARISON_TYPE\""),
		},
		{
			name: "valid",
			req: &UpdateConstraintRequest{
				Id:         "1",
				SegmentKey: "segmentKey",
				Type:       ComparisonType_STRING_COMPARISON_TYPE,
				Property:   "foo",
				Operator:   "eq",
				Value:      "bar",
			},
		},
		{
			name: "emptyValue string valid",
			req: &UpdateConstraintRequest{
				Id:         "1",
				SegmentKey: "segmentKey",
				Type:       ComparisonType_STRING_COMPARISON_TYPE,
				Property:   "foo",
				Operator:   "notempty",
			},
		},
		{
			name: "emptyValue string not allowed",
			req: &UpdateConstraintRequest{
				Id:         "1",
				SegmentKey: "segmentKey",
				Type:       ComparisonType_STRING_COMPARISON_TYPE,
				Property:   "foo",
				Operator:   "eq",
			},
			wantErr: errors.EmptyFieldError("value"),
		},
		{
			name: "emptyValue boolean valid",
			req: &UpdateConstraintRequest{
				Id:         "1",
				SegmentKey: "segmentKey",
				Type:       ComparisonType_BOOLEAN_COMPARISON_TYPE,
				Property:   "foo",
				Operator:   "true",
			},
		},
		{
			name: "emptyValue number valid",
			req: &UpdateConstraintRequest{
				Id:         "1",
				SegmentKey: "segmentKey",
				Type:       ComparisonType_NUMBER_COMPARISON_TYPE,
				Property:   "foo",
				Operator:   "present",
			},
		},
	}

	for _, tt := range tests {
		var (
			req     = tt.req
			wantErr = tt.wantErr
		)

		t.Run(tt.name, func(t *testing.T) {
			err := req.Validate()
			assert.Equal(t, wantErr, err)
		})
	}
}

func TestValidate_DeleteConstraintRequest(t *testing.T) {
	tests := []struct {
		name    string
		req     *DeleteConstraintRequest
		wantErr error
	}{
		{
			name:    "emptyID",
			req:     &DeleteConstraintRequest{Id: "", SegmentKey: "segmentKey"},
			wantErr: errors.EmptyFieldError("id"),
		},
		{
			name:    "emptySegmentKey",
			req:     &DeleteConstraintRequest{Id: "id", SegmentKey: ""},
			wantErr: errors.EmptyFieldError("segmentKey"),
		},
		{
			name: "valid",
			req:  &DeleteConstraintRequest{Id: "id", SegmentKey: "segmentKey"},
		},
	}

	for _, tt := range tests {
		var (
			req     = tt.req
			wantErr = tt.wantErr
		)

		t.Run(tt.name, func(t *testing.T) {
			err := req.Validate()
			assert.Equal(t, wantErr, err)
		})
	}
}
