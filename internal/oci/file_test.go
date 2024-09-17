package oci

import (
	"bytes"
	"context"
	"embed"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"log"
	"net/http"
	"net/http/httptest"
	"path"
	"strings"
	"testing"
	"time"

	containerregistry "github.com/google/go-containerregistry/pkg/registry"
	"github.com/opencontainers/go-digest"
	v1 "github.com/opencontainers/image-spec/specs-go/v1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap/zaptest"
	"oras.land/oras-go/v2"
	"oras.land/oras-go/v2/content/oci"
	"oras.land/oras-go/v2/registry"
)

const repo = "testrepo"

func TestParseReference(t *testing.T) {
	for _, test := range []struct {
		name        string
		reference   string
		expected    Reference
		expectedErr error
	}{
		{
			name:        "unexpected scheme",
			reference:   "fake://local/something:latest",
			expectedErr: errors.New(`unexpected repository scheme: "fake" should be one of [http|https|flipt]`),
		},
		{
			name:        "invalid local reference",
			reference:   "flipt://invalid/something:latest",
			expectedErr: errors.New(`unexpected local reference: "invalid/something:latest"`),
		},
		{
			name:      "valid local",
			reference: "flipt://local/something:latest",
			expected: Reference{
				Reference: registry.Reference{
					Registry:   "local",
					Repository: "something",
					Reference:  "latest",
				},
				Scheme: "flipt",
			},
		},
		{
			name:      "valid bare local",
			reference: "something:latest",
			expected: Reference{
				Reference: registry.Reference{
					Registry:   "local",
					Repository: "something",
					Reference:  "latest",
				},
				Scheme: "flipt",
			},
		},
		{
			name:      "valid insecure remote",
			reference: "http://remote/something:latest",
			expected: Reference{
				Reference: registry.Reference{
					Registry:   "remote",
					Repository: "something",
					Reference:  "latest",
				},
				Scheme: "http",
			},
		},
		{
			name:      "valid remote",
			reference: "https://remote/something:latest",
			expected: Reference{
				Reference: registry.Reference{
					Registry:   "remote",
					Repository: "something",
					Reference:  "latest",
				},
				Scheme: "https",
			},
		},
		{
			name:      "valid bare remote",
			reference: "remote/something:latest",
			expected: Reference{
				Reference: registry.Reference{
					Registry:   "remote",
					Repository: "something",
					Reference:  "latest",
				},
				Scheme: "https",
			},
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			ref, err := ParseReference(test.reference)
			if test.expectedErr != nil {
				require.Equal(t, test.expectedErr, err)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, test.expected, ref)
		})
	}
}

func TestStore_Fetch_InvalidMediaType(t *testing.T) {
	dir := testRepository(t,
		layer("default", `{"namespace":"default"}`, "unexpected.media.type"),
	)

	ref, err := ParseReference(fmt.Sprintf("flipt://local/%s:latest", repo))
	require.NoError(t, err)

	store, err := NewStore(zaptest.NewLogger(t), dir)
	require.NoError(t, err)

	ctx := context.Background()
	_, err = store.Fetch(ctx, ref)
	require.EqualError(t, err, "layer \"sha256:85ee577ad99c62f314abca9f43ad87c2ee8818513e6383a77690df56d0352748\": type \"unexpected.media.type\": unexpected media type")

	dir = testRepository(t,
		layer("default", `{"namespace":"default"}`, MediaTypeFliptNamespace+"+unknown"),
	)

	store, err = NewStore(zaptest.NewLogger(t), dir)
	require.NoError(t, err)

	_, err = store.Fetch(ctx, ref)
	require.EqualError(t, err, "layer \"sha256:85ee577ad99c62f314abca9f43ad87c2ee8818513e6383a77690df56d0352748\": unexpected layer encoding: \"unknown\"")
}

