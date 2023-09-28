//go:build go1.18
// +build go1.18

package cue

import (
	"bytes"
	"os"
	"testing"
)

func FuzzValidate(f *testing.F) {
	testcases := []string{"testdata/valid.yml", "testdata/invalid.yml"}

	for _, tc := range testcases {
		fi, _ := os.ReadFile(tc)
		f.Add(fi)
	}

	f.Fuzz(func(t *testing.T, in []byte) {
		validator, err := NewFeaturesValidator()
		if err != nil {
			// only care about errors from Validating
			t.Skip()
		}

		if err := validator.Validate("foo", bytes.NewBuffer(in)); err != nil {
			// we only care about panics
			t.Skip()
		}
	})
}
