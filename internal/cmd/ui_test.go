package cmd

import (
	"embed"
	"fmt"
	"io"
	"io/fs"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

//go:embed all:testdata/*
var data embed.FS

func Test_newNextFS(t *testing.T) {
	testdata, err := fs.Sub(data, "testdata")
	require.NoError(t, err)

	wrapped, err := newNextFS(testdata)
	require.NoError(t, err)

	for _, test := range []struct {
		path         string
		expectedFile []byte
		expectedErr  error
	}{
		{
			path:        "does-not-exists.txt",
			expectedErr: fs.ErrNotExist,
		},
		{
			path:         "index.html",
			expectedFile: []byte("<html></html>\n"),
		},
		{
			path:         "foo/index.html",
			expectedFile: []byte("<html><title>foo</title></html>\n"),
		},
		// ensure eplicit path are preserved
		{
			path:         "bar/baz/index.html",
			expectedFile: []byte("<html><title>baz</title></html>\n"),
		},
		// ensure dynamic paths are supported
		{
			path:         "bar/some-dynamic-path/index.html",
			expectedFile: []byte("<html><title>[pid]</title></html>\n"),
		},
	} {
		test := test
		t.Run(fmt.Sprintf("Open(%s)", test.path), func(t *testing.T) {
			fi, err := wrapped.Open(test.path)
			if test.expectedErr != nil {
				require.ErrorIs(t, err, test.expectedErr)
				return
			}

			require.NoError(t, err)

			data, err := io.ReadAll(fi)
			require.NoError(t, err)

			assert.Equal(t, test.expectedFile, data)
		})
	}
}
