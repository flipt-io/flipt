//go:build go1.18
// +build go1.18

package flipt

import (
	"errors"
	"testing"

	errs "go.flipt.io/flipt/errors"
)

func FuzzValidateJsonParameter(f *testing.F) {
	seeds := []string{`{"key":"value"}`, `{"foo":"bar"}`, `{"foo":"bar","key":"value"}`, `{"foo":"bar","key":"value","baz":"qux"}`}

	for _, s := range seeds {
		f.Add(s)
	}

	f.Fuzz(func(t *testing.T, in string) {
		err := validateJsonParameter(in, "attachment")
		if err != nil {
			var verr errs.ErrValidation
			if errors.As(err, &verr) {
				t.Skip()
			}
			t.Fail()
		}
	})
}

func FuzzValidateArrayValue(f *testing.F) {
	seeds := []string{`["hello", "world"]`, `["foo", "bar", "testing", "more"]`, `[]`}

	for _, s := range seeds {
		f.Add(s)
	}

	f.Fuzz(func(t *testing.T, in string) {
		err := validateArrayValue(ComparisonType_STRING_COMPARISON_TYPE, in, "foobar")
		if err != nil {
			var verr errs.ErrValidation
			if errors.As(err, &verr) {
				t.Skip()
			}
			t.Fail()
		}
	})
}
