package log

import (
	"context"
	"io"
	"os"
	"path"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.flipt.io/flipt/internal/config"
	"go.flipt.io/flipt/internal/server/audit"
	"go.flipt.io/flipt/rpc/flipt"
)

func TestSink(t *testing.T) {
	file, err := os.CreateTemp("", "test*.log")
	require.NoError(t, err)

	defer func() {
		if file != nil {
			file.Close()
			os.Remove(file.Name())
		}
	}()

	s, err := NewSink(WithPath(file.Name()), WithEncoding(config.LogEncodingJSON))
	require.NoError(t, err)

	require.Equal(t, "log", s.String())

	err = s.SendAudits(context.TODO(), []audit.Event{
		{
			Version: "0.2",
			Type:    string(flipt.SubjectFlag),
			Action:  string(flipt.ActionCreate),
			Status:  string(flipt.StatusSuccess),
		},
		{
			Version: "0.2",
			Type:    string(flipt.SubjectConstraint),
			Action:  string(flipt.ActionUpdate),
			Status:  string(flipt.StatusSuccess),
		},
	})

	require.NoError(t, err)
	require.NoError(t, s.Close())

	_, err = os.Stat(file.Name())
	require.NoError(t, err)

	log, err := io.ReadAll(file)
	require.NoError(t, err)

	lines := strings.Split(string(log), "\n")

	assert.NotEmpty(t, lines)
	assert.NotEmpty(t, lines[0])

	assert.JSONEq(t, `{"version": "0.2", "type": "flag", "action": "create", "metadata": {}, "payload": null, "timestamp": "", "status": "success"}`, lines[0])
}

func TestSink_DirNotExists(t *testing.T) {
	tests := []struct {
		name string
		path string
	}{
		{
			name: "one level",
			path: "foo/doesnotexist/test.log",
		},
		{
			name: "two levels",
			path: "foo/doesnotexist/bar/test.log",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			path := path.Join(os.TempDir(), tt.path)

			defer os.RemoveAll(path)

			s, err := NewSink(WithPath(path), WithEncoding(config.LogEncodingJSON))
			require.NoError(t, err)

			err = s.SendAudits(context.TODO(), []audit.Event{
				{
					Version: "0.2",
					Type:    string(flipt.SubjectFlag),
					Action:  string(flipt.ActionCreate),
					Status:  string(flipt.StatusSuccess),
				},
				{
					Version: "0.2",
					Type:    string(flipt.SubjectConstraint),
					Action:  string(flipt.ActionUpdate),
					Status:  string(flipt.StatusSuccess),
				},
			})

			require.NoError(t, err)
			require.NoError(t, s.Close())

			file, err := os.Open(path)
			require.NoError(t, err)
			defer file.Close()

			log, err := io.ReadAll(file)
			require.NoError(t, err)

			lines := strings.Split(string(log), "\n")

			assert.NotEmpty(t, lines)
			assert.NotEmpty(t, lines[0])

			assert.JSONEq(t, `{"version": "0.2", "type": "flag", "action": "create", "metadata": {}, "payload": null, "timestamp": "", "status": "success"}`, lines[0])
		})
	}
}
