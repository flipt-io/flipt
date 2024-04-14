package object

import (
	"context"
	"embed"
	"fmt"
	"io/fs"
	"net/url"
	"os"
	"strconv"
	"testing"
	"time"

	gstorage "cloud.google.com/go/storage"
	"github.com/Azure/azure-sdk-for-go/sdk/storage/azblob"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/stretchr/testify/require"
	"go.flipt.io/flipt/internal/storage"
	storagefs "go.flipt.io/flipt/internal/storage/fs"
	"go.uber.org/zap/zaptest"
	gcblob "gocloud.dev/blob"
	"gocloud.dev/blob/azureblob"
	"gocloud.dev/blob/fileblob"
	"gocloud.dev/blob/gcsblob"
	"gocloud.dev/blob/memblob"
)

var (
	bucketName = "testdata"
	//go:embed all:testdata/*
	testdata embed.FS
)

func Test_Store(t *testing.T) {
	t.Run(memblob.Scheme, func(t *testing.T) {
		testStore(t, func(t *testing.T) string {
			t.Helper()
			return fmt.Sprintf("%s://%s", memblob.Scheme, bucketName)
		})
	})

	t.Run(fileblob.Scheme, func(t *testing.T) {
		testStore(t, func(t *testing.T) string {
			t.Helper()
			return fmt.Sprintf("%s://%s", fileblob.Scheme, t.TempDir())
		})
	})

	t.Run("s3", func(t *testing.T) {
		minioURL := os.Getenv("TEST_S3_ENDPOINT")
		if minioURL == "" {
			t.Skip("set TEST_S3_ENDPOINT to run this case")
			return
		}

		ctx := context.Background()
		client := s3Client(t, minioURL)
		_, err := client.CreateBucket(ctx, &s3.CreateBucketInput{
			Bucket: &bucketName,
		})
		require.NoError(t, err)

		t.Cleanup(func() {
			_, _ = client.DeleteBucket(ctx, &s3.DeleteBucketInput{
				Bucket: &bucketName,
			})
		})

		testStore(t, func(t *testing.T) string {
			t.Helper()
			q := url.Values{}
			q.Set("awssdk", "v2")
			q.Set("region", "minio")
			q.Set("endpoint", minioURL)
			s3 := &url.URL{
				Scheme:   "s3",
				Host:     bucketName,
				RawQuery: q.Encode(),
			}

			return s3.String()
		})
	})

	t.Run("azure", func(t *testing.T) {
		azuriteURL := os.Getenv("TEST_AZURE_ENDPOINT")
		if azuriteURL == "" {
			t.Skip("set TEST_AZURE_ENDPOINT to run this case")
			return
		}

		u, err := url.Parse(azuriteURL)
		require.NoError(t, err)

		os.Setenv("AZURE_STORAGE_PROTOCOL", u.Scheme)
		os.Setenv("AZURE_STORAGE_IS_LOCAL_EMULATOR", strconv.FormatBool(u.Scheme == "http"))
		os.Setenv("AZURE_STORAGE_DOMAIN", u.Host)

		ctx := context.Background()
		client := azureClient(t, azuriteURL)
		_, err = client.CreateContainer(ctx, bucketName, &azblob.CreateContainerOptions{})
		require.NoError(t, err)

		t.Cleanup(func() {
			_, _ = client.DeleteContainer(ctx, bucketName, &azblob.DeleteContainerOptions{})
		})

		testStore(t, func(t *testing.T) string {
			t.Helper()

			s3 := &url.URL{
				Scheme: azureblob.Scheme,
				Host:   bucketName,
			}

			return s3.String()
		})
	})

	t.Run("gcs", func(t *testing.T) {
		emulatorURL := os.Getenv("STORAGE_EMULATOR_HOST")
		if emulatorURL == "" {
			t.Skip("set STORAGE_EMULATOR_HOST to run this case")
			return
		}

		ctx := context.Background()

		client := gcsClient(t)
		bucket := client.Bucket(bucketName)
		require.NoError(t, bucket.Create(ctx, "", nil))

		t.Cleanup(func() {
			_ = bucket.Delete(ctx)
		})

		testStore(t, func(t *testing.T) string {
			t.Helper()
			return fmt.Sprintf("%s://%s", gcsblob.Scheme, bucketName)
		})
	})
}

