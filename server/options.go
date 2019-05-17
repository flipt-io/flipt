package server

import "github.com/markphelps/flipt/storage/cache"

// Option is a server option
type Option func(s *Server)

// WithCache sets the cache to be used for the server
func WithCache(c cache.Cacher) Option {
	return func(s *Server) {
		s.cache = c
	}
}
