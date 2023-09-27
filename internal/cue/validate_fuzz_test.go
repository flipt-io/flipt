//go:build go1.18
// +build go1.18

package cue

import (
	"io"
	"os"
	"testing"
)

func FuzzValidate(f *testing.F) {
	testcases := []string{"testdata/valid.yml", "testdata/invalid.yml"}

	for _, tc := range testcases {
		fi, _ := os.Open(tc)
		f.Add(fi)
	}

	f.Fuzz(func(t *testing.T, in io.Reader) {
		validator, err := NewFeaturesValidator()
		if err != nil {
			// only care about errors from Validating
			t.Skip()
		}

		if err := validator.Validate("foo", in); err != nil {
			// we only care about panics
			t.Skip()
		}
	})
}
