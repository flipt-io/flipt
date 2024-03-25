//go:build go1.18
// +build go1.18

package validation

import (
	"bytes"
	"os"
	"testing"
)

func FuzzValidate(f *testing.F) {
	testcases := []string{"testdata/valid.yml", "testdata/invalid.yml", "testdata/valid_v1.yml", "testdata/valid_segments_v2.yml", "testdata/valid_yaml_stream.yml"}

	for _, tc := range testcases {
		fi, _ := os.ReadFile(tc)
		f.Add(fi)
	}

	f.Fuzz(func(t *testing.T, in []byte) {
		validator, err := NewFeaturesValidator(WithSchemaExtension(in))
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
