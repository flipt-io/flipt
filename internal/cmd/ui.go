package cmd

import (
	"fmt"
	"io/fs"
	"net/http"
	"path"
	"strings"

	"go.flipt.io/flipt/ui"
)

func uiHandler() (http.Handler, error) {
	u, err := fs.Sub(ui.UI, "dist")
	if err != nil {
		return nil, fmt.Errorf("mounting UI: %w", err)
	}

	return http.FileServer(http.FS(newNextFS(u))), nil
}

type nextFS struct {
	fs.FS

	root node
}

func newNextFS(filesystem fs.FS) fs.FS {
	n := &nextFS{FS: filesystem}

	fs.WalkDir(filesystem, ".", func(path string, _ fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		n.root.insert(path)

		return nil
	})

	return n
}

// Open opens the named file.
// When Open returns an error, it should be of type *PathError
// with the Op field set to "open", the Path field set to name,
// and the Err field describing the problem.
//
// Open should reject attempts to open names that do not satisfy
// ValidPath(name), returning a *PathError with Err set to
// ErrInvalid or ErrNotExist.
func (n *nextFS) Open(name string) (fs.File, error) {
	// if the path is invalid then immediately delegate
	// to the underlying filesystem and let it error
	// appropriately.
	if !fs.ValidPath(name) {
		return n.FS.Open(name)
	}

	// next we try to rewrite the path if it matches any
	// dynamic paths.
	parts := strings.Split(name, "/")

	node := &n.root
	for i, part := range parts {
		if part == "" {
			continue
		}

		child, ok := node.children[part]
		if ok {
			node = child
			continue
		}

		// if the current node has a dynamic child
		// entry then replace the matching entry
		// in the path with the name of the dynamic file.
		// This supports next.js style /path/to/[pid]/index.html.
		if node.dynamic != nil {
			node = node.dynamic
			parts[i] = node.name
			continue
		}

		// we dont have a matching part so dont attempt to rewrite it
		return n.FS.Open(name)
	}

	return n.FS.Open(path.Join(parts...))
}

type node struct {
	name     string
	dynamic  *node
	children map[string]*node
}

func (n *node) insert(path string) {
	next := n
	for _, part := range strings.Split(path, "/") {
		if part == "" {
			continue
		}

		next = next.getOrAdd(part)
	}
}

func (n *node) getOrAdd(name string) *node {
	if isDynamic(name) {
		if n.dynamic == nil {
			n.dynamic = &node{name: name}
		}

		return n.dynamic
	}

	if n.children == nil {
		n.children = map[string]*node{}
	}

	if next, ok := n.children[name]; ok {
		return next
	}

	c := &node{name: name}

	n.children[name] = c

	return c
}

func isDynamic(name string) bool {
	if len(name) < 2 {
		return false
	}

	return name[0] == '[' && name[len(name)-1] == ']'
}
