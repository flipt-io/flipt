package filesystem

import (
	"bytes"
	"context"
	"encoding/json"
	"os"
	"time"

	"go.flipt.io/flipt/internal/server/authz"
)

var _ authz.DataSource = (*DataSource)(nil)
var _ authz.PolicySource = (*PolicySource)(nil)

// PolicySource is an implementation of authz.PolicySource
// It sources policy definitions from the location filesystem
type PolicySource struct {
	path string
}

// PolicySourceFromPath builds an instance pof *PolicySourceFromPath which reads the provided
// path and returns the located policy
func PolicySourceFromPath(path string) *PolicySource {
	return &PolicySource{path}
}

// Get reads the *PolicySource path as and returns it as a byte slice
// If the mod time of the filepath has not changed based on the provided previously seen
// mod time then it returns authz.ErrNotModified
func (p *PolicySource) Get(_ context.Context, seen []byte) ([]byte, []byte, error) {
	return read(p.path, seen)
}

// DataSource is an implementation of authz.DataSource
// It sources additional data for the policy engine from the local filesystem
type DataSource struct {
	path string
}

// DataSourceFromPath builds an instance of *DataSource which reads the provided
// path and parses it as JSON on calls to Get
func DataSourceFromPath(path string) *DataSource {
	return &DataSource{path: path}
}

// Get fetches and parses the *DataSource path as JSON and returns it
// modelled as a go map of string to any type
// If the mod time of the filepath has not changed based on the provided previously seen
// mod time then it returns authz.ErrNotModified
func (d *DataSource) Get(_ context.Context, seen []byte) (map[string]any, []byte, error) {
	b, mod, err := read(d.path, seen)
	if err != nil {
		return nil, nil, err
	}

	data := map[string]any{}

	return data, mod, json.Unmarshal(b, &data)
}

func read(path string, seen []byte) (data, mod []byte, err error) {
	info, err := os.Stat(path)
	if err != nil {
		return nil, nil, err
	}

	mod = []byte(info.ModTime().Format(time.RFC3339))
	if seen != nil && bytes.Equal(seen, mod) {
		return nil, nil, authz.ErrNotModified
	}

	data, err = os.ReadFile(path)
	return
}
