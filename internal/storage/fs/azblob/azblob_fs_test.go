package azblob

import (
	"context"
	"io"
	"io/fs"
	"strings"
	"testing"
	"time"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/runtime"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/to"
	"github.com/Azure/azure-sdk-for-go/sdk/storage/azblob"
	"github.com/Azure/azure-sdk-for-go/sdk/storage/azblob/blob"
	"github.com/Azure/azure-sdk-for-go/sdk/storage/azblob/bloberror"
	"github.com/Azure/azure-sdk-for-go/sdk/storage/azblob/container"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"go.flipt.io/flipt/internal/storage/fs/azblob/mocks"
	"go.uber.org/zap/zaptest"
)

func newBlob(name string) *container.BlobItem {
	modified := time.Now()
	return &container.BlobItem{
		Name: to.Ptr(name),
		Properties: &container.BlobProperties{
			ContentLength: to.Ptr(int64(len(name + "data"))),
			LastModified:  &modified,
		},
	}
}

func newDownloadStreamResponse(name string) blob.DownloadStreamResponse {
	b := newBlob(name)
	response := blob.DownloadStreamResponse{}
	response.Body = io.NopCloser(strings.NewReader(name + "data"))
	response.ContentLength = b.Properties.ContentLength
	response.LastModified = b.Properties.LastModified
	return response
}

func Test_FS(t *testing.T) {
	containerName := "test-container"
	logger := zaptest.NewLogger(t)
	// run with no prefix, returning all files
	t.Run("Ensure invalid and non existent paths produce an error", func(t *testing.T) {
		// setup mocks
		mockClient := mocks.NewClientAPI(t)
		azfs, err := NewFS(logger, mockClient, containerName)
		require.NoError(t, err)

		var options *blob.DownloadStreamOptions
		mockClient.On("DownloadStream", mock.Anything, containerName, "zero.txt", options).
			Return(blob.DownloadStreamResponse{}, &azcore.ResponseError{ErrorCode: string(bloberror.BlobNotFound)})
		// running test
		_, err = azfs.Open("..")
		require.Equal(t, &fs.PathError{
			Op:   "Open",
			Path: "..",
			Err:  fs.ErrInvalid,
		}, err)

		_, err = azfs.Open("zero.txt")
		require.Equal(t, &fs.PathError{
			Op:   "Open",
			Path: "zero.txt",
			Err:  fs.ErrNotExist,
		}, err)

		mockClient.AssertExpectations(t)
	})

	t.Run("Ensure files exist with expected contents", func(t *testing.T) {
		// setup the mock
		mockClient := mocks.NewClientAPI(t)
		azfs, err := NewFS(logger, mockClient, containerName)
		require.NoError(t, err)
		objectChunks := [][]string{{"one"}, {"two"}}

		walker := runtime.PagingHandler[azblob.ListBlobsFlatResponse]{}
		walker.More = func(r azblob.ListBlobsFlatResponse) bool {
			return len(objectChunks) > 0
		}
		walker.Fetcher = func(ctx context.Context, r *azblob.ListBlobsFlatResponse) (azblob.ListBlobsFlatResponse, error) {
			b := container.ListBlobsFlatResponse{}
			b.Segment = &container.BlobFlatListSegment{}
			chunk := objectChunks[0]
			objectChunks = objectChunks[1:]

			b.Segment.BlobItems = []*container.BlobItem{}
			for _, c := range chunk {
				b.Segment.BlobItems = append(b.Segment.BlobItems, newBlob(c))
			}
			return b, nil
		}
		pager := runtime.NewPager[azblob.ListBlobsFlatResponse](walker)
		var options *container.ListBlobsFlatOptions
		mockClient.On("NewListBlobsFlatPager", containerName, options).Return(pager)

		for _, chunk := range objectChunks {
			for _, name := range chunk {
				var downloadOptions *blob.DownloadStreamOptions
				mockClient.On("DownloadStream", mock.Anything, containerName, name, downloadOptions).
					Return(newDownloadStreamResponse(name), nil)
			}
		}

		// running test
		seen := map[string]string{}
		err = fs.WalkDir(azfs, ".", func(path string, d fs.DirEntry, err error) error {
			if err != nil {
				return err
			}

			fi, err := azfs.Open(path)
			require.NoError(t, err)

			defer fi.Close()

			contents, err := io.ReadAll(fi)
			require.NoError(t, err)

			seen[path] = string(contents)

			return nil
		})
		require.NoError(t, err)

		expected := map[string]string{
			".":   "",
			"one": "onedata",
			"two": "twodata",
		}
		require.Equal(t, expected, seen)
		mockClient.AssertExpectations(t)
	})
}
