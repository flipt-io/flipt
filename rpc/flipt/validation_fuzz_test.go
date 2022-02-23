//go:build go1.18
// +build go1.18

package flipt

import (
	"testing"
)

func FuzzValidateAttachment(f *testing.F) {
	testcases := []string{`{"key":"value"}`, `{"foo":"bar"}`, `{"foo":"bar","key":"value"}`, `{"foo":"bar","key":"value","baz":"qux"}`}

	for _, tc := range testcases {
		f.Add(tc)
	}

	f.Fuzz(func(t *testing.T, in string) {
		if err := validateAttachment(in); err != nil {
			t.Skip()
		}
	})
}
