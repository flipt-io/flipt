package flipt

import (
	"context"
	"fmt"
	"io"
	"os"
	"strings"

	"go.flipt.io/flipt/errors"
	"go.flipt.io/flipt/internal/ext"
	environmentsfs "go.flipt.io/flipt/internal/storage/environments/fs"
	"go.flipt.io/flipt/rpc/flipt"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/anypb"
	"gopkg.in/yaml.v3"
)


func getDocsAndNamespace(ctx context.Context, fs environmentsfs.Filesystem, key string) (docs []*ext.Document, idx int, err error) {
	docs, err = parseNamespace(ctx, fs, key)
	if err != nil {
		return nil, -1, err
	}

	var (
		found bool
		doc   *ext.Document
	)

	for idx, doc = range docs {
		if found = doc.Namespace.GetKey() == key; found {
			break
		}
	}

	if !found {
		if key != flipt.DefaultNamespace {
			return nil, -1, fmt.Errorf("namespace not found: %q", key)
		}

		// we only support autocreating the default namespace as it should
		// always exist from an API perspective
		// otherwise, we expect the caller to create it
		idx = 0
		docs = append(docs, environmentsfs.NewDocumentForNS(ext.DefaultNamespace))
	}

	return
}

func parseNamespace(_ context.Context, fs environmentsfs.Filesystem, namespace string) (docs []*ext.Document, err error) {
	// Try to open features file with both .yaml and .yml extensions
	fi, _, err := environmentsfs.TryOpenFeaturesFile(fs, namespace)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, nil
		}

		return nil, err
	}
	defer fi.Close()

	decoder := yaml.NewDecoder(fi)
	for {
		doc := &ext.Document{}
		if err = decoder.Decode(doc); err != nil {
			if errors.Is(err, io.EOF) {
				err = nil
				break
			}

			return nil, err
		}

		// set namespace to default if empty in document
		if doc.Namespace.GetKey() == "" {
			doc.Namespace = ext.DefaultNamespace
		}

		docs = append(docs, doc)
	}

	return
}

func newAny(msg proto.Message) (*anypb.Any, error) {
	a, err := anypb.New(msg)
	if err != nil {
		return nil, err
	}

	a.TypeUrl = strings.TrimPrefix(a.TypeUrl, "type.googleapis.com/")
	return a, nil
}
