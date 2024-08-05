package oci

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"

	"github.com/opencontainers/go-digest"
	v1 "github.com/opencontainers/image-spec/specs-go/v1"
	"go.flipt.io/flipt/internal/containers"
	"go.flipt.io/flipt/internal/ext"
	"go.flipt.io/flipt/internal/storage"
	storagefs "go.flipt.io/flipt/internal/storage/fs"
	"go.uber.org/zap"
	"oras.land/oras-go/v2"
	"oras.land/oras-go/v2/content"
	"oras.land/oras-go/v2/content/memory"
	"oras.land/oras-go/v2/content/oci"
	"oras.land/oras-go/v2/errdef"
	"oras.land/oras-go/v2/registry"
	"oras.land/oras-go/v2/registry/remote"
	"oras.land/oras-go/v2/registry/remote/auth"
	"oras.land/oras-go/v2/registry/remote/retry"
)

const (
	SchemeHTTP  = "http"
	SchemeHTTPS = "https"
	SchemeFlipt = "flipt"
)

type credentialFunc func(registry string) auth.CredentialFunc

// Store is a type which can retrieve Flipt feature files from a target repository and reference
// Repositories can be local (OCI layout directories on the filesystem) or a remote registry
type Store struct {
	opts   StoreOptions
	logger *zap.Logger
	local  oras.Target
}

// NewStore constructs and configures an instance of *Store for the provided config
func NewStore(logger *zap.Logger, dir string, opts ...containers.Option[StoreOptions]) (*Store, error) {
	store := &Store{
		opts: StoreOptions{
			bundleDir:       dir,
			manifestVersion: oras.PackManifestVersion1_1,
		},
		logger: logger,
		local:  memory.New(),
	}

	containers.ApplyAll(&store.opts, opts...)

	return store, nil
}

type Reference struct {
	registry.Reference
	Scheme string
}

func ParseReference(repository string) (Reference, error) {
	scheme, repository, match := strings.Cut(repository, "://")
	// support empty scheme as remote and https
	if !match {
		repository = scheme
		scheme = SchemeHTTPS
	}

	if !strings.Contains(repository, "/") {
		repository = "local/" + repository
		scheme = SchemeFlipt
	}

	ref, err := registry.ParseReference(repository)
	if err != nil {
		return Reference{}, err
	}

	switch scheme {
	case SchemeHTTP, SchemeHTTPS:
	case SchemeFlipt:
		if ref.Registry != "local" {
			return Reference{}, fmt.Errorf("unexpected local reference: %q", ref)
		}
	default:
		return Reference{}, fmt.Errorf("unexpected repository scheme: %q should be one of [http|https|flipt]", scheme)
	}

	return Reference{
		Reference: ref,
		Scheme:    scheme,
	}, nil
}

func (s *Store) getTarget(ref Reference) (oras.Target, error) {
	switch ref.Scheme {
	case SchemeHTTP, SchemeHTTPS:
		remote, err := remote.NewRepository(fmt.Sprintf("%s/%s", ref.Registry, ref.Repository))
		if err != nil {
			return nil, err
		}

		remote.PlainHTTP = ref.Scheme == "http"

		if s.opts.auth != nil {
			remote.Client = &auth.Client{
				Credential: s.opts.auth(ref.Registry),
				Cache:      s.opts.authCache,
				Client:     retry.DefaultClient,
			}
		}

		return remote, nil
	case SchemeFlipt:
		// build the store once to ensure it is valid
		store, err := oci.New(path.Join(s.opts.bundleDir, ref.Repository))
		if err != nil {
			return nil, err
		}

		store.AutoSaveIndex = true

		return store, nil
	}

	return nil, fmt.Errorf("unexpected repository scheme: %q should be one of [http|https|flipt]", ref.Scheme)
}

// FetchOptions configures a call to Fetch
type FetchOptions struct {
	IfNoMatch digest.Digest
}

// FetchResponse contains any fetched files for the given tracked reference
// If Matched == true, then the supplied IfNoMatch digest matched and Files should be nil
type FetchResponse struct {
	Digest  digest.Digest
	Files   []fs.File
	Matched bool
}

// IfNoMatch configures the call to Fetch to return early if the supplied
// digest matches the target manifest pointed at by the underlying reference
// This is a cache optimization to skip re-fetching resources if the contents
// has already been seen by the caller
func IfNoMatch(digest digest.Digest) containers.Option[FetchOptions] {
	return func(fo *FetchOptions) {
		fo.IfNoMatch = digest
	}
}

