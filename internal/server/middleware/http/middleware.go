package http_middleware

import (
	"context"
	"net/http"
	"strconv"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"google.golang.org/protobuf/proto"
)

func HttpResponseModifier(ctx context.Context, w http.ResponseWriter, _ proto.Message) error {
	md, ok := runtime.ServerMetadataFromContext(ctx)
	if !ok {
		return nil
	}

	// set etag header if it exists
	if vals := md.HeaderMD.Get("x-etag"); len(vals) > 0 {
		// delete the headers to not expose any grpc-metadata in http response
		delete(md.HeaderMD, "x-etag")
		delete(w.Header(), "Grpc-Metadata-X-Etag")
		w.Header().Set("Etag", vals[0])
	}

	// check if we set a custom status code
	if vals := md.HeaderMD.Get("x-http-code"); len(vals) > 0 {
		// delete the headers to not expose any grpc-metadata in http response
		delete(md.HeaderMD, "x-http-code")
		delete(w.Header(), "Grpc-Metadata-X-Http-Code")

		code, _ := strconv.Atoi(vals[0])
		w.WriteHeader(code)
	}

	return nil
}