func TestStore_Fetch(t *testing.T) {
	dir := testRepository(t,
		layer("default", `{"namespace":"default"}`, MediaTypeFliptNamespace),
		layer("other", `namespace: other`, MediaTypeFliptNamespace+"+yaml"),
	)

	ref, err := ParseReference(fmt.Sprintf("flipt://local/%s:latest", repo))
	require.NoError(t, err)

	store, err := NewStore(zaptest.NewLogger(t), dir)
	require.NoError(t, err)

	ctx := context.Background()
	resp, err := store.Fetch(ctx, ref)
	require.NoError(t, err)

	require.False(t, resp.Matched, "matched an empty digest unexpectedly")
	// should remain consistent with contents
	const manifestDigest = digest.Digest("sha256:7cd89519a7f44605a0964cb96e72fef972ebdc0fa4153adac2e8cd2ed5b0e90a")
	assert.Equal(t, manifestDigest, resp.Digest)

	var (
		expected = map[string]string{
			"85ee577ad99c62f314abca9f43ad87c2ee8818513e6383a77690df56d0352748.json": `{"namespace":"default"}`,
			"bbc859ba2a5e9ecc9469a06ae8770b7c0a6e2af2bf16f6bb9184d0244ffd79da.yaml": `namespace: other`,
		}
		found = map[string]string{}
	)

	for _, fi := range resp.Files {
		defer fi.Close()

		stat, err := fi.Stat()
		require.NoError(t, err)

		bytes, err := io.ReadAll(fi)
		require.NoError(t, err)

		found[stat.Name()] = string(bytes)
	}

	assert.Equal(t, expected, found)

	t.Run("IfNoMatch", func(t *testing.T) {
		resp, err = store.Fetch(ctx, ref, IfNoMatch(manifestDigest))
		require.NoError(t, err)

		require.True(t, resp.Matched)
		assert.Equal(t, manifestDigest, resp.Digest)
		assert.Empty(t, resp.Files)
	})
}

//go:embed testdata/*
var testdata embed.FS

func TestStore_Build(t *testing.T) {
	ctx := context.TODO()
	dir := testRepository(t)

	ref, err := ParseReference(fmt.Sprintf("flipt://local/%s:latest", repo))
	require.NoError(t, err)

	store, err := NewStore(zaptest.NewLogger(t), dir)
	require.NoError(t, err)

	testdata, err := fs.Sub(testdata, "testdata")
	require.NoError(t, err)

	bundle, err := store.Build(ctx, testdata, ref)
	require.NoError(t, err)

	assert.Equal(t, repo, bundle.Repository)
	assert.Equal(t, "latest", bundle.Tag)
	assert.NotEmpty(t, bundle.Digest)
	assert.NotEmpty(t, bundle.CreatedAt)

	resp, err := store.Fetch(ctx, ref)
	require.NoError(t, err)
	require.False(t, resp.Matched)

	assert.Len(t, resp.Files, 2)
}

func TestStore_List(t *testing.T) {
	ctx := context.TODO()
	dir := testRepository(t)

	ref, err := ParseReference(fmt.Sprintf("%s:latest", repo))
	require.NoError(t, err)

	store, err := NewStore(zaptest.NewLogger(t), dir)
	require.NoError(t, err)

	bundles, err := store.List(ctx)
	require.NoError(t, err)
	require.Empty(t, bundles)

	testdata, err := fs.Sub(testdata, "testdata")
	require.NoError(t, err)

	bundle, err := store.Build(ctx, testdata, ref)
	require.NoError(t, err)

	t.Log("bundle created digest:", bundle.Digest)

	// sleep long enough for 1 second to pass
	// to bump the timestamp on next build
	time.Sleep(1 * time.Second)

	bundle, err = store.Build(ctx, testdata, ref)
	require.NoError(t, err)

	t.Log("bundle created digest:", bundle.Digest)

	bundles, err = store.List(ctx)
	require.NoError(t, err)
	require.Len(t, bundles, 2)

	assert.Equal(t, "latest", bundles[0].Tag)
	assert.Empty(t, bundles[1].Tag)
}