// Fetch retrieves the associated files for the tracked repository and reference
// It can optionally be configured to skip fetching given the caller has a digest
// that matches the current reference target
func (s *Store) Fetch(ctx context.Context, ref Reference, opts ...containers.Option[FetchOptions]) (*FetchResponse, error) {
	var options FetchOptions
	containers.ApplyAll(&options, opts...)

	store, err := s.getTarget(ref)
	if err != nil {
		return nil, err
	}

	desc, err := oras.Copy(ctx,
		store,
		ref.Reference.Reference,
		s.local,
		ref.Reference.Reference,
		oras.DefaultCopyOptions)
	if err != nil {
		return nil, err
	}

	bytes, err := content.FetchAll(ctx, s.local, desc)
	if err != nil {
		return nil, err
	}

	var manifest v1.Manifest
	if err = json.Unmarshal(bytes, &manifest); err != nil {
		return nil, err
	}

	var d digest.Digest
	{
		// shadow manifest so that we can safely
		// strip annotations before calculating
		// the digest
		manifest := manifest
		manifest.Annotations = map[string]string{}
		bytes, err := json.Marshal(&manifest)
		if err != nil {
			return nil, err
		}

		d = digest.FromBytes(bytes)
		if d == options.IfNoMatch {
			return &FetchResponse{Matched: true, Digest: d}, nil
		}
	}

	files, err := s.fetchFiles(ctx, store, manifest)
	if err != nil {
		return nil, err
	}

	return &FetchResponse{Files: files, Digest: d}, nil
}

// fetchFiles retrieves the associated flipt feature content files from the content fetcher.
// It traverses the provided manifests and returns a slice of file instances with appropriate
// content type extensions.
func (s *Store) fetchFiles(ctx context.Context, store oras.ReadOnlyTarget, manifest v1.Manifest) ([]fs.File, error) {
	var files []fs.File

	created, err := time.Parse(time.RFC3339, manifest.Annotations[v1.AnnotationCreated])
	if err != nil {
		return nil, err
	}

	for _, layer := range manifest.Layers {
		mediaType, encoding, err := getMediaTypeAndEncoding(layer)
		if err != nil {
			return nil, fmt.Errorf("layer %q: %w", layer.Digest, err)
		}

		if mediaType != MediaTypeFliptNamespace {
			return nil, fmt.Errorf("layer %q: type %q: %w", layer.Digest, mediaType, ErrUnexpectedMediaType)
		}

		switch encoding {
		case "", "json", "yaml", "yml":
		default:
			return nil, fmt.Errorf("layer %q: unexpected layer encoding: %q", layer.Digest, encoding)
		}

		rc, err := store.Fetch(ctx, layer)
		if err != nil {
			return nil, err
		}

		files = append(files, &File{
			ReadCloser: rc,
			info: FileInfo{
				desc:     layer,
				encoding: encoding,
				mod:      created,
			},
		})
	}

	return files, nil
}

// Bundle is a record of an existing Flipt feature bundle
type Bundle struct {
	Digest     digest.Digest
	Repository string
	Tag        string
	CreatedAt  time.Time
}

// List returns a slice of bundles available on the host
func (s *Store) List(ctx context.Context) (bundles []Bundle, _ error) {
	fi, err := os.Open(s.opts.bundleDir)
	if err != nil {
		return nil, err
	}

	defer fi.Close()

	entries, err := fi.ReadDir(-1)
	if err != nil {
		return nil, err
	}

	for _, entry := range entries {
		bytes, err := os.ReadFile(filepath.Join(s.opts.bundleDir, entry.Name(), v1.ImageIndexFile))
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
			path := filepath.Join(s.opts.bundleDir, entry.Name(), "blobs", digest.Algorithm().String(), digest.Hex())
			bytes, err := os.ReadFile(path)
			if err != nil {
				return nil, err
			}

			var man v1.Manifest
			if err := json.Unmarshal(bytes, &man); err != nil {
				return nil, err
			}

			bundle := Bundle{
				Digest:     manifest.Digest,
				Repository: entry.Name(),
				Tag:        manifest.Annotations[v1.AnnotationRefName],
			}

			bundle.CreatedAt, err = parseCreated(man.Annotations)
			if err != nil {
				return nil, err
			}

			bundles = append(bundles, bundle)
		}
	}

	return
}

// Build bundles the target directory Flipt feature state into the target configured on the Store
// It returns a Bundle which contains metadata regarding the resulting bundle details
func (s *Store) Build(ctx context.Context, src fs.FS, ref Reference) (Bundle, error) {
	store, err := s.getTarget(ref)
	if err != nil {
		return Bundle{}, err
	}

	layers, err := s.buildLayers(ctx, store, src)
	if err != nil {
		return Bundle{}, err
	}

	desc, err := oras.PackManifest(ctx, store, s.opts.manifestVersion, MediaTypeFliptFeatures, oras.PackManifestOptions{
		ManifestAnnotations: map[string]string{},
		Layers:              layers,
	})
	if err != nil {
		return Bundle{}, err
	}

	if ref.Reference.Reference != "" {
		if err := store.Tag(ctx, desc, ref.Reference.Reference); err != nil {
			return Bundle{}, err
		}
	}

	bundle := Bundle{
		Digest:     desc.Digest,
		Repository: ref.Repository,
		Tag:        ref.Reference.Reference,
	}

	bundle.CreatedAt, err = parseCreated(desc.Annotations)
	if err != nil {
		return Bundle{}, err
	}

	return bundle, nil
}

