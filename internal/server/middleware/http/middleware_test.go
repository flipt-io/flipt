package http_middleware

import (
	"context"
	"net/http/httptest"
	"testing"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/metadata"
	pb "google.golang.org/protobuf/types/known/emptypb"
)

func TestHttpResponseModifier(t *testing.T) {
	t.Run("etag header exists", func(t *testing.T) {
		md := runtime.ServerMetadata{
			HeaderMD: metadata.Pairs(
				"foo", "bar",
				"baz", "qux",
				"x-etag", "etag",
			),
		}

		var (
			ctx  = runtime.NewServerMetadataContext(context.Background(), md)
			resp = httptest.NewRecorder()
			msg  = &pb.Empty{}
		)

		err := HttpResponseModifier(ctx, resp, msg)
		require.NoError(t, err)

		w := resp.Result()
		defer w.Body.Close()

		assert.NotEmpty(t, w.Header)

		assert.Equal(t, "etag", w.Header.Get("Etag"))
		assert.Empty(t, w.Header.Get("Grpc-Metadata-X-Etag"))
	})

	t.Run("http code header exists", func(t *testing.T) {
		md := runtime.ServerMetadata{
			HeaderMD: metadata.Pairs(
				"foo", "bar",
				"baz", "qux",
				"x-http-code", "300",
			),
		}

		var (
			ctx  = runtime.NewServerMetadataContext(context.Background(), md)
			resp = httptest.NewRecorder()
			msg  = &pb.Empty{}
		)

		err := HttpResponseModifier(ctx, resp, msg)
		require.NoError(t, err)

		w := resp.Result()
		defer w.Body.Close()

		assert.Empty(t, w.Header)
		assert.Equal(t, 300, w.StatusCode)
	})
}