func TestStore_Copy(t *testing.T) {
	ctx := context.TODO()
	dir := testRepository(t)

	src, err := ParseReference("flipt://local/source:latest")
	require.NoError(t, err)

	store, err := NewStore(zaptest.NewLogger(t), dir)
	require.NoError(t, err)

	testdata, err := fs.Sub(testdata, "testdata")
	require.NoError(t, err)

	_, err = store.Build(ctx, testdata, src)
	require.NoError(t, err)

	for _, test := range []struct {
		name         string
		src          string
		dst          string
		expectedRepo string
		expectedTag  string
		expectedErr  error
	}{
		{
			name:         "valid",
			src:          "flipt://local/source:latest",
			dst:          "flipt://local/target:latest",
			expectedRepo: "target",
			expectedTag:  "latest",
		},
		{
			name:        "invalid source (no reference)",
			src:         "flipt://local/source",
			dst:         "flipt://local/target:latest",
			expectedErr: fmt.Errorf("source bundle: %w", ErrReferenceRequired),
		},
		{
			name:        "invalid destination (no reference)",
			src:         "flipt://local/source:latest",
			dst:         "flipt://local/target",
			expectedErr: fmt.Errorf("destination bundle: %w", ErrReferenceRequired),
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			src, err := ParseReference(test.src)
			require.NoError(t, err)

			dst, err := ParseReference(test.dst)
			require.NoError(t, err)

			bundle, err := store.Copy(ctx, src, dst)
			if test.expectedErr != nil {
				require.Equal(t, test.expectedErr, err)
				return
			}

			require.NoError(t, err)

			assert.Equal(t, test.expectedRepo, bundle.Repository)
			assert.Equal(t, test.expectedTag, bundle.Tag)
			assert.NotEmpty(t, bundle.Digest)
			assert.NotEmpty(t, bundle.CreatedAt)

			resp, err := store.Fetch(ctx, dst)
			require.NoError(t, err)
			require.False(t, resp.Matched)

			assert.Len(t, resp.Files, 2)
		})
	}
}

func TestFile(t *testing.T) {
	var (
		rd   = strings.NewReader("contents")
		mod  = time.Date(2023, 11, 9, 12, 0, 0, 0, time.UTC)
		info = FileInfo{
			desc: v1.Descriptor{
				Digest: digest.FromString("contents"),
				Size:   rd.Size(),
			},
			encoding: "json",
			mod:      mod,
		}
		fi = File{
			ReadCloser: readCloseSeeker{rd},
			info:       info,
		}
	)

	stat, err := fi.Stat()
	require.NoError(t, err)

	assert.Equal(t, "d1b2a59fbea7e20077af9f91b27e95e865061b270be03ff539ab3b73587882e8.json", stat.Name())
	assert.Equal(t, fs.ModePerm, stat.Mode())
	assert.Equal(t, mod, stat.ModTime())
	assert.False(t, stat.IsDir())
	assert.Nil(t, stat.Sys())

	count, err := fi.Seek(3, io.SeekStart)
	require.NoError(t, err)
	assert.Equal(t, int64(3), count)

	data, err := io.ReadAll(fi)
	require.NoError(t, err)
	assert.Equal(t, "tents", string(data))

	// rewind reader
	_, err = rd.Seek(0, io.SeekStart)
	require.NoError(t, err)

	// ensure seeker cannot seek
	fi = File{
		ReadCloser: io.NopCloser(rd),
		info:       info,
	}

	_, err = fi.Seek(3, io.SeekStart)
	require.EqualError(t, err, "seeker cannot seek")
}

