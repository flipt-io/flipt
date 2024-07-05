package http_middleware

import (
	"bytes"
	"crypto/sha1"
	"fmt"
	"hash"
	"io"
	"net/http"
)

type etagResponseWriter struct {
	http.ResponseWriter
	buf  bytes.Buffer
	hash hash.Hash
	w    io.Writer
}

func (e *etagResponseWriter) Write(p []byte) (int, error) {
	return e.w.Write(p)
}

func Etag(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ew := &etagResponseWriter{
			ResponseWriter: w,
			buf:            bytes.Buffer{},
			hash:           sha1.New(),
		}

		ew.w = io.MultiWriter(&ew.buf, ew.hash)

		next.ServeHTTP(ew, r)

		sum := fmt.Sprintf("%x", ew.hash.Sum(nil))
		w.Header().Set("ETag", sum)

		if r.Header.Get("If-None-Match") == sum {
			w.WriteHeader(http.StatusNotModified)
			return
		}

		_, _ = ew.buf.WriteTo(w)
	})
}
