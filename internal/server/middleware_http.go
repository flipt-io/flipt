package server

import "net/http"

// StripSlashes is a middleware that strips trailing slashes from the request path, this is because GRPC Gateway
// doesn't allow for trailing slashes and will 404/route to the wrong endpoint.
func StripSlashes(next http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Path
		if len(path) > 1 && path[len(path)-1] == '/' {
			newPath := path[:len(path)-1]
			r.URL.Path = newPath
		}
		next.ServeHTTP(w, r)
	}
	return http.HandlerFunc(fn)
}
