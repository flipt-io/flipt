package containers

// Option is any function which receives a pointer to any type T.
type Option[T any] func(*T)

// ApplyAll takes a pointer to type T and it passes it onto each supplied Option
// function in the supplied order.
func ApplyAll[T any](t *T, opts ...Option[T]) {
	for _, opt := range opts {
		opt(t)
	}
}
