package flipt

import (
	"testing"

	"github.com/markphelps/flipt/errors"
	"github.com/stretchr/testify/assert"
)

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
			name: "emptyName",
			req: &CreateSegmentRequest{
				Key:         "key",
				Name:        "",
				Description: "desc",
			},
			wantErr: errors.EmptyFieldError("name"),
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
