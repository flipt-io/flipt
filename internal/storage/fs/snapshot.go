package fs

import (
	"errors"
	"io"
)

type storeSnapshot struct{}

// snapshotFromFS constructs a storeSnapshot from the provided
// fs.FS implementation.
func snapshotFromFS(sources ...io.Reader) (*storeSnapshot, error) {
	return nil, errors.New("not implemented")
}
