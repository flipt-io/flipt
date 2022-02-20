//go:build go1.18
// +build go1.18

package ext

import (
	"bytes"
	"context"
	"io/ioutil"
	"testing"
)

func FuzzImport(f *testing.F) {
	testcases := []string{"testdata/import.yml", "testdata/import_no_attachment.yml"}

	for _, tc := range testcases {
		b, _ := ioutil.ReadFile(tc)
		f.Add(b)
	}

	f.Fuzz(func(t *testing.T, in []byte) {
		importer := NewImporter(&mockCreator{})
		if err := importer.Import(context.Background(), bytes.NewReader(in)); err != nil {
			t.Skip()
		}
	})
}
