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

// HandleNoBodyResponse is a response modifier that does not write a body if the response is a 204 No Content or 304 Not Modified.
func HandleNoBodyResponse(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		nmw := &noBodyResponseWriter{ResponseWriter: w}
		next.ServeHTTP(nmw, r)
	})
}

type noBodyResponseWriter struct {
	wroteHeader bool
	code        int
	http.ResponseWriter
}

func (w *noBodyResponseWriter) WriteHeader(code int) {
	w.code = code
	w.wroteHeader = true
	w.ResponseWriter.WriteHeader(code)
}

func (w *noBodyResponseWriter) Write(b []byte) (int, error) {
	if !w.wroteHeader {
		w.WriteHeader(http.StatusOK)
	}
	if w.code == http.StatusNotModified || w.code == http.StatusNoContent {
		return 0, nil
	}
	return w.ResponseWriter.Write(b)
}
