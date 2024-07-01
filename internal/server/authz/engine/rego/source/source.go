package source

import "errors"

// ErrNotModified is returned from a source when the data has not
// been modified, identified based on the provided hash value
var ErrNotModified = errors.New("not modified")

type Hash []byte
