package oci

import "errors"

const (
	// MediaTypeFliptFeatures is the OCI media type for a flipt features artifact
	MediaTypeFliptFeatures = "application/vnd.io.flipt.features.v1"
	// MediaTypeFliptNamespace is the OCI media type for a flipt features namespace artifact
	MediaTypeFliptNamespace = "application/vnd.io.flipt.features.namespace.v1"

	// AnnotationFliptNamespace is an OCI annotation key which identifies the namespace key
	// of the annotated flipt namespace artifact
	AnnotationFliptNamespace = "io.flipt.features.namespace"
)

var (
	// ErrMissingMediaType is returned when a descriptor is presented
	// without a media type
	ErrMissingMediaType = errors.New("missing media type")
	// ErrUnexpectedMediaType is returned when an unexpected media type
	// is found on a target manifest or descriptor
	ErrUnexpectedMediaType = errors.New("unexpected media type")
	// ErrReferenceRequired is returned when a referenced is required for
	// a particular operation
	ErrReferenceRequired = errors.New("reference required")
)
