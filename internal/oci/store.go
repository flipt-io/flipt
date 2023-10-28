package oci

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/opencontainers/go-digest"
	v1 "github.com/opencontainers/image-spec/specs-go/v1"
	"go.flipt.io/flipt/internal/containers"
	"go.uber.org/zap"
	"oras.land/oras-go/v2"
	"oras.land/oras-go/v2/content/oci"
	"oras.land/oras-go/v2/registry"
	"oras.land/oras-go/v2/registry/remote"
)

type Bunch struct {
	Digest     digest.Digest
	Repository string
	Tag        string
	CreatedAt  time.Time
}

type Store struct {
	dir string
}

func NewStore(dir string) (*Store, error) {
	stat, err := os.Stat(dir)
	if err != nil {
		return nil, err
	}

	if !stat.IsDir() {
		return nil, fmt.Errorf("expected directory at %q, found %q", dir, stat.Mode())
	}

	return &Store{dir}, nil
}

func (s *Store) List() (bunches []Bunch, _ error) {
	fi, err := os.Open(s.dir)
	if err != nil {
		return nil, err
	}

	defer fi.Close()

	entries, err := fi.ReadDir(-1)
	if err != nil {
		return nil, err
	}

	for _, entry := range entries {
		bytes, err := os.ReadFile(filepath.Join(s.dir, entry.Name(), v1.ImageIndexFile))
		if err != nil {
			if errors.Is(err, os.ErrNotExist) {
				return nil, nil
			}

			return nil, err
		}

		var index v1.Index
		if err := json.Unmarshal(bytes, &index); err != nil {
			return nil, err
		}

		for _, manifest := range index.Manifests {
			digest := manifest.Digest
			path := filepath.Join(s.dir, entry.Name(), "blobs", digest.Algorithm().String(), digest.Hex())
			bytes, err := os.ReadFile(path)
			if err != nil {
				return nil, err
			}

			var man v1.Manifest
			if err := json.Unmarshal(bytes, &man); err != nil {
				return nil, err
			}

			bunch := Bunch{
				Digest:     manifest.Digest,
				Repository: entry.Name(),
				Tag:        manifest.Annotations[v1.AnnotationRefName],
			}

			bunch.CreatedAt, err = parseCreated(man.Annotations)
			if err != nil {
				return nil, err
			}

			bunches = append(bunches, bunch)
		}
	}

	return
}

type CreateOption struct {
	Ref string
}

func WithRef(ref string) containers.Option[CreateOption] {
	return func(co *CreateOption) {
		co.Ref = ref
	}
}

func (s *Store) Create(ctx context.Context, src fs.FS, opts ...containers.Option[CreateOption]) (Bunch, error) {
	var options CreateOption
	containers.ApplyAll(&options, opts...)

	repo, tag, _ := strings.Cut(options.Ref, ":")

	store, err := oci.New(filepath.Join(s.dir, repo))
	if err != nil {
		return Bunch{}, err
	}

	store.AutoSaveIndex = true

	layers, err := NewPackager(zap.NewNop()).Package(ctx, store, src)
	if err != nil {
		return Bunch{}, err
	}

	desc, err := oras.PackManifest(ctx, store, oras.PackManifestVersion1_1_RC4, MediaTypeFliptState, oras.PackManifestOptions{
		ManifestAnnotations: map[string]string{},
		Layers:              layers,
	})
	if err != nil {
		return Bunch{}, err
	}

	if tag != "" {
		if err := store.Tag(ctx, desc, tag); err != nil {
			return Bunch{}, err
		}
	}

	bunch := Bunch{
		Digest:     desc.Digest,
		Repository: repo,
		Tag:        tag,
	}

	bunch.CreatedAt, err = parseCreated(desc.Annotations)
	if err != nil {
		return Bunch{}, err
	}

	return bunch, nil
}

func (s *Store) Push(ctx context.Context, ref string) error {
	reference, err := registry.ParseReference(ref)
	if err != nil {
		return err
	}

	repository, err := remote.NewRepository(fmt.Sprintf("%s/%s", reference.Registry, reference.Repository))
	if err != nil {
		return err
	}

	repository.PlainHTTP = true

	store, err := oci.New(filepath.Join(s.dir, reference.Repository))
	if err != nil {
		return err
	}

	store.AutoSaveIndex = true

	desc, err := oras.Copy(ctx, store, reference.Reference, repository, reference.Reference, oras.DefaultCopyOptions)
	if err != nil {
		return err
	}

	fmt.Println(desc.Digest)

	return nil
}

func (s *Store) Pull(ctx context.Context, ref string) error {
	reference, err := registry.ParseReference(ref)
	if err != nil {
		return err
	}

	repository, err := remote.NewRepository(fmt.Sprintf("%s/%s", reference.Registry, reference.Repository))
	if err != nil {
		return err
	}

	repository.PlainHTTP = true

	store, err := oci.New(filepath.Join(s.dir, reference.Repository))
	if err != nil {
		return err
	}

	store.AutoSaveIndex = true

	desc, err := oras.Copy(ctx, repository, reference.Reference, store, reference.Reference, oras.DefaultCopyOptions)
	if err != nil {
		return err
	}

	fmt.Println(desc.Digest)

	return nil
}

func parseCreated(annotations map[string]string) (time.Time, error) {
	return time.Parse(time.RFC3339, annotations[v1.AnnotationCreated])
}
