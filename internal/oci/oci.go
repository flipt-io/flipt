package oci

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"io/fs"

	"github.com/opencontainers/go-digest"
	v1 "github.com/opencontainers/image-spec/specs-go/v1"
	"go.flipt.io/flipt/internal/cue"
	"go.flipt.io/flipt/internal/ext"
	storefs "go.flipt.io/flipt/internal/storage/fs"
	"go.uber.org/zap"
	"gopkg.in/yaml.v3"
	"oras.land/oras-go/v2"
)

const (
	// MediaTypeFliptState is the OCI media type for a flipt state artifact
	MediaTypeFliptState = "application/vnd.io.flipt.state+json"
	// MediaTypeFliptNamespace is the OCI media type for a flipt state namespace artifact
	MediaTypeFliptNamespace = "application/vnd.io.flipt.state.namespace+json"

	// AnnotationFliptNamespace is an OCI annotation key which identifies the namespace key
	// of the annotated flipt namespace artifact
	AnnotationFliptNamespace = "io.flipt.state.namespace"
)

type Packager struct {
	logger *zap.Logger
}

func NewPackager(logger *zap.Logger) *Packager {
	return &Packager{logger}
}

func (p *Packager) Package(ctx context.Context, store oras.Target, src fs.FS) (v1.Descriptor, error) {
	validator, err := cue.NewFeaturesValidator()
	if err != nil {
		return v1.Descriptor{}, err
	}

	paths, err := storefs.ListStateFiles(p.logger, src)
	if err != nil {
		return v1.Descriptor{}, err
	}

	var layers []v1.Descriptor

	for _, file := range paths {
		p.logger.Debug("opening state file", zap.String("path", file))

		fi, err := src.Open(file)
		if err != nil {
			return v1.Descriptor{}, err
		}
		defer fi.Close()

		buf := &bytes.Buffer{}
		reader := io.TeeReader(fi, buf)

		if err := validator.Validate(file, reader); err != nil {
			return v1.Descriptor{}, err
		}

		decoder := yaml.NewDecoder(buf)
		for {
			var doc ext.Document

			if err := decoder.Decode(&doc); err != nil {
				if errors.Is(err, io.EOF) {
					break
				}
				return v1.Descriptor{}, err
			}

			// set namespace to default if empty in document
			if doc.Namespace == "" {
				doc.Namespace = "default"
			}

			payload, err := json.Marshal(&doc)
			if err != nil {
				return v1.Descriptor{}, nil
			}

			desc := v1.Descriptor{
				Digest: digest.FromBytes(payload),
				Size:   int64(len(payload)),
				Annotations: map[string]string{
					AnnotationFliptNamespace: doc.Namespace,
				},
			}

			p.logger.Debug("adding layer", zap.String("digest", desc.Digest.Hex()), zap.String("namespace", doc.Namespace))

			if err := store.Push(ctx, desc, bytes.NewReader(payload)); err != nil {
				return v1.Descriptor{}, err
			}

			layers = append(layers, desc)
		}
	}

	return oras.PackManifest(ctx, store, oras.PackManifestVersion1_1_RC4, MediaTypeFliptState, oras.PackManifestOptions{
		ManifestAnnotations: map[string]string{},
		Layers:              layers,
	})
}
