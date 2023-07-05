//go:build go1.18
// +build go1.18

package ext

import (
	"bytes"
	"context"
	"os"
	"testing"
)

func FuzzImport(f *testing.F) {
	testcases := []string{"testdata/import.yml", "testdata/import_no_attachment.yml", "testdata/export.yml"}

	for _, tc := range testcases {
		b, _ := os.ReadFile(tc)
		f.Add(b)
	}

	f.Fuzz(func(t *testing.T, in []byte) {
		importer := NewImporter(&mockCreator{})
		if err := importer.Import(context.Background(), bytes.NewReader(in)); err != nil {
			// we only care about panics
			t.Skip()
		}
	})
}
