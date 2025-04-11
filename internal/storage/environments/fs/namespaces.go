package fs

import (
	"context"
	"fmt"
	"io"
	"os"
	"path"

	"go.flipt.io/flipt/errors"
	"go.flipt.io/flipt/internal/ext"
	"go.flipt.io/flipt/rpc/flipt"
	rpcenvironments "go.flipt.io/flipt/rpc/v2/environments"
	"go.uber.org/zap"
)

const (
	defaultKey       = flipt.DefaultNamespace
	FeaturesFilename = "features.yaml"
)

var defaultNamespace = &rpcenvironments.Namespace{
	Key:         defaultKey,
	Name:        "default",
	Description: ptr("The default namespace"),
	Protected:   ptr(true),
}

type NamespaceStorage struct {
	logger *zap.Logger
}

func NewNamespaceStorage(logger *zap.Logger) *NamespaceStorage {
	return &NamespaceStorage{logger: logger}
}

func (s *NamespaceStorage) GetNamespace(ctx context.Context, fs Filesystem, key string) (*rpcenvironments.Namespace, error) {
	fi, err := fs.OpenFile(path.Join(key, FeaturesFilename), os.O_RDONLY, 0644)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			if key == defaultKey {
				return defaultNamespace, nil
			}

			return nil, errors.ErrNotFoundf("namespace %q", key)
		}

		return nil, err
	}

	defer fi.Close()

	dec := ext.EncodingYAML.NewDecoder(fi)
	for {
		var doc ext.Document
		if err := dec.Decode(&doc); err != nil {
			if errors.Is(err, io.EOF) {
				break
			}

			return nil, err
		}

		if doc.Namespace.GetKey() == key {
			return rpcNamespaceFor(doc.Namespace), nil
		}
	}

	return nil, errors.ErrNotFoundf("namespace %q", key)
}

func (s *NamespaceStorage) ListNamespaces(ctx context.Context, fs Filesystem) (items []*rpcenvironments.Namespace, err error) {
	entries, err := fs.ReadDir(".")
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return []*rpcenvironments.Namespace{}, nil
		}

		return nil, fmt.Errorf("list namespaces: %w", err)
	}

	var defaultSeen bool
	for _, info := range entries {
		if !info.IsDir() {
			continue
		}

		fi, err := fs.OpenFile(path.Join(info.Name(), FeaturesFilename), os.O_RDONLY, 0644)
		if err != nil {
			if !errors.Is(err, os.ErrNotExist) {
				return nil, err
			}

			// disregard directories without features.yaml files
			continue
		}

		defer fi.Close()

		dec := ext.EncodingYAML.NewDecoder(fi)

		for {
			var doc ext.Document
			if err := dec.Decode(&doc); err != nil {
				if errors.Is(err, io.EOF) {
					break
				}

				return nil, err
			}

			defaultSeen = defaultSeen || doc.Namespace.GetKey() == defaultKey

			items = append(items, rpcNamespaceFor(doc.Namespace))
		}
	}

	if !defaultSeen {
		items = append([]*rpcenvironments.Namespace{defaultNamespace}, items...)
	}

	return
}

func (s *NamespaceStorage) PutNamespace(ctx context.Context, fs Filesystem, ns *rpcenvironments.Namespace) error {
	if err := fs.MkdirAll(ns.Key, 0755); err != nil {
		return fmt.Errorf("creating directory %q: %w", ns.Key, err)
	}

	path := path.Join(ns.Key, FeaturesFilename)
	_, err := fs.Stat(path)
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		return err
	}

	fi, err := fs.OpenFile(path, os.O_CREATE|os.O_RDWR|os.O_TRUNC, 0644)
	if err != nil {
		return err
	}

	defer fi.Close()

	return ext.EncodingYAML.NewEncoder(fi).Encode(NewDocumentForNS(extNamespaceFor(ns)))
}

func (s *NamespaceStorage) DeleteNamespace(ctx context.Context, fs Filesystem, key string) error {
	if key == defaultKey {
		return errors.ErrInvalid(`namespace "default" is protected`)
	}

	path := path.Join(key, FeaturesFilename)
	_, err := fs.Stat(path)
	if err != nil {
		// already removed
		if errors.Is(err, os.ErrNotExist) {
			return nil
		}

		return err
	}

	if err := fs.Remove(path); err != nil {
		return err
	}

	return err
}

func extNamespaceFor(ns *rpcenvironments.Namespace) *ext.NamespaceEmbed {
	return &ext.NamespaceEmbed{
		IsNamespace: &ext.Namespace{
			Key:         ns.Key,
			Name:        ns.Name,
			Description: derefOrZero(ns.Description),
		},
	}
}

func rpcNamespaceFor(ns *ext.NamespaceEmbed) *rpcenvironments.Namespace {
	var (
		name        = ns.GetKey()
		description string
	)

	if n, ok := ns.IsNamespace.(*ext.Namespace); ok {
		if n.Name != "" {
			name = n.Name
		}
		description = n.Description
	}

	return &rpcenvironments.Namespace{
		Key:         ns.GetKey(),
		Name:        name,
		Description: &description,
		Protected:   ptr(ns.GetKey() == defaultKey),
	}
}

func ptr[T any](t T) *T {
	return &t
}

func derefOrZero[T any](t *T) (v T) {
	if t == nil {
		return
	}

	return *t
}

func NewDocumentForNS(ns *ext.NamespaceEmbed) *ext.Document {
	version := ext.LatestVersion.FinalizeVersion()
	if ext.LatestVersion.Patch == 0 {
		version = fmt.Sprintf("%d.%d", ext.LatestVersion.Major, ext.LatestVersion.Minor)
	}

	return &ext.Document{
		Version:   version,
		Namespace: ns,
	}
}
