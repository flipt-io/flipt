package filesystem

import (
	"bytes"
	"context"
	"encoding/json"
	"os"
	"time"

	"go.flipt.io/flipt/internal/server/authz/engine/rego/source"
)

// LocalPolicySource is an implementation of PolicySource
// It sources policy definitions from the location filesystem
type LocalPolicySource struct {
	path string
}

// LocalPolicySourceFromPath builds an instance of *PolicySourceFromPath which reads the provided
// path and returns the located policy
func PolicySourceFromPath(path string) *LocalPolicySource {
	return &LocalPolicySource{path}
}

// Get reads the *LocalPolicySource path as and returns it as a byte slice
// If the mod time of the filepath has not changed based on the provided previously seen
// mod time then it returns authz.ErrNotModified
func (p *LocalPolicySource) Get(_ context.Context, seen source.Hash) ([]byte, source.Hash, error) {
	return read(p.path, seen)
}

// LocalDataSource is an implementation of DataSource
// It sources additional data for the policy engine from the local filesystem
type LocalDataSource struct {
	path string
}

// DataSourceFromPath builds an instance of *LocalDataSource which reads the provided
// path and parses it as JSON on calls to Get
func DataSourceFromPath(path string) *LocalDataSource {
	return &LocalDataSource{path: path}
}

// Get fetches and parses the *LocalDataSource path as JSON and returns it
// modelled as a go map of string to any type
// If the mod time of the filepath has not changed based on the provided previously seen
// mod time then it returns authz.ErrNotModified
func (d *LocalDataSource) Get(_ context.Context, seen source.Hash) (map[string]any, source.Hash, error) {
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
		return nil, nil, source.ErrNotModified
	}

	data, err = os.ReadFile(path)
	return
}
