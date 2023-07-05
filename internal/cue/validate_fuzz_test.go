//go:build go1.18
// +build go1.18

package cue

import (
	"os"
	"testing"
)

func FuzzValidate(f *testing.F) {
	testcases := []string{"testdata/valid.yml", "testdata/invalid.yml"}

	for _, tc := range testcases {
		b, _ := os.ReadFile(tc)
		f.Add(b)
	}

	f.Fuzz(func(t *testing.T, in []byte) {
		validator, err := NewFeaturesValidator()
		if err != nil {
			// only care about errors from Validating
			t.Skip()
		}

		if _, err := validator.Validate("foo", in); err != nil {
			// we only care about panics
			t.Skip()
		}
	})
}