func (s *Store) buildLayers(ctx context.Context, store oras.Target, src fs.FS) (layers []v1.Descriptor, _ error) {
	if err := storagefs.WalkDocuments(s.logger, src, func(doc *ext.Document) error {
		payload, err := json.Marshal(&doc)
		if err != nil {
			return err
		}

		var namespaceKey string
		if doc.Namespace == nil {
			namespaceKey = storage.DefaultNamespace
		} else {
			namespaceKey = doc.Namespace.GetKey()
		}

		desc := v1.Descriptor{
			Digest:    digest.FromBytes(payload),
			Size:      int64(len(payload)),
			MediaType: MediaTypeFliptNamespace,
			Annotations: map[string]string{
				AnnotationFliptNamespace: namespaceKey,
			},
		}

		s.logger.Debug("adding layer", zap.String("digest", desc.Digest.Hex()), zap.String("namespace", namespaceKey))

		if err := store.Push(ctx, desc, bytes.NewReader(payload)); err != nil && !errors.Is(err, errdef.ErrAlreadyExists) {
			return err
		}

		layers = append(layers, desc)
		return nil
	}); err != nil {
		return nil, err
	}
	return layers, nil
}

func (s *Store) Copy(ctx context.Context, src, dst Reference) (Bundle, error) {
	if src.Reference.Reference == "" {
		return Bundle{}, fmt.Errorf("source bundle: %w", ErrReferenceRequired)
	}

	if dst.Reference.Reference == "" {
		return Bundle{}, fmt.Errorf("destination bundle: %w", ErrReferenceRequired)
	}

	srcTarget, err := s.getTarget(src)
	if err != nil {
		return Bundle{}, err
	}

	dstTarget, err := s.getTarget(dst)
	if err != nil {
		return Bundle{}, err
	}

	desc, err := oras.Copy(
		ctx,
		srcTarget,
		src.Reference.Reference,
		dstTarget,
		dst.Reference.Reference,
		oras.DefaultCopyOptions)
	if err != nil {
		return Bundle{}, err
	}

	rd, err := dstTarget.Fetch(ctx, desc)
	if err != nil {
		return Bundle{}, err
	}

	var man v1.Manifest
	if err := json.NewDecoder(rd).Decode(&man); err != nil {
		return Bundle{}, err
	}

	bundle := Bundle{
		Digest:     desc.Digest,
		Repository: dst.Repository,
		Tag:        dst.Reference.Reference,
	}

	bundle.CreatedAt, err = parseCreated(man.Annotations)
	if err != nil {
		return Bundle{}, err
	}

	return bundle, nil
}

func getMediaTypeAndEncoding(layer v1.Descriptor) (mediaType, encoding string, _ error) {
	var ok bool
	if mediaType = layer.MediaType; mediaType == "" {
		return "", "", ErrMissingMediaType
	}

	if mediaType, encoding, ok = strings.Cut(mediaType, "+"); !ok {
		encoding = "json"
	}

	return
}

// File is a wrapper around a flipt feature state files contents.
type File struct {
	io.ReadCloser
	info FileInfo
}

// Seek attempts to seek the embedded read-closer.
// If the embedded read closer implements seek, then it delegates
// to that instances implementation. Alternatively, it returns
// an error signifying that the File cannot be seeked.
func (f *File) Seek(offset int64, whence int) (int64, error) {
	if seek, ok := f.ReadCloser.(io.Seeker); ok {
		return seek.Seek(offset, whence)
	}

	return 0, errors.New("seeker cannot seek")
}

func (f *File) Stat() (fs.FileInfo, error) {
	return &f.info, nil
}

// FileInfo describes a flipt features state file instance.
type FileInfo struct {
	desc     v1.Descriptor
	encoding string
	mod      time.Time
}

func (f FileInfo) Name() string {
	return f.desc.Digest.Hex() + "." + f.encoding
}

func (f FileInfo) Size() int64 {
	return f.desc.Size
}

func (f FileInfo) Mode() fs.FileMode {
	return fs.ModePerm
}

func (f FileInfo) ModTime() time.Time {
	return f.mod
}

func (f FileInfo) IsDir() bool {
	return false
}

func (f FileInfo) Sys() any {
	return nil
}

func parseCreated(annotations map[string]string) (time.Time, error) {
	return time.Parse(time.RFC3339, annotations[v1.AnnotationCreated])
}