func testStore(t *testing.T, fn func(t *testing.T) string) {
	t.Helper()

	ctx := context.TODO()

	t.Run("WithoutPrefix", func(t *testing.T) {
		ch := make(chan struct{})

		dest := fn(t)

		t.Log("opening bucket", dest)
		u, err := url.Parse(dest)
		require.NoError(t, err)
		bucket, err := OpenBucket(ctx, u)
		require.NoError(t, err)

		t.Cleanup(func() { _ = bucket.Close() })

		writeTestDataToBucket(t, ctx, bucket)

		t.Log("Creating store to test")

		store, err := NewSnapshotStore(
			ctx,
			zaptest.NewLogger(t),
			u.Scheme,
			bucket,
			WithPollOptions(
				storagefs.WithInterval(time.Second),
				storagefs.WithNotify(t, func(modified bool) {
					if modified {
						close(ch)
					}
				}),
			),
		)
		require.NoError(t, err)

		t.Cleanup(func() {
			_ = store.Close()
		})

		// both the production and prefix namespaces should be present
		// however, both should be empty
		require.NoError(t, store.View(ctx, func(s storage.ReadOnlyStore) error {
			_, err := s.GetNamespace(ctx, storage.NewNamespace("production"))
			require.NoError(t, err)

			_, err = s.GetNamespace(ctx, storage.NewNamespace("prefix"))
			require.NoError(t, err)

			_, err = s.GetFlag(ctx, storage.NewResource("production", "foo"))
			require.Error(t, err, "flag should not be defined yet")

			return nil
		}))

		// update features.yml
		path := "features.yml"
		require.NoError(t, bucket.WriteAll(context.TODO(), path,
			[]byte(`namespace: production
flags:
    - key: foo
      name: Foo`),
			&gcblob.WriterOptions{}))

		// assert matching state
		select {
		case <-ch:
		case <-time.After(time.Minute):
			t.Fatal("timed out waiting for update")
		}

		t.Log("received new snapshot")

		require.NoError(t, store.View(ctx, func(s storage.ReadOnlyStore) error {
			_, err = s.GetNamespace(context.TODO(), storage.NewNamespace("production"))
			if err != nil {
				return err
			}

			_, err = s.GetFlag(context.TODO(), storage.NewResource("production", "foo"))
			if err != nil {
				return err
			}

			_, err = s.GetNamespace(context.TODO(), storage.NewNamespace("prefix"))
			return err
		}))
	})

	t.Run("WithPrefix", func(t *testing.T) {
		dest := fn(t)

		t.Log("opening bucket", dest)
		u, err := url.Parse(dest)
		require.NoError(t, err)
		bucket, err := OpenBucket(ctx, u)
		require.NoError(t, err)

		t.Cleanup(func() { _ = bucket.Close() })

		writeTestDataToBucket(t, ctx, bucket)

		t.Log("Creating store to test")

		store, err := NewSnapshotStore(
			ctx,
			zaptest.NewLogger(t),
			u.Scheme,
			bucket,
			WithPrefix("prefix"),
		)
		require.NoError(t, err)

		t.Cleanup(func() {
			_ = store.Close()
		})

		// both the production and prefix namespaces should be present
		// however, both should be empty
		require.NoError(t, store.View(ctx, func(s storage.ReadOnlyStore) error {
			_, err := s.GetNamespace(ctx, storage.NewNamespace("production"))
			require.Error(t, err, "should not exist in prefixed bucket")

			_, err = s.GetFlag(ctx, storage.NewResource("production", "foo"))
			require.Error(t, err, "flag should not be defined yet")

			_, err = s.GetNamespace(ctx, storage.NewNamespace("prefix"))
			require.NoError(t, err)

			return nil
		}))
	})
}

func writeTestDataToBucket(t *testing.T, ctx context.Context, bucket *gcblob.Bucket) {
	t.Helper()

	t.Log("Adding testdata contents to target bucket")

	src, err := fs.Sub(testdata, "testdata")
	require.NoError(t, err)

	// copy testdata into target bucket
	err = fs.WalkDir(src, ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if d.IsDir() {
			return nil
		}

		t.Log("Adding key to bucket", path)

		bytes, err := fs.ReadFile(src, path)
		require.NoError(t, err)

		err = bucket.WriteAll(ctx, path, bytes, &gcblob.WriterOptions{})
		require.NoError(t, err)

		return nil
	})
	require.NoError(t, err)
}

func s3Client(t *testing.T, endpoint string) *s3.Client {
	t.Helper()

	cfg, err := config.LoadDefaultConfig(context.Background(),
		config.WithRegion("minio"))
	require.NoError(t, err)

	var s3Opts []func(*s3.Options)
	if endpoint != "" {
		s3Opts = append(s3Opts, func(o *s3.Options) {
			o.BaseEndpoint = &endpoint
			o.UsePathStyle = true
			o.Region = "minio"
		})
	}
	return s3.NewFromConfig(cfg, s3Opts...)
}

func azureClient(t *testing.T, endpoint string) *azblob.Client {
	t.Helper()
	account := os.Getenv("AZURE_STORAGE_ACCOUNT")
	sharedKey := os.Getenv("AZURE_STORAGE_KEY")
	credentials, err := azblob.NewSharedKeyCredential(account, sharedKey)
	require.NoError(t, err)
	client, err := azblob.NewClientWithSharedKeyCredential(endpoint, credentials, nil)
	require.NoError(t, err)
	return client
}

func gcsClient(t *testing.T) *gstorage.Client {
	t.Helper()
	client, err := gstorage.NewClient(context.Background())
	require.NoError(t, err)
	return client
}