func TestStore_FetchWithECR(t *testing.T) {
	registryURI, endpoint := testEcrStub(t, time.Second)

	dir := testRepository(t)

	src, err := ParseReference("flipt://local/source:latest")
	require.NoError(t, err)

	store, err := NewStore(zaptest.NewLogger(t), dir, WithAWSECRCredentials(endpoint))
	require.NoError(t, err)

	testdata, err := fs.Sub(testdata, "testdata")
	require.NoError(t, err)

	_, err = store.Build(context.Background(), testdata, src)
	require.NoError(t, err)

	dst, err := ParseReference(fmt.Sprintf("%s/%s:latest", registryURI, repo))
	require.NoError(t, err)
	_, err = store.Copy(context.Background(), src, dst)
	require.NoError(t, err)

	_, err = store.Fetch(context.Background(), dst)
	require.NoError(t, err)
	time.Sleep(time.Second)
	_, err = store.Fetch(context.Background(), dst)
	require.NoError(t, err)
}

type readCloseSeeker struct {
	io.ReadSeeker
}

func (r readCloseSeeker) Close() error { return nil }

func layer(ns, payload, mediaType string) func(*testing.T, oras.Target) v1.Descriptor {
	return func(t *testing.T, store oras.Target) v1.Descriptor {
		t.Helper()

		desc := v1.Descriptor{
			Digest:    digest.FromString(payload),
			Size:      int64(len(payload)),
			MediaType: mediaType,
			Annotations: map[string]string{
				AnnotationFliptNamespace: ns,
			},
		}

		require.NoError(t, store.Push(context.TODO(), desc, bytes.NewReader([]byte(payload))))

		return desc
	}
}

func testRepository(t *testing.T, layerFuncs ...func(*testing.T, oras.Target) v1.Descriptor) (dir string) {
	t.Helper()

	dir = t.TempDir()

	t.Log("test OCI directory", dir, repo)

	store, err := oci.New(path.Join(dir, repo))
	require.NoError(t, err)

	store.AutoSaveIndex = true

	ctx := context.TODO()

	if len(layerFuncs) == 0 {
		return
	}

	var layers []v1.Descriptor
	for _, fn := range layerFuncs {
		layers = append(layers, fn(t, store))
	}

	desc, err := oras.PackManifest(ctx, store, oras.PackManifestVersion1_1, MediaTypeFliptFeatures, oras.PackManifestOptions{
		ManifestAnnotations: map[string]string{},
		Layers:              layers,
	})
	require.NoError(t, err)

	require.NoError(t, store.Tag(ctx, desc, "latest"))

	return
}

// testEcrStub is a stub for AWS ECR private service. It allows to get the auth token and
// push/pull from registry.
func testEcrStub(t testing.TB, tokenTtl time.Duration) (registry string, endpoint string) {
	t.Helper()
	t.Setenv("AWS_ACCESS_KEY_ID", "key")
	t.Setenv("AWS_SECRET_ACCESS_KEY", "secret")
	token := base64.RawStdEncoding.EncodeToString([]byte("user_name:password"))

	authMiddleware := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Header.Get("Authorization") != "Basic "+token {
				w.Header().Add("Www-Authenticate", "Basic realm=private-ecr")
				w.WriteHeader(http.StatusUnauthorized)
				return
			}
			next.ServeHTTP(w, r)
		})
	}

	mux := http.NewServeMux()
	// registry endpoint
	mux.Handle("/v2/", authMiddleware(containerregistry.New(containerregistry.Logger(log.New(io.Discard, "", 0)))))
	// aws ecr endpoint
	mux.HandleFunc("POST /aws-api", func(w http.ResponseWriter, r *http.Request) {
		amzDate := r.Header.Get("X-Amz-Date")
		requestTime, _ := time.Parse("20060102T150405Z", amzDate)
		requestTime = requestTime.Add(tokenTtl)
		amzTarget := r.Header.Get("X-Amz-Target")
		switch amzTarget {
		case "AmazonEC2ContainerRegistry_V20150921.GetAuthorizationToken":
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(fmt.Sprintf(`{"authorizationData": [{"authorizationToken": "%s", "expiresAt": %d}]}`, token, requestTime.Unix())))
		default:
			w.WriteHeader(http.StatusBadRequest)
		}
	})
	httpServer := httptest.NewServer(mux)
	t.Cleanup(httpServer.Close)
	registry = httpServer.URL
	endpoint = httpServer.URL + "/aws-api"
	return
}
