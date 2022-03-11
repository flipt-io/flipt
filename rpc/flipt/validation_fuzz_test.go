//go:build go1.18
// +build go1.18

package flipt

import (
	"errors"
	"testing"

	errs "github.com/markphelps/flipt/errors"
)

func FuzzValidateAttachment(f *testing.F) {
	seeds := []string{`{"key":"value"}`, `{"foo":"bar"}`, `{"foo":"bar","key":"value"}`, `{"foo":"bar","key":"value","baz":"qux"}`}

	for _, s := range seeds {
		f.Add(s)
	}

	f.Fuzz(func(t *testing.T, in string) {
		err := validateAttachment(in)
		if err != nil {
			var verr errs.ErrValidation
			if errors.As(err, &verr) {
				t.Skip()
			}
			t.Fail()
		}
	})
}
